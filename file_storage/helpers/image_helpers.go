package helpers

import (
	"bytes"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
)

type FileResult struct {
	ID   string
	Data []byte
}

const MaxImageSize = 10 << 20 // 10MB
const ImageDimension = 640

func ParseImageFromRequest(r *http.Request) (*FileResult, *ErrorResult) {
	vars := mux.Vars(r)
	validated := ValidateUUID(vars["id"])
	if !validated {
		return nil, &ErrorResult{Message: "Invalid id provided", Status: http.StatusBadRequest}
	}

	id := vars["id"]

	// Get image file
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, &ErrorResult{Message: "failed to read multipart data", Status: http.StatusBadRequest}
	}
	part, err := reader.NextPart()
	if err != nil {
		return nil, &ErrorResult{Message: "failed to get file part", Status: http.StatusBadRequest}
	}

	if part.FormName() != "image" {
		return nil, &ErrorResult{Message: "expected 'image' field", Status: http.StatusBadRequest}
	}

	// Validate filename exists
	filename := part.FileName()
	if filename == "" {
		return nil, &ErrorResult{Message: "missing filename", Status: http.StatusBadRequest}
	}

	// Limit reader to prevent memory exhaustion
	limitedReader := io.LimitReader(part, int64(MaxImageSize)+1)
	imgData, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, &ErrorResult{Message: "failed to read image data", Status: http.StatusBadRequest}
	}

	// Check if size limit was exceeded
	if len(imgData) == MaxImageSize {
		return nil, &ErrorResult{Message: "image exceeds maximum size", Status: http.StatusRequestEntityTooLarge}
	}

	// Decode image once and validate format
	img, format, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, &ErrorResult{Message: "invalid image file", Status: http.StatusBadRequest}
	}

	// Validate format
	validFormats := map[string]bool{
		"jpeg": true,
		"png":  true,
		"gif":  true,
		"webp": true,
	}
	if !validFormats[format] {
		return nil, &ErrorResult{Message: "unsupported image format", Status: http.StatusBadRequest}
	}

	// Check size
	bounds := img.Bounds()
	if bounds.Dx() != bounds.Dy() {
		return nil, &ErrorResult{Message: "image must be square", Status: http.StatusBadRequest}
	}
	if bounds.Dx() != ImageDimension || bounds.Dy() != ImageDimension {
		img = resize.Resize(ImageDimension, ImageDimension, img, resize.Lanczos3)
	}

	// Convert to JPEG
	jpegData, err2 := encodeToJPEG(img)
	if err2 != nil {
		return nil, err2
	}

	return &FileResult{ID: id, Data: jpegData}, nil
}

func encodeToJPEG(img image.Image) ([]byte, *ErrorResult) {
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	if err != nil {
		return nil, &ErrorResult{Message: "failed to encode as JPEG", Status: http.StatusInternalServerError}
	}
	return buf.Bytes(), nil
}

type RetrievalResult struct {
	Name    string
	ModTime time.Time
	File    *os.File
}

func RetrieveImage(id string, baseDir string) (*RetrievalResult, *ErrorResult) {
	validated := ValidateUUID(id)
	if !validated {
		return nil, &ErrorResult{Message: "Invalid id provided", Status: http.StatusBadRequest}
	}

	file, err := os.Open(filepath.Join(baseDir, id+".jpeg"))
	if err != nil {
		return nil, &ErrorResult{Message: "failed to open file", Status: http.StatusNotFound}
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, &ErrorResult{Message: "failed to stat file", Status: http.StatusInternalServerError}
	}

	return &RetrievalResult{Name: stat.Name(), ModTime: stat.ModTime(), File: file}, nil

}

func RetrieveDefaultImage(id string) (*RetrievalResult, *ErrorResult) {
	baseDir := GetDataFolder("default")
	file, err := os.Open(filepath.Join(baseDir, id))
	if err != nil {
		return nil, &ErrorResult{Message: "Default not found", Status: http.StatusNotFound}
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, &ErrorResult{Message: "Stat failed", Status: http.StatusInternalServerError}
	}

	return &RetrievalResult{Name: stat.Name(), ModTime: stat.ModTime(), File: file}, nil
}
