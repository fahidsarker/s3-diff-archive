package archiver

import (
	"os"
	"s3-diff-archive/utils"
	"strings"
)

type DiffZipTask struct {
	utils.TaskConfig
	MaxZipSizeInMBytes int64
	zipper             *Zipper
	TotalScannedFiles  int
	TotalChangedFiles  int
	ZipFilePaths       []string
}

func NewDiffZipTask(config *utils.Config, taskId string) *DiffZipTask {
	task, err := config.GetTask(taskId)
	if err != nil {
		panic(err)
	}
	newZipOutputFile := config.NewZipFileNameForTask(taskId, 0)
	return &DiffZipTask{
		TotalScannedFiles:  0,
		TotalChangedFiles:  0,
		MaxZipSizeInMBytes: config.MaxZipSize,
		TaskConfig:         *task,
		ZipFilePaths:       []string{},
		zipper:             NewZipper(newZipOutputFile),
	}
}

func (c *DiffZipTask) nZipper(newFileStat os.FileInfo) *Zipper {

	if c.zipper == nil {
		newZipperPath := c.NewZipFileNameForTask(c.ID, len(c.ZipFilePaths))
		c.zipper = NewZipper(newZipperPath)
	}

	if ((c.zipper.totalSizeInBytes + newFileStat.Size()) / 1024 / 1024) > c.MaxZipSizeInMBytes {
		newPath := c.zipper.flush()
		if newPath != "" {
			c.ZipFilePaths = append(c.ZipFilePaths, newPath)
		}
		newZipperPath := c.NewZipFileNameForTask(c.ID, len(c.ZipFilePaths))
		c.zipper = NewZipper(newZipperPath)
	}
	return c.zipper
}

func (c *DiffZipTask) Zip(filePath string, fileStat os.FileInfo) {
	fileName := filePath
	if strings.HasPrefix(filePath, c.Task.Dir) {
		// remove baseDir from filePath
		fileName = strings.Replace(filePath, c.Task.Dir, "", 1)
	}
	c.nZipper(fileStat).zip(filePath, fileName, fileStat, c.Task.Password)
}

func (c *DiffZipTask) flush() {
	newPath := c.zipper.flush()
	if newPath != "" {
		c.ZipFilePaths = append(c.ZipFilePaths, newPath)
	}
	c.zipper = nil
}
