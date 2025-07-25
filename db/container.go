package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"s3-diff-archive/types"
	"s3-diff-archive/utils"

	badger "github.com/dgraph-io/badger/v4"
)

type DBContainer struct {
	dir    string
	db     *badger.DB
	closed bool
}

func NewDBInDir(workingDir string) *DBContainer {
	dbPath := path.Join(workingDir, fmt.Sprintf("db-%s-%s", utils.GenerateRandString(5), utils.NowTime()))
	return &DBContainer{dir: dbPath}
}

func NewDBFromPath(dbPath string) *DBContainer {
	return &DBContainer{dir: dbPath}
}

func (c *DBContainer) Close() {
	c.closed = true
	if c.db != nil {
		c.db.Close()
		c.db = nil
	}
	_ = os.RemoveAll(c.dir)
}

func (c *DBContainer) CloseNoDel() string {
	c.closed = true
	if c.db != nil {
		c.db.Close()
		c.db = nil
	}
	return c.dir
}

func (c *DBContainer) CloseAndZip(password string) (string, error) {
	c.closed = true
	if c.db != nil {
		c.db.Close()
		c.db = nil
	}
	zippedRes, err := archiveDB(c.dir, password)
	_ = os.RemoveAll(c.dir)
	return zippedRes, err
}

func (c *DBContainer) GetDB() *badger.DB {
	if c.db == nil {
		if c.closed {
			panic("DB is closed and cannot be used")
		}
		c.db = getDB(c.dir)
	}
	return c.db
}

func (c *DBContainer) InsertSfilesToDB(files []*types.SFile) {
	const batchSize = 1000
	for i := 0; i < len(files); i += batchSize {
		end := min(i+batchSize, len(files))

		err := c.GetDB().Update(func(txn *badger.Txn) error {

			for _, file := range files[i:end] {
				fileJson, err := json.Marshal(file)
				if err != nil {
					return err
				}
				err = txn.Set([]byte(file.RelativePath), fileJson)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}
