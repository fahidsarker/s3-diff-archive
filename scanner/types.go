package scanner

import "s3-diff-archive/types"

type ScannedResult struct {
	UpdatedFiles   []*types.SFile
	SkippedFiles   []string
	UnChangedFiles []*types.SFile
}
