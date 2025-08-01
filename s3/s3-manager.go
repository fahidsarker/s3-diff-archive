package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	lg "s3-diff-archive/logger"
	"strings"

	"bytes"

	nTypes "s3-diff-archive/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const multipartThreshold = 100 * 1024 * 1024 // 100 MB

// UploadFileToS3 uploads a file to S3 using PutObject or Multipart depending on size.
func UploadFileToS3(cnfg *nTypes.S3Config, ctx context.Context, nKey string, filePath string) error {
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

	lg.Logs.Info("Uploading: %s to s3://%s/%s", filePath, cnfg.S3Bucket, key)
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
	lg.Logs.Info("File size: %d", fileInfo.Size())
	fileSize := fileInfo.Size()

	// Use Multipart Upload if file is large
	if fileSize > multipartThreshold {
		lg.Logs.Info("Using multipart upload")
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

	lg.Logs.Info("Uploaded (simple): %s -> s3://%s/%s", filePath, cnfg.S3Bucket, key)
	return nil
}

func multipartUpload(ctx context.Context, client *s3.Client, file *os.File, fileSize int64, bucket, key string, sc types.StorageClass) (err error) {
	const partSize = int64(90 * 1024 * 1024) // 90 MB

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

		lg.Logs.Info("Uploading (%s) part %d/%d (%.2f%%)", key, partNumber, totalParts, float64(offset+curPartSize)*100/float64(fileSize))
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

	lg.Logs.Info("Uploaded (multipart): %s -> s3://%s/%s\n", file.Name(), bucket, key)
	return nil
}

// DownloadFileFromS3 downloads a file from S3 and saves it to the specified local path.
func DownloadFileFromS3(cnfg *nTypes.S3Config, ctx context.Context, nKey, destinationPath string) (err error) {
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

	lg.Logs.Info("Downloading: %s", key)

	// Get object from S3
	resp, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &cnfg.S3Bucket,
		Key:    &key,
	})

	// check if file exists
	if err != nil {
		if strings.Contains(err.Error(), "StatusCode: 404") {
			lg.Logs.Warn("S3 File not found: %s", key)
			return fmt.Errorf("not-found")
		}
		lg.Logs.Error("Error downloading file: %s, Error: %s", key, err.Error())
		return err
	}

	lg.Logs.Info("✔︎ Downloaded: %s", key)
	lg.Logs.Info("Size: %d", resp.ContentLength)

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

	lg.Logs.Info("Downloaded s3://%s/%s to %s", cnfg.S3Bucket, key, destinationPath)
	return nil
}
