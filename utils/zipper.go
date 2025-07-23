package utils

import (
	// "archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexmullins/zip"
)

func ZipFile(filePath string, filename string, fileStat os.FileInfo, zipWriter *zip.Writer, password string) {
	fileToZip, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open file %s: %v", filePath, err)
	}
	defer fileToZip.Close()

	var w io.Writer
	header, err := zip.FileInfoHeader(fileStat)
	if err != nil {
		log.Fatalf("failed to create zip header: %v", err)
	}
	header.Name = filename
	header.Method = zip.Deflate
	if password != "" {
		header.SetPassword(password)
	}
	w, err = zipWriter.CreateHeader(header)
	if err != nil {
		log.Fatalf("failed to create zip entry: %v", err)
	}

	_, err = io.Copy(w, fileToZip)
	if err != nil {
		log.Fatalf("failed to copy file data to zip: %v", err)
	}
}

func Unzip(zipPath, destDir, password string) error {
	readCloser, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file %s: %w", zipPath, err)
	}
	defer readCloser.Close()

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
	}

	for _, file := range readCloser.File {
		// Construct the full path for the extracted file
		// Sanitize the file path to prevent directory traversal attacks
		filePath := filepath.Join(destDir, file.Name)
		if !strings.HasPrefix(filePath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in zip: %s", file.Name)
		}

		// Check if it's a directory
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, file.Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", filePath, err)
			}
			continue
		}

		// Handle password if set
		if file.IsEncrypted() && password != "" {
			file.SetPassword(password)
		} else if file.IsEncrypted() && password == "" {
			fmt.Printf("Warning: File '%s' is password-protected but no password was provided. Skipping or it may fail.\n", file.Name)
			// You might choose to skip this file or return an error here
			// For this example, we'll let file.Open() potentially fail.
		}

		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s in zip: %w", file.Name, err)
		}
		defer rc.Close() // Defer closing inside the loop for each file

		// Ensure the directory for the current file exists
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for file %s: %w", filePath, err)
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", filePath, err)
		}
		defer outFile.Close() // Defer closing inside the loop for each file

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return fmt.Errorf("failed to copy data for file %s: %w", filePath, err)
		}
	}

	return nil
}
