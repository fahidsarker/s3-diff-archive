package restorer

import (
	"context"
	"os"
	"path"
	"s3-diff-archive/db"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/s3"
	"s3-diff-archive/utils"
	"strings"
)

func RestoreTask(task *utils.TaskConfig) error {
	dbc := db.FetchRemoteDB(task)
	defer dbc.Close()

	fileReg, err := db.FetchRegOfTask(task)
	if err != nil {
		panic(err)
	}

	if fileReg == "" {
		lg.Logs.Warn("No reg file found for task %s", task.ID)
		return nil
	}

	fileList := strings.Split(fileReg, "\n")
	lg.Logs.Info("Restoring task %s, files: %d...", task.ID, len(fileList))
	restoredPath := path.Join(task.WorkingDir, task.ID, "restored")
	for _, file := range fileList {
		if strings.TrimSpace(file) == "" {
			continue
		}
		downloadPath := path.Join(task.WorkingDir, task.ID, file)
		_ = os.RemoveAll(restoredPath)
		err := os.MkdirAll(restoredPath, 0755)
		if err != nil {
			return err
		}

		err = s3.DownloadFileFromS3(task.CreateS3Config(task.StorageClass), context.TODO(), file, downloadPath)
		if err != nil {
			return err
		}

		lg.Logs.Info("Downloaded file: %s", file)

		lg.Logs.Info("Extracting file: %s", file)
		utils.Unzip(downloadPath, restoredPath, task.Password)
		// _ = os.Remove(downloadPath)
		lg.Logs.Info("Extracted file: %s", file)
		lg.Logs.Info("File restored: %s", file)
		lg.Logs.Break()
	}
	lg.Logs.Info("Task %s restored in %s", task.ID, restoredPath)
	lg.Logs.Info("Validating task %s", task.ID)
	isEq, err := DirsEqual(task.Dir, restoredPath, task.Excludes)
	if err != nil {
		return err
	}
	if !isEq {
		lg.Logs.Error("Task %s is not equal to original task %s", task.ID, task.ID)
		return err
	}
	return nil
}
