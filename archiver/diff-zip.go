package archiver

import (
	// "archive/zip"
	"fmt"
	"os"
	"path"
	"s3-diff-archive/logger"
	"s3-diff-archive/utils"

	badger "github.com/dgraph-io/badger/v4"
)

func ZipDiff(config utils.Config) int {

	totalFiles := 0

	for _, taskConfig := range config.Tasks {
		logger.Logs.Info("Executing task: " + taskConfig.ID)
		dbpath := path.Join(config.WorkingDir, taskConfig.ID, "db")
		db := GetDB(dbpath)
		task := NewDiffZipTask(config, taskConfig.ID)
		stats, _ := os.Stat(task.Dir)
		if !stats.IsDir() {
			logger.Logs.Error("baseDir is not a directory")
			panic("baseDir is not a directory")
		}

		totalFiles += zipIterator(db, task, task.Dir)
		println("")
		logger.Logs.Info("Total Zip file created in task " + task.ID + ": " + fmt.Sprint(len(task.ZipFilePaths)))
		logger.Logs.Info("Total files in task " + task.ID + ": " + fmt.Sprint(task.TotalScannedFiles) + ", New files: " + fmt.Sprint(task.TotalChangedFiles))
		db.Close()
		task.flush()
		ArchiveDB(dbpath, config.DBConfig.Encrypt, NewZipper(config.NewZipFileNameForTask(task.ID, 0, "_db")))
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
			fileUpdated := hasFileUpdated(db, dirPath+"/"+file.Name(), stats)
			logger.ScanLog.Info(task.ID + "\t" + dirPath + "/" + file.Name() + ", File Updated: " + fmt.Sprint(fileUpdated) + ", Size: " + fmt.Sprint(stats.Size()))
			if fileUpdated {
				task.TotalChangedFiles++
				task.Zip(dirPath+"/"+file.Name(), stats)
			}
		}
		fmt.Printf("\r>>> Scanned: %d files, New: %d files, Zipped: %d files...", task.TotalScannedFiles, task.TotalChangedFiles, task.TotalChangedFiles)
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
	logger.Logs.Info("Total files in DB: " + fmt.Sprint(len(files)))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
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
