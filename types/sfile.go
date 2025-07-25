package types

type SFile struct {
	RelativePath string `json:"path"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	Mtime        int64  `json:"mtime"`
}

func SfilesToNames(sfiles []*SFile) []string {
	var names []string
	for _, sfile := range sfiles {
		names = append(names, sfile.Name)
	}
	return names
}
