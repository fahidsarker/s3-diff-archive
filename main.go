package main

import (
	"fmt"
	"os"
	"s3-diff-archive/archiver"
	"s3-diff-archive/db"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/restorer"
	"s3-diff-archive/s3"
	"s3-diff-archive/scanner"
	"s3-diff-archive/utils"
)

func runArchiner(config *utils.Config) {
	errors := 0
	lg.Logs.Info("Archiver started")
	for i := range config.Tasks {
		lg.Logs.Break()
		task, err := config.GetTask(config.Tasks[i].ID)
		if err != nil {
			errors++
			lg.Logs.Error("%s", err.Error())
			continue
		}
		lg.Logs.Info("Processing task %s, dir: %s, s3 StorageClass: %s", task.ID, task.Dir, task.StorageClass)
		refDB := db.FetchRemoteDB(task)
		defer refDB.Close()

		scannedRes := scanner.ScanTask(refDB.GetDB(), task)
		zipPaths := archiver.ArchiveToZip(task, scannedRes)

		writeDB := db.NewDBInDir(task.WorkingDir)
		writeDB.InsertSfilesToDB(scannedRes.UpdatedFiles)
		writeDB.InsertSfilesToDB(scannedRes.UnChangedFiles)
		zippedDBPath, err := writeDB.CloseAndZip(task.Password)

		if err != nil {
			lg.Logs.Fatal("%s", err.Error())
		}

		uploader := &s3.TaskUploader{
			Task:          task,
			ArchivedFiles: zipPaths,
			DBZipPath:     zippedDBPath,
		}
		err = uploader.UploadAndDelete()
		if err != nil {
			lg.Logs.Fatal("%s", err.Error())
		}
		db.UpdateRegOfTask(task, zipPaths)
	}
	lg.Logs.Info("Archiver completed. Total tasks: %d. Error occured: %d", len(config.Tasks), errors)
}

func runScanner(config *utils.Config) {
	lg.Logs.Info("Scanner started")
	errors := 0
	for i := range config.Tasks {
		lg.Logs.Break()
		task, err := config.GetTask(config.Tasks[i].ID)
		if err != nil {
			errors++
			lg.Logs.Error("%s", err.Error())
			continue
		}
		lg.Logs.Info("Processing task %s, dir: %s, s3 StorageClass: %s", task.ID, task.Dir, task.StorageClass)
		refDB := db.FetchRemoteDB(task)
		defer refDB.Close()

		scannedRes := scanner.ScanTask(refDB.GetDB(), task)
		lg.Logs.Info("Scanned %d files in task %s. Skipped %d files, Changed %d files", len(scannedRes.UpdatedFiles)+len(scannedRes.UnChangedFiles)+len(scannedRes.SkippedFiles), task.ID, len(scannedRes.SkippedFiles), len(scannedRes.UpdatedFiles))
	}
	lg.Logs.Info("Scanner completed. Total tasks: %d. Error occured: %d", len(config.Tasks), errors)
}
func runRestorer(config *utils.Config) {
	lg.Logs.Info("Restorer started")
	errors := 0
	for i := range config.Tasks {
		lg.Logs.Break()
		task, err := config.GetTask(config.Tasks[i].ID)
		if err != nil {
			errors++
			lg.Logs.Error("%s", err.Error())
			continue
		}
		lg.Logs.Info("Processing task %s, dir: %s, s3 StorageClass: %s", task.ID, task.Dir, task.StorageClass)
		err = restorer.RestoreTask(task)
		if err != nil {
			errors++
			lg.Logs.Error("%s", err.Error())
			continue
		}
		lg.Logs.Info("Task %s restored", task.ID)
	}
	lg.Logs.Info("Restorer completed. Total tasks: %d. Error occured: %d", len(config.Tasks), errors)
}

func main() {

	args := os.Args
	if len(args) < 3 {
		fmt.Println("Usage: s3-diff-archive <command> <config-file-path>")
		os.Exit(1)
	}
	config := utils.GetConfig(args[2])
	err := lg.InitLoggers(config)
	if err != nil {
		panic(err)
	}
	defer lg.CloseGlobalLoggers()
	lg.Logs.Info("Config file: %s", args[2])
	lg.Logs.Info("S3 Bucket: %s", config.S3Bucket)
	lg.Logs.Info("S3 BasePath: %s", config.S3BasePath)
	lg.Logs.Info("S3 Max Zip Size: %d", config.MaxZipSize)
	lg.Logs.Info("Working Dir: %s", config.WorkingDir)
	lg.Logs.Info("Logs Dir: %s", config.LogsDir)
	lg.Logs.Info("Tasks: %d", len(config.Tasks))
	lg.Logs.Break()

	switch args[1] {
	case "scan":
		runScanner(config)
	case "archive":
		runArchiner(config)
	case "restore":
		runRestorer(config)
	case "view":
		if len(args) < 5 {
			fmt.Println("Usage: s3-diff-archive view <config-file-path> --task <task-id>")
			os.Exit(1)
		}
		taskId := args[4]
		task, err := config.GetTask(taskId)
		if err != nil {
			panic(err)
		}
		db.ViewDB(task)
	default:
		fmt.Println("Unknown command")
		os.Exit(1)

	}

}
