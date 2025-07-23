package main

import (
	"fmt"
	"s3-diff-archive/archiver"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/utils"
)

func main() {
	config := utils.GetConfig()
	err := lg.InitLoggers(config)
	if err != nil {
		panic(err)
	}
	defer lg.CloseGlobalLoggers()

	// fmt.Println(utils.ToJson(config))
	zipped := archiver.ZipDiff(config)
	fmt.Println(utils.ToJson(zipped))

}
