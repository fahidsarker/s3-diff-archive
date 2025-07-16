package utils

import (
	"errors"
	"fmt"
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

func Err(message string, placeholders ...any) {
	fmt.Printf(message+"\n", placeholders...)
	os.Exit(1)
}
