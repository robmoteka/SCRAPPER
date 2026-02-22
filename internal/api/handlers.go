package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/user/scrapper/internal/export"
	"github.com/user/scrapper/internal/models"
	"github.com/user/scrapper/internal/scraper"
)

// Global state for running scrapers (in production, use Redis/DB)
var (
	activeProjects = make(map[string]*scraper.Scraper)
	projectsMutex  sync.RWMutex
	dataDir        = getEnvOrDefault("DATA_DIR", "./data")
)

// HandleScrape starts a new scraping job
func HandleScrape(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req models.ScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.URL == "" {
		respondError(w, http.StatusBadRequest, "URL is required")
		return
	}

	if req.Depth < 1 || req.Depth > 5 {
		respondError(w, http.StatusBadRequest, "Depth must be between 1 and 5")
		return
	}

	urlPrefix, err := scraper.ValidateAndNormalizeScopePrefix(req.URL, req.URLPrefix)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate filters
	if err := scraper.ValidateFilters(req.Filters); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid filters: %v", err))
		return
	}

	// Create project
	project := &models.Project{
		ID:        uuid.New().String(),
		URL:       req.URL,
		URLPrefix: urlPrefix,
		Depth:     req.Depth,
		Filters:   req.Filters,
		Status:    models.StatusStarted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create scraper
	s, err := scraper.NewScraper(project, dataDir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create scraper: %v", err))
		return
	}

	// Register scraper
	projectsMutex.Lock()
	activeProjects[project.ID] = s
	projectsMutex.Unlock()

	// Start scraping async
	go runScraper(s, project.ID)

	// Response
	response := models.ScrapeResponse{
		ProjectID: project.ID,
		Status:    models.StatusStarted,
	}

	respondJSON(w, http.StatusAccepted, response)
}

// runScraper executes scraping in background
func runScraper(s *scraper.Scraper, projectID string) {
	// Track project
	globalTracker.TrackProject(projectID, s)

	defer func() {
		// Cleanup after completion
		globalTracker.UntrackProject(projectID)

		projectsMutex.Lock()
		delete(activeProjects, projectID)
		projectsMutex.Unlock()
	}()

	// Run scraping with periodic status updates
	if err := s.Run(); err != nil {
		s.Project.Status = models.StatusFailed
		s.Project.Errors = append(s.Project.Errors, err.Error())
		s.SaveProject() // Save error state
	}
}

// HandleStatus returns scraping job status
func HandleStatus(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	// Check active scrapers first
	projectsMutex.RLock()
	s, isActive := activeProjects[projectID]
	projectsMutex.RUnlock()

	if isActive {
		// Return live status
		response := models.StatusResponse{
			Status:     s.Project.Status,
			Progress:   calculateProgress(s),
			Downloaded: s.Project.Downloaded,
			Total:      s.Project.Total,
			CurrentURL: s.Project.CurrentURL,
			Errors:     s.Project.Errors,
		}
		respondJSON(w, http.StatusOK, response)
		return
	}

	// Load from disk if not active
	project, err := scraper.LoadProject(projectID, dataDir)
	if err != nil {
		respondError(w, http.StatusNotFound, "Project not found")
		return
	}

	response := models.StatusResponse{
		Status:     project.Status,
		Progress:   100, // Completed or failed
		Downloaded: project.Downloaded,
		Total:      project.Total,
		CurrentURL: project.CurrentURL,
		Errors:     project.Errors,
	}

	respondJSON(w, http.StatusOK, response)
}

// calculateProgress computes progress percentage
func calculateProgress(s *scraper.Scraper) int {
	if s.Project.Total == 0 {
		return 0
	}
	return (s.Project.Downloaded * 100) / s.Project.Total
}

// HandleExportZip exports project as ZIP
func HandleExportZip(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	// Check if project exists
	if !scraper.ProjectExists(projectID, dataDir) {
		respondError(w, http.StatusNotFound, "Project not found")
		return
	}

	// Load project to check status
	project, err := scraper.LoadProject(projectID, dataDir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to load project")
		return
	}

	// Only export completed projects
	if project.Status != models.StatusCompleted {
		respondError(w, http.StatusBadRequest, "Project is not completed yet")
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", projectID))

	// Stream ZIP directly to response
	if err := export.StreamZipToWriter(w, projectID, dataDir); err != nil {
		// Can't send error response after streaming started
		log.Printf("ZIP export error for project %s: %v", projectID, err)
	}
}

// HandleExportPDF generates and exports project as PDF
func HandleExportPDF(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	// Check if project exists
	if !scraper.ProjectExists(projectID, dataDir) {
		respondError(w, http.StatusNotFound, "Project not found")
		return
	}

	// Load project to check status
	project, err := scraper.LoadProject(projectID, dataDir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to load project")
		return
	}

	// Only export completed projects
	if project.Status != models.StatusCompleted {
		respondError(w, http.StatusBadRequest, "Project is not completed yet")
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", projectID))

	// Generate and stream PDF
	if err := export.StreamPDFToWriter(w, projectID, dataDir); err != nil {
		log.Printf("PDF export error for project %s: %v", projectID, err)
	}
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func getEnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
