package archiver

import (
	// "archive/zip"
	"fmt"
	"os"
	"path"
	"s3-diff-archive/db"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/utils"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
)

type ZipDiffTaskResult struct {
	TaskConfig        *utils.TaskConfig
	ZippedFiles       int
	TotalScannedFiles int
	ArchivesPaths     []string
	DBZipPath         string
}

func ZipDiff(config *utils.Config) []ZipDiffTaskResult {

	var results []ZipDiffTaskResult

	for i := range config.Tasks {
		taskConfig, err := config.GetTask(config.Tasks[i].ID)
		if err != nil {
			lg.Logs.Fatal("%s", err.Error())
		}
		lg.Logs.Info("Executing task: " + taskConfig.ID)

		dbpath := path.Join(config.WorkingDir, taskConfig.ID, "db")
		refDBPath := db.FetchRemoteDB(taskConfig)
		refDB := db.GetDB(refDBPath)
		writeDB := db.GetDB(dbpath)

		utils.CopyFile(path.Join(refDBPath, ".history"), path.Join(dbpath, ".history"))

		task := NewDiffZipTask(config, taskConfig.ID)
		stats, _ := os.Stat(task.Dir)
		if !stats.IsDir() {
			lg.Logs.Error("baseDir is not a directory")
			panic("baseDir is not a directory")
		}

		zipIterator(refDB, writeDB, task, task.Dir)

		println("")
		refDB.Close()
		writeDB.Close()
		task.flush()

		lg.Logs.Info("Total Zip file created in task " + task.ID + ": " + fmt.Sprint(len(task.ZipFilePaths)))
		lg.Logs.Info("Total files in task " + task.ID + ": " + fmt.Sprint(task.TotalScannedFiles) + ", New files: " + fmt.Sprint(task.TotalChangedFiles))

		zippedDBPath := config.NewZipFileNameForTask(task.ID, 0, "_db")
		regKepper, err := lg.CreateLogger(path.Join(dbpath, ".history"), false, true)
		if err != nil {
			lg.Logs.Fatal("%s", err.Error())
		}
		for i := range task.ZipFilePaths {
			lg.Log(regKepper, utils.FileNameFromPath(task.ZipFilePaths[i]))
		}
		regKepper.Close()

		ArchiveDB(dbpath, task.Password, NewZipper(zippedDBPath))
		result := ZipDiffTaskResult{
			TaskConfig:        taskConfig,
			ZippedFiles:       len(task.ZipFilePaths),
			TotalScannedFiles: task.TotalScannedFiles,
			ArchivesPaths:     task.ZipFilePaths,
			DBZipPath:         zippedDBPath,
		}
		results = append(results, result)
		UploadZipDiffResult(result)
		Cleanup(result)
	}

	return results
}

func zipIterator(rdb *badger.DB, wdb *badger.DB, task *DiffZipTask, dirPath string) int {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() {
			zipIterator(rdb, wdb, task, dirPath+"/"+file.Name())
			// println("Total zipped files: For dir: ", file.Name(), totalZippedFiles)
		} else {
			task.TotalScannedFiles++

			// ignore .DS_Store files
			filePath := dirPath + "/" + file.Name()
			relativeFilePath := dirPath + "/" + file.Name()

			skip := false
			for _, patterm := range task.Excludes {
				if utils.MatchPattern(patterm, relativeFilePath) {
					task.TotalSkippedFiles++
					lg.Logs.Warn("Skipping file %s due to exclude pattern %s", filePath, patterm)
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			stats, err := os.Stat(dirPath + "/" + file.Name())
			if err != nil {
				panic(err)
			}

			if strings.HasPrefix(filePath, task.Task.Dir) {
				// remove baseDir from filePath
				relativeFilePath = strings.Replace(filePath, task.Task.Dir, "", 1)
			}
			fileUpdated := db.HasFileUpdated(rdb, wdb, filePath, relativeFilePath, stats)
			lg.ScanLog.Info(task.ID + "\t" + relativeFilePath + ", File Updated: " + fmt.Sprint(fileUpdated) + ", Size: " + fmt.Sprint(stats.Size()))
			if fileUpdated {
				task.TotalChangedFiles++
				task.Zip(dirPath+"/"+file.Name(), stats)
			}
		}
		fmt.Printf("\r>>> Scanned: %d files, New: %d files, Skipped: %d files, Zipped: %d files...", task.TotalScannedFiles, task.TotalChangedFiles, task.TotalSkippedFiles, task.TotalChangedFiles)
	}
	return task.zipper.fileCounts
}

func ArchiveDB(dbPath string, encryptPass string, zipper *Zipper) {
	files, err := os.ReadDir(dbPath)

	if err != nil {
		panic(err)
	}
	lg.Logs.Info("%s", "Total files in DB: "+fmt.Sprint(len(files)))
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
		// println(file.Name())
		zipper.zip(filePath, file.Name(), stats, encryptPass)
	}

	zipper.flush()

}
