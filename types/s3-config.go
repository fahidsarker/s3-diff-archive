package types

import (
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	S3Bucket        string
	S3BasePath      string
	StorageClass    types.StorageClass
}
