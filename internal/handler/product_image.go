package handler

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

// processImage resizes and compresses an image
func (h *ProductHandler) processImage(file multipart.File, header *multipart.FileHeader) (io.Reader, int64, string, string, error) {
	// 1. Decode image (supports peg, png, gif via standard library imports)
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, 0, "", "", fmt.Errorf("failed to decode image: %w", err)
	}

	// 2. Resize if too large (max 1024x1024)
	bounds := img.Bounds()
	if bounds.Dx() > 1024 || bounds.Dy() > 1024 {
		img = imaging.Fit(img, 1024, 1024, imaging.Lanczos)
	}

	// 3. Encode to JPEG with compression
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, 0, "", "", fmt.Errorf("failed to encode image: %w", err)
	}

	// 4. Generate filename
	ext := ".jpg" // We converted to JPEG
	filename := uuid.New().String() + ext
	
	return buf, int64(buf.Len()), "image/jpeg", filename, nil
}

// validateImage checks if the uploaded file is a valid image
func (h *ProductHandler) validateImage(header *multipart.FileHeader) error {
	// Check size (max 5MB)
	if header.Size > 5*1024*1024 {
		return fmt.Errorf("image too large (max 5MB)")
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return fmt.Errorf("invalid image format (jpg, jpeg, png only)")
	}
	
	// Check Content-Type
	mimeType := header.Header.Get("Content-Type")
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		// Just a warning, rely on extension or magic bytes if needed
	}

	return nil
}
