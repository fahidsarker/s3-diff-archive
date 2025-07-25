package db

import (
	"context"
	"os"
	"path"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/s3"
	"s3-diff-archive/utils"

	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	badger "github.com/dgraph-io/badger/v4"
)

func getDB(dbPath string) *badger.DB {
	opts := badger.DefaultOptions(dbPath).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		lg.Logs.Fatal("%s", err.Error())
	}
	// defer db.Close() // Ensure the database is closed when the function exits
	return db
}

func FetchRemoteDB(task *utils.TaskConfig) *DBContainer {
	tempDBPath := path.Join(task.WorkingDir, task.ID, "db.zip")
	refDBPath := path.Join(task.WorkingDir, task.ID, "db-remote")
	err := s3.DownloadFileFromS3(task.CreateS3Config(s3Types.StorageClassStandard), context.TODO(), "db.zip", tempDBPath)
	if err != nil {
		if err.Error() == "not-found" {
			lg.Logs.Warn("Remote DB not found, Treating as a new backup task")
		} else {
			lg.Logs.Fatal("%s", err.Error())
		}
	} else {
		err := utils.Unzip(tempDBPath, refDBPath, task.Password)
		if err != nil {
			lg.Logs.Fatal("%s", err.Error())
		} else {
			lg.Logs.Info("DB unzipped")
		}
	}
	_ = os.Remove(tempDBPath)
	return &DBContainer{dir: refDBPath}
}
