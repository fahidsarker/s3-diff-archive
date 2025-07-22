package archiver

import (
	"os"
	"s3-diff-archive/utils"
)

type DiffZipTask struct {
	utils.TaskConfig
	MaxZipSizeInMBytes int64
	currentZWIdx       int
	zipper             *Zipper
	TotalScannedFiles  int
	TotalChangedFiles  int
}

func NewDiffZipTask(config utils.Config, taskId string) *DiffZipTask {
	task, err := config.GetTask(taskId)
	if err != nil {
		panic(err)
	}
	return &DiffZipTask{
		TotalScannedFiles:  0,
		TotalChangedFiles:  0,
		MaxZipSizeInMBytes: config.MaxZipSize,
		currentZWIdx:       0,
		TaskConfig:         *task,
		zipper:             NewZipper(config.NewZipFileNameForTask(taskId, 0)),
	}
}

func (c *DiffZipTask) nZipper(newFileStat os.FileInfo) *Zipper {
	if c.zipper == nil {
		c.currentZWIdx++
		c.zipper = NewZipper(c.NewZipFileNameForTask(c.ID, c.currentZWIdx))
	}

	if ((c.zipper.totalSizeInBytes + newFileStat.Size()) / 1024 / 1024) > c.MaxZipSizeInMBytes {
		c.zipper.flush()
		c.currentZWIdx++
		c.zipper = NewZipper(c.NewZipFileNameForTask(c.ID, c.currentZWIdx))
	}
	return c.zipper
}

func (c *DiffZipTask) Zip(filePath string, fileStat os.FileInfo) {
	c.nZipper(fileStat).zip(filePath, filePath, fileStat, c.Task.Password)
}

func (c *DiffZipTask) flush() {
	c.zipper.flush()
	c.zipper = nil
}
