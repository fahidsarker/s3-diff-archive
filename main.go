package main

import (
	"encoding/json"
	"s3-diff-archive/crypto"
	l "s3-diff-archive/logger"
	"s3-diff-archive/registery"
)

func main() {
	// zipped := archiver.ZipDiff(utils.GetConfig())
	// println("Total zipped files: ", zipped)
	config := registery.DummyRegistry()
	jsonConfig, _ := json.Marshal(config)
	encrypted, _ := crypto.EncryptString(string(jsonConfig), "asdasd")
	logger, _ := l.CreateLogger("scan.reg")

	for i := 0; i < 1000; i++ {
		l.Log(logger, encrypted)
	}

	l.CloseLogger(logger)
	lastReg, _ := l.ReadLastLine("scan.reg")
	println(lastReg)
	decrypted, _ := crypto.DecryptString(lastReg, "asdasd")
	println(decrypted)
}
