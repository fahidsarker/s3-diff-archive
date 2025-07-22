package main

import (
	"s3-diff-archive/archiver"
	l "s3-diff-archive/logger"
	"s3-diff-archive/utils"
)

func main() {
	config := utils.GetConfig()
	err := l.InitLoggers(&config)
	if err != nil {
		panic(err)
	}
	defer l.ScanLog.Close()
	defer l.Logs.Close()
	// defer l.DiffLog.Close()
	zipped := archiver.ZipDiff(config)
	println("Total zipped files: ", zipped)
}
