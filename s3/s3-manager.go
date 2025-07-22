package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"bytes"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// // UploadFileToS3 uploads a file to S3 with the specified bucket, key, and storage class.
// func UploadFileToS3(ctx context.Context, filePath, bucketName, key string, storageClass types.StorageClass) error {
// 	// Load AWS config from environment or shared config
// 	cfg, err := config.LoadDefaultConfig(ctx)
// 	if err != nil {
// 		return fmt.Errorf("unable to load AWS config: %w", err)
// 	}

// 	// Open the file
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return fmt.Errorf("unable to open file %s: %w", filePath, err)
// 	}
// 	defer file.Close()

// 	// Create the S3 client
// 	s3Client := s3.NewFromConfig(cfg)

// 	// Upload the file
// 	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
// 		Bucket:       aws.String(bucketName),
// 		Key:          aws.String(key),
// 		Body:         file,
// 		StorageClass: storageClass,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to upload file: %w", err)
// 	}

// 	fmt.Printf("File %s uploaded to s3://%s/%s with storage class %s\n", filepath.Base(filePath), bucketName, key, storageClass)
// 	return nil
// }

// const multipartThreshold = 100 * 1024 * 1024 // 100 MB
const multipartThreshold = 10 * 1024 * 1024 // 10 MB

// UploadFileToS3 uploads a file to S3 using PutObject or Multipart depending on size.
func UploadFileToS3(cnfg *S3Config, ctx context.Context, nKey string, filePath string) error {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cnfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cnfg.AccessKeyID, cnfg.SecretAccessKey, ""),
		))
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}
	s3Client := s3.NewFromConfig(cfg)

	cnfg.S3BasePath = strings.TrimSuffix(cnfg.S3BasePath, "/")
	key := cnfg.S3BasePath + "/" + nKey
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("unable to stat file: %w", err)
	}
	fileSize := fileInfo.Size()

	// Use Multipart Upload if file is large
	if fileSize > multipartThreshold {
		return multipartUpload(ctx, s3Client, file, fileSize, cnfg.S3Bucket, key, cnfg.StorageClass)
	}

	// Small file: use PutObject
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(cnfg.S3Bucket),
		Key:          aws.String(key),
		Body:         file,
		StorageClass: cnfg.StorageClass,
	})
	if err != nil {
		return fmt.Errorf("PutObject failed: %w", err)
	}

	fmt.Printf("Uploaded (simple): %s -> s3://%s/%s\n", filePath, cnfg.S3Bucket, key)
	return nil
}

func multipartUpload(ctx context.Context, client *s3.Client, file *os.File, fileSize int64, bucket, key string, sc types.StorageClass) (err error) {
	const partSize = int64(10 * 1024 * 1024) // 10 MB

	var parts []types.CompletedPart
	partNumber := int32(1)

	// Step 1: initiate multipart upload
	createResp, err := client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(key),
		StorageClass: sc,
	})
	if err != nil {
		return fmt.Errorf("failed to initiate multipart upload: %w", err)
	}
	uploadID := createResp.UploadId

	defer func() {
		if err != nil {
			client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(bucket),
				Key:      aws.String(key),
				UploadId: uploadID,
			})
		}
	}()

	totalParts := int32(fileSize / partSize)
	if fileSize%partSize != 0 {
		totalParts++
	}
	// Step 2: upload parts
	for offset := int64(0); offset < fileSize; offset += partSize {
		curPartSize := partSize
		if offset+curPartSize > fileSize {
			curPartSize = fileSize - offset
		}

		partBuf := make([]byte, curPartSize)
		_, err := file.ReadAt(partBuf, offset)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read file chunk: %w", err)
		}

		partResp, err := client.UploadPart(ctx, &s3.UploadPartInput{
			Bucket:     aws.String(bucket),
			Key:        aws.String(key),
			UploadId:   uploadID,
			PartNumber: aws.Int32(partNumber),

			// Body: bytesReader(partBuf),
			Body: bytes.NewReader(partBuf),
		})
		if err != nil {
			return fmt.Errorf("failed to upload part %d: %w", partNumber, err)
		}

		parts = append(parts, types.CompletedPart{
			ETag:       partResp.ETag,
			PartNumber: aws.Int32(partNumber),
		})

		// fmt.Printf("Uploaded part %d/%d\n", partNumber, (fileSize+partSize-1)/partSize)
		fmt.Printf("Uploaded part %d/%d (%.2f%%)\n", partNumber, totalParts, float64(offset+curPartSize)*100/float64(fileSize))
		partNumber++
	}

	// Step 3: complete multipart upload
	_, err = client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: uploadID,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: parts,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	fmt.Printf("Uploaded (multipart): %s -> s3://%s/%s\n", file.Name(), bucket, key)
	return nil
}

// DownloadFileFromS3 downloads a file from S3 and saves it to the specified local path.
func DownloadFileFromS3(cnfg *S3Config, ctx context.Context, nKey, destinationPath string) (err error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cnfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cnfg.AccessKeyID, cnfg.SecretAccessKey, ""),
		))
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}
	s3Client := s3.NewFromConfig(cfg)

	cnfg.S3BasePath = strings.TrimSuffix(cnfg.S3BasePath, "/")
	key := cnfg.S3BasePath + "/" + nKey

	println("Downloading: ", key)

	// Get object from S3
	fmt.Println("Calling GetObject...")
	resp, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &cnfg.S3Bucket,
		Key:    &key,
	})
	fmt.Println("GetObject done")
	fmt.Printf("Content-Length: %d\n", resp.ContentLength)
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer resp.Body.Close()

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Create local file
	outFile, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", destinationPath, err)
	}
	defer outFile.Close()

	// Write S3 content to file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Downloaded s3://%s/%s to %s\n", cnfg.S3Bucket, key, destinationPath)
	return nil
}
