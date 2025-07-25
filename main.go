package main

import (
	"fmt"
	"s3-diff-archive/archiver"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/utils"
)

func archive(config *utils.Config) {
	archiver.ZipDiff(config)
	lg.Logs.Info("✔︎✔︎ DONE")
}

func main() {

	config := utils.GetConfig("config.yaml")
	fmt.Println(utils.ToJson(config))

	return
	// args := os.Args
	// if len(args) < 3 {
	// 	fmt.Println("Usage: s3-diff-archive <command> <config-file-path>")
	// 	os.Exit(1)
	// }

	// config := utils.GetConfig(args[2])
	// err := lg.InitLoggers(config)
	// if err != nil {
	// 	panic(err)
	// }
	// defer lg.CloseGlobalLoggers()

	// switch args[1] {
	// case "archive":
	// 	archive(config)
	// case "view":
	// 	if len(args) < 5 {
	// 		fmt.Println("Usage: s3-diff-archive view <config-file-path> --task <task-id>")
	// 		os.Exit(1)
	// 	}
	// 	taskId := args[4]
	// 	task, err := config.GetTask(taskId)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	db.ViewDB(task)
	// default:
	// 	fmt.Println("Unknown command")
	// 	os.Exit(1)

	// }

}
