package main

import (
	"flag"
	"fmt"
	"os"
	"path"
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

		scannedRes, err := scanner.ScanTask(refDB.GetDB(), task)
		if err != nil {
			errors++
			lg.Logs.Error("%s", err.Error())
			task.Notify("scan", "error", err.Error())
			continue
		}
		task.Notify("scan", "success", fmt.Sprintf("Scanned %d files in task %s. Skipped %d files, Changed %d files", scannedRes.TotalScanned(), task.ID, len(scannedRes.SkippedFiles), len(scannedRes.UpdatedFiles)))
		zipPaths := archiver.ArchiveToZip(task, scannedRes)
		task.Notify("archive", "success", fmt.Sprintf("Archived %d files in task %s to %d zip files", len(scannedRes.UpdatedFiles), task.ID, len(zipPaths)))

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
			task.Notify("scan", "error", err.Error())
			continue
		}
		lg.Logs.Info("Processing task %s, dir: %s, s3 StorageClass: %s", task.ID, task.Dir, task.StorageClass)
		refDB := db.FetchRemoteDB(task)
		defer refDB.Close()

		scannedRes, err := scanner.ScanTask(refDB.GetDB(), task)
		if err != nil {
			errors++
			lg.Logs.Error("%s", err.Error())
			task.Notify("scan", "error", err.Error())
			continue
		}
		message := fmt.Sprintf("Scanned %d files in task %s. Skipped %d files, Changed %d files", scannedRes.TotalScanned(), task.ID, len(scannedRes.SkippedFiles), len(scannedRes.UpdatedFiles))
		lg.Logs.Info("%s", message)
		task.Notify("scan", "success", message)
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
		zips, err := restorer.ExperimentalDownloadArchivedZips(task)
		if err != nil {
			errors++
			lg.Logs.Error("%s", err.Error())
			continue
		}
		restorePath := path.Join(task.WorkingDir, task.ID, "restored_"+utils.NowTime())
		_ = os.RemoveAll(restorePath)
		err = restorer.RestoreFromZips(zips, restorePath, task.Password)
		if err != nil {
			errors++
			lg.Logs.Error("%s", err.Error())
			continue
		}
		utils.DeleteFils(zips)
		lg.Logs.Info("Task %s restored in %s", task.ID, restorePath)
	}
	lg.Logs.Info("Restorer completed. Total tasks: %d. Error occured: %d", len(config.Tasks), errors)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Define command-specific flag sets
	switch command {
	case "scan":
		runScanCommand()
	case "archive":
		runArchiveCommand()
	case "restore":
		runRestoreCommand()
	case "view":
		runViewCommand()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: s3-diff-archive <command> [flags]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  scan     - Scan directories for changes")
	fmt.Println("  archive  - Archive changed files to S3")
	fmt.Println("  restore  - Restore files from S3")
	fmt.Println("  view     - View database for a specific task")
	fmt.Println("")
	fmt.Println("Use 's3-diff-archive <command> -h' for command-specific help")
}

func runScanCommand() {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to configuration file (required)")
	envPath := fs.String("env", ".env", "Path to environment file")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s scan [flags]\n\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "Scan directories for changes\n\n")
		fmt.Fprintf(fs.Output(), "Flags:\n")
		fs.PrintDefaults()
	}

	fs.Parse(os.Args[2:])

	if *configPath == "" {
		fmt.Println("Error: -config flag is required")
		fs.Usage()
		os.Exit(1)
	}

	config := utils.GetConfig(*configPath, *envPath)
	initLoggersAndRun(config, func() {
		runScanner(config)
	})
}

func runArchiveCommand() {
	fs := flag.NewFlagSet("archive", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to configuration file (required)")
	envPath := fs.String("env", ".env", "Path to environment file")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s archive [flags]\n\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "Archive changed files to S3\n\n")
		fmt.Fprintf(fs.Output(), "Flags:\n")
		fs.PrintDefaults()
	}

	fs.Parse(os.Args[2:])

	if *configPath == "" {
		fmt.Println("Error: -config flag is required")
		fs.Usage()
		os.Exit(1)
	}

	config := utils.GetConfig(*configPath, *envPath)
	initLoggersAndRun(config, func() {
		runArchiner(config)
	})
}

func runRestoreCommand() {
	fs := flag.NewFlagSet("restore", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to configuration file (required)")
	envPath := fs.String("env", ".env", "Path to environment file")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s restore [flags]\n\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "Restore files from S3\n\n")
		fmt.Fprintf(fs.Output(), "Flags:\n")
		fs.PrintDefaults()
	}

	fs.Parse(os.Args[2:])

	if *configPath == "" {
		fmt.Println("Error: -config flag is required")
		fs.Usage()
		os.Exit(1)
	}

	config := utils.GetConfig(*configPath, *envPath)
	initLoggersAndRun(config, func() {
		runRestorer(config)
	})
}

func runViewCommand() {
	fs := flag.NewFlagSet("view", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to configuration file (required)")
	envPath := fs.String("env", ".env", "Path to environment file")
	taskId := fs.String("task", "", "Task ID to view (required)")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s view [flags]\n\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "View database for a specific task\n\n")
		fmt.Fprintf(fs.Output(), "Flags:\n")
		fs.PrintDefaults()
	}

	fs.Parse(os.Args[2:])

	if *configPath == "" {
		fmt.Println("Error: -config flag is required")
		fs.Usage()
		os.Exit(1)
	}

	if *taskId == "" {
		fmt.Println("Error: -task flag is required")
		fs.Usage()
		os.Exit(1)
	}

	config := utils.GetConfig(*configPath, *envPath)
	task, err := config.GetTask(*taskId)
	if err != nil {
		panic(err)
	}
	db.ViewDB(task)
}

func initLoggersAndRun(config *utils.Config, runFunc func()) {
	err := lg.InitLoggers(config)
	if err != nil {
		panic(err)
	}
	defer lg.CloseGlobalLoggers()

	lg.Logs.Info("S3 Bucket: %s", config.S3Bucket)
	lg.Logs.Info("S3 BasePath: %s", config.S3BasePath)
	lg.Logs.Info("S3 Max Zip Size: %d", config.MaxZipSize)
	lg.Logs.Info("Working Dir: %s", config.WorkingDir)
	lg.Logs.Info("Logs Dir: %s", config.LogsDir)
	lg.Logs.Info("Tasks: %d", len(config.Tasks))
	lg.Logs.Break()

	runFunc()
}
