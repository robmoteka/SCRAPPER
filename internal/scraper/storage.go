package scraper

import (
	"encoding/json"
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
		filepath.Join(projectDir, "assets", "img"),
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

// ProjectSnapshot represents a point-in-time state of the project
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

// CreateProjectSnapshot saves a summary of project state
func (s *Scraper) CreateProjectSnapshot() *ProjectSnapshot {
	return &ProjectSnapshot{
		ProjectID:        s.Project.ID,
		URL:              s.Project.URL,
		Status:           s.Project.Status,
		TotalPages:       len(s.Pages),
		TotalAssets:      len(s.Assets),
		DownloadedPages:  s.countDownloadedPages(),
		DownloadedAssets: s.countDownloadedAssets(),
		ErrorCount:       len(s.Project.Errors),
		CreatedAt:        s.Project.CreatedAt,
		UpdatedAt:        time.Now(),
	}
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
