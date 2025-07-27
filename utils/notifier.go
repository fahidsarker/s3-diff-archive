package utils

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func Notify(script, operation, status, message string) {
	if script == "" {
		return
	}

	switch status {
	case "success":
		script = strings.ReplaceAll(script, "%icon%", "✅")
	case "error":
		script = strings.ReplaceAll(script, "%icon%", "❌")
	case "fatal":
		script = strings.ReplaceAll(script, "%icon%", "❌⚠️❌⚠️")
	case "warn":
		script = strings.ReplaceAll(script, "%icon%", "⚠️")
	default:
		script = strings.ReplaceAll(script, "%icon%", "ℹ️")
	}
	script = strings.ReplaceAll(script, "%source%", operation)
	script = strings.ReplaceAll(script, "%status%", status)
	script = strings.ReplaceAll(script, "%message%", message)

	// Execute the script through the appropriate shell based on OS
	var err error
	if runtime.GOOS == "windows" {
		err = exec.Command("cmd", "/C", script).Run()
	} else {
		err = exec.Command("sh", "-c", script).Run()
	}

	if err != nil {
		fmt.Printf("Failed to execute notify script: %s, error: %v\n", script, err)
	}
}
