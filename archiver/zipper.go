package archiver

import (
	"fmt"
	"os"
	"path/filepath"
	"s3-diff-archive/utils"
	"time"

	"github.com/alexmullins/zip"
)

type Zipper struct {
	file             *os.File
	zw               *zip.Writer
	totalSizeInBytes int64
	fileCounts       int
}

func (c *Zipper) flush() {
	if c.file != nil {
		defer c.file.Close()
	}
	if c.zw != nil {
		defer c.zw.Close()
	}
	if c.fileCounts <= 0 {
		defer os.Remove(c.file.Name())
	}

	c.file = nil
	c.zw = nil
	c.totalSizeInBytes = 0
}

func (c *Zipper) zip(file string, fileStat os.FileInfo) {
	utils.ZipFile(file, c.zw)
	c.totalSizeInBytes += fileStat.Size()
	c.fileCounts++
}

func NewZipper(workingdir string, taskId string, idx int) *Zipper {
	// check if the workingdir exists, create if not
	if _, err := os.Stat(workingdir); os.IsNotExist(err) {
		err := os.Mkdir(workingdir, 0755)
		if err != nil {
			panic(err)
		}
	}

	zipSuffix := time.Now().UTC().Format("2006-01-02_15-04-05")
	if idx > 0 {
		zipSuffix = zipSuffix + "_" + fmt.Sprintf("%d", idx)
	}
	outputFile := filepath.Join(workingdir, taskId+"_"+zipSuffix+".zip")
	println("Output file: ", outputFile)
	outFile, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}

	return &Zipper{
		file:             outFile,
		zw:               zip.NewWriter(outFile),
		totalSizeInBytes: 0,
	}
}
