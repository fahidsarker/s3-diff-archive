package main

import (
	"s3-diff-archive/archiver"
	"s3-diff-archive/db"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/s3"
	"s3-diff-archive/scanner"
	"s3-diff-archive/utils"
)

func main() {

	config := utils.GetConfig("config.yaml")
	err := lg.InitLoggers(config)
	if err != nil {
		panic(err)
	}
	defer lg.CloseGlobalLoggers()

	task, _ := config.GetTask("photos")
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

	// args := os.Args
	// if len(args) < 3 {
	// 	fmt.Println("Usage: s3-diff-archive <command> <config-file-path>")
	// 	os.Exit(1)
	// }

	// config := utils.GetConfig(args[2])
	// err := lg.InitLoggers(config)
	// if err != nil {
	// 	panic(err)
	// }
	// defer lg.CloseGlobalLoggers()

	// switch args[1] {
	// case "archive":
	// 	archive(config)
	// case "view":
	// 	if len(args) < 5 {
	// 		fmt.Println("Usage: s3-diff-archive view <config-file-path> --task <task-id>")
	// 		os.Exit(1)
	// 	}
	// 	taskId := args[4]
	// 	task, err := config.GetTask(taskId)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	db.ViewDB(task)
	// default:
	// 	fmt.Println("Unknown command")
	// 	os.Exit(1)

	// }

}
