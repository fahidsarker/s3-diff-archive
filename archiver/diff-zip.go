package archiver

import (
	// "archive/zip"
	"fmt"
	"os"
	"s3-diff-archive/utils"

	badger "github.com/dgraph-io/badger/v4"
)

func ZipDiff(config utils.Config) int {

	db := GetDB()
	defer db.Close()
	totalFiles := 0

	for _, taskConfig := range config.Tasks {
		println(">>> Executing task: ", taskConfig.ID)
		task := NewDiffZipTask(config, taskConfig.ID)
		stats, _ := os.Stat(task.BaseDir)
		if !stats.IsDir() {
			panic("baseDir is not a directory")
		}

		totalFiles += zipIterator(db, task, task.BaseDir)
		task.flush()
		println("")
	}

	return totalFiles
}

func zipIterator(db *badger.DB, task *DiffZipTask, dirPath string) int {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() {
			zipIterator(db, task, dirPath+"/"+file.Name())
			// println("Total zipped files: For dir: ", file.Name(), totalZippedFiles)
		} else {
			if file.Name() == ".DS_Store" {
				continue
			}
			task.TotalScannedFiles++
			stats, err := os.Stat(dirPath + "/" + file.Name())
			if err != nil {
				panic(err)
			}
			if hasFileUpdated(db, dirPath+"/"+file.Name(), stats) {
				// totalZippedFiles++
				task.TotalChangedFiles++
				task.Zip(dirPath+"/"+file.Name(), stats)
				// utils.ZipFile(baseDir+"/"+file.Name(), zipWriter)
			}
		}
		fmt.Printf("\r>>> Scanned: %d files, New: %d files, Zipped: %d files...", task.TotalScannedFiles, task.TotalChangedFiles, task.zipper.fileCounts)
	}
	return task.zipper.fileCounts
}

// func ArchiveDB(outputPath string) {
// 	dbDir := "./tmp/db"
// 	outFile, err := os.Create(outputPath)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer outFile.Close()
// 	zipWriter := zip.NewWriter(outFile)
// 	defer zipWriter.Close()

// 	files, err := os.ReadDir(dbDir)
// 	if err != nil {
// 		panic(err)
// 	}
// 	for _, file := range files {
// 		if file.IsDir() {
// 			continue
// 		}
// 		fileStat:= os.Stat(dbDir+"/"+file.Name())
// 		utils.ZipFile(dbDir+"/"+file.Name(), zipWriter, "golang")
// 	}
// 	println("Total zipped files: ", len(files))
// }
