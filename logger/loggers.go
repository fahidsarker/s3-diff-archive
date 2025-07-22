package logger

import (
	"fmt"
	"s3-diff-archive/utils"
)

var ScanLog, _ = CreateLogger(fmt.Sprintf("scan_%s.log", utils.NowTime()), false)
var Logs, _ = CreateLogger(fmt.Sprintf("logs_%s.log", utils.NowTime()), true)
