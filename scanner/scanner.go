package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"s3-diff-archive/types"
	"s3-diff-archive/utils"

	lg "s3-diff-archive/logger"

	badger "github.com/dgraph-io/badger/v4"
)

func ScanTask(db *badger.DB, task *utils.TaskConfig) *ScannedResult {
	result := &ScannedResult{
		UpdatedFiles:   []*types.SFile{},
		SkippedFiles:   []string{},
		UnChangedFiles: []*types.SFile{},
	}
	lg.Logs.Info("Scanning task %s", task.ID)
	lg.ScanLog.Info("Scanning task %s", task.ID)
	iterator(db, task, result, task.Dir)
	println("")
	return result
}

func iterator(rdb *badger.DB, task *utils.TaskConfig, res *ScannedResult, dirPath string) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() {
			iterator(rdb, task, res, dirPath+"/"+file.Name())
		} else {

			// change to relative path to task.Dir
			relativeFilePath := utils.RelativePath(dirPath+"/"+file.Name(), task.Dir)

			// ignore any excluded files
			skip := false
			for _, patterm := range task.Excludes {
				if utils.MatchPattern(patterm, relativeFilePath) {
					// lg.Logs.Warn("Skipping file %s due to exclude pattern %s", path.Join(task.Dir, relativeFilePath), patterm)
					lg.ScanLog.Info("Slipped file: %s, due to exclude pattern %s", path.Join(task.Dir, relativeFilePath), patterm)
					skip = true
					break
				}
			}
			if skip {
				res.SkippedFiles = append(res.SkippedFiles, relativeFilePath)
				continue
			}

			stats, err := os.Stat(dirPath + "/" + file.Name())
			if err != nil {
				panic(err)
			}

			file, fileUpdated := hasFileUpdated(rdb, relativeFilePath, &stats)
			lg.ScanLog.Info("%s\t%s, File Updated: %t, Size: %d", task.ID, relativeFilePath, fileUpdated, stats.Size())

			if fileUpdated {
				res.UpdatedFiles = append(res.UpdatedFiles, file)
			} else {
				res.UnChangedFiles = append(res.UnChangedFiles, file)
			}
		}
		fmt.Printf("\r>>> Scanned: %d files, Change Detected: %d files, Skipped: %d files", len(res.UpdatedFiles)+len(res.UnChangedFiles)+len(res.SkippedFiles), len(res.UpdatedFiles), len(res.SkippedFiles))
	}
}

func hasFileUpdated(rdb *badger.DB, relativePath string, statsPointer *os.FileInfo) (*types.SFile, bool) {
	stats := *statsPointer
	var file types.SFile
	err := rdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(relativePath))
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

	newSfile := &types.SFile{RelativePath: relativePath, Name: stats.Name(), Size: stats.Size(), Mtime: stats.ModTime().Unix()}
	if err != nil {
		return newSfile, true
	}

	if stats.Size() != file.Size || stats.ModTime().Unix() != file.Mtime || stats.Name() != file.Name {
		return newSfile, true
	}

	return newSfile, false
}
