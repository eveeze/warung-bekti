package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // Register PNG decoder
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	appConfig "github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/pkg/logger"
)

// R2Client wraps the S3 client for Cloudflare R2
type R2Client struct {
	client *s3.Client
	config *appConfig.R2Config
}

// NewR2 creates a new R2 connection
func NewR2(cfg *appConfig.R2Config) (*R2Client, error) {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID),
		}, nil
	})

	// Debug credentials (print length only for safety)
	fmt.Printf("DEBUG: R2 AccountID Len: %d\n", len(cfg.AccountID))
	fmt.Printf("DEBUG: R2 AccessKeyID Len: %d\n", len(cfg.AccessKeyID))
	fmt.Printf("DEBUG: R2 SecretAccessKey Len: %d\n", len(cfg.SecretAccessKey))

	sdkConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		config.WithRegion("auto"), // R2 uses 'auto'
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %w", err)
	}

	client := s3.NewFromConfig(sdkConfig)

	// Verify connection (optional but good practice)
	// We skip strict verification here to startup faster, 
	// but health check should verify it.
	logger.Info("R2 client initialized for bucket: %s", cfg.BucketName)

	return &R2Client{
		client: client,
		config: cfg,
	}, nil
}

// Health checks the R2 connectivity
func (r *R2Client) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.config.BucketName),
	})
	return err
}

// UploadImage processes (resizes/compresses) and uploads an image to R2
func (r *R2Client) UploadImage(ctx context.Context, objectName string, originalReader io.Reader) (string, error) {
	// 1. Decode
	img, _, err := image.Decode(originalReader)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// 2. Resize (Max 800x800, keep aspect ratio)
	// Use imaging.Linear for speed (Lanczos is too slow for large images)
	maxDimension := 800
	if img.Bounds().Dx() > maxDimension || img.Bounds().Dy() > maxDimension {
		img = imaging.Fit(img, maxDimension, maxDimension, imaging.Linear)
	}

	// 3. Encode to JPEG with compression (Quality 75)
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 75})
	if err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	// 4. Force .jpg extension
	// Since we are encoding to JPEG, we should always use .jpg
	ext := filepath.Ext(objectName)
	objectName = strings.TrimSuffix(objectName, ext) + ".jpg"
	
	// 5. Upload to R2
	input := &s3.PutObjectInput{
		Bucket:      aws.String(r.config.BucketName),
		Key:         aws.String(objectName),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/jpeg"),
		// Aggressive Caching for Performance
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	}

	_, err = r.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	return r.GetPublicURL(objectName), nil
}

// GetPublicURL returns the public URL for an object
func (r *R2Client) GetPublicURL(objectName string) string {
	// Cloudflare R2 public URL format usually: https://pub-domain.com/key
	// Or if using worker: https://worker-domain.com/key
	baseURL := r.config.PublicURL
	if baseURL == "" {
		// Fallback (unlikely to work directly without public access enabled on bucket)
		// Assuming user has a domain map or similar.
		return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", r.config.AccountID, r.config.BucketName, objectName)
	}
	
	// Remove trailing slash if present
	baseURL = strings.TrimRight(baseURL, "/")
	
	// Ensure objectName doesn't have leading slash
	objectName = strings.TrimLeft(objectName, "/")

	return fmt.Sprintf("%s/%s", baseURL, objectName)
}

// GetKeyFromURL extracts the object key from a full URL
func (r *R2Client) GetKeyFromURL(url string) (string, error) {
	// Flexible extraction parsing not dependent on specific domain parts
	// format: https://domain.com/products/file.jpg -> products/file.jpg
	
	// Split by protocol
	parts := strings.Split(url, "://")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid url format")
	}
	
	// Get path part (after the first slash following domain)
	domainAndPath := parts[1]
	slashIndex := strings.Index(domainAndPath, "/")
	if slashIndex == -1 {
		return "", fmt.Errorf("no path in url")
	}
	
	key := domainAndPath[slashIndex+1:]
	return key, nil
}

// DeleteFile deletes a file from R2
func (r *R2Client) DeleteFile(ctx context.Context, objectName string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.config.BucketName),
		Key:    aws.String(objectName),
	})
	return err
}

// GenerateProductImagePath generates a consistent path
func GenerateProductImagePath(productID string) string {
	// Using directory structure for organization
	return fmt.Sprintf("products/%s.jpg", productID) 
}
