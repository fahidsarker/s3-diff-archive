package main

import (
	"s3-diff-archive/archiver"
	"s3-diff-archive/utils"
)

func main() {
	// totalFiles := utils.CreateRandDirFiles("./test-files", 5, 4, 0)
	// println("Total files created: ", totalFiles)
	totalZipped := archiver.ZipDiff(utils.GetConfig())
	println("Total zipped files: ", totalZipped)
	// archiver.ArchiveDB("./tmp/test-db.zip")

	// utils.ParseConfig()
	// diffZip := archiver.NewDiffZip(utils.GetConfig(), "photos")
	// println("Scanning: ", diffZip.BaseDir)
	// config := utils.GetConfig()
	// jsn, err := json.Marshal(config)
	// if err != nil {
	// 	panic(err)
	// }
	// for _, task := range config.Tasks {
	// 	println("Password for task: ", task.ID, " is: ", task.Password)
	// }
	// println(string(jsn))
}
