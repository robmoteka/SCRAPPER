# AGENT 5: Export Features

**Phase**: Export Features  
**Zadania**: 11-13  
**Dependencies**: Agent 4 (API layer)  
**Estimated Time**: 40-50 minut

---

## Cel Agenta

Implementacja eksportu projektów do ZIP oraz PDF (konsolidacja wszystkich stron w jeden dokument).

---

## Prerequisites

Przed rozpoczęciem sprawdź:
- [x] Agent 4 ukończony (API działa)
- [x] Projects zapisują się do `data/{id}/`
- [x] Dependencies: `gofpdf` zainstalowany

---

## Zadania do Wykonać

### ✅ Zadanie 11: Implement ZIP export

**Cel**: Rekurencyjne pakowanie folderu projektu do ZIP z streaming.

**Plik**: `internal/export/zip.go`

```go
package export

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CreateZipArchive creates a ZIP file from project directory
func CreateZipArchive(projectID, dataDir string) (string, error) {
	projectDir := filepath.Join(dataDir, projectID)
	zipPath := filepath.Join(dataDir, projectID+".zip")

	// Create ZIP file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	// Create ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk project directory
	err = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the project directory itself
		if path == projectDir {
			return nil
		}

		// Get relative path for ZIP entry
		relPath, err := filepath.Rel(projectDir, path)
		if err != nil {
			return err
		}

		// Create ZIP entry header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Use forward slashes for ZIP compatibility
		header.Name = filepath.ToSlash(relPath)

		// Handle directories
		if info.IsDir() {
			header.Name += "/"
		} else {
			// Set compression method
			header.Method = zip.Deflate
		}

		// Create entry writer
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Write file content (if not directory)
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(writer, file); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to archive project: %w", err)
	}

	return zipPath, nil
}

// StreamZipToWriter streams ZIP archive directly to HTTP response
func StreamZipToWriter(w io.Writer, projectID, dataDir string) error {
	projectDir := filepath.Join(dataDir, projectID)

	// Create ZIP writer directly to response
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// Walk and stream
	return filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == projectDir {
			return nil
		}

		relPath, err := filepath.Rel(projectDir, path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(relPath)

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(writer, file); err != nil {
				return err
			}
		}

		return nil
	})
}

// CleanupZipFile removes temporary ZIP file
func CleanupZipFile(zipPath string) error {
	return os.Remove(zipPath)
}
```

**Verification**:
```bash
go build ./internal/export
```

---

### ✅ Zadanie 12: Implement PDF export (consolidated)

**Cel**: Generowanie jednego PDF z wszystkimi stronami jako rozdziały.

**Plik**: `internal/export/pdf.go`

```go
package export

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// CreateConsolidatedPDF generates a single PDF from all HTML pages
func CreateConsolidatedPDF(projectID, dataDir string) (string, error) {
	projectDir := filepath.Join(dataDir, projectID)
	pdfPath := filepath.Join(dataDir, projectID+".pdf")

	// Initialize PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetFont("Arial", "", 12)

	// Find all HTML files
	htmlFiles, err := findHTMLFiles(projectDir)
	if err != nil {
		return "", fmt.Errorf("failed to find HTML files: %w", err)
	}

	if len(htmlFiles) == 0 {
		return "", fmt.Errorf("no HTML files found in project")
	}

	// Process each HTML file as a chapter
	for i, htmlPath := range htmlFiles {
		// Determine chapter title
		chapterTitle := getChapterTitle(htmlPath, projectDir, i)

		// Add chapter
		if err := addChapterToPDF(pdf, htmlPath, chapterTitle); err != nil {
			return "", fmt.Errorf("failed to add chapter %s: %w", chapterTitle, err)
		}
	}

	// Save PDF
	if err := pdf.OutputFileAndClose(pdfPath); err != nil {
		return "", fmt.Errorf("failed to save PDF: %w", err)
	}

	return pdfPath, nil
}

// findHTMLFiles recursively finds all HTML files in project
func findHTMLFiles(projectDir string) ([]string, error) {
	var htmlFiles []string

	// Priority: index.html first
	indexPath := filepath.Join(projectDir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		htmlFiles = append(htmlFiles, indexPath)
	}

	// Then pages directory
	pagesDir := filepath.Join(projectDir, "pages")
	if _, err := os.Stat(pagesDir); err == nil {
		err := filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
				htmlFiles = append(htmlFiles, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return htmlFiles, nil
}

// getChapterTitle determines chapter title from file path
func getChapterTitle(htmlPath, projectDir string, index int) string {
	relPath, _ := filepath.Rel(projectDir, htmlPath)
	
	if relPath == "index.html" {
		return "Main Page"
	}

	// Use filename without extension
	base := filepath.Base(htmlPath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	
	return fmt.Sprintf("Chapter %d: %s", index+1, name)
}

// addChapterToPDF adds HTML content as PDF chapter
func addChapterToPDF(pdf *gofpdf.Fpdf, htmlPath, title string) error {
	// Read HTML
	htmlBytes, err := os.ReadFile(htmlPath)
	if err != nil {
		return err
	}

	// Convert HTML to plain text
	text := htmlToText(string(htmlBytes))

	// Add page with chapter title
	pdf.AddPage()
	
	// Chapter heading
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, title, "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Content
	pdf.SetFont("Arial", "", 11)
	
	// Split text into lines and add to PDF
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			pdf.Ln(4)
			continue
		}

		// MultiCell for text wrapping
		pdf.MultiCell(0, 5, line, "", "L", false)
	}

	return nil
}

// htmlToText strips HTML tags and returns plain text
func htmlToText(html string) string {
	// Remove script and style tags with content
	reScript := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	html = reScript.ReplaceAllString(html, "")

	reStyle := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	html = reStyle.ReplaceAllString(html, "")

	// Remove HTML comments
	reComment := regexp.MustCompile(`<!--.*?-->`)
	html = reComment.ReplaceAllString(html, "")

	// Remove all HTML tags
	reTag := regexp.MustCompile(`<[^>]*>`)
	text := reTag.ReplaceAllString(html, "")

	// Decode HTML entities (basic)
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	// Clean up whitespace
	text = strings.TrimSpace(text)
	
	// Normalize multiple spaces
	reSpaces := regexp.MustCompile(`\s+`)
	text = reSpaces.ReplaceAllString(text, " ")

	// Normalize newlines
	reNewlines := regexp.MustCompile(`\n{3,}`)
	text = reNewlines.ReplaceAllString(text, "\n\n")

	return text
}

// StreamPDFToWriter generates and streams PDF directly to writer
func StreamPDFToWriter(w io.Writer, projectID, dataDir string) error {
	// Generate PDF to temp file first (gofpdf limitation)
	pdfPath, err := CreateConsolidatedPDF(projectID, dataDir)
	if err != nil {
		return err
	}
	defer os.Remove(pdfPath) // Cleanup temp file

	// Stream to writer
	file, err := os.Open(pdfPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	return err
}
```

**Verification**:
```bash
go build ./internal/export
```

---

### ✅ Zadanie 13: Add export API handlers

**Cel**: Połączenie export logic z API endpoints.

**Update**: `internal/api/handlers.go` - replace export stubs

```go
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
```

**Add import**:
```go
import (
	"github.com/user/scrapper/internal/export"
	"log"
)
```

**Verification**:
```bash
go build ./internal/api
go build ./cmd/server
```

---

## Integration Test

```bash
# Start server
go run cmd/server/main.go &

# Create scraping job
curl -X POST http://localhost:8080/api/scrape \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","depth":1,"filters":[]}'

# Save project ID
PROJECT_ID="uuid-from-response"

# Wait for completion
sleep 30

# Test ZIP export
curl -o project.zip http://localhost:8080/api/project/$PROJECT_ID/export/zip

# Verify ZIP
unzip -l project.zip
# Expected: index.html, assets/, pages/, etc.

# Test PDF export
curl -o project.pdf http://localhost:8080/api/project/$PROJECT_ID/export/pdf

# Verify PDF
file project.pdf
# Expected: PDF document

# Open PDF (Linux)
xdg-open project.pdf

pkill -f "cmd/server/main.go"
```

---

## Expected Output Files

Po ukończeniu Agenta 5:

```
✅ internal/export/zip.go
✅ internal/export/pdf.go
✅ Updated internal/api/handlers.go
✅ ZIP export działa
✅ PDF export działa (consolidated)
✅ Export API endpoints działają
```

---

## Verification Checklist

- [ ] `go build ./...` kompiluje cały projekt
- [ ] GET /api/project/{id}/export/zip zwraca valid ZIP
- [ ] POST /api/project/{id}/export/pdf zwraca valid PDF
- [ ] ZIP zawiera wszystkie pliki projektu
- [ ] PDF zawiera wszystkie strony jako rozdziały
- [ ] Export działa tylko dla completed projects
- [ ] Error handling (project not found, not completed)

---

## Common Issues & Solutions

### Issue 1: PDF empty or corrupted
**Symptom**: PDF file opens but is blank  
**Solution**: Check HTML parsing, ensure `htmlToText` preserves content

### Issue 2: ZIP download fails
**Symptom**: `unexpected EOF` during download  
**Solution**: Ensure all file handles closed, check streaming logic

### Issue 3: Memory issues with large projects
**Symptom**: OOM during export  
**Solution**: Use streaming (`StreamZipToWriter`), avoid loading all in memory

---

## Next Agent

Po ukończeniu **Agent 5**, przejdź do:
👉 **AGENT_06_FRONTEND.md** (Web UI implementacja)

**Prerequisites verified**:
- ✅ Export features działają
- ✅ ZIP i PDF streaming implemented
- ✅ API complete (scrape → status → export)

---

**Agent Status**: ⏳ TODO  
**Last Updated**: 17 lutego 2026
