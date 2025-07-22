package utils

import (
	// "archive/zip"
	"io"
	"log"
	"os"

	"github.com/alexmullins/zip"
)

func ZipFile(filePath string, filename string, fileStat os.FileInfo, zipWriter *zip.Writer, password string) {
	fileToZip, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open file %s: %v", filePath, err)
	}
	defer fileToZip.Close()

	var w io.Writer
	if password == "" {
		header, err := zip.FileInfoHeader(fileStat)
		if err != nil {
			log.Fatalf("failed to create zip header: %v", err)
		}
		header.Name = filename
		header.Method = zip.Deflate

		w, err = zipWriter.CreateHeader(header)
		if err != nil {
			log.Fatalf("failed to create zip entry: %v", err)
		}
	} else {
		// Assumes zipWriter.Encrypt is available from a 3rd-party library like "github.com/alexmullins/zip"
		w, err = zipWriter.Encrypt(filePath, password)
		if err != nil {
			log.Fatalf("failed to create encrypted zip entry: %v", err)
		}
	}

	_, err = io.Copy(w, fileToZip)
	if err != nil {
		log.Fatalf("failed to copy file data to zip: %v", err)
	}
}
