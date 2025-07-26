package restorer

import (
	"fmt"
	"os"
	"path/filepath"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/utils"
	"sort"
	"strings"
	"time"
)

func DirsEqual(dir1, dir2 string, skips []string) (bool, error) {
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
		lg.Logs.Error("One of the dirs is not a directory")
		return false, nil
	}

	return compareDirsRecursive(absDir1, absDir2, skips)
}

// compareDirsRecursive is a helper function to recursively compare directories.
func compareDirsRecursive(path1, path2 string, skips []string) (bool, error) {
	entries1, err := os.ReadDir(path1)
	if err != nil {
		lg.Logs.Error("%s", err.Error())
		return false, err
	}
	entries2, err := os.ReadDir(path2)
	if err != nil {
		lg.Logs.Error("%s", err.Error())
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
		lg.Logs.Error("Number of entries are not equal. %s=%d, %s=%d", path1, len(entries1), path2, len(entries2))
		names1 := make([]string, len(entries1))
		for i, entry := range entries1 {
			names1[i] = entry.Name()
		}
		names2 := make([]string, len(entries2))
		for i, entry := range entries2 {
			names2[i] = entry.Name()
		}
		lg.Logs.Error("Files in %s are:\n%s", path1, strings.Join(names1, "\n"))
		lg.Logs.Error("Files in %s are:\n%s", path2, strings.Join(names2, "\n"))
		return false, nil
	}

	for i := range entries1 {
		entry1 := entries1[i]
		entry2 := entries2[i]

		if entry1.Name() == ".DS_Store" || entry2.Name() == ".DS_Store" {
			continue
		}
		skip := false
		for _, skipFile := range skips {
			if utils.MatchPattern(skipFile, entry1.Name()) {
				skip = true
				break
			}
		}
		if skip {
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
			lg.Logs.Error("%s not found", info1.Name())
			return false, err
		}
		info2, err := entry2.Info()
		if err != nil {
			lg.Logs.Error("%s not found", info2.Name())
			return false, err
		}

		// Check if entry types (file/dir) are the same
		if info1.IsDir() != info2.IsDir() {
			fmt.Printf("Dirs are not equal: %s, %s\n", path1, path2)
			return false, nil
		}

		if info1.IsDir() {
			// Recursively compare subdirectories
			equal, err := compareDirsRecursive(fullPath1, fullPath2, skips)
			if err != nil {
				return false, err
			}
			if !equal {
				return false, nil // Early exit
			}
		} else {
			// Compare files: size and modification time
			if info1.Size() != info2.Size() {
				lg.Logs.Error("File Sizes are not equal: %s, %s\n", info1.Name(), info2.Name())
				return false, nil
			}

			// Compare modification times. Truncate to second precision to
			// allow a threshold of 2 seconds for timestamp differences
			threshold := 1 * time.Second
			// account for file system differences in timestamp granularity.
			if time1, time2 := info1.ModTime().Truncate(time.Second), info2.ModTime().Truncate(time.Second); time1.After(time2.Add(threshold)) || time2.After(time1.Add(threshold)) {
				lg.Logs.Error("Mod times are not within threshold: %s : %s, %s : %s\n", info1.Name(), info1.ModTime(), info2.Name(), info2.ModTime())
				return false, nil
			}
		}
	}

	return true, nil
}
