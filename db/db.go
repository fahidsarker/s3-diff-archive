package db

import (
	"context"
	"encoding/json"
	"os"
	"path"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/s3"
	"s3-diff-archive/types"
	"s3-diff-archive/utils"

	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	badger "github.com/dgraph-io/badger/v4"
)

func GetDB(dbPath string) *badger.DB {
	opts := badger.DefaultOptions(dbPath).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		lg.Logs.Fatal("%s", err.Error())
	}
	// defer db.Close() // Ensure the database is closed when the function exits
	return db
}

func FetchRemoteDB(task *utils.TaskConfig) string {
	localDBPath := path.Join(task.WorkingDir, task.ID, "db.zip")
	refDBPath := path.Join(task.WorkingDir, task.ID, "db-ref")
	err := s3.DownloadFileFromS3(task.CreateS3Config(s3Types.StorageClassStandard), context.TODO(), "db.zip", localDBPath)
	if err != nil {
		if err.Error() == "not-found" {
			lg.Logs.Warn("Remote DB not found, Treating as a new backup task")
			// create the ref DBPath if not exists
			if _, err := os.Stat(refDBPath); os.IsNotExist(err) {
				err := os.MkdirAll(refDBPath, 0755)
				if err != nil {
					lg.Logs.Fatal("%s", err.Error())
				}
			}
			// create an empty .history file in ref DBPath
			historyFilePath := path.Join(refDBPath, ".history")
			err := os.WriteFile(historyFilePath, []byte(""), 0644)
			if err != nil {
				lg.Logs.Fatal("%s", err.Error())
			}
		} else {
			lg.Logs.Fatal("%s", err.Error())
		}
	} else {
		err := utils.Unzip(localDBPath, refDBPath, task.Password)
		if err != nil {
			lg.Logs.Fatal("%s", err.Error())
		} else {
			lg.Logs.Info("DB unzipped")
		}
	}
	return refDBPath
}

func storeFileToDB(db *badger.DB, file types.SFile) {
	jsonData, err := json.Marshal(file)
	if err != nil {
		panic(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(file.Path), jsonData)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func HasFileUpdated(rdb *badger.DB, wdb *badger.DB, fileName string, stats os.FileInfo) bool {
	var file types.SFile
	err := rdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fileName))
		if err != nil {
			// println("___", err.Error())
			return err
		}
		err = item.Value(func(val []byte) error {
			err := json.Unmarshal(val, &file)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})

	newSfile := types.SFile{Path: fileName, Name: stats.Name(), Size: stats.Size(), Mtime: stats.ModTime().Unix()}
	storeFileToDB(wdb, newSfile)
	if err != nil {
		// logger.Logs.Error(err.Error())
		// storeFileToDB(rdb, newSfile)
		return true
	}

	if stats.Size() != file.Size || stats.ModTime().Unix() != file.Mtime || stats.Name() != file.Name {
		// storeFileToDB(rdb, newSfile)
		return true
	}

	return false
}
