package archiver

import (
	"fmt"
	"os"
	"path"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/scanner"
	"s3-diff-archive/utils"
)

func ArchiveToZip(task *utils.TaskConfig, scanRes *scanner.ScannedResult) []string {

	lg.Logs.Info("Total files to zip in task %s: %d", task.ID, len(scanRes.UpdatedFiles))

	if len(scanRes.UpdatedFiles) == 0 {
		lg.Logs.Info("No files to zip in task %s", task.ID)
		return []string{}
	}

	maxZipSizeInBytes := task.MaxZipSize * 1024 * 1024
	currentZippedFileSizeInBytes := int64(0)
	totalZippedFilesSizeInBytes := int64(0)
	zipFilePaths := []string{}

	zipper := NewZipper(task.NewZipFileNameForTask(task.ID, 0))

	totalFilesToZip := len(scanRes.UpdatedFiles)

	for i := range totalFilesToZip {
		file := scanRes.UpdatedFiles[i]
		if currentZippedFileSizeInBytes+file.Size > maxZipSizeInBytes {
			newPath := zipper.Flush()
			if newPath != "" {
				zipFilePaths = append(zipFilePaths, newPath)
			}
			zipper = NewZipper(task.NewZipFileNameForTask(task.ID, len(zipFilePaths)))
			currentZippedFileSizeInBytes = 0
		}

		fileStat, err := os.Stat(path.Join(task.Dir, file.RelativePath))
		if err != nil {
			lg.Logs.Fatal("%s", err.Error())
		}
		zipper.Zip(path.Join(task.Dir, file.RelativePath), file.Name, &fileStat, task.Password)
		currentZippedFileSizeInBytes += fileStat.Size()
		totalZippedFilesSizeInBytes += fileStat.Size()
		fmt.Printf("\r>>> Zipped: %d / %d files, Total Size: %d bytes", i+1, totalFilesToZip, totalZippedFilesSizeInBytes)
	}
	println("")
	newPath := zipper.Flush()
	if newPath != "" {
		zipFilePaths = append(zipFilePaths, newPath)
	}
	lg.Logs.Info("Total Zip file created in task %s: %d", task.ID, len(zipFilePaths))

	return zipFilePaths
}
