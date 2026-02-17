# AGENT 4: API Layer

**Phase**: API Layer  
**Zadania**: 9-10  
**Dependencies**: Agent 1-3 (wszystkie core features)  
**Estimated Time**: 40-50 minut

---

## Cel Agenta

Implementacja REST API handlers z async scraping, status tracking, oraz integracja wszystkich wcześniejszych komponentów.

---

## Prerequisites

Przed rozpoczęciem sprawdź:
- [x] Agent 1-3 ukończone
- [x] Scraper działa end-to-end
- [x] Models, filtering, storage gotowe
- [x] Chi router setup z stubs

---

## Zadania do Wykonania

### ✅ Zadanie 9: Implement API handlers.go (full implementation)

**Cel**: Pełna implementacja wszystkich API endpoints.

**Plik**: `internal/api/handlers.go` (replace stubs)

```go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

	// Validate filters
	if err := scraper.ValidateFilters(req.Filters); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid filters: %v", err))
		return
	}

	// Create project
	project := &models.Project{
		ID:        uuid.New().String(),
		URL:       req.URL,
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
	defer func() {
		// Cleanup after completion
		projectsMutex.Lock()
		delete(activeProjects, projectID)
		projectsMutex.Unlock()
	}()

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

// HandleExportZip exports project as ZIP (placeholder for Agent 5)
func HandleExportZip(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	// Check if project exists
	if !scraper.ProjectExists(projectID, dataDir) {
		respondError(w, http.StatusNotFound, "Project not found")
		return
	}

	// TODO: Implement in Agent 5
	respondError(w, http.StatusNotImplemented, "ZIP export not yet implemented")
}

// HandleExportPDF generates and exports project as PDF (placeholder for Agent 5)
func HandleExportPDF(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	// Check if project exists
	if !scraper.ProjectExists(projectID, dataDir) {
		respondError(w, http.StatusNotFound, "Project not found")
		return
	}

	// TODO: Implement in Agent 5
	respondError(w, http.StatusNotImplemented, "PDF export not yet implemented")
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
```

**Add missing import**:
```go
import "os"
```

**Verification**:
```bash
go build ./internal/api
```

---

### ✅ Zadanie 10: Add async scraping & status tracking

**Cel**: Status endpoint z real-time updates + goroutine safety.

**Rozszerzenie**: `internal/api/status.go` (dedicated status logic)

```go
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
	Project      *models.Project
	LastUpdate   time.Time
	IsActive     bool
	Scraper      *scraper.Scraper
}

var globalTracker = &StatusTracker{
	projects: make(map[string]*ProjectStatus),
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

// StartCleanuproutine runs periodic cleanup
func (st *StatusTracker) StartCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			st.CleanupStaleProjects(30 * time.Minute)
		}
	}()
}
```

**Integration**: Update `runScraper` to use tracker

```go
// Updated runScraper with status tracking
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
		s.SaveProject()
	}
}
```

**Update main.go**: Start cleanup routine

```go
// In cmd/server/main.go, add before starting server:

// Start background cleanup
globalTracker := api.GetGlobalTracker() // export tracker
globalTracker.StartCleanupRoutine()
```

**Verification**:
```bash
go build ./internal/api
go build ./cmd/server
```

---

## Integration Test

**Manual API test**:

```bash
# Start server
go run cmd/server/main.go &

# Test scrape endpoint
curl -X POST http://localhost:8080/api/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "depth": 2,
    "filters": [
      {"start": "<script", "end": "</script>"}
    ]
  }'

# Expected response:
# {"project_id":"uuid-here","status":"started"}

# Save project_id from response
PROJECT_ID="uuid-from-above"

# Check status
curl http://localhost:8080/api/status/$PROJECT_ID

# Expected:
# {
#   "status": "in_progress",
#   "progress": 50,
#   "pages_downloaded": 5,
#   "total_pages": 10,
#   "current_url": "https://example.com/page",
#   "errors": []
# }

# Wait for completion, then check again
sleep 30
curl http://localhost:8080/api/status/$PROJECT_ID

# Expected:
# {"status": "completed", "progress": 100, ...}

# Verify files
ls -la data/$PROJECT_ID/
# Should see: index.html, pages/, assets/, project.json, filters.json

pkill -f "cmd/server/main.go"
```

---

## Expected Output Files

Po ukończeniu Agenta 4:

```
✅ internal/api/handlers.go (full implementation)
✅ internal/api/status.go (tracking logic)
✅ Updated cmd/server/main.go
✅ API endpoints działają
✅ Async scraping działa
✅ Status tracking real-time
```

---

## Verification Checklist

- [ ] `go build ./...` kompiluje cały projekt
- [ ] POST /api/scrape zwraca project_id
- [ ] GET /api/project/{id}/status zwraca live status
- [ ] Scraping wykonuje się asynchronicznie
- [ ] Progress updates w czasie rzeczywistym
- [ ] Status tracking działa po zakończeniu (loaded from disk)
- [ ] Error handling graceful (invalid input, not found, etc.)

---

## Common Issues & Solutions

### Issue 1: Race conditions
**Symptom**: `fatal error: concurrent map read and write`  
**Solution**: Ensure all map access protected by mutex (RLock/Lock)

### Issue 2: Goroutine leaks
**Symptom**: Memory usage grows over time  
**Solution**: Verify cleanup in defer, check goroutine termination

### Issue 3: Status not updating
**Symptom**: Progress stuck at 0  
**Solution**: Add periodic status updates in scraper loop

---

## Next Agent

Po ukończeniu **Agent 4**, przejdź do:
👉 **AGENT_05_EXPORT.md** (ZIP and PDF export features)

**Prerequisites verified**:
- ✅ API endpoints działają
- ✅ Async scraping implemented
- ✅ Status tracking real-time
- ✅ Projects persist to disk

---

**Agent Status**: ⏳ TODO  
**Last Updated**: 17 lutego 2026
