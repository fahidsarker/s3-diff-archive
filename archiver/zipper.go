package archiver

import (
	"os"
	"s3-diff-archive/utils"

	"github.com/alexmullins/zip"
)

type Zipper struct {
	file             *os.File
	zw               *zip.Writer
	totalSizeInBytes int64
	fileCounts       int
}

func (c *Zipper) flush() string {
	filePath := c.file.Name()
	if c.fileCounts == 0 {
		filePath = ""
	}
	if c.file != nil {
		defer c.file.Close()
	}
	if c.zw != nil {
		defer c.zw.Close()
	}
	if c.fileCounts <= 0 && c.file != nil {
		defer os.Remove(c.file.Name())
	}

	// After defers
	c.file = nil
	c.zw = nil
	c.totalSizeInBytes = 0
	return filePath
}

func (c *Zipper) zip(filePath string, filename string, fileStat os.FileInfo, password string) {
	utils.ZipFile(filePath, filename, fileStat, c.zw, password)
	c.totalSizeInBytes += fileStat.Size()
	c.fileCounts++
}

func NewZipper(outputFile string) *Zipper {
	// println("Output file: ", outputFile)
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
