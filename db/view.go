package db

import (
	lg "s3-diff-archive/logger"
	"s3-diff-archive/utils"

	badger "github.com/dgraph-io/badger/v4"
)

func ViewDB(task *utils.TaskConfig) {
	dbc := FetchRemoteDB(task)
	defer dbc.Close()

	// log all the entries in DB using a cursor
	err := dbc.GetDB().View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			lg.Logs.Info("%s", string(value))
		}
		return nil
	})
	if err != nil {
		lg.Logs.Fatal("%s", err.Error())
	}

}
