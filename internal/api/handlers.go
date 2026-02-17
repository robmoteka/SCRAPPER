package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/robmoteka/scrapper/internal/export"
	"github.com/robmoteka/scrapper/internal/models"
	"github.com/robmoteka/scrapper/internal/scraper"
)

// Handler manages API requests
type Handler struct {
	dataDir  string
	projects map[string]*models.Project
	mu       sync.RWMutex
}

// NewHandler creates a new API handler
func NewHandler(dataDir string) *Handler {
	return &Handler{
		dataDir:  dataDir,
		projects: make(map[string]*models.Project),
	}
}

// HandleScrape starts a new scraping job
func (h *Handler) HandleScrape(w http.ResponseWriter, r *http.Request) {
	var req models.ScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}
	if req.Depth < 1 || req.Depth > 5 {
		http.Error(w, "Depth must be between 1 and 5", http.StatusBadRequest)
		return
	}

	// Generate project ID
	projectID := uuid.New().String()
	projectPath := filepath.Join(h.dataDir, projectID)

	// Create project
	project := &models.Project{
		ID:       projectID,
		URL:      req.URL,
		Depth:    req.Depth,
		Filters:  req.Filters,
		DataPath: projectPath,
		Status: &models.ProjectStatus{
			Status:    "in_progress",
			StartTime: time.Now(),
		},
		CreatedAt: time.Now(),
	}

	// Store project
	h.mu.Lock()
	h.projects[projectID] = project
	h.mu.Unlock()

	// Start scraping in background
	go h.runScraping(project)

	// Return response
	response := models.ScrapeResponse{
		ProjectID: projectID,
		Status:    "started",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// runScraping executes the scraping operation
func (h *Handler) runScraping(project *models.Project) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Scraping panic for project %s: %v", project.ID, r)
			h.updateProjectStatus(project.ID, func(status *models.ProjectStatus) {
				status.Status = "failed"
				status.Errors = append(status.Errors, fmt.Sprintf("Internal error: %v", r))
				status.EndTime = time.Now()
			})
		}
	}()

	// Create scraper
	s := scraper.NewScraper(project.DataPath, project.Depth, project.URL)

	// Setup progress callback
	s.SetProgressCallback(func(current, total int, url string) {
		h.updateProjectStatus(project.ID, func(status *models.ProjectStatus) {
			status.PagesDownloaded = current
			status.TotalPages = total
			status.CurrentURL = url
			if total > 0 {
				status.Progress = (current * 100) / total
			}
		})
	})

	// Run scraping
	if err := s.Scrape(project.URL); err != nil {
		log.Printf("Scraping failed for project %s: %v", project.ID, err)
		h.updateProjectStatus(project.ID, func(status *models.ProjectStatus) {
			status.Status = "failed"
			status.Errors = append(status.Errors, err.Error())
			status.EndTime = time.Now()
		})
		return
	}

	// Apply filters if provided
	if len(project.Filters) > 0 {
		if err := scraper.ApplyFiltersToProject(project.DataPath, project.Filters); err != nil {
			log.Printf("Filter application failed for project %s: %v", project.ID, err)
			h.updateProjectStatus(project.ID, func(status *models.ProjectStatus) {
				status.Errors = append(status.Errors, fmt.Sprintf("Filter error: %v", err))
			})
		}
	}

	// Mark as completed
	h.updateProjectStatus(project.ID, func(status *models.ProjectStatus) {
		status.Status = "completed"
		status.Progress = 100
		status.EndTime = time.Now()
	})

	log.Printf("Scraping completed for project %s", project.ID)
}

// updateProjectStatus safely updates project status
func (h *Handler) updateProjectStatus(projectID string, updateFn func(*models.ProjectStatus)) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if project, exists := h.projects[projectID]; exists {
		updateFn(project.Status)
	}
}

// HandleStatus returns the status of a scraping job
func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	h.mu.RLock()
	project, exists := h.projects[projectID]
	h.mu.RUnlock()

	if !exists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project.Status)
}

// HandleZipExport exports project as ZIP
func (h *Handler) HandleZipExport(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	h.mu.RLock()
	project, exists := h.projects[projectID]
	h.mu.RUnlock()

	if !exists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if project.Status.Status != "completed" {
		http.Error(w, "Project not completed", http.StatusBadRequest)
		return
	}

	// Set headers for ZIP download
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", projectID))

	// Create ZIP and stream to response
	if err := export.CreateZip(project.DataPath, w); err != nil {
		log.Printf("ZIP export failed for project %s: %v", projectID, err)
		http.Error(w, "Failed to create ZIP", http.StatusInternalServerError)
		return
	}
}

// HandlePDFExport generates and exports project as PDF
func (h *Handler) HandlePDFExport(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	h.mu.RLock()
	project, exists := h.projects[projectID]
	h.mu.RUnlock()

	if !exists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if project.Status.Status != "completed" {
		http.Error(w, "Project not completed", http.StatusBadRequest)
		return
	}

	// Set headers for PDF download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", projectID))

	// Generate PDF and stream to response
	if err := export.CreatePDF(project.DataPath, w); err != nil {
		log.Printf("PDF export failed for project %s: %v", projectID, err)
		http.Error(w, "Failed to create PDF", http.StatusInternalServerError)
		return
	}
}
