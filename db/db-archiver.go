package db

import (
	"fmt"
	"os"
	"path"
	"s3-diff-archive/archiver"
	"s3-diff-archive/utils"
)

func archiveDB(dbPath string, encryptPass string) (string, error) {
	parentOfDBPath := path.Dir(dbPath)
	zipper := archiver.NewZipper(path.Join(parentOfDBPath, fmt.Sprintf("db-%s-%s.zip", utils.GenerateRandString(5), utils.NowTime())))
	filesOfDB, err := os.ReadDir(dbPath)
	if err != nil {
		return "", err
	}
	for _, file := range filesOfDB {
		if file.IsDir() {
			panic("DB archiver does not support directories")
		}
		filePath := dbPath + "/" + file.Name()
		stats, err := os.Stat(filePath)
		if err != nil {
			return "", err
		}
		zipper.Zip(filePath, file.Name(), &stats, encryptPass)
	}
	newPath := zipper.Flush()
	if newPath != "" {
		return newPath, nil
	}
	return "", fmt.Errorf("failed to create zip file")
}
