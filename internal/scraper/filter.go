package scraper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/robmoteka/scrapper/internal/models"
)

// ApplyFilters applies filter rules to HTML content
func ApplyFilters(html string, filters []models.FilterRule) string {
	result := html

	for _, filter := range filters {
		result = applyFilter(result, filter)
	}

	return result
}

// applyFilter applies a single filter rule to HTML
func applyFilter(html string, filter models.FilterRule) string {
	result := html

	for {
		// Find start pattern
		startIdx := strings.Index(result, filter.Start)
		if startIdx == -1 {
			break
		}

		// Find end pattern after start
		searchFrom := startIdx + len(filter.Start)
		endIdx := strings.Index(result[searchFrom:], filter.End)
		if endIdx == -1 {
			break
		}

		// Calculate actual end position
		endIdx = searchFrom + endIdx + len(filter.End)

		// Remove the content between (including markers)
		result = result[:startIdx] + result[endIdx:]
	}

	return result
}

// ApplyFiltersToProject applies filters to all HTML files in a project
func ApplyFiltersToProject(projectPath string, filters []models.FilterRule) error {
	if len(filters) == 0 {
		return nil
	}

	// Save filters to project
	filtersPath := filepath.Join(projectPath, "filters.json")
	filtersData, err := json.MarshalIndent(filters, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}
	if err := os.WriteFile(filtersPath, filtersData, 0644); err != nil {
		return fmt.Errorf("failed to save filters: %w", err)
	}

	// Apply filters to index.html
	indexPath := filepath.Join(projectPath, "index.html")
	if err := applyFiltersToFile(indexPath, filters); err != nil {
		fmt.Printf("Warning: failed to apply filters to index.html: %v\n", err)
	}

	// Apply filters to all pages
	pagesDir := filepath.Join(projectPath, "pages")
	if _, err := os.Stat(pagesDir); err == nil {
		err = filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && (strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".htm")) {
				if err := applyFiltersToFile(path, filters); err != nil {
					fmt.Printf("Warning: failed to apply filters to %s: %v\n", path, err)
				}
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to walk pages directory: %w", err)
		}
	}

	return nil
}

// applyFiltersToFile applies filters to a single HTML file
func applyFiltersToFile(path string, filters []models.FilterRule) error {
	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Apply filters
	filtered := ApplyFilters(string(content), filters)

	// Write back
	if err := os.WriteFile(path, []byte(filtered), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
