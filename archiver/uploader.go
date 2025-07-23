package archiver

import (
	"context"
	lg "s3-diff-archive/logger"
	"s3-diff-archive/s3"
	"s3-diff-archive/utils"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func UploadZipDiffResult(result ZipDiffTaskResult) {
	// err := s3.UploadFileToS3(result.TaskConfig.CreateS3Config(), context.TODO(), result.DBZipPath, result.ArchivesPaths)

	for i := range result.ArchivesPaths {
		err := s3.UploadFileToS3(result.TaskConfig.CreateS3Config(result.TaskConfig.StorageClass), context.TODO(), utils.FileNameFromPath(result.ArchivesPaths[i]), result.ArchivesPaths[i])
		if err != nil {
			lg.Logs.Fatal("Error Uploading (%s) :%s", result.ArchivesPaths[i], err.Error())
		}
	}
	err := s3.UploadFileToS3(result.TaskConfig.CreateS3Config(types.StorageClassStandard), context.TODO(), "db.zip", result.DBZipPath)
	if err != nil {
		lg.Logs.Fatal("Error Uploading (%s) :%s", result.DBZipPath, err.Error())
	}
}
