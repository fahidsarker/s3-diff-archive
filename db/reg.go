package db

import (
	"context"
	"fmt"
	"os"
	"path"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/s3"
	"s3-diff-archive/utils"

	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func FetchRegOfTask(task *utils.TaskConfig) (string, error) {
	localRegPath := path.Join(task.WorkingDir, fmt.Sprintf("%s-%s.txt", task.ID, utils.RandAndTime(5)))
	err := s3.DownloadFileFromS3(task.CreateS3Config(s3Types.StorageClassStandard), context.TODO(), fmt.Sprintf("reg-%s.txt", task.ID), localRegPath)
	if err != nil {
		if err.Error() == "not-found" {
			return "", nil
		}
		return "", err
	}
	fileStr, err := os.ReadFile(localRegPath)
	if err != nil {
		return "", err
	}
	_ = os.Remove(localRegPath)
	return string(fileStr), nil
}

func UpdateRegOfTask(task *utils.TaskConfig, newArchived []string) error {
	if len(newArchived) == 0 {
		lg.Logs.Info("No new files to update reg file for task %s. Continuing...", task.ID)
		return nil
	}
	lg.Logs.Info("Updating reg file for task %s, new files: %d", task.ID, len(newArchived))
	reg, err := FetchRegOfTask(task)
	if err != nil {
		return err
	}
	for _, file := range newArchived {
		reg += fmt.Sprintf("%s\n", utils.FileNameFromPath(file))
	}

	// write to a file
	filePath := path.Join(task.WorkingDir, fmt.Sprintf("%s-%s.txt", task.ID, utils.RandAndTime(5)))
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	_, err = file.WriteString(reg)
	if err != nil {
		return err
	}
	file.Close()

	err = s3.UploadFileToS3(task.CreateS3Config(s3Types.StorageClassStandard), context.TODO(), fmt.Sprintf("reg-%s.txt", task.ID), filePath)
	if err != nil {
		return err
	}

	lg.Logs.Info("Reg file updated for task %s", task.ID)

	// remove the file
	_ = os.Remove(file.Name())
	return nil
}
