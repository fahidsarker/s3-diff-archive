package utils

import (
	// "archive/zip"
	"io"
	"log"
	"os"

	"github.com/alexmullins/zip"
)

func ZipFiles(filePaths []string, outputFile string) {
	// Need to open the outputFile for writing.
	outFile, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	for _, file := range filePaths {
		fileToZip, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer fileToZip.Close()

		// Get the file information
		info, err := fileToZip.Stat()
		if err != nil {
			panic(err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			panic(err)
		}

		// Using FileInfoHeader() above only uses the basename of the file. If we want
		// to preserve the folder structure we can overwrite this with the full path.
		header.Name = file

		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(writer, fileToZip)
		if err != nil {
			panic(err)
		}

	}
}

// func main() {
// 	contents := []byte("Hello World")
// 	fzip, err := os.Create(`./test.zip`)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	zipw := zip.NewWriter(fzip)
// 	defer zipw.Close()
// 	w, err := zipw.Encrypt(`test.txt`, `golang`)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	_, err = io.Copy(w, bytes.NewReader(contents))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	zipw.Flush()
// }

func ZipFile(file string, zipWriter *zip.Writer) {
	// Need to open the outputFile for writing.

	fileToZip, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer fileToZip.Close()

	// Get the file information
	// info, err := fileToZip.Stat()
	// if err != nil {
	// 	panic(err)
	// }

	// header, err := zip.FileInfoHeader(info)
	// if err != nil {
	// 	panic(err)
	// }

	// to preserve the folder structure we can overwrite this with the full path.
	// header.Name = file

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	// header.Method = zip.Deflate

	// writer, err := zipWriter.CreateHeader()(header)
	// if err != nil {
	// 	panic(err)
	// }

	w, err := zipWriter.Encrypt(file, `golang`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(w, fileToZip)
	if err != nil {
		panic(err)
	}

}
