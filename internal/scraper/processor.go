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
