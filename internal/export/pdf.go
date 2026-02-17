package export

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jung-kurt/gofpdf"
)

// CreatePDF generates a consolidated PDF from all HTML files in the project
func CreatePDF(projectPath string, w io.Writer) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)

	// Process index.html first
	indexPath := filepath.Join(projectPath, "index.html")
	if err := addHTMLToPDF(pdf, indexPath, "Main Page"); err != nil {
		fmt.Printf("Warning: failed to add index.html to PDF: %v\n", err)
	}

	// Process all pages in the pages directory
	pagesDir := filepath.Join(projectPath, "pages")
	if _, err := os.Stat(pagesDir); err == nil {
		err = filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && (strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".htm")) {
				// Use filename as section title
				title := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
				if err := addHTMLToPDF(pdf, path, title); err != nil {
					fmt.Printf("Warning: failed to add %s to PDF: %v\n", path, err)
				}
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to walk pages directory: %w", err)
		}
	}

	// Write PDF to output
	if err := pdf.Output(w); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	return nil
}

// addHTMLToPDF adds a single HTML file to the PDF as a new section
func addHTMLToPDF(pdf *gofpdf.Fpdf, htmlPath, title string) error {
	// Read HTML file
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		return fmt.Errorf("failed to read HTML file: %w", err)
	}

	// Parse HTML and extract text
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(content)))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Add new page
	pdf.AddPage()

	// Add section title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, title)
	pdf.Ln(12)

	// Reset font for content
	pdf.SetFont("Arial", "", 11)

	// Extract and add text content
	// Remove script and style tags
	doc.Find("script, style").Remove()

	// Get body text
	bodyText := doc.Find("body").Text()

	// Clean up whitespace
	lines := strings.Split(bodyText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Encode text for PDF (handle special characters)
		line = strings.Map(func(r rune) rune {
			// Keep only printable ASCII characters
			if r >= 32 && r < 127 {
				return r
			}
			return ' '
		}, line)

		// Add line to PDF with word wrapping
		pdf.MultiCell(0, 5, line, "", "", false)
	}

	return nil
}
