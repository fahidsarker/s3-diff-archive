package utils

import (
	"math/rand"
	"os"
)

func randInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func generateRandString(length int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func writeRandTxtFile(dir string) {
	// if dir doesn't exist, create it
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}
	fileName := generateRandString(randInt(10, 20))
	filePath := dir + "/" + fileName + ".txt"
	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(generateRandString(randInt(1000, 10000)))
	if err != nil {
		panic(err)
	}
}

func CreateRandDirFiles(baseDir string, fileCount int, depth int, currentDepth int) int {
	// create dir if not exist
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		os.Mkdir(baseDir, 0755)
	}

	if currentDepth > depth {
		return 0
	}

	filesCreated := 0
	for i := 0; i < fileCount; i++ {
		// create a new file
		writeRandTxtFile(baseDir)
		filesCreated++
	}
	if currentDepth < depth {
		for i := 0; i < fileCount; i++ {
			// create a new directory
			newDirName := generateRandString(randInt(10, 20))
			newDirPath := baseDir + "/" + newDirName
			os.Mkdir(newDirPath, 0755)
			filesCreated += CreateRandDirFiles(newDirPath, fileCount, depth, currentDepth+1)
		}
	}
	return filesCreated
}

// recursively check sub-dirs to find all the file count under this dir
func FlatFileCount(baseDir string) int {
	filesAvailable := 0
	files, err := os.ReadDir(baseDir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			filesAvailable += FlatFileCount(baseDir + "/" + file.Name())
		} else {
			filesAvailable++
		}
	}
	return filesAvailable
}
