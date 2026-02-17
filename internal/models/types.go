package models

import "time"

// ScrapeRequest represents the input for a scraping operation
type ScrapeRequest struct {
	URL     string       `json:"url"`
	Depth   int          `json:"depth"`
	Filters []FilterRule `json:"filters"`
}

// FilterRule defines a pattern to remove content between start and end markers
type FilterRule struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// ScrapeResponse is returned when a scrape job is started
type ScrapeResponse struct {
	ProjectID string `json:"project_id"`
	Status    string `json:"status"`
}

// ProjectStatus represents the current state of a scraping job
type ProjectStatus struct {
	Status          string    `json:"status"` // in_progress, completed, failed
	Progress        int       `json:"progress"`
	PagesDownloaded int       `json:"pages_downloaded"`
	TotalPages      int       `json:"total_pages"`
	CurrentURL      string    `json:"current_url"`
	Errors          []string  `json:"errors"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time,omitempty"`
}

// Project represents metadata about a scraping project
type Project struct {
	ID        string
	URL       string
	Depth     int
	Status    *ProjectStatus
	DataPath  string
	Filters   []FilterRule
	CreatedAt time.Time
}
