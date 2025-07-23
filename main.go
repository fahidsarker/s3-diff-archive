package main

import (
	"fmt"
	"os"
	"s3-diff-archive/archiver"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/utils"
)

func main() {
	args := os.Args
	if len(args) < 3 {
		fmt.Println("Usage: s3-diff-archive <command> <config-file-path>")
		os.Exit(1)
	}
	if args[1] != "archive" {
		fmt.Println("Command not yet supported")
		os.Exit(1)
	}
	config := utils.GetConfig(args[2])
	err := lg.InitLoggers(config)
	if err != nil {
		panic(err)
	}
	defer lg.CloseGlobalLoggers()

	// fmt.Println(utils.ToJson(config))
	archiver.ZipDiff(config)

	lg.Logs.Info("✔︎✔︎ DONE")
	// fmt.Println(utils.ToJson(zipped))

}
