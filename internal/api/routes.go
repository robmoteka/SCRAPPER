package api

import (
	"github.com/go-chi/chi/v5"
)

// SetupRoutes configures all API routes
func SetupRoutes(r chi.Router, dataDir string) {
	h := NewHandler(dataDir)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Scraping endpoints
		r.Post("/scrape", h.HandleScrape)
		r.Get("/project/{id}/status", h.HandleStatus)

		// Export endpoints
		r.Get("/project/{id}/export/zip", h.HandleZipExport)
		r.Post("/project/{id}/export/pdf", h.HandlePDFExport)
	})
}
