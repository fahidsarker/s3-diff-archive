package restorer

import (
	"context"
	"path"
	"s3-diff-archive/db"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/s3"
	"s3-diff-archive/utils"
	"strings"
)

func RestoreFromZips(zipPaths []string, outputPath string, password string) error {
	for _, zipPath := range zipPaths {
		err := utils.Unzip(zipPath, outputPath, password)
		if err != nil {
			return err
		}
	}
	return nil
}

func ExperimentalDownloadArchivedZips(task *utils.TaskConfig) ([]string, error) {
	fileReg, err := db.FetchRegOfTask(task)
	zipPaths := []string{}
	if err != nil {
		panic(err)
	}

	if fileReg == "" {
		lg.Logs.Warn("No reg file found for task %s", task.ID)
		return zipPaths, nil
	}

	fileList := strings.Split(fileReg, "\n")
	lg.Logs.Info("Downlaoding Archived Zips for task %s, files: %d...", task.ID, len(fileList))
	for _, file := range fileList {
		if strings.TrimSpace(file) == "" {
			continue
		}
		downloadPath := path.Join(task.WorkingDir, task.ID, file)

		err = s3.DownloadFileFromS3(task.CreateS3Config(task.StorageClass), context.TODO(), file, downloadPath)
		if err != nil {
			return []string{}, err
		}

		lg.Logs.Info("Downloaded file: %s", file)
		zipPaths = append(zipPaths, downloadPath)
	}
	lg.Logs.Info("Task %s downloaded in %s", task.ID, strings.Join(zipPaths, ", "))

	return zipPaths, nil
}
