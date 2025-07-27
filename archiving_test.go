package main

import (
	"s3-diff-archive/archiver"
	"s3-diff-archive/db"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/restorer"
	"s3-diff-archive/scanner"
	"s3-diff-archive/utils"
	"testing"
)

func TestArchiving(t *testing.T) {
	config := utils.GetConfig("./config.yaml", ".env")
	err := lg.InitLoggers(config)
	if err != nil {
		panic(err)
	}
	defer lg.CloseGlobalLoggers()

	utils.CreateRandDirFiles("./test-files", 3, 5, 0)

	task, err := config.GetTask("photos")
	if err != nil {
		panic(err)
	}

	tempDB := db.NewDBInDir("./tmp/test-db")
	defer tempDB.Close()
	scanned, _ := scanner.ScanTask(tempDB.GetDB(), task)
	println(scanned.SkippedFiles)

	archived := archiver.ArchiveToZip(task, scanned)
	println(archived)

	// err := restorer.RestoreFromZips([]string{"tmp/photos_2025_07_26_05_42_35.zip", "tmp/photos_2025_07_26_05_42_40_1.zip", "tmp/photos_2025_07_26_05_42_44_2.zip"}, "./tmp/restored", "PASasdSWORD")
	err = restorer.RestoreFromZips(archived, "./tmp/restored", "PASasdSWORD")
	if err != nil {
		panic(err)
	}

	isEq, err := restorer.DirsEqual("./test-files", "./tmp/restored", []string{".DS_Store"})
	if err != nil {
		panic(err)
	}
	println(isEq)

}
