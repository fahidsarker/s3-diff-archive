package main

import "s3-archive/archiver"

func main() {
	// totalFiles := utils.CreateRandDirFiles("./test-files", 5, 4, 0)
	// println("Total files created: ", totalFiles)
	totalZipped := archiver.ZipDiff("./test-files", "./tmp/test.zip")
	println("Total zipped files: ", totalZipped)
	archiver.ArchiveDB("./tmp/test-db.zip")
}
