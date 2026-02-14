package service

import (
	"bytes"
	"file-storage/internal/general"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
)

type FileResult struct {
	ID     string
	Data   []byte
	Bucket string
}

const MaxImageSize = 10 << 20 // 10MB
const ImageDimension = 640

func ParseImageFromRequest(r *http.Request) (*FileResult, *general.ErrorResult) {
	vars := mux.Vars(r)
	validated := general.ValidateUUID(vars["id"])
	if !validated {
		return nil, &general.ErrorResult{Message: "Invalid id provided", Status: http.StatusBadRequest}
	}

	id := vars["id"]
	bucketName := vars["folder"]

	// Get image file
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, &general.ErrorResult{Message: "failed to read multipart data", Status: http.StatusBadRequest}
	}
	part, err := reader.NextPart()
	if err != nil {
		return nil, &general.ErrorResult{Message: "failed to get file part", Status: http.StatusBadRequest}
	}

	if part.FormName() != "image" {
		return nil, &general.ErrorResult{Message: "expected 'image' field", Status: http.StatusBadRequest}
	}

	// Validate filename exists
	filename := part.FileName()
	if filename == "" {
		return nil, &general.ErrorResult{Message: "missing filename", Status: http.StatusBadRequest}
	}

	// Limit reader to prevent memory exhaustion
	limitedReader := io.LimitReader(part, int64(MaxImageSize)+1)
	imgData, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, &general.ErrorResult{Message: "failed to read image data", Status: http.StatusBadRequest}
	}

	// Check if size limit was exceeded
	if len(imgData) == MaxImageSize {
		return nil, &general.ErrorResult{Message: "image exceeds maximum size", Status: http.StatusRequestEntityTooLarge}
	}

	// Decode image once and validate format
	img, format, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, &general.ErrorResult{Message: "invalid image file", Status: http.StatusBadRequest}
	}

	// Validate format
	validFormats := map[string]bool{
		"jpeg": true,
		"png":  true,
		"gif":  true,
		"webp": true,
	}
	if !validFormats[format] {
		return nil, &general.ErrorResult{Message: "unsupported image format", Status: http.StatusBadRequest}
	}

	// Check size
	bounds := img.Bounds()
	if bounds.Dx() != bounds.Dy() {
		return nil, &general.ErrorResult{Message: "image must be square", Status: http.StatusBadRequest}
	}
	if bounds.Dx() != ImageDimension || bounds.Dy() != ImageDimension {
		img = resize.Resize(ImageDimension, ImageDimension, img, resize.Lanczos3)
	}

	// Convert to JPEG
	jpegData, err2 := encodeToJPEG(img)
	if err2 != nil {
		return nil, err2
	}

	return &FileResult{ID: id, Data: jpegData, Bucket: bucketName}, nil
}

func encodeToJPEG(img image.Image) ([]byte, *general.ErrorResult) {
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	if err != nil {
		return nil, &general.ErrorResult{Message: "failed to encode as JPEG", Status: http.StatusInternalServerError}
	}
	return buf.Bytes(), nil
}
