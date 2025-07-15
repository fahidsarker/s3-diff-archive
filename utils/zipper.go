package utils

import (
	// "archive/zip"
	"io"
	"log"
	"os"

	"github.com/alexmullins/zip"
)

func ZipFile(file string, zipWriter *zip.Writer) {
	// Need to open the outputFile for writing.
	fileToZip, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer fileToZip.Close()

	w, err := zipWriter.Encrypt(file, `golang`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(w, fileToZip)
	if err != nil {
		panic(err)
	}

}
