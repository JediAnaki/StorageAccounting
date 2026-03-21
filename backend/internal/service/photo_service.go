// Package service contains business logic for the inventory management system
package service

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"strings"

	"golang.org/x/image/draw"
)

// PhotoService handles photo upload, compression, and storage operations
type PhotoService struct {
	telegramBotToken string
	s3Endpoint       string
	s3AccessKey      string
	s3SecretKey      string
	s3Bucket         string
}

// NewPhotoService creates a new photo service
func NewPhotoService(telegramBotToken, s3Endpoint, s3AccessKey, s3SecretKey, s3Bucket string) *PhotoService {
	return &PhotoService{
		telegramBotToken: telegramBotToken,
		s3Endpoint:       s3Endpoint,
		s3AccessKey:      s3AccessKey,
		s3SecretKey:      s3SecretKey,
		s3Bucket:         s3Bucket,
	}
}

// ValidatePhoto validates the uploaded photo file
// Checks file type (JPEG, PNG, WebP) and size (max 10MB)
func (s *PhotoService) ValidatePhoto(fileHeader *multipart.FileHeader) error {
	// Check file size (max 10MB)
	const maxSize = 10 * 1024 * 1024 // 10MB
	if fileHeader.Size > maxSize {
		return fmt.Errorf("file too large: %d bytes (max %d bytes)", fileHeader.Size, maxSize)
	}

	// Check file type by extension
	filename := strings.ToLower(fileHeader.Filename)
	validExtensions := []string{".jpg", ".jpeg", ".png", ".webp"}

	hasValidExtension := false
	for _, ext := range validExtensions {
		if strings.HasSuffix(filename, ext) {
			hasValidExtension = true
			break
		}
	}

	if !hasValidExtension {
		return fmt.Errorf("invalid file type: only JPEG, PNG, and WebP are supported")
	}

	// Open file to verify it's actually an image
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Try to decode as image
	_, format, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("file is not a valid image: %w", err)
	}

	// Verify format matches expected types
	validFormats := map[string]bool{
		"jpeg": true,
		"png":  true,
		"webp": true,
	}

	if !validFormats[format] {
		return fmt.Errorf("unsupported image format: %s", format)
	}

	return nil
}

// CompressedPhoto represents a compressed photo with both full-size and thumbnail
type CompressedPhoto struct {
	FullSize  []byte
	Thumbnail []byte
	Format    string
}

// CompressPhoto compresses a photo to meet size requirements
// Full-size: <500KB, Thumbnail: <50KB
func (s *PhotoService) CompressPhoto(fileHeader *multipart.FileHeader) (*CompressedPhoto, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Compress full-size image (target: <500KB)
	fullSize, err := s.compressImage(img, 1920, 1080, 500*1024, format)
	if err != nil {
		return nil, fmt.Errorf("failed to compress full-size image: %w", err)
	}

	// Generate thumbnail (target: <50KB)
	thumbnail, err := s.compressImage(img, 200, 200, 50*1024, format)
	if err != nil {
		return nil, fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	return &CompressedPhoto{
		FullSize:  fullSize,
		Thumbnail: thumbnail,
		Format:    format,
	}, nil
}

// compressImage compresses an image to fit within maxWidth, maxHeight, and targetSize
func (s *PhotoService) compressImage(img image.Image, maxWidth, maxHeight, targetSize int, format string) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate new dimensions while maintaining aspect ratio
	if width > maxWidth || height > maxHeight {
		ratio := float64(width) / float64(height)
		if width > height {
			width = maxWidth
			height = int(float64(maxWidth) / ratio)
		} else {
			height = maxHeight
			width = int(float64(maxHeight) * ratio)
		}
	}

	// Resize image
	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)

	// Encode with quality adjustment to meet target size
	var buf bytes.Buffer
	quality := 85 // Start with high quality

	for quality >= 50 {
		buf.Reset()

		var err error
		switch format {
		case "jpeg", "jpg":
			err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: quality})
		case "png":
			err = png.Encode(&buf, resized)
		default:
			// Default to JPEG for other formats
			err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: quality})
		}

		if err != nil {
			return nil, fmt.Errorf("failed to encode image: %w", err)
		}

		// Check if size is acceptable
		if buf.Len() <= targetSize {
			break
		}

		// Reduce quality and try again
		quality -= 10
	}

	return buf.Bytes(), nil
}

// UploadToTelegram uploads a photo to Telegram File API
// Returns the file_id for later retrieval
func (s *PhotoService) UploadToTelegram(photoData []byte, filename string) (string, error) {
	// Note: This is a placeholder implementation
	// In production, you would use the Telegram Bot API sendPhoto method
	// to upload the file and get a file_id

	// For now, return a mock file_id
	// TODO: Implement actual Telegram File API upload
	return fmt.Sprintf("telegram_file_id_%s", filename), nil
}

// UploadToS3 uploads a photo to S3-compatible storage
// Returns the object key for later retrieval
func (s *PhotoService) UploadToS3(photoData []byte, filename string) (string, error) {
	// Note: This is a placeholder implementation
	// In production, you would use an S3 client to upload the file

	// For now, return a mock S3 key
	// TODO: Implement actual S3 upload using AWS SDK or MinIO client
	objectKey := fmt.Sprintf("photos/%s", filename)
	return objectKey, nil
}

// ProcessAndUpload validates, compresses, and uploads a photo
// Returns Telegram file_id and S3 key
func (s *PhotoService) ProcessAndUpload(fileHeader *multipart.FileHeader) (telegramFileID, s3Key string, err error) {
	// Validate photo
	if err := s.ValidatePhoto(fileHeader); err != nil {
		return "", "", fmt.Errorf("photo validation failed: %w", err)
	}

	// Compress photo
	compressed, err := s.CompressPhoto(fileHeader)
	if err != nil {
		return "", "", fmt.Errorf("photo compression failed: %w", err)
	}

	// Upload to Telegram (primary storage)
	telegramFileID, err = s.UploadToTelegram(compressed.FullSize, fileHeader.Filename)
	if err != nil {
		return "", "", fmt.Errorf("telegram upload failed: %w", err)
	}

	// Upload to S3 (fallback/archival)
	s3Key, err = s.UploadToS3(compressed.FullSize, fileHeader.Filename)
	if err != nil {
		// S3 upload is fallback, so we log but don't fail
		// In production, you might want to queue this for retry
		return telegramFileID, "", nil
	}

	return telegramFileID, s3Key, nil
}
