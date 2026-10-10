package handler

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"testing"
)

type mockFile struct {
	*bytes.Reader
}

func (m *mockFile) Close() error { return nil }

func TestProductImageAndUUIDHelpers(t *testing.T) {
	h := &ProductHandler{}

	// validateImage: size > 5MB
	largeHdr := &multipart.FileHeader{
		Size:     6 * 1024 * 1024,
		Filename: "test.jpg",
	}
	if err := h.validateImage(largeHdr); err == nil {
		t.Errorf("expected error for image > 5MB")
	}

	// validateImage: bad extension
	badExtHdr := &multipart.FileHeader{
		Size:     1024,
		Filename: "test.exe",
	}
	if err := h.validateImage(badExtHdr); err == nil {
		t.Errorf("expected error for non-image extension")
	}

	// validateImage: good header
	goodHdr := &multipart.FileHeader{
		Size:     1024,
		Filename: "photo.jpg",
	}
	goodHdr.Header = make(map[string][]string)
	goodHdr.Header.Set("Content-Type", "image/jpeg")
	if err := h.validateImage(goodHdr); err != nil {
		t.Errorf("expected valid image header, got %v", err)
	}

	// processImage test
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var imgBuf bytes.Buffer
	_ = jpeg.Encode(&imgBuf, img, nil)
	mf := &mockFile{Reader: bytes.NewReader(imgBuf.Bytes())}
	r, lenBytes, mimeType, filename, err := h.processImage(mf, goodHdr)
	if err != nil || r == nil || lenBytes == 0 || mimeType != "image/jpeg" || filename == "" {
		t.Errorf("processImage failed: %v", err)
	}

	// parseUUID helper
	validUUID := "c56a4180-65aa-42ec-a945-5fd21dec0538"
	if parseUUID(validUUID).String() != validUUID {
		t.Errorf("parseUUID failed")
	}
}
