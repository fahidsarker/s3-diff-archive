package s3

import (
	"context"
	"os"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/utils"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type TaskUploader struct {
	Task          *utils.TaskConfig
	ArchivedFiles []string
	DBZipPath     string
}

func (t *TaskUploader) Upload() error {
	if len(t.ArchivedFiles) == 0 {
		lg.Logs.Info("No files to upload in task %s. Continuing...", t.Task.ID)
		return nil
	}
	lg.Logs.Info("Uploading task %s, files: %d", t.Task.ID, len(t.ArchivedFiles))
	for _, file := range t.ArchivedFiles {
		err := UploadFileToS3(t.Task.CreateS3Config(t.Task.StorageClass), context.TODO(), utils.FileNameFromPath(file), file)
		if err != nil {
			return err
		}
	}
	err := UploadFileToS3(t.Task.CreateS3Config(types.StorageClassStandard), context.TODO(), "db.zip", t.DBZipPath)
	if err != nil {
		return err
	}
	lg.Logs.Info("Task %s uploaded", t.Task.ID)
	return nil
}

func (t *TaskUploader) UploadAndDelete() error {
	err := t.Upload()
	if err != nil {
		return err
	}

	lg.Logs.Info("Deleting temp files of task %s, files: %d", t.Task.ID, len(t.ArchivedFiles))
	err = os.Remove(t.DBZipPath)
	if err != nil {
		return err
	}
	for _, file := range t.ArchivedFiles {
		err = os.Remove(file)
		if err != nil {
			return err
		}
	}
	lg.Logs.Info("Task %s temp files deleted", t.Task.ID)
	return nil
}
