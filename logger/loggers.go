package lg

import (
	"fmt"
	"os"
	"s3-diff-archive/utils"
	"strings"
)

// var ScanLog, _ = CreateLogger(fmt.Sprintf("scan_%s.log", utils.NowTime()), false)
// var Logs, _ = CreateLogger(fmt.Sprintf("logs_%s.log", utils.NowTime()), true)

var ScanLog *BufferedLogger
var Logs *BufferedLogger

func InitLoggers(config *utils.Config) error {
	defPrintToFile := config.LogsDir != ""
	if config.LogsDir != "" {
		// make sure logs dir exists // make all dirs in path
		if _, err := os.Stat(config.LogsDir); os.IsNotExist(err) {
			// Change os.Mkdir to os.MkdirAll
			err := os.MkdirAll(config.LogsDir, 0755) // This will create all intermediate directories
			if err != nil {
				return err // Handle the error (e.g., permissions issues)
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

	Logs.Info("Loggers initialized")
	Logs.Info("init: %s", strings.Join(os.Args, " "))
	return nil
}

func CloseGlobalLoggers() {
	CloseLogger(ScanLog)
	CloseLogger(Logs)
}
