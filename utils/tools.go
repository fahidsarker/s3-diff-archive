package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

func Where[T any](list []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range list {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

func Find[T any](list []T, predicate func(T) bool) (*T, error) {
	for _, item := range list {
		if predicate(item) {
			return &item, nil
		}
	}
	return nil, errors.New("not found")
}

func GenerateRandString(length int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RandAndTime(length int) string {
	return fmt.Sprintf("%s_%s", GenerateRandString(length), NowTime())
}

func Err(message string, placeholders ...any) {
	fmt.Printf(message+"\n", placeholders...)
	os.Exit(1)
}

func IsPathDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func IsPathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsDirEmpty(path string) bool {
	files, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	return len(files) == 0
}

func IsWorkingDirValid(workingDir string) bool {
	if !IsPathExists(workingDir) {
		return true
	}
	if !IsPathDir(workingDir) {
		Err("Working dir is not a directory")
		return false
	}
	if !IsDirEmpty(workingDir) {
		Err("Working dir is not empty")
		return false
	}
	return true
}

func NowTime() string {
	// 2025-07-22 12:20:44
	return time.Now().Format("2006-01-02 15:04:05")
}

func ToJson(data any) string {
	dataJson, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(dataJson)
}

func FileNameFromPath(path string) string {
	return filepath.Base(path)
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func MatchPattern(pattern string, path string) bool {

	match, err := doublestar.PathMatch(pattern, path)
	if err != nil {
		panic(err)
	}

	return match
}

func RelativePath(abs, dir string) string {
	relativeFilePath := abs
	if strings.HasPrefix(abs, dir) {
		relativeFilePath = strings.Replace(abs, dir, "", 1)
	}
	if (strings.HasPrefix(relativeFilePath, "/") || strings.HasPrefix(relativeFilePath, "\\")) && relativeFilePath != "\\" && relativeFilePath != "/" {
		relativeFilePath = relativeFilePath[1:]
	}
	return relativeFilePath
}
