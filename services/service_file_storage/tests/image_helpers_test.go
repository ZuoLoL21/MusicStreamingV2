package tests

import (
	"bytes"
	"file-storage/internal/service"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
)

func makeImageRequest(t *testing.T, fieldName, filename string, data []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename))
	h.Set("Content-Type", "image/jpeg")
	part, err := mw.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(data)
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func squareJPEG(t *testing.T, size int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func squarePNG(t *testing.T, size int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestParseImageFromRequest_ValidJPEG(t *testing.T) {
	data := squareJPEG(t, service.ImageDimension)
	req := makeImageRequest(t, "image", "photo.jpg", data)
	result, err := service.ParseImageFromRequest(req, "test-id", "profile_pictures")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %q", result.ID)
	}
	if result.Bucket != "profile_pictures" {
		t.Errorf("expected bucket 'profile_pictures', got %q", result.Bucket)
	}
}

func TestParseImageFromRequest_ValidPNG_Resizes(t *testing.T) {
	data := squarePNG(t, 200) // square but smaller than ImageDimension, will be resized
	req := makeImageRequest(t, "image", "photo.png", data)
	result, err := service.ParseImageFromRequest(req, "test-id", "music_pictures")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestParseImageFromRequest_WrongFieldName(t *testing.T) {
	data := squareJPEG(t, service.ImageDimension)
	req := makeImageRequest(t, "file", "photo.jpg", data)
	_, err := service.ParseImageFromRequest(req, "test-id", "profile_pictures")
	if err == nil {
		t.Fatal("expected error for wrong field name, got nil")
	}
	if err.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.Status)
	}
}

func TestParseImageFromRequest_MissingFilename(t *testing.T) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="image"`) // no filename
	part, _ := mw.CreatePart(h)
	_, _ = part.Write(squareJPEG(t, 100))
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	_, err := service.ParseImageFromRequest(req, "test-id", "profile_pictures")
	if err == nil {
		t.Fatal("expected error for missing filename, got nil")
	}
	if err.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.Status)
	}
}

func TestParseImageFromRequest_InvalidImageData(t *testing.T) {
	req := makeImageRequest(t, "image", "photo.jpg", []byte("not an image"))
	_, err := service.ParseImageFromRequest(req, "test-id", "profile_pictures")
	if err == nil {
		t.Fatal("expected error for invalid image data, got nil")
	}
	if err.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.Status)
	}
}

func TestParseImageFromRequest_NonSquareImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 200))
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, nil)

	req := makeImageRequest(t, "image", "photo.jpg", buf.Bytes())
	_, err := service.ParseImageFromRequest(req, "test-id", "profile_pictures")
	if err == nil {
		t.Fatal("expected error for non-square image, got nil")
	}
	if err.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.Status)
	}
}

func TestParseImageFromRequest_NotMultipart(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not multipart"))
	req.Header.Set("Content-Type", "application/json")
	_, err := service.ParseImageFromRequest(req, "test-id", "profile_pictures")
	if err == nil {
		t.Fatal("expected error for non-multipart request, got nil")
	}
}
