package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/pkg/logger"
)

// MinioClient wraps the minio.Client with additional functionality
type MinioClient struct {
	*minio.Client
	config *config.MinioConfig
}

// NewMinio creates a new Minio connection
func NewMinio(cfg *config.MinioConfig) (*MinioClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure bucket exists
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		logger.Info("Created Minio bucket: %s", cfg.BucketName)

		// Set bucket policy to allow public read for product images
		policy := `{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::` + cfg.BucketName + `/products/*"]
				}
			]
		}`
		err = client.SetBucketPolicy(ctx, cfg.BucketName, policy)
		if err != nil {
			logger.Warn("Failed to set bucket policy: %v", err)
		}
	}

	logger.Info("Minio connection established successfully")

	return &MinioClient{
		Client: client,
		config: cfg,
	}, nil
}

// Health checks the Minio health
func (m *MinioClient) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := m.Client.BucketExists(ctx, m.config.BucketName)
	return err
}

// UploadFile uploads a file to Minio
func (m *MinioClient) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := m.Client.PutObject(ctx, m.config.BucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return the object URL
	objectURL := m.GetPublicURL(objectName)
	return objectURL, nil
}

// GetPublicURL returns the public URL for an object
func (m *MinioClient) GetPublicURL(objectName string) string {
	protocol := "http"
	if m.config.UseSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, m.config.Endpoint, m.config.BucketName, objectName)
}

// GetPresignedURL returns a presigned URL for temporary access
func (m *MinioClient) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	presignedURL, err := m.Client.PresignedGetObject(ctx, m.config.BucketName, objectName, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

// DeleteFile deletes a file from Minio
func (m *MinioClient) DeleteFile(ctx context.Context, objectName string) error {
	err := m.Client.RemoveObject(ctx, m.config.BucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// FileExists checks if a file exists in Minio
func (m *MinioClient) FileExists(ctx context.Context, objectName string) (bool, error) {
	_, err := m.Client.StatObject(ctx, m.config.BucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GenerateProductImagePath generates a path for product image
func GenerateProductImagePath(productID, filename string) string {
	return fmt.Sprintf("products/%s/%s", productID, filename)
}

// GenerateReceiptPath generates a path for receipt file
func GenerateReceiptPath(transactionID, filename string) string {
	return fmt.Sprintf("receipts/%s/%s", transactionID, filename)
}
