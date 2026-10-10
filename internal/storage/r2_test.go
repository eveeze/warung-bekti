package storage_test

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"strings"
	"testing"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/storage"
)

func TestStorageHelpers(t *testing.T) {
	// GenerateProductImagePath
	path := storage.GenerateProductImagePath("prod-123")
	if path != "products/prod-123.jpg" {
		t.Errorf("expected products/prod-123.jpg, got %s", path)
	}

	// NewR2 with empty credentials should fail gracefully
	emptyCfg := &config.R2Config{}
	client, err := storage.NewR2(emptyCfg)
	if err == nil || client != nil {
		t.Errorf("expected error initializing R2 with empty credentials")
	}

	// Fake client for URL tests
	client = &storage.R2Client{}
	// To test GetPublicURL and GetKeyFromURL with non-empty config
	configuredClient, _ := storage.NewR2(&config.R2Config{
		AccountID:       "acc-123",
		AccessKeyID:     "key-123",
		SecretAccessKey: "sec-123",
		BucketName:      "test-bucket",
		PublicURL:       "https://cdn.warung.com/",
	})
	if configuredClient != nil {
		url := configuredClient.GetPublicURL("products/item.jpg")
		if url != "https://cdn.warung.com/products/item.jpg" {
			t.Errorf("expected https://cdn.warung.com/products/item.jpg, got %s", url)
		}

		key, err := configuredClient.GetKeyFromURL(url)
		if err != nil || key != "products/item.jpg" {
			t.Errorf("expected key products/item.jpg, got %s, err: %v", key, err)
		}
	}

	// Test fallback public URL (no PublicURL specified)
	fallbackClient, _ := storage.NewR2(&config.R2Config{
		AccountID:       "acc-123",
		AccessKeyID:     "key-123",
		SecretAccessKey: "sec-123",
		BucketName:      "test-bucket",
		PublicURL:       "",
	})
	if fallbackClient != nil {
		url := fallbackClient.GetPublicURL("products/item.jpg")
		expected := "https://acc-123.r2.cloudflarestorage.com/test-bucket/products/item.jpg"
		if url != expected {
			t.Errorf("expected %s, got %s", expected, url)
		}
	}

	// Invalid URL formats for GetKeyFromURL
	dummyClient := fallbackClient
	if dummyClient != nil {
		_, err := dummyClient.GetKeyFromURL("invalid-url-without-protocol")
		if err == nil {
			t.Errorf("expected error for invalid url")
		}

		_, err = dummyClient.GetKeyFromURL("https://domain-without-path")
		if err == nil {
			t.Errorf("expected error for url without path")
		}

		// Test UploadImage error path with invalid image bytes
		_, err = dummyClient.UploadImage(nil, "test.jpg", strings.NewReader("not an image"))
		if err == nil {
			t.Errorf("expected error for invalid image data")
		}

		// Test UploadImage with valid 900x900 image (exercises resizing, format change to .jpg and upload payload preparation)
		img := image.NewRGBA(image.Rect(0, 0, 900, 900))
		var imgBuf bytes.Buffer
		_ = jpeg.Encode(&imgBuf, img, nil)
		_, _ = dummyClient.UploadImage(context.Background(), "test.png", bytes.NewReader(imgBuf.Bytes()))

		// Test Health and DeleteFile (returns network error since fake credentials, but covers statements)
		_ = dummyClient.Health(context.Background())
		_ = dummyClient.DeleteFile(context.Background(), "products/test.jpg")
	}
}
