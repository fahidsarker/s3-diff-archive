package db

import (
	lg "s3-diff-archive/logger"
	"s3-diff-archive/utils"

	badger "github.com/dgraph-io/badger/v4"
)

func ViewDB(task *utils.TaskConfig) {
	dbPath := FetchRemoteDB(task)
	lg.Logs.Info("DB path: " + dbPath)
	db := GetDB(dbPath)
	defer db.Close()

	// log all the entries in DB using a cursor
	err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			lg.Logs.Info(string(value))
		}
		return nil
	})
	if err != nil {
		lg.Logs.Fatal("%s", err.Error())
	}

}
