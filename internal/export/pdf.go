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
