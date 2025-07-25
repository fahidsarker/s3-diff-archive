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

func ZipFile(filePath string, filename string, fileStat *os.FileInfo, zipWriter *zip.Writer, password string) {
	fileToZip, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open file %s: %v", filePath, err)
	}
	defer fileToZip.Close()

	var w io.Writer
	header, err := zip.FileInfoHeader(*fileStat)
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

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
	}

	for _, file := range readCloser.File {
		filePath := filepath.Join(destDir, file.Name)
		if !strings.HasPrefix(filePath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in zip: %s", file.Name)
		}

		modTime := file.ModTime()

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, file.Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", filePath, err)
			}
			// Set modification time
			if err := os.Chtimes(filePath, modTime, modTime); err != nil {
				return fmt.Errorf("failed to set mod time for directory %s: %w", filePath, err)
			}
			continue
		}

		if file.IsEncrypted() && password != "" {
			file.SetPassword(password)
		}

		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s in zip: %w", file.Name, err)
		}

		// Ensure parent dir exists
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			rc.Close()
			return fmt.Errorf("failed to create directory for file %s: %w", filePath, err)
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create output file %s: %w", filePath, err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return fmt.Errorf("failed to copy data for file %s: %w", filePath, err)
		}

		// Set the mod time after writing
		if err := os.Chtimes(filePath, modTime, modTime); err != nil {
			return fmt.Errorf("failed to set mod time for file %s: %w", filePath, err)
		}
	}

	return nil
}
