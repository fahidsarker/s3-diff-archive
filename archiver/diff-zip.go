package archiver

import (
	// "archive/zip"
	"fmt"
	"os"
	"path"
	"s3-diff-archive/utils"

	badger "github.com/dgraph-io/badger/v4"
)

func ZipDiff(config utils.Config) int {

	totalFiles := 0

	for _, taskConfig := range config.Tasks {
		dbpath := path.Join(config.WorkingDir, taskConfig.ID, "db")
		db := GetDB(dbpath)
		println(">>> Executing task: ", taskConfig.ID)
		task := NewDiffZipTask(config, taskConfig.ID)
		stats, _ := os.Stat(task.Dir)
		if !stats.IsDir() {
			panic("baseDir is not a directory")
		}

		totalFiles += zipIterator(db, task, task.Dir)
		db.Close()
		task.flush()
		ArchiveDB(dbpath, config.DBConfig.Encrypt, NewZipper(config.NewZipFileNameForTask(task.ID, 0, "_db")))
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
			// ignore .DS_Store files
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

func ArchiveDB(dbPath string, encrypt bool, zipper *Zipper) {
	files, err := os.ReadDir(dbPath)
	password := ""
	if encrypt {
		// password = utils.GenerateRandString(16)
	}
	if err != nil {
		panic(err)
	}
	println("Total files in DB: ", len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		println("Zipping file: ", file.Name())
		filePath := dbPath + "/" + file.Name()
		stats, err := os.Stat(filePath)
		if err != nil {
			panic(err)
		}
		// utils.ZipFile(filePath, stats, zipper.zw, password)
		zipper.zip(filePath, file.Name(), stats, password)
	}

	zipper.flush()

}
