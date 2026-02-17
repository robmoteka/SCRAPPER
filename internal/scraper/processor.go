package scraper

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// TransformLinks converts absolute URLs to relative paths in HTML
func TransformLinks(html, currentURL, baseURL, baseDomain string) string {
	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html
	}

	// Transform href attributes
	doc.Find("[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}

		transformed := transformURL(href, currentURL, baseURL, baseDomain)
		s.SetAttr("href", transformed)
	})

	// Transform src attributes
	doc.Find("[src]").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if !exists || src == "" {
			return
		}

		transformed := transformURL(src, currentURL, baseURL, baseDomain)
		s.SetAttr("src", transformed)
	})

	// Transform srcset attributes (for responsive images)
	doc.Find("[srcset]").Each(func(i int, s *goquery.Selection) {
		srcset, exists := s.Attr("srcset")
		if !exists || srcset == "" {
			return
		}

		// srcset can have multiple URLs separated by commas
		parts := strings.Split(srcset, ",")
		var transformed []string
		for _, part := range parts {
			part = strings.TrimSpace(part)
			// Each part can be "url descriptor"
			urlParts := strings.Fields(part)
			if len(urlParts) > 0 {
				transformedURL := transformURL(urlParts[0], currentURL, baseURL, baseDomain)
				if len(urlParts) > 1 {
					transformed = append(transformed, transformedURL+" "+strings.Join(urlParts[1:], " "))
				} else {
					transformed = append(transformed, transformedURL)
				}
			}
		}
		s.SetAttr("srcset", strings.Join(transformed, ", "))
	})

	// Transform data-src attributes (lazy loading)
	doc.Find("[data-src]").Each(func(i int, s *goquery.Selection) {
		dataSrc, exists := s.Attr("data-src")
		if !exists || dataSrc == "" {
			return
		}

		transformed := transformURL(dataSrc, currentURL, baseURL, baseDomain)
		s.SetAttr("data-src", transformed)
	})

	// Get transformed HTML
	transformedHTML, err := doc.Html()
	if err != nil {
		return html
	}

	return transformedHTML
}

// transformURL converts a URL to a relative path if it's within the same domain
func transformURL(rawURL, currentURL, baseURL, baseDomain string) string {
	// Skip empty URLs
	if rawURL == "" {
		return rawURL
	}

	// Skip anchors and javascript
	if strings.HasPrefix(rawURL, "#") || strings.HasPrefix(rawURL, "javascript:") || strings.HasPrefix(rawURL, "mailto:") {
		return rawURL
	}

	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// If it's a relative URL, make it absolute first
	if !parsedURL.IsAbs() {
		baseURLParsed, err := url.Parse(currentURL)
		if err != nil {
			return rawURL
		}
		parsedURL = baseURLParsed.ResolveReference(parsedURL)
	}

	// Check if URL is from the same domain
	if parsedURL.Host != baseDomain {
		// External URL - keep as absolute
		return parsedURL.String()
	}

	// Convert to relative path
	path := parsedURL.Path
	if path == "" || path == "/" {
		return "index.html"
	}

	// Determine if it's an asset or a page
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".css":
		return "assets/css/" + filepath.Base(path)
	case ".js":
		return "assets/js/" + filepath.Base(path)
	case ".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp", ".ico":
		return "assets/img/" + filepath.Base(path)
	case ".woff", ".woff2", ".ttf", ".eot":
		return "assets/fonts/" + filepath.Base(path)
	default:
		// Treat as HTML page
		cleanPath := strings.TrimPrefix(path, "/")
		if cleanPath == "" {
			return "index.html"
		}
		if !strings.HasSuffix(cleanPath, ".html") && !strings.HasSuffix(cleanPath, ".htm") {
			cleanPath = cleanPath + ".html"
		}
		return "pages/" + cleanPath
	}
}
