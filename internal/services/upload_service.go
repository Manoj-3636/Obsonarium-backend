package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxFileSize = 10 * 1024 * 1024 // 10MB
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

type UploadService struct {
	uploadDir string
}

func NewUploadService() *UploadService {
	return &UploadService{
		uploadDir: "./uploads/products",
	}
}

func (s *UploadService) SaveProductImage(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		return "", fmt.Errorf("invalid file extension: %s. Allowed extensions: .jpg, .jpeg, .png, .webp", ext)
	}

	// Validate file size
	if header.Size > maxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of 10MB")
	}

	// Generate unique filename using UnixNano timestamp
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(s.uploadDir, filename)

	// Ensure upload directory exists
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		os.Remove(filePath) // Clean up on error
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Return public URL
	publicURL := fmt.Sprintf("/api/uploads/products/%s", filename)
	return publicURL, nil
}
