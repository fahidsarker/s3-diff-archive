package main

import (
	"s3-diff-archive/archiver"
	l "s3-diff-archive/logger"
	"s3-diff-archive/utils"
)

func main() {
	defer l.ScanLog.Close()
	defer l.Logs.Close()
	// defer l.DiffLog.Close()
	zipped := archiver.ZipDiff(utils.GetConfig())
	println("Total zipped files: ", zipped)

}
