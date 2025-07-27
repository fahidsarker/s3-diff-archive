package scanner

import (
	"fmt"
	"s3-diff-archive/types"
)

type ScannedResult struct {
	UpdatedFiles   []*types.SFile
	SkippedFiles   []string
	UnChangedFiles []*types.SFile
}

type TaskScanSummary struct {
	TaskID         string
	TotalScanned   int
	UpdatedFiles   int
	SkippedFiles   int
	UnChangedFiles int
}

func (sr *ScannedResult) TotalScanned() int {
	return len(sr.UpdatedFiles) + len(sr.SkippedFiles) + len(sr.UnChangedFiles)
}

func (sr *ScannedResult) Summary(taskId string) *TaskScanSummary {
	return &TaskScanSummary{
		TaskID:         taskId,
		TotalScanned:   sr.TotalScanned(),
		UpdatedFiles:   len(sr.UpdatedFiles),
		SkippedFiles:   len(sr.SkippedFiles),
		UnChangedFiles: len(sr.UnChangedFiles),
	}
}

func (ts *TaskScanSummary) Message() string {
	return fmt.Sprintf("Task: %s, Total: %d, Updated: %d, Skipped: %d, Unchanged: %d",
		ts.TaskID, ts.TotalScanned, ts.UpdatedFiles, ts.SkippedFiles, ts.UnChangedFiles)
}
