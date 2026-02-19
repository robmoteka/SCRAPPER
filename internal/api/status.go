package api

import (
	"sync"
	"time"

	"github.com/user/scrapper/internal/models"
	"github.com/user/scrapper/internal/scraper"
)

// StatusTracker manages project status updates
type StatusTracker struct {
	projects map[string]*ProjectStatus
	mu       sync.RWMutex
}

// ProjectStatus holds runtime status info
type ProjectStatus struct {
	Project    *models.Project
	LastUpdate time.Time
	IsActive   bool
	Scraper    *scraper.Scraper
}

var globalTracker = &StatusTracker{
	projects: make(map[string]*ProjectStatus),
}

// GetGlobalTracker returns the singleton tracker instance
func GetGlobalTracker() *StatusTracker {
	return globalTracker
}

// TrackProject registers a project for status tracking
func (st *StatusTracker) TrackProject(projectID string, s *scraper.Scraper) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.projects[projectID] = &ProjectStatus{
		Project:    s.Project,
		LastUpdate: time.Now(),
		IsActive:   true,
		Scraper:    s,
	}
}

// UntrackProject removes project from active tracking
func (st *StatusTracker) UntrackProject(projectID string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if status, exists := st.projects[projectID]; exists {
		status.IsActive = false
		status.LastUpdate = time.Now()
	}
}

// GetStatus retrieves current project status
func (st *StatusTracker) GetStatus(projectID string) (*ProjectStatus, bool) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	status, exists := st.projects[projectID]
	return status, exists
}

// UpdateProgress updates project progress (called periodically by scraper)
func (st *StatusTracker) UpdateProgress(projectID string, downloaded, total int, currentURL string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if status, exists := st.projects[projectID]; exists {
		status.Project.Downloaded = downloaded
		status.Project.Total = total
		status.Project.CurrentURL = currentURL
		status.LastUpdate = time.Now()
	}
}

// CleanupStaleProjects removes old inactive projects from memory
func (st *StatusTracker) CleanupStaleProjects(maxAge time.Duration) {
	st.mu.Lock()
	defer st.mu.Unlock()

	now := time.Now()
	for id, status := range st.projects {
		if !status.IsActive && now.Sub(status.LastUpdate) > maxAge {
			delete(st.projects, id)
		}
	}
}

// StartCleanupRoutine runs periodic cleanup
func (st *StatusTracker) StartCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			st.CleanupStaleProjects(30 * time.Minute)
		}
	}()
}
