package scanner

import "s3-diff-archive/types"

type ScannedResult struct {
	UpdatedFiles   []*types.SFile
	SkippedFiles   []string
	UnChangedFiles []*types.SFile
}

func (r *ScannedResult) TotalScanned() int64 {
	return int64(len(r.UpdatedFiles) + len(r.SkippedFiles) + len(r.UnChangedFiles))
}
