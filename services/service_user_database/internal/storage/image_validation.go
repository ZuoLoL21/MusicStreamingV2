package storage

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"golang.org/x/image/draw"
)

// ImageValidationConfig holds image validation requirements
type ImageValidationConfig struct {
	MaxFileSizeMB     int  // Maximum file size in MB
	RequireSquare     bool // Require width == height
	TargetResolution  int  // Target width/height for resizing (0 = no resize)
	AllowedFormats    []string
	EnforceJPEGOutput bool // Convert all images to JPEG
}

// DefaultProfileImageConfig returns standard config for profile images
func DefaultProfileImageConfig() ImageValidationConfig {
	return ImageValidationConfig{
		MaxFileSizeMB:     10,
		RequireSquare:     true,
		TargetResolution:  512, // 512x512
		AllowedFormats:    []string{"jpeg", "jpg", "png", "webp"},
		EnforceJPEGOutput: true,
	}
}

// DefaultMusicImageConfig returns standard config for music/album/playlist images
func DefaultMusicImageConfig() ImageValidationConfig {
	return ImageValidationConfig{
		MaxFileSizeMB:     10,
		RequireSquare:     true,
		TargetResolution:  1024, // 1024x1024
		AllowedFormats:    []string{"jpeg", "jpg", "png", "webp"},
		EnforceJPEGOutput: true,
	}
}

// ValidateAndProcessImage validates and optionally transforms an image
// Returns the processed image data as a reader and the final format
func ValidateAndProcessImage(data io.Reader, config ImageValidationConfig) (io.Reader, string, error) {
	buf := &bytes.Buffer{}
	written, err := io.Copy(buf, data)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %w", err)
	}

	maxSize := int64(config.MaxFileSizeMB) << 20
	if written > maxSize {
		return nil, "", fmt.Errorf("file size %d bytes exceeds maximum %d MB", written, config.MaxFileSizeMB)
	}

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, "", fmt.Errorf("invalid image format: %w", err)
	}

	// Validate format if restrictions are specified
	if len(config.AllowedFormats) > 0 {
		allowed := false
		for _, f := range config.AllowedFormats {
			if format == f {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, "", fmt.Errorf("unsupported image format: %s (allowed: %v)", format, config.AllowedFormats)
		}
	}

	// Check if image is square
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if config.RequireSquare && width != height {
		return nil, "", fmt.Errorf("image must be square (got %dx%d)", width, height)
	}

	// Resize if target resolution is specified
	if config.TargetResolution > 0 && (width != config.TargetResolution || height != config.TargetResolution) {
		img = resizeImage(img, config.TargetResolution, config.TargetResolution)
	}

	// Done
	outputBuf := &bytes.Buffer{}
	finalFormat := format

	if config.EnforceJPEGOutput && format != "jpeg" {
		if err := encodeJPEG(outputBuf, img); err != nil {
			return nil, "", fmt.Errorf("failed to encode as JPEG: %w", err)
		}
		finalFormat = "jpeg"
	} else if config.TargetResolution > 0 {
		if err := encodeImage(outputBuf, img, format); err != nil {
			return nil, "", fmt.Errorf("failed to re-encode image: %w", err)
		}
	} else {
		return bytes.NewReader(buf.Bytes()), format, nil
	}

	return outputBuf, finalFormat, nil
}

func resizeImage(img image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

func encodeJPEG(w io.Writer, img image.Image) error {
	return jpeg.Encode(w, img, &jpeg.Options{Quality: 90})
}

func encodeImage(w io.Writer, img image.Image, format string) error {
	switch format {
	case "jpeg", "jpg":
		return encodeJPEG(w, img)
	case "png":
		return png.Encode(w, img)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
