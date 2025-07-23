package archiver

import (
	"os"
	lg "s3-diff-archive/logger"
)

func Cleanup(result ZipDiffTaskResult) {
	for i := range result.ArchivesPaths {
		err := os.Remove(result.ArchivesPaths[i])
		if err != nil {
			lg.Logs.Error("Error Cleaning up (%s) :%s", result.ArchivesPaths[i], err.Error())
		}
	}
	if result.DBZipPath != "" {
		err := os.Remove(result.DBZipPath)
		if err != nil {
			lg.Logs.Error("Error Cleaning up (%s) :%s", result.DBZipPath, err.Error())
		}
	}
}
