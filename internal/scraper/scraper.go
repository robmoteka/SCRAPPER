package scraper

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/user/scrapper/internal/models"
)

// Scraper manages web scraping operations
type Scraper struct {
	Project     *models.Project
	Collector   *colly.Collector
	BaseURL     *url.URL
	BaseDomain  string
	ScopePrefix string
	Pages       map[string]*models.Page  // URL -> Page
	Assets      map[string]*models.Asset // URL -> Asset
	mu          sync.RWMutex
	DataDir     string
	MaxDepth    int
}

// NewScraper creates a configured scraper instance
func NewScraper(project *models.Project, dataDir string) (*Scraper, error) {
	baseURL, err := url.Parse(project.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	scopePrefix, err := ValidateAndNormalizeScopePrefix(project.URL, project.URLPrefix)
	if err != nil {
		return nil, err
	}
	project.URLPrefix = scopePrefix

	s := &Scraper{
		Project:     project,
		BaseURL:     baseURL,
		BaseDomain:  baseURL.Hostname(),
		ScopePrefix: scopePrefix,
		Pages:       make(map[string]*models.Page),
		Assets:      make(map[string]*models.Asset),
		DataDir:     dataDir,
		MaxDepth:    project.Depth,
	}

	// Configure Colly
	s.Collector = colly.NewCollector(
		colly.MaxDepth(project.Depth),
		colly.AllowedDomains(s.BaseDomain),
		colly.Async(true),
	)

	// Set custom User-Agent
	s.Collector.UserAgent = "WebScraper/1.0 (+https://github.com/user/scrapper)"

	// Limit parallelism
	s.Collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       0, // No artificial delay
	})

	s.setupCallbacks()

	return s, nil
}

// setupCallbacks configures Colly event handlers
func (s *Scraper) setupCallbacks() {
	// On HTML page
	s.Collector.OnHTML("html", func(e *colly.HTMLElement) {
		pageURL := e.Request.URL.String()
		depth := e.Request.Depth

		s.mu.Lock()
		// Initialize page if not exists (could be pre-created)
		if _, exists := s.Pages[pageURL]; !exists {
			s.Pages[pageURL] = &models.Page{
				URL:        pageURL,
				Depth:      depth,
				Downloaded: true,
			}
		}
		page := s.Pages[pageURL]
		page.HTML = string(e.Response.Body)
		page.Downloaded = true
		s.mu.Unlock()

		// Extract and follow links
		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			link := el.Request.AbsoluteURL(el.Attr("href"))
			if s.shouldVisit(link) {
				el.Request.Visit(link)
			}
		})

		// Extract assets
		s.extractAssets(e)
	})

	// On request
	s.Collector.OnRequest(func(r *colly.Request) {
		if !s.isWithinScope(r.URL.String()) {
			r.Abort()
			return
		}

		s.mu.Lock()
		s.Project.CurrentURL = r.URL.String()
		s.mu.Unlock()
	})

	// On response
	s.Collector.OnResponse(func(r *colly.Response) {
		s.mu.Lock()
		s.Project.Downloaded++
		s.mu.Unlock()
	})

	// On error
	s.Collector.OnError(func(r *colly.Response, err error) {
		s.mu.Lock()
		errMsg := fmt.Sprintf("Failed to scrape %s: %v", r.Request.URL, err)
		s.Project.Errors = append(s.Project.Errors, errMsg)
		s.mu.Unlock()
	})
}

// extractAssets finds and queues asset downloads
func (s *Scraper) extractAssets(e *colly.HTMLElement) {
	// Images
	e.ForEach("img[src]", func(_ int, el *colly.HTMLElement) {
		src := el.Request.AbsoluteURL(el.Attr("src"))
		s.addAsset(src, "img")
	})

	// CSS
	e.ForEach("link[rel=stylesheet]", func(_ int, el *colly.HTMLElement) {
		href := el.Request.AbsoluteURL(el.Attr("href"))
		s.addAsset(href, "css")
	})

	// JavaScript
	e.ForEach("script[src]", func(_ int, el *colly.HTMLElement) {
		src := el.Request.AbsoluteURL(el.Attr("src"))
		s.addAsset(src, "js")
	})

	// Fonts (from CSS or direct links)
	e.ForEach("link[rel=preload][as=font]", func(_ int, el *colly.HTMLElement) {
		href := el.Request.AbsoluteURL(el.Attr("href"))
		s.addAsset(href, "font")
	})
}

// addAsset registers an asset for download
func (s *Scraper) addAsset(assetURL, assetType string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Assets[assetURL]; exists {
		return // Already tracked
	}

	// Simple validation
	if assetURL == "" {
		return
	}

	if !s.isWithinScope(assetURL) {
		return
	}

	s.Assets[assetURL] = &models.Asset{
		URL:  assetURL,
		Type: assetType,
	}
}

// shouldVisit checks if URL should be scraped
func (s *Scraper) shouldVisit(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Same domain only
	if parsedURL.Hostname() != s.BaseDomain {
		return false
	}

	if !s.isWithinScope(urlStr) {
		return false
	}

	// Skip common non-HTML extensions
	ext := strings.ToLower(filepath.Ext(parsedURL.Path))
	skipExts := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".zip", ".css", ".js", ".svg", ".ico", ".woff", ".woff2", ".ttf"}
	for _, skipExt := range skipExts {
		if ext == skipExt {
			return false
		}
	}

	return true
}

// ValidateAndNormalizeScopePrefix validates urlPrefix and returns normalized absolute prefix.
// If prefix is empty, startURL is used as default prefix.
func ValidateAndNormalizeScopePrefix(startURL, urlPrefix string) (string, error) {
	baseURL, err := url.Parse(startURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	prefix := strings.TrimSpace(urlPrefix)
	if prefix == "" {
		prefix = fmt.Sprintf("%s://%s", baseURL.Scheme, baseURL.Host)
	}

	prefixURL, err := url.Parse(prefix)
	if err != nil {
		return "", fmt.Errorf("invalid url_prefix: %w", err)
	}

	if !prefixURL.IsAbs() {
		prefixURL = baseURL.ResolveReference(prefixURL)
	}

	if prefixURL.Hostname() != baseURL.Hostname() {
		return "", fmt.Errorf("url_prefix must be in the same domain as url")
	}

	normalized := normalizeURLForScope(prefixURL.String())
	if normalized == "" {
		return "", fmt.Errorf("url_prefix cannot be empty")
	}

	return normalized, nil
}

func (s *Scraper) isWithinScope(rawURL string) bool {
	return strings.HasPrefix(normalizeURLForScope(rawURL), s.ScopePrefix)
}

func normalizeURLForScope(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}

	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/")
}

// Run starts the scraping process
func (s *Scraper) Run() error {
	s.mu.Lock()
	s.Project.Status = models.StatusInProgress
	s.Project.UpdatedAt = time.Now()
	s.mu.Unlock()

	// Initialize project directory
	if err := InitializeProjectDirectory(s.Project.ID, s.DataDir); err != nil {
		s.mu.Lock()
		s.Project.Status = models.StatusFailed
		s.Project.Errors = append(s.Project.Errors, fmt.Sprintf("Failed to initialize project: %v", err))
		s.mu.Unlock()
		return fmt.Errorf("failed to initialize project: %w", err)
	}

	// Start from base URL
	if err := s.Collector.Visit(s.Project.URL); err != nil {
		s.mu.Lock()
		s.Project.Status = models.StatusFailed
		s.Project.Errors = append(s.Project.Errors, fmt.Sprintf("Failed to start scraping: %v", err))
		s.Project.UpdatedAt = time.Now()
		s.mu.Unlock()
		return fmt.Errorf("failed to start scraping: %w", err)
	}

	// Wait for completion
	s.Collector.Wait()

	// Download assets
	if err := s.downloadAssets(); err != nil {
		s.mu.Lock()
		s.Project.Errors = append(s.Project.Errors, fmt.Sprintf("Asset download errors: %v", err))
		s.mu.Unlock()
	}

	// Save pages to disk
	if err := s.savePages(); err != nil {
		s.mu.Lock()
		s.Project.Status = models.StatusFailed
		s.Project.Errors = append(s.Project.Errors, fmt.Sprintf("Failed to save pages: %v", err))
		s.Project.UpdatedAt = time.Now()
		s.mu.Unlock()
		return fmt.Errorf("failed to save pages: %w", err)
	}

	// Process links (transformation)
	if err := s.ProcessLinks(); err != nil {
		s.mu.Lock()
		s.Project.Errors = append(s.Project.Errors, fmt.Sprintf("Link processing errors: %v", err))
		s.mu.Unlock()
	}

	// Apply filters
	if err := s.ApplyFiltersToProject(); err != nil {
		s.mu.Lock()
		s.Project.Errors = append(s.Project.Errors, fmt.Sprintf("Filter errors: %v", err))
		s.mu.Unlock()
	}

	s.mu.Lock()
	s.Project.Status = models.StatusCompleted
	s.Project.Total = len(s.Pages)
	s.Project.UpdatedAt = time.Now()
	s.Project.Progress = 100
	s.mu.Unlock()

	// Save project metadata
	if err := s.SaveProject(); err != nil {
		return fmt.Errorf("failed to save project metadata: %w", err)
	}

	return nil
}

// downloadAssets downloads all tracked assets
func (s *Scraper) downloadAssets() error {
	projectDir := filepath.Join(s.DataDir, s.Project.ID)
	assetsDir := filepath.Join(projectDir, "assets")

	// Create map copy for iteration to avoid locking issues during long operations
	var assetsToDownload []*models.Asset
	s.mu.RLock()
	for _, asset := range s.Assets {
		assetsToDownload = append(assetsToDownload, asset)
	}
	s.mu.RUnlock()

	for _, asset := range assetsToDownload {
		localPath, err := s.downloadAsset(asset.URL, assetsDir, asset.Type)
		if err != nil {
			s.mu.Lock()
			asset.Error = err.Error()
			s.mu.Unlock()
			continue
		}
		s.mu.Lock()
		asset.LocalPath = localPath
		asset.Downloaded = true
		s.mu.Unlock()
	}

	return nil
}

// downloadAsset downloads single asset to local path
func (s *Scraper) downloadAsset(assetURL, assetsDir, assetType string) (string, error) {
	resp, err := http.Get(assetURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	// Create type-specific subdirectory
	typeDir := filepath.Join(assetsDir, assetType)
	if err := os.MkdirAll(typeDir, 0755); err != nil {
		return "", err
	}

	// Generate filename from URL hash + extension
	parsedURL, _ := url.Parse(assetURL)
	ext := filepath.Ext(parsedURL.Path)
	if ext == "" {
		// Try to guess extension or default (could be improved)
		ct := resp.Header.Get("Content-Type")
		if strings.Contains(ct, "image/jpeg") {
			ext = ".jpg"
		} else if strings.Contains(ct, "image/png") {
			ext = ".png"
		} else if strings.Contains(ct, "text/css") {
			ext = ".css"
		} else if strings.Contains(ct, "javascript") {
			ext = ".js"
		}
	}

	filename := generateFilename(assetURL) + ext
	localPath := filepath.Join(typeDir, filename)

	// Save file
	file, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", err
	}

	// Return absolute path so we can calculate relative later or just return full path
	return localPath, nil
}

// savePages writes HTML pages to disk
func (s *Scraper) savePages() error {
	projectDir := filepath.Join(s.DataDir, s.Project.ID)
	pagesDir := filepath.Join(projectDir, "pages")

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for pageURL, page := range s.Pages {
		// Determine save path
		var savePath string
		// Normalize URLs for comparison
		normalizedPageURL := strings.TrimRight(pageURL, "/")
		normalizedProjectURL := strings.TrimRight(s.Project.URL, "/")

		if normalizedPageURL == normalizedProjectURL {
			// Main page
			savePath = filepath.Join(projectDir, "index.html")
		} else {
			// Subpage
			filename := generateFilename(pageURL) + ".html"
			savePath = filepath.Join(pagesDir, filename)
		}

		page.LocalPath = savePath // Store absolute path

		// Write HTML
		if err := os.WriteFile(savePath, []byte(page.HTML), 0644); err != nil {
			page.Error = err.Error()
			continue
		}
	}

	return nil
}

// generateFilename creates a safe filename from URL
func generateFilename(urlStr string) string {
	hash := md5.Sum([]byte(urlStr))
	return hex.EncodeToString(hash[:])[:16]
}
