package tests

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"file-storage/internal/di"

	"go.uber.org/zap"
)

const testUUID = "550e8400-e29b-41d4-a716-446655440000"

func newTestDeps(t *testing.T) (*zap.Logger, *di.Config, *di.LocalStorageManager, *di.ReturnManager) {
	t.Helper()
	logger := zap.NewNop()
	cfg := &di.Config{
		StorageLocation: t.TempDir(),
		RequestIDKey:    "request_id",
	}
	storage := di.GetLocalStorageManager(logger, cfg)
	storage.InitStorage()
	returns := di.GetReturnManager(logger, cfg)
	return logger, cfg, storage, returns
}

func makeMultipartReq(t *testing.T, method, url, fieldName, filename string, data []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename))
	part, err := mw.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(data)
	_ = mw.Close()
	req := httptest.NewRequest(method, url, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func id3Data() []byte {
	return append([]byte{0x49, 0x44, 0x33}, make([]byte, 100)...)
}

func squareJPEGData(t *testing.T, size int) []byte {
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
