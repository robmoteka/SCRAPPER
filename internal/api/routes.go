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

	// CORS for frontend
	r.Use(corsMiddleware)

	// Static files (web UI)
	fileServer := http.FileServer(http.Dir("./web"))
	r.Handle("/*", fileServer) // Serve index.html and assets

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
