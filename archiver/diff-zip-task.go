package archiver

import (
	"fmt"
	"os"
	"s3-diff-archive/utils"
)

type DiffZipTask struct {
	utils.Task
	TaskName           string
	WorkingDir         string
	MaxZipSizeInMBytes int64
	currentZWIdx       int
	zipper             *Zipper
	TotalScannedFiles  int
	TotalChangedFiles  int
}

func NewDiffZipTask(config utils.Config, taskId string) *DiffZipTask {
	task, err := utils.Find(config.Tasks, func(task utils.Task) bool {
		return task.ID == taskId
	})
	if err != nil {
		panic(err)
	}
	return &DiffZipTask{
		TaskName:           config.TaskName,
		TotalScannedFiles:  0,
		TotalChangedFiles:  0,
		WorkingDir:         config.WorkingDir,
		MaxZipSizeInMBytes: config.MaxZipSize,
		currentZWIdx:       0,
		Task:               *task,
		zipper:             NewZipper(config.WorkingDir, fmt.Sprintf("%s_%s", config.TaskName, taskId), 0),
	}
}

func (c *DiffZipTask) nZipper(newFileStat os.FileInfo) *Zipper {
	if c.zipper == nil {
		c.currentZWIdx++
		c.zipper = NewZipper(c.WorkingDir, fmt.Sprintf("%s_%s", c.TaskName, c.Task.ID), c.currentZWIdx)
	}

	if ((c.zipper.totalSizeInBytes + newFileStat.Size()) / 1024 / 1024) > c.MaxZipSizeInMBytes {
		c.zipper.flush()
		c.currentZWIdx++
		c.zipper = NewZipper(c.WorkingDir, fmt.Sprintf("%s_%s", c.TaskName, c.Task.ID), c.currentZWIdx)
	}
	return c.zipper
}

func (c *DiffZipTask) Zip(filePath string, fileStat os.FileInfo) {
	c.nZipper(fileStat).zip(filePath, fileStat)
}

func (c *DiffZipTask) flush() {
	c.zipper.flush()
	c.zipper = nil
}
