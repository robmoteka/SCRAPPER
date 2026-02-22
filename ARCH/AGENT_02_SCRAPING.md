# AGENT 2: Core Scraping Engine

**Phase**: Core Scraping  
**Zadania**: 5-6  
**Dependencies**: Agent 1 (models, folder structure)  
**Estimated Time**: 45-60 minut

---

## Cel Agenta

Implementacja silnika scrapingu wykorzystującego Colly z kontrolą głębokości oraz mechanizmu transformacji linków dla przenośności offline.

---

## Prerequisites

Przed rozpoczęciem sprawdź:
- [x] Agent 1 ukończony (`internal/models/types.go` istnieje)
- [x] Dependencies zainstalowane (`colly`, `goquery`)
- [x] Folder `internal/scraper/` istnieje

---

## Zadania do Wykonania

### ✅ Zadanie 5: Implement scraper.go (Colly integration)

**Cel**: Zbudowanie core scraping logic z Colly, depth control, domain filtering, asset downloading.

**Plik**: `internal/scraper/scraper.go`

```go
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

	"github.com/gocolly/colly/v2"
	"github.com/user/scrapper/internal/models"
)

// Scraper manages web scraping operations
type Scraper struct {
	Project     *models.Project
	Collector   *colly.Collector
	BaseURL     *url.URL
	BaseDomain  string
	Pages       map[string]*models.Page // URL -> Page
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

	s := &Scraper{
		Project:    project,
		BaseURL:    baseURL,
		BaseDomain: baseURL.Hostname(),
		Pages:      make(map[string]*models.Page),
		Assets:     make(map[string]*models.Asset),
		DataDir:    dataDir,
		MaxDepth:   project.Depth,
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
		page := &models.Page{
			URL:        pageURL,
			Depth:      depth,
			HTML:       e.Response.Body,
			Downloaded: true,
		}
		s.Pages[pageURL] = page
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
		s.addAsset(src, "image")
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

	// Skip common non-HTML extensions
	ext := strings.ToLower(filepath.Ext(parsedURL.Path))
	skipExts := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".zip", ".css", ".js"}
	for _, skipExt := range skipExts {
		if ext == skipExt {
			return false
		}
	}

	return true
}

// Run starts the scraping process
func (s *Scraper) Run() error {
	s.Project.Status = models.StatusInProgress

	// Start from base URL
	if err := s.Collector.Visit(s.Project.URL); err != nil {
		s.Project.Status = models.StatusFailed
		return fmt.Errorf("failed to start scraping: %w", err)
	}

	// Wait for completion
	s.Collector.Wait()

	// Download assets
	if err := s.downloadAssets(); err != nil {
		s.Project.Errors = append(s.Project.Errors, fmt.Sprintf("Asset download errors: %v", err))
	}

	// Save pages to disk
	if err := s.savePages(); err != nil {
		s.Project.Status = models.StatusFailed
		return fmt.Errorf("failed to save pages: %w", err)
	}

	s.Project.Status = models.StatusCompleted
	s.Project.Total = len(s.Pages)
	return nil
}

// downloadAssets downloads all tracked assets
func (s *Scraper) downloadAssets() error {
	projectDir := filepath.Join(s.DataDir, s.Project.ID)
	assetsDir := filepath.Join(projectDir, "assets")

	for assetURL, asset := range s.Assets {
		localPath, err := s.downloadAsset(assetURL, assetsDir, asset.Type)
		if err != nil {
			asset.Error = err.Error()
			continue
		}
		asset.LocalPath = localPath
		asset.Downloaded = true
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

	// Generate filename from URL hash
	filename := generateFilename(assetURL)
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

	// Return relative path from project root
	relPath, _ := filepath.Rel(filepath.Join(s.DataDir, s.Project.ID), localPath)
	return relPath, nil
}

// savePages writes HTML pages to disk
func (s *Scraper) savePages() error {
	projectDir := filepath.Join(s.DataDir, s.Project.ID)
	pagesDir := filepath.Join(projectDir, "pages")

	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		return err
	}

	for pageURL, page := range s.Pages {
		// Determine save path
		var savePath string
		if pageURL == s.Project.URL {
			// Main page
			savePath = filepath.Join(projectDir, "index.html")
		} else {
			// Subpage
			filename := generateFilename(pageURL) + ".html"
			savePath = filepath.Join(pagesDir, filename)
		}

		page.LocalPath = savePath

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
```

**Verification**:
```bash
go build ./internal/scraper
```

---

### ✅ Zadanie 6: Implement processor.go (Link transformation)

**Cel**: Przekształcanie absolutnych URLi na względne ścieżki dla przenośności offline.

**Plik**: `internal/scraper/processor.go`

```go
package scraper

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/scrapper/internal/models"
)

// ProcessLinks transforms absolute URLs to relative paths in all HTML files
func (s *Scraper) ProcessLinks() error {
	for _, page := range s.Pages {
		if err := s.processPageLinks(page); err != nil {
			page.Error = fmt.Sprintf("Link processing failed: %v", err)
			continue
		}
		page.Processed = true
	}
	return nil
}

// processPageLinks transforms links in a single HTML page
func (s *Scraper) processPageLinks(page *models.Page) error {
	// Read HTML file
	htmlBytes, err := os.ReadFile(page.LocalPath)
	if err != nil {
		return err
	}

	// Parse with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return err
	}

	// Transform href attributes
	doc.Find("[href]").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists || href == "" {
			return
		}

		newHref := s.transformURL(href, page.URL)
		sel.SetAttr("href", newHref)
	})

	// Transform src attributes
	doc.Find("[src]").Each(func(i int, sel *goquery.Selection) {
		src, exists := sel.Attr("src")
		if !exists || src == "" {
			return
		}

		newSrc := s.transformURL(src, page.URL)
		sel.SetAttr("src", newSrc)
	})

	// Transform srcset attributes (responsive images)
	doc.Find("[srcset]").Each(func(i int, sel *goquery.Selection) {
		srcset, exists := sel.Attr("srcset")
		if !exists || srcset == "" {
			return
		}

		newSrcset := s.transformSrcset(srcset, page.URL)
		sel.SetAttr("srcset", newSrcset)
	})

	// Transform data-src (lazy loading)
	doc.Find("[data-src]").Each(func(i int, sel *goquery.Selection) {
		dataSrc, exists := sel.Attr("data-src")
		if !exists || dataSrc == "" {
			return
		}

		newDataSrc := s.transformURL(dataSrc, page.URL)
		sel.SetAttr("data-src", newDataSrc)
	})

	// Transform CSS url() in style attributes
	doc.Find("[style]").Each(func(i int, sel *goquery.Selection) {
		style, exists := sel.Attr("style")
		if !exists || !strings.Contains(style, "url(") {
			return
		}

		newStyle := s.transformStyleURLs(style, page.URL)
		sel.SetAttr("style", newStyle)
	})

	// Get modified HTML
	modifiedHTML, err := doc.Html()
	if err != nil {
		return err
	}

	// Save back to file
	return os.WriteFile(page.LocalPath, []byte(modifiedHTML), 0644)
}

// transformURL converts absolute URL to relative path or keeps external
func (s *Scraper) transformURL(urlStr, pageURL string) string {
	// Skip empty, anchors, and data URLs
	if urlStr == "" || strings.HasPrefix(urlStr, "#") || strings.HasPrefix(urlStr, "data:") {
		return urlStr
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr // Keep original on error
	}

	// Make absolute if relative
	if !parsedURL.IsAbs() {
		baseURL, _ := url.Parse(pageURL)
		parsedURL = baseURL.ResolveReference(parsedURL)
	}

	// Check if same domain
	if parsedURL.Hostname() != s.BaseDomain {
		return parsedURL.String() // Keep external URLs absolute
	}

	// Find local path
	localPath := s.findLocalPath(parsedURL.String())
	if localPath == "" {
		return urlStr // Not downloaded, keep original
	}

	return localPath
}

// findLocalPath returns relative path for a downloaded resource
func (s *Scraper) findLocalPath(urlStr string) string {
	// Check in pages
	if page, exists := s.Pages[urlStr]; exists {
		return s.makeRelativePath(page.LocalPath)
	}

	// Check in assets
	if asset, exists := s.Assets[urlStr]; exists && asset.Downloaded {
		return s.makeRelativePath(asset.LocalPath)
	}

	return ""
}

// makeRelativePath converts absolute path to relative from project root
func (s *Scraper) makeRelativePath(absPath string) string {
	projectDir := filepath.Join(s.DataDir, s.Project.ID)
	relPath, err := filepath.Rel(projectDir, absPath)
	if err != nil {
		return absPath
	}

	// Ensure forward slashes for web compatibility
	return filepath.ToSlash(relPath)
}

// transformSrcset handles responsive image srcset attribute
func (s *Scraper) transformSrcset(srcset, pageURL string) string {
	parts := strings.Split(srcset, ",")
	transformed := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split "url descriptor"
		tokens := strings.Fields(part)
		if len(tokens) == 0 {
			continue
		}

		urlPart := tokens[0]
		descriptor := ""
		if len(tokens) > 1 {
			descriptor = " " + strings.Join(tokens[1:], " ")
		}

		newURL := s.transformURL(urlPart, pageURL)
		transformed = append(transformed, newURL+descriptor)
	}

	return strings.Join(transformed, ", ")
}

// transformStyleURLs handles CSS url() in inline styles
func (s *Scraper) transformStyleURLs(style, pageURL string) string {
	// Simple regex-like replacement for url(...)
	// This is a basic implementation; production might need proper CSS parser
	
	result := style
	startIdx := 0

	for {
		urlIdx := strings.Index(result[startIdx:], "url(")
		if urlIdx == -1 {
			break
		}

		urlIdx += startIdx
		startIdx = urlIdx + 4

		// Find closing )
		endIdx := strings.Index(result[startIdx:], ")")
		if endIdx == -1 {
			break
		}

		endIdx += startIdx

		// Extract URL
		urlContent := strings.TrimSpace(result[startIdx:endIdx])
		urlContent = strings.Trim(urlContent, `"'`)

		// Transform
		newURL := s.transformURL(urlContent, pageURL)

		// Replace in result
		result = result[:startIdx] + `"` + newURL + `"` + result[endIdx:]
		startIdx = endIdx + len(newURL)
	}

	return result
}
```

**Verification**:
```bash
go build ./internal/scraper
```

---

## Integration Test

**Test script**: `test_scraper.go` (optional, for manual testing)

```go
package main

import (
	"log"
	"github.com/user/scrapper/internal/models"
	"github.com/user/scrapper/internal/scraper"
)

func main() {
	project := &models.Project{
		ID:    "test-project",
		URL:   "https://example.com",
		Depth: 2,
	}

	s, err := scraper.NewScraper(project, "./data")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting scrape...")
	if err := s.Run(); err != nil {
		log.Fatal(err)
	}

	log.Println("Processing links...")
	if err := s.ProcessLinks(); err != nil {
		log.Fatal(err)
	}

	log.Printf("✅ Done! Pages: %d, Assets: %d", len(s.Pages), len(s.Assets))
}
```

**Run test**:
```bash
go run test_scraper.go
ls -la data/test-project/
# Expected: index.html, pages/, assets/css/, assets/js/, assets/image/
```

---

## Expected Output Files

Po ukończeniu Agenta 2:

```
✅ internal/scraper/scraper.go
✅ internal/scraper/processor.go
✅ Core scraping działa (test manual)
✅ Link transformation działa
```

---

## Verification Checklist

- [ ] `go build ./internal/scraper` kompiluje bez błędów
- [ ] Test scraping simple URL (np. example.com) tworzy pliki
- [ ] Folder `data/{project-id}/` zawiera `index.html`
- [ ] Assets pobrane do `assets/css/`, `assets/js/`, `assets/image/`
- [ ] Linki w HTML przekształcone na względne
- [ ] Zewnętrzne linki pozostają absolutne

---

## Common Issues & Solutions

### Issue 1: Colly timeout
**Symptom**: `context deadline exceeded`  
**Solution**: Zwiększ timeout w collector configuration

### Issue 2: Asset download fails
**Symptom**: Assets nie są pobierane  
**Solution**: Sprawdź network connectivity, CORS issues

### Issue 3: Link transformation breaks
**Symptom**: Linki wskazują na nieistniejące pliki  
**Solution**: Debug `findLocalPath()` - verify URL matching

---

## Next Agent

Po ukończeniu **Agent 2**, przejdź do:
👉 **AGENT_03_FILTERING.md** (HTML/JS filtering system)

**Prerequisites verified**:
- ✅ Scraper działa i pobiera strony
- ✅ Assets są zapisywane
- ✅ Link transformation implementowany

---

**Agent Status**: ⏳ TODO  
**Last Updated**: 17 lutego 2026
