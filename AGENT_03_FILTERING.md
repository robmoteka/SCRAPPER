	# AGENT 3: Filtering & Storage

**Phase**: Filtering & Storage  
**Zadania**: 7-8  
**Dependencies**: Agent 2 (scraper logic)  
**Estimated Time**: 30-40 minut

---

## Cel Agenta

Implementacja systemu filtrowania HTML/JS z użyciem wzorców "start|||end" oraz logiki persystencji projektów na dysku z obsługą JSON.

---

## Prerequisites

Przed rozpoczęciem sprawdź:
- [x] Agent 2 ukończony (`internal/scraper/scraper.go` działa)
- [x] Pages są zapisywane do `data/{project-id}/`
- [x] Modele zawierają `FilterRule` struct

---

## Zadania do Wykonania

### ✅ Zadanie 7: Implement filter.go (HTML/JS filtering)

**Cel**: System do usuwania fragmentów HTML/JS na podstawie wzorców start/end.

**Plik**: `internal/scraper/filter.go`

```go
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
```

**Verification**:
```bash
go build ./internal/scraper
```

**Unit test example** (optional):
```go
// Add to filter_test.go
package scraper

import (
	"testing"
	"github.com/user/scrapper/internal/models"
)

func TestApplyFilter(t *testing.T) {
	html := `<html><body><script>alert('test')</script><div>Keep this</div></body></html>`
	filter := models.FilterRule{
		Start: "<script",
		End:   "</script>",
	}

	result := applyFilter(html, filter)
	
	if strings.Contains(result, "<script") {
		t.Error("Filter did not remove script tag")
	}
	
	if !strings.Contains(result, "Keep this") {
		t.Error("Filter removed valid content")
	}
}

func TestApplyFilters(t *testing.T) {
	html := `<html><script>bad</script><style>bad</style><div>good</div></html>`
	filters := []models.FilterRule{
		{Start: "<script", End: "</script>"},
		{Start: "<style", End: "</style>"},
	}

	result := ApplyFilters(html, filters)

	if strings.Contains(result, "<script") || strings.Contains(result, "<style") {
		t.Error("Filters did not remove both tags")
	}
}
```

---

### ✅ Zadanie 8: Create file storage logic

**Cel**: Rozszerzenie scrapera o kompletny workflow persystencji + projekt metadata.

**Rozszerzenie**: `internal/scraper/storage.go`

```go
package scraper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/scrapper/internal/models"
)

// SaveProject saves project metadata to JSON file
func (s *Scraper) SaveProject() error {
	projectDir := filepath.Join(s.DataDir, s.Project.ID)
	metadataPath := filepath.Join(projectDir, "project.json")

	// Update timestamp
	s.Project.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(s.Project, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

// LoadProject loads project metadata from JSON file
func LoadProject(projectID, dataDir string) (*models.Project, error) {
	metadataPath := filepath.Join(dataDir, projectID, "project.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var project models.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// ProjectExists checks if project directory exists
func ProjectExists(projectID, dataDir string) bool {
	projectDir := filepath.Join(dataDir, projectID)
	info, err := os.Stat(projectDir)
	return err == nil && info.IsDir()
}

// InitializeProjectDirectory creates project folder structure
func InitializeProjectDirectory(projectID, dataDir string) error {
	projectDir := filepath.Join(dataDir, projectID)

	// Create main directory
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return err
	}

	// Create subdirectories
	dirs := []string{
		filepath.Join(projectDir, "pages"),
		filepath.Join(projectDir, "assets", "css"),
		filepath.Join(projectDir, "assets", "js"),
		filepath.Join(projectDir, "assets", "image"),
		filepath.Join(projectDir, "assets", "font"),
		filepath.Join(projectDir, "assets", "other"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// DeleteProject removes project directory and all contents
func DeleteProject(projectID, dataDir string) error {
	projectDir := filepath.Join(dataDir, projectID)
	return os.RemoveAll(projectDir)
}

// ListProjects returns list of all project IDs in data directory
func ListProjects(dataDir string) ([]string, error) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	projects := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			projects = append(projects, entry.Name())
		}
	}

	return projects, nil
}

// GetProjectSize calculates total size of project directory in bytes
func GetProjectSize(projectID, dataDir string) (int64, error) {
	projectDir := filepath.Join(dataDir, projectID)
	var size int64

	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// CreateProjectSnapshot saves a summary of project state
func (s *Scraper) CreateProjectSnapshot() *ProjectSnapshot {
	return &ProjectSnapshot{
		ProjectID:      s.Project.ID,
		URL:            s.Project.URL,
		Status:         s.Project.Status,
		TotalPages:     len(s.Pages),
		TotalAssets:    len(s.Assets),
		DownloadedPages: s.countDownloadedPages(),
		DownloadedAssets: s.countDownloadedAssets(),
		ErrorCount:     len(s.Project.Errors),
		CreatedAt:      s.Project.CreatedAt,
		UpdatedAt:      time.Now(),
	}
}

// ProjectSnapshot represents a point-in-time state
type ProjectSnapshot struct {
	ProjectID        string              `json:"project_id"`
	URL              string              `json:"url"`
	Status           models.ProjectStatus `json:"status"`
	TotalPages       int                 `json:"total_pages"`
	TotalAssets      int                 `json:"total_assets"`
	DownloadedPages  int                 `json:"downloaded_pages"`
	DownloadedAssets int                 `json:"downloaded_assets"`
	ErrorCount       int                 `json:"error_count"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

func (s *Scraper) countDownloadedPages() int {
	count := 0
	for _, page := range s.Pages {
		if page.Downloaded {
			count++
		}
	}
	return count
}

func (s *Scraper) countDownloadedAssets() int {
	count := 0
	for _, asset := range s.Assets {
		if asset.Downloaded {
			count++
		}
	}
	return count
}
```

**Verification**:
```bash
go build ./internal/scraper
```

---

## Integration: Update Scraper.Run()

**Modify**: `internal/scraper/scraper.go` - dodaj storage i filtering do workflow

```go
// Run starts the complete scraping workflow
func (s *Scraper) Run() error {
	s.Project.Status = models.StatusInProgress
	s.Project.CreatedAt = time.Now()

	// Initialize project directory
	if err := InitializeProjectDirectory(s.Project.ID, s.DataDir); err != nil {
		return fmt.Errorf("failed to initialize project: %w", err)
	}

	// Start scraping
	if err := s.Collector.Visit(s.Project.URL); err != nil {
		s.Project.Status = models.StatusFailed
		return fmt.Errorf("failed to start scraping: %w", err)
	}

	s.Collector.Wait()

	// Download assets
	if err := s.downloadAssets(); err != nil {
		s.Project.Errors = append(s.Project.Errors, fmt.Sprintf("Asset errors: %v", err))
	}

	// Save pages
	if err := s.savePages(); err != nil {
		s.Project.Status = models.StatusFailed
		return fmt.Errorf("failed to save pages: %w", err)
	}

	// Process links (make relative)
	if err := s.ProcessLinks(); err != nil {
		s.Project.Errors = append(s.Project.Errors, fmt.Sprintf("Link processing errors: %v", err))
	}

	// Apply filters
	if err := s.ApplyFiltersToProject(); err != nil {
		s.Project.Errors = append(s.Project.Errors, fmt.Sprintf("Filter errors: %v", err))
	}

	// Save project metadata
	s.Project.Status = models.StatusCompleted
	s.Project.Total = len(s.Pages)
	if err := s.SaveProject(); err != nil {
		return fmt.Errorf("failed to save project metadata: %w", err)
	}

	return nil
}
```

---

## Expected Output Files

Po ukończeniu Agenta 3:

```
✅ internal/scraper/filter.go
✅ internal/scraper/storage.go
✅ Updated scraper.go with full workflow
✅ Filter application działa
✅ Project persistence działa
```

---

## Verification Checklist

- [ ] `go build ./internal/scraper` kompiluje
- [ ] Filter rules parsują poprawnie z formatu `START|||END`
- [ ] Filtering usuwa content między patterns
- [ ] Multiple filters aplikują się sekwencyjnie
- [ ] Project metadata zapisuje się do `project.json`
- [ ] Filters zapisują się do `filters.json`
- [ ] Project directory ma pełną strukturę folderów

---

## Manual Test

```bash
# Test filtering
cat > test_filter.go <<EOF
package main

import (
	"fmt"
	"github.com/user/scrapper/internal/models"
	"github.com/user/scrapper/internal/scraper"
)

func main() {
	html := \`<html><body><script>alert('ads')</script><div>Content</div></body></html>\`
	
	filters := []models.FilterRule{
		{Start: "<script", End: "</script>"},
	}
	
	result := scraper.ApplyFilters(html, filters)
	fmt.Println(result)
	// Expected: <html><body><div>Content</div></body></html>
}
EOF

go run test_filter.go
```

---

## Common Issues & Solutions

### Issue 1: Filter removes too much
**Symptom**: Więcej treści usuniętej niż expected  
**Solution**: Check if End pattern occurs multiple times; adjust pattern specificity

### Issue 2: JSON marshal error
**Symptom**: `json: unsupported type`  
**Solution**: Ensure all struct fields are exportable (capitalized)

### Issue 3: File permission errors
**Symptom**: `permission denied` during save  
**Solution**: Check directory permissions, ensure `os.MkdirAll` with 0755

---

## Next Agent

Po ukończeniu **Agent 3**, przejdź do:
👉 **AGENT_04_API.md** (REST API handlers + async scraping)

**Prerequisites verified**:
- ✅ Filtering system działa
- ✅ Project persistence implemented
- ✅ Complete scraping workflow (scrape → transform → filter → save)

---

**Agent Status**: ⏳ TODO  
**Last Updated**: 17 lutego 2026
