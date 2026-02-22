package models

import "time"

// ScrapeRequest represents incoming scraping request from API
type ScrapeRequest struct {
	URL       string       `json:"url"`
	URLPrefix string       `json:"url_prefix,omitempty"`
	Depth     int          `json:"depth"`
	Filters   []FilterRule `json:"filters"`
}

// FilterRule defines HTML/JS filtering pattern
type FilterRule struct {
	Start string `json:"start"` // Start pattern (e.g., "<script")
	End   string `json:"end"`   // End pattern (e.g., "</script>")
}

// ProjectStatus represents project execution state
type ProjectStatus string

const (
	StatusStarted    ProjectStatus = "started"
	StatusInProgress ProjectStatus = "in_progress"
	StatusCompleted  ProjectStatus = "completed"
	StatusFailed     ProjectStatus = "failed"
)

// Project represents a scraping project
type Project struct {
	ID         string        `json:"project_id"`
	URL        string        `json:"url"`
	URLPrefix  string        `json:"url_prefix,omitempty"`
	Depth      int           `json:"depth"`
	Status     ProjectStatus `json:"status"`
	Filters    []FilterRule  `json:"filters"`
	Progress   int           `json:"progress"`
	Downloaded int           `json:"pages_downloaded"`
	Total      int           `json:"total_pages"`
	CurrentURL string        `json:"current_url"`
	Errors     []string      `json:"errors"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// Asset represents a downloadable resource (image, CSS, JS, etc.)
type Asset struct {
	URL        string // Original URL
	LocalPath  string // Path in project folder
	Type       string // "image", "css", "js", "font", "other"
	Downloaded bool
	Error      string
}

// Page represents a scraped HTML page
type Page struct {
	URL        string
	LocalPath  string // Relative path in project
	Depth      int
	ParentURL  string
	HTML       string
	Assets     []Asset
	Links      []string // Extracted links
	Downloaded bool
	Processed  bool // Link transformation done
	Filtered   bool // Filters applied
	Error      string
}

// ScrapeResponse returned after starting scrape
type ScrapeResponse struct {
	ProjectID string        `json:"project_id"`
	Status    ProjectStatus `json:"status"`
}

// StatusResponse for status endpoint
type StatusResponse struct {
	Status     ProjectStatus `json:"status"`
	Progress   int           `json:"progress"`
	Downloaded int           `json:"pages_downloaded"`
	Total      int           `json:"total_pages"`
	CurrentURL string        `json:"current_url"`
	Errors     []string      `json:"errors"`
}
