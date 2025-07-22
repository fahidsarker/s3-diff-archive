package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
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
