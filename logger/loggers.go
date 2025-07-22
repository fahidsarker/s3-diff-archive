package logger

import (
	"fmt"
	"os"
	"s3-diff-archive/utils"
)

// var ScanLog, _ = CreateLogger(fmt.Sprintf("scan_%s.log", utils.NowTime()), false)
// var Logs, _ = CreateLogger(fmt.Sprintf("logs_%s.log", utils.NowTime()), true)

var ScanLog *BufferedLogger
var Logs *BufferedLogger

func InitLoggers(config *utils.Config) error {
	defPrintToFile := config.LogsDir != ""
	if config.LogsDir != "" {
		// make sure logs dir exists
		if _, err := os.Stat(config.LogsDir); os.IsNotExist(err) {
			err := os.Mkdir(config.LogsDir, 0755)
			if err != nil {
				return err
			}
		}
	}
	var err error
	Logs, err = CreateLogger(fmt.Sprintf("%s/logs_%s.log", config.LogsDir, utils.NowTime()), true, defPrintToFile)
	if err != nil {
		return err
	}

	ScanLog, err = CreateLogger(fmt.Sprintf("%s/scan_%s.log", config.LogsDir, utils.NowTime()), false, defPrintToFile)
	if err != nil {
		return err
	}
	return nil
}
