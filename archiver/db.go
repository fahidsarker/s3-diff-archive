package archiver

import (
	"encoding/json"
	"log"
	"os"

	badger "github.com/dgraph-io/badger/v4"
)

func GetDB() *badger.DB {
	opts := badger.DefaultOptions("./tmp/db") // Directory to store the database
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close() // Ensure the database is closed when the function exits
	return db
}

func storeFileToS3(db *badger.DB, file SFile) {
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

func hasFileUpdated(db *badger.DB, filePath string, stats os.FileInfo) bool {
	var file SFile
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(filePath))
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

	newSfile := SFile{Path: filePath, Name: stats.Name(), Size: stats.Size(), Mtime: stats.ModTime().Unix()}
	if err != nil {
		storeFileToS3(db, newSfile)
		return true
	}

	if stats.Size() != file.Size || stats.ModTime().Unix() != file.Mtime || stats.Name() != file.Name {
		storeFileToS3(db, newSfile)
		return true
	}

	return false
}
