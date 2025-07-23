package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func DirsEqual(dir1, dir2 string) (bool, error) {
	// Canonicalize paths to ensure consistent comparison
	absDir1, err := filepath.Abs(dir1)
	if err != nil {
		return false, err
	}
	absDir2, err := filepath.Abs(dir2)
	if err != nil {
		return false, err
	}

	// Get file info for the root directories
	info1, err := os.Stat(absDir1)
	if err != nil {
		return false, err
	}
	info2, err := os.Stat(absDir2)
	if err != nil {
		return false, err
	}

	// Ensure both are directories
	if !info1.IsDir() || !info2.IsDir() {
		return false, nil
	}

	return compareDirsRecursive(absDir1, absDir2)
}

// compareDirsRecursive is a helper function to recursively compare directories.
func compareDirsRecursive(path1, path2 string) (bool, error) {
	entries1, err := os.ReadDir(path1)
	if err != nil {
		return false, err
	}
	entries2, err := os.ReadDir(path2)
	if err != nil {
		return false, err
	}

	// Sort entries by name to ensure consistent comparison regardless of OS
	sort.Slice(entries1, func(i, j int) bool {
		return entries1[i].Name() < entries1[j].Name()
	})
	sort.Slice(entries2, func(i, j int) bool {
		return entries2[i].Name() < entries2[j].Name()
	})

	// Check if the number of entries is the same
	if len(entries1) != len(entries2) {
		return false, nil
	}

	for i := range entries1 {
		entry1 := entries1[i]
		entry2 := entries2[i]

		if entry1.Name() == ".DS_Store" || entry2.Name() == ".DS_Store" {
			continue
		}

		// Check if filenames are the same
		if entry1.Name() != entry2.Name() {
			return false, nil
		}

		fullPath1 := filepath.Join(path1, entry1.Name())
		fullPath2 := filepath.Join(path2, entry2.Name())

		info1, err := entry1.Info()
		if err != nil {
			return false, err
		}
		info2, err := entry2.Info()
		if err != nil {
			return false, err
		}

		// Check if entry types (file/dir) are the same
		if info1.IsDir() != info2.IsDir() {
			fmt.Printf("Dirs are not equal: %s, %s\n", path1, path2)
			return false, nil
		}

		if info1.IsDir() {
			// Recursively compare subdirectories
			equal, err := compareDirsRecursive(fullPath1, fullPath2)
			if err != nil {
				return false, err
			}
			if !equal {
				return false, nil // Early exit
			}
		} else {
			// Compare files: size and modification time
			if info1.Size() != info2.Size() {
				fmt.Printf("File Sizes are not equal: %s, %s\n", info1.Name(), info2.Name())
				return false, nil
			}

			// Compare modification times. Truncate to second precision to
			// account for file system differences in timestamp granularity.
			if !info1.ModTime().Truncate(time.Second).Equal(info2.ModTime().Truncate(time.Second)) {
				fmt.Printf("Mod times are not equal: %s, %s\n", info1.Name(), info2.Name())
				return false, nil
			}
		}
	}

	return true, nil
}
