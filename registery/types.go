package registery

import "time"

type Registry struct {
	Timestamp       int64
	FileName        string
	FileSize        int64
	EncryptPassword string
	S3Path          string
	S3Bucket        string
	S3Region        string
	StorageClass    string
}

func DummyRegistry() *Registry {
	return &Registry{
		Timestamp:       time.Now().Unix(),
		FileName:        "photos_2025-07-20_02_03_39.zip",
		FileSize:        56*1024*1024 + 1,
		EncryptPassword: "asdoniasd7nyi8wdhiuwnd7&^&S",
		S3Path:          "backups/photos/photos_2025-07-20_02_03_39.zip",
		S3Bucket:        "fsds225p-backup",
		S3Region:        "us-east-1",
		StorageClass:    "DEEP_ARCHIVE",
	}
}
