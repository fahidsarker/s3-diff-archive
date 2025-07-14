package archiver

import (
	// "archive/zip"
	"fmt"
	"os"
	"s3-archive/utils"

	"github.com/alexmullins/zip"

	badger "github.com/dgraph-io/badger/v4"
)

func ZipDiff(baseDir string, outputFile string) int {
	db := GetDB()
	defer db.Close()
	stats, _ := os.Stat(baseDir)
	if !stats.IsDir() {
		panic("baseDir is not a directory")
	}

	outFile, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}

	defer outFile.Close()
	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	res := zipIterator(db, baseDir, zipWriter)
	zipWriter.Flush()
	println("")
	return res
}

var totalZippedFiles = 0
var totalScannedFiles = 0

func zipIterator(db *badger.DB, baseDir string, zipWriter *zip.Writer) int {
	files, err := os.ReadDir(baseDir)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() {
			zipIterator(db, baseDir+"/"+file.Name(), zipWriter)
			// println("Total zipped files: For dir: ", file.Name(), totalZippedFiles)
		} else {
			if file.Name() == ".DS_Store" {
				continue
			}
			totalScannedFiles++
			if hasFileUpdated(db, baseDir+"/"+file.Name()) {
				totalZippedFiles++
				utils.ZipFile(baseDir+"/"+file.Name(), zipWriter)
			}
		}
		fmt.Printf("\r>>> Scanned: %d files, New: %d files, Zipped: %d files...", totalScannedFiles, totalZippedFiles, totalZippedFiles)
	}
	return totalZippedFiles
}

func ArchiveDB(outputPath string) {
	dbDir := "./tmp/db"
	outFile, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()
	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	files, err := os.ReadDir(dbDir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		utils.ZipFile(dbDir+"/"+file.Name(), zipWriter)
	}
	println("Total zipped files: ", len(files))
}
