# AGENT 1: Foundation & Bootstrap

**Phase**: Foundation  
**Zadania**: 1-4  
**Dependencies**: Żadne  
**Estimated Time**: 20-30 minut

---

## Cel Agenta

Przygotowanie podstawowej struktury projektu Go, inicjalizacja modułu, instalacja zależności, zdefiniowanie podstawowych struktur danych oraz setup Chi routera z podstawowymi endpointami.

---

## Zadania do Wykonania

### ✅ Zadanie 1: Initialize Go module & dependencies

**Akcje**:
1. Inicjalizacja modułu Go
2. Instalacja wszystkich wymaganych pakietów
3. Weryfikacja `go.mod` i `go.sum`

**Komendy**:
```bash
cd /home/robert/1_DEVELOPMENT/HDEVS/SCRAPPER
go mod init github.com/user/scrapper
go get github.com/gocolly/colly/v2
go get github.com/PuerkitoBio/goquery
go get github.com/go-chi/chi/v5
go get github.com/jung-kurt/gofpdf
```

**Verification**:
```bash
go mod tidy
cat go.mod  # sprawdź czy wszystkie dependencies present
```

---

### ✅ Zadanie 2: Create project folder structure

**Akcje**:
Utworzenie pełnej struktury folderów zgodnie z AGENTS.md:

```
/scrapper
├── cmd/
│   └── server/
├── internal/
│   ├── api/
│   ├── scraper/
│   ├── export/
│   └── models/
├── web/
└── data/  (Git-ignored)
```

**Komendy**:
```bash
mkdir -p cmd/server
mkdir -p internal/api
mkdir -p internal/scraper
mkdir -p internal/export
mkdir -p internal/models
mkdir -p web
mkdir -p data
echo "data/" >> .gitignore
echo "*.log" >> .gitignore
echo "scrapper" >> .gitignore  # binary
```

**Verification**:
```bash
tree -L 3 -d  # pokaż strukturę folderów
```

---

### ✅ Zadanie 3: Implement models/types.go

**Akcje**:
Zdefiniowanie wszystkich podstawowych struktur danych używanych w aplikacji.

**Plik**: `internal/models/types.go`

```go
package models

import "time"

// ScrapeRequest represents incoming scraping request from API
type ScrapeRequest struct {
	URL     string       `json:"url"`
	Depth   int          `json:"depth"`
	Filters []FilterRule `json:"filters"`
}

// FilterRule defines HTML/JS filtering pattern
type FilterRule struct {
	Start string `json:"start"` // Start pattern (e.g., "<script")
	End   string `json:"end"`   // End pattern (e.g., "</script>")
}

// Project represents a scraping project
type Project struct {
	ID         string       `json:"project_id"`
	URL        string       `json:"url"`
	Depth      int          `json:"depth"`
	Status     ProjectStatus `json:"status"`
	Filters    []FilterRule `json:"filters"`
	Progress   int          `json:"progress"`
	Downloaded int          `json:"pages_downloaded"`
	Total      int          `json:"total_pages"`
	CurrentURL string       `json:"current_url"`
	Errors     []string     `json:"errors"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

// ProjectStatus represents project execution state
type ProjectStatus string

const (
	StatusStarted    ProjectStatus = "started"
	StatusInProgress ProjectStatus = "in_progress"
	StatusCompleted  ProjectStatus = "completed"
	StatusFailed     ProjectStatus = "failed"
)

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

// Asset represents a downloadable resource (image, CSS, JS, etc.)
type Asset struct {
	URL          string // Original URL
	LocalPath    string // Path in project folder
	Type         string // "image", "css", "js", "font", "other"
	Downloaded   bool
	Error        string
}

// Page represents a scraped HTML page
type Page struct {
	URL           string
	LocalPath     string // Relative path in project
	Depth         int
	ParentURL     string
	HTML          string
	Assets        []Asset
	Links         []string // Extracted links
	Downloaded    bool
	Processed     bool // Link transformation done
	Filtered      bool // Filters applied
	Error         string
}
```

**Verification**:
```bash
go build ./internal/models
```

---

### ✅ Zadanie 4: Setup Chi router + basic endpoints

**Akcje**:
1. Utworzenie `internal/api/routes.go` z konfiguracją Chi routera
2. Utworzenie `cmd/server/main.go` jako entry point
3. Setup podstawowych endpoints (stubs)

**Plik**: `internal/api/routes.go`

```go
package api

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SetupRoutes configures all HTTP routes
func SetupRoutes() *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// CORS for frontend (if needed)
	r.Use(corsMiddleware)

	// Static files (web UI)
	fileServer := http.FileServer(http.Dir("./web"))
	r.Handle("/*", fileServer)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Post("/scrape", HandleScrape)
		r.Get("/project/{id}/status", HandleStatus)
		r.Get("/project/{id}/export/zip", HandleExportZip)
		r.Post("/project/{id}/export/pdf", HandleExportPDF)
	})

	return r
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}
```

**Plik**: `internal/api/handlers.go` (stubs)

```go
package api

import (
	"encoding/json"
	"net/http"
	"github.com/go-chi/chi/v5"
)

// HandleScrape starts a new scraping job
func HandleScrape(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Agent 4
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Scrape endpoint - TODO",
	})
}

// HandleStatus returns scraping job status
func HandleStatus(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Agent 4
	projectID := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"project_id": projectID,
		"message":    "Status endpoint - TODO",
	})
}

// HandleExportZip exports project as ZIP
func HandleExportZip(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Agent 5
	projectID := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"project_id": projectID,
		"message":    "ZIP export - TODO",
	})
}

// HandleExportPDF generates and exports project as PDF
func HandleExportPDF(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Agent 5
	projectID := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"project_id": projectID,
		"message":    "PDF export - TODO",
	})
}
```

**Plik**: `cmd/server/main.go`

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"context"

	"github.com/user/scrapper/internal/api"
)

func main() {
	// Environment variables
	port := getEnv("PORT", "8080")
	dataDir := getEnv("DATA_DIR", "./data")

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Setup routes
	router := api.SetupRoutes()

	// HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped")
}

// getEnv retrieves environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
```

**Verification**:
```bash
go build ./cmd/server
./cmd/server/main.go &
curl http://localhost:8080/api/scrape
# Expected: JSON response {"message": "Scrape endpoint - TODO"}
pkill -f "cmd/server/main.go"
```

---

## Expected Output Files

Po ukończeniu Agenta 1 powinny istnieć:

```
✅ go.mod
✅ go.sum
✅ .gitignore
✅ cmd/server/main.go
✅ internal/models/types.go
✅ internal/api/routes.go
✅ internal/api/handlers.go
✅ Folder structure (cmd/, internal/, web/, data/)
```

---

## Verification Checklist

- [ ] `go mod tidy` wykonuje się bez błędów
- [ ] `go build ./...` kompiluje wszystkie pakiety
- [ ] `go run cmd/server/main.go` uruchamia serwer
- [ ] Serwer odpowiada na `http://localhost:8080`
- [ ] Wszystkie 4 API endpoints zwracają JSON (nawet jako stubs)
- [ ] Folder `data/` istnieje i jest w `.gitignore`

---

## Common Issues & Solutions

### Issue 1: Import path errors
**Symptom**: `package github.com/user/scrapper/internal/api: cannot find package`  
**Solution**: 
```bash
# Fix go.mod module path
go mod edit -module github.com/user/scrapper
# Lub zmień wszystkie importy na właściwą ścieżkę
```

### Issue 2: Port already in use
**Symptom**: `bind: address already in use`  
**Solution**:
```bash
# Zmień port
export PORT=8081
go run cmd/server/main.go
```

### Issue 3: Dependencies not found
**Symptom**: `cannot find package "github.com/gocolly/colly/v2"`  
**Solution**:
```bash
go mod download
go mod tidy
```

---

## Next Agent

Po ukończeniu **Agent 1**, możesz przejść do:
👉 **AGENT_02_SCRAPING.md** (Implementacja core scraping engine)

**Prerequisites verified**:
- ✅ Modele zdefiniowane (`models/types.go`)
- ✅ Struktura folderów gotowa
- ✅ Router działa

---

**Agent Status**: ⏳ TODO  
**Last Updated**: 17 lutego 2026
