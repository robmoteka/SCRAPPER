package scraper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/scrapper/internal/models"
)

// ApplyFilters applies all filter rules to HTML content sequentially
func ApplyFilters(html string, filters []models.FilterRule) string {
	result := html

	for _, filter := range filters {
		result = applyFilter(result, filter)
	}

	return result
}

// applyFilter applies a single filter rule to HTML
// Removes all content between Start and End patterns (inclusive)
func applyFilter(html string, filter models.FilterRule) string {
	if filter.Start == "" || filter.End == "" {
		return html
	}

	result := html

	for {
		// Find start pattern
		startIdx := strings.Index(result, filter.Start)
		if startIdx == -1 {
			break // No more matches
		}

		// Find end pattern after start
		endIdx := strings.Index(result[startIdx:], filter.End)
		if endIdx == -1 {
			// End not found, skip this start
			break
		}

		// Calculate absolute end position
		endIdx = startIdx + endIdx + len(filter.End)

		// Remove content from start to end (inclusive)
		result = result[:startIdx] + result[endIdx:]
	}

	return result
}

// ApplyFiltersToProject applies filters to all HTML files in project
func (s *Scraper) ApplyFiltersToProject() error {
	if len(s.Project.Filters) == 0 {
		return nil // No filters to apply
	}

	for _, page := range s.Pages {
		if err := s.applyFiltersToPage(page); err != nil {
			// Log error but continue
			page.Error = fmt.Sprintf("Filter application failed: %v", err)
			continue
		}
		page.Filtered = true
	}

	// Save filter configuration
	if err := s.SaveFilters(); err != nil {
		return fmt.Errorf("failed to save filters: %w", err)
	}

	return nil
}

// applyFiltersToPage applies filters to a single HTML file
func (s *Scraper) applyFiltersToPage(page *models.Page) error {
	// Read current HTML
	htmlBytes, err := os.ReadFile(page.LocalPath)
	if err != nil {
		return err
	}

	// Apply filters
	filteredHTML := ApplyFilters(string(htmlBytes), s.Project.Filters)

	// Write back
	return os.WriteFile(page.LocalPath, []byte(filteredHTML), 0644)
}

// SaveFilters persists filter rules to JSON file in project directory
func (s *Scraper) SaveFilters() error {
	projectDir := filepath.Join(s.DataDir, s.Project.ID)
	filtersPath := filepath.Join(projectDir, "filters.json")

	data, err := json.MarshalIndent(s.Project.Filters, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filtersPath, data, 0644)
}

// LoadFilters reads filter rules from JSON file
func LoadFilters(projectID, dataDir string) ([]models.FilterRule, error) {
	filtersPath := filepath.Join(dataDir, projectID, "filters.json")

	// Check if file exists
	if _, err := os.Stat(filtersPath); os.IsNotExist(err) {
		return []models.FilterRule{}, nil // No filters file, return empty
	}

	data, err := os.ReadFile(filtersPath)
	if err != nil {
		return nil, err
	}

	var filters []models.FilterRule
	if err := json.Unmarshal(data, &filters); err != nil {
		return nil, err
	}

	return filters, nil
}

// ValidateFilters checks if filter rules are valid
func ValidateFilters(filters []models.FilterRule) error {
	for i, filter := range filters {
		if filter.Start == "" {
			return fmt.Errorf("filter %d: start pattern is empty", i)
		}
		if filter.End == "" {
			return fmt.Errorf("filter %d: end pattern is empty", i)
		}
		// Warning: Check if patterns might remove too much
		if filter.Start == filter.End {
			return fmt.Errorf("filter %d: start and end patterns are identical", i)
		}
	}
	return nil
}

// ParseFiltersFromLines parses filter rules from text lines
// Format: "START|||END" per line
func ParseFiltersFromLines(lines []string) ([]models.FilterRule, error) {
	filters := make([]models.FilterRule, 0, len(lines))

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue // Skip empty lines
		}

		parts := strings.Split(line, "|||")
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: invalid format, expected 'START|||END'", i+1)
		}

		filter := models.FilterRule{
			Start: strings.TrimSpace(parts[0]),
			End:   strings.TrimSpace(parts[1]),
		}

		filters = append(filters, filter)
	}

	return filters, nil
}
