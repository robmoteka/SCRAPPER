package export

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
