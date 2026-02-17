package export

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CreateZip creates a ZIP archive of the project directory
func CreateZip(projectPath string, w io.Writer) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// Walk through all files in the project directory
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(projectPath, path)
		if err != nil {
			return err
		}

		// Create ZIP entry
		zipFile, err := zipWriter.Create(relPath)
		if err != nil {
			return fmt.Errorf("failed to create zip entry for %s: %w", relPath, err)
		}

		// Open source file
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer file.Close()

		// Copy file content to ZIP
		if _, err := io.Copy(zipFile, file); err != nil {
			return fmt.Errorf("failed to write file %s to zip: %w", relPath, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create zip: %w", err)
	}

	return nil
}
