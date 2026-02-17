package scraper

import (
	"crypto/md5"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

// Scraper handles web scraping operations
type Scraper struct {
	collector        *colly.Collector
	dataPath         string
	depth            int
	baseURL          string
	baseDomain       string
	visitedURLs      map[string]bool
	mu               sync.Mutex
	progressCallback func(current, total int, url string)
	pageCount        int
	assetCount       int
}

// NewScraper creates a new scraper instance
func NewScraper(dataPath string, depth int, baseURL string) *Scraper {
	c := colly.NewCollector(
		colly.MaxDepth(depth),
		colly.Async(false),
	)

	// Set user agent
	c.UserAgent = "WebScraper/1.0"

	// Parse base URL to get domain
	parsedURL, _ := url.Parse(baseURL)
	baseDomain := parsedURL.Host

	return &Scraper{
		collector:   c,
		dataPath:    dataPath,
		depth:       depth,
		baseURL:     baseURL,
		baseDomain:  baseDomain,
		visitedURLs: make(map[string]bool),
	}
}

// SetProgressCallback sets a callback for progress updates
func (s *Scraper) SetProgressCallback(callback func(current, total int, url string)) {
	s.progressCallback = callback
}

// Scrape starts the scraping process
func (s *Scraper) Scrape(startURL string) error {
	// Create directory structure
	if err := os.MkdirAll(filepath.Join(s.dataPath, "pages"), 0755); err != nil {
		return fmt.Errorf("failed to create pages directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(s.dataPath, "assets", "css"), 0755); err != nil {
		return fmt.Errorf("failed to create assets/css directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(s.dataPath, "assets", "js"), 0755); err != nil {
		return fmt.Errorf("failed to create assets/js directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(s.dataPath, "assets", "img"), 0755); err != nil {
		return fmt.Errorf("failed to create assets/img directory: %w", err)
	}

	// Setup collectors
	s.setupCallbacks()

	// Start scraping
	if err := s.collector.Visit(startURL); err != nil {
		return fmt.Errorf("failed to start scraping: %w", err)
	}

	return nil
}

// setupCallbacks configures the collector callbacks
func (s *Scraper) setupCallbacks() {
	// Handle HTML pages
	s.collector.OnHTML("html", func(e *colly.HTMLElement) {
		s.mu.Lock()
		currentURL := e.Request.URL.String()

		// Skip if already visited
		if s.visitedURLs[currentURL] {
			s.mu.Unlock()
			return
		}
		s.visitedURLs[currentURL] = true
		s.pageCount++
		s.mu.Unlock()

		// Save HTML content
		s.saveHTMLPage(e)

		// Report progress
		if s.progressCallback != nil {
			s.progressCallback(s.pageCount, s.pageCount+1, currentURL)
		}
	})

	// Follow links
	s.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(link)

		// Parse URL
		parsedURL, err := url.Parse(absoluteURL)
		if err != nil {
			return
		}

		// Only visit links within the same domain
		if parsedURL.Host == s.baseDomain {
			e.Request.Visit(link)
		}
	})

	// Download CSS files
	s.collector.OnHTML("link[rel=stylesheet]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		s.downloadAsset(e.Request, href, "css")
	})

	// Download JavaScript files
	s.collector.OnHTML("script[src]", func(e *colly.HTMLElement) {
		src := e.Attr("src")
		s.downloadAsset(e.Request, src, "js")
	})

	// Download images
	s.collector.OnHTML("img[src]", func(e *colly.HTMLElement) {
		src := e.Attr("src")
		s.downloadAsset(e.Request, src, "img")
	})

	// Error handling
	s.collector.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error scraping %s: %v\n", r.Request.URL, err)
	})
}

// saveHTMLPage saves an HTML page to disk
func (s *Scraper) saveHTMLPage(e *colly.HTMLElement) {
	currentURL := e.Request.URL.String()
	parsedURL, _ := url.Parse(currentURL)

	var filename string
	if parsedURL.Path == "/" || parsedURL.Path == "" {
		filename = "index.html"
	} else {
		// Create filename from path
		filename = strings.TrimPrefix(parsedURL.Path, "/")
		if !strings.HasSuffix(filename, ".html") && !strings.HasSuffix(filename, ".htm") {
			filename = filename + ".html"
		}
		filename = filepath.Join("pages", filename)
	}

	// Ensure directory exists
	dir := filepath.Dir(filepath.Join(s.dataPath, filename))
	os.MkdirAll(dir, 0755)

	// Save HTML
	fullPath := filepath.Join(s.dataPath, filename)
	html := e.DOM.Find("html").Text()
	if html == "" {
		html = string(e.Response.Body)
	}

	// Transform links to relative paths
	transformedHTML := TransformLinks(string(e.Response.Body), currentURL, s.baseURL, s.baseDomain)

	if err := os.WriteFile(fullPath, []byte(transformedHTML), 0644); err != nil {
		fmt.Printf("Failed to save HTML %s: %v\n", fullPath, err)
	}
}

// downloadAsset downloads and saves an asset (CSS, JS, image)
func (s *Scraper) downloadAsset(req *colly.Request, assetURL string, assetType string) {
	// Get absolute URL
	absoluteURL := req.AbsoluteURL(assetURL)

	// Parse URL
	parsedURL, err := url.Parse(absoluteURL)
	if err != nil {
		return
	}

	// Generate filename from URL
	filename := s.generateAssetFilename(parsedURL, assetType)
	fullPath := filepath.Join(s.dataPath, "assets", assetType, filename)

	// Skip if already downloaded
	if _, err := os.Stat(fullPath); err == nil {
		return
	}

	// Download asset
	c := colly.NewCollector()
	c.OnResponse(func(r *colly.Response) {
		if err := os.WriteFile(fullPath, r.Body, 0644); err != nil {
			fmt.Printf("Failed to save asset %s: %v\n", fullPath, err)
		} else {
			s.assetCount++
		}
	})

	c.Visit(absoluteURL)
}

// generateAssetFilename creates a filename for an asset
func (s *Scraper) generateAssetFilename(parsedURL *url.URL, assetType string) string {
	path := parsedURL.Path

	// Get the base filename
	filename := filepath.Base(path)

	// If no proper filename, generate one from hash
	if filename == "" || filename == "." || filename == "/" {
		hash := md5.Sum([]byte(parsedURL.String()))
		ext := getExtensionForType(assetType)
		filename = fmt.Sprintf("%x%s", hash, ext)
	}

	return filename
}

// getExtensionForType returns the default extension for an asset type
func getExtensionForType(assetType string) string {
	switch assetType {
	case "css":
		return ".css"
	case "js":
		return ".js"
	case "img":
		return ".jpg"
	default:
		return ""
	}
}
