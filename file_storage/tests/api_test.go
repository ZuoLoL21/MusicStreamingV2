package tests

import (
	"bytes"
	"file-storage/internal/app"
	"file-storage/internal/di"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Test server setup
// ---------------------------------------------------------------------------

func newTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()

	tmpDir := t.TempDir()
	logger := zap.NewNop()

	config := &di.Config{
		StorageLocation: tmpDir,
		RequestIDKey:    "request_id",
	}
	storage := di.GetLocalStorageManager(logger, config)
	storage.InitStorage()
	returns := di.GetReturnManager(logger, config)

	application := app.New(logger, config, storage, returns)
	srv := httptest.NewServer(application.Router())
	t.Cleanup(srv.Close)

	return srv, tmpDir
}

// ---------------------------------------------------------------------------
// Test data helpers
// ---------------------------------------------------------------------------

// minimalMP3 returns a small but valid-looking MP3 byte slice (ID3v2 header).
func minimalMP3() []byte {
	// ID3v2 magic bytes followed by enough padding to look like a real file.
	header := []byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00}
	// ID3v2 size field (4 bytes, synchsafe): size=0 means no frames.
	header = append(header, 0x00, 0x00, 0x00, 0x00)
	// Pad to a reasonable size.
	header = append(header, bytes.Repeat([]byte{0x00}, 512)...)
	return header
}

// squareJPEG returns the bytes of a 640x640 white JPEG image.
func squareJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 640, 640))
	for y := 0; y < 640; y++ {
		for x := 0; x < 640; x++ {
			img.Set(x, y, color.White)
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("failed to encode test JPEG: %v", err)
	}
	return buf.Bytes()
}

// nonSquareJPEG returns the bytes of a 100x200 (non-square) JPEG.
func nonSquareJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 100, 200))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("failed to encode non-square JPEG: %v", err)
	}
	return buf.Bytes()
}

// multipartAudio builds a multipart/form-data body with an "audio" field.
func multipartAudio(t *testing.T, content []byte, fieldName string) (body *bytes.Buffer, contentType string) {
	t.Helper()
	body = &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile(fieldName, "test.mp3")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err = part.Write(content); err != nil {
		t.Fatalf("failed to write audio content: %v", err)
	}
	_ = w.Close()
	return body, w.FormDataContentType()
}

// multipartImage builds a multipart/form-data body with an "image" field.
func multipartImage(t *testing.T, content []byte, fieldName string) (body *bytes.Buffer, contentType string) {
	t.Helper()
	body = &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile(fieldName, "test.jpg")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err = part.Write(content); err != nil {
		t.Fatalf("failed to write image content: %v", err)
	}
	_ = w.Close()
	return body, w.FormDataContentType()
}

// seedAudioFile writes a fake MP3 directly into the music storage folder.
func seedAudioFile(t *testing.T, tmpDir string, id string) {
	t.Helper()
	path := filepath.Join(tmpDir, "music", id+".mp3")
	if err := os.WriteFile(path, minimalMP3(), 0o644); err != nil {
		t.Fatalf("failed to seed audio file: %v", err)
	}
}

// seedImageFile writes a JPEG directly into the given bucket folder.
func seedImageFile(t *testing.T, tmpDir, bucket, id string) {
	t.Helper()
	path := filepath.Join(tmpDir, bucket, id+".jpeg")
	if err := os.WriteFile(path, squareJPEG(t), 0o644); err != nil {
		t.Fatalf("failed to seed image file: %v", err)
	}
}

func newID() string { return uuid.New().String() }

// ---------------------------------------------------------------------------
// Audio: GET /music/{id}
// ---------------------------------------------------------------------------

func TestStreamAudio_Success(t *testing.T) {
	srv, tmpDir := newTestServer(t)
	id := newID()
	seedAudioFile(t, tmpDir, id)

	resp, err := http.Get(srv.URL + "/music/" + id)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "audio/mpeg" {
		t.Errorf("expected Content-Type audio/mpeg, got %s", ct)
	}
}

func TestStreamAudio_InvalidUUID(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, err := http.Get(srv.URL + "/music/not-a-uuid")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestStreamAudio_NotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	id := newID()

	resp, err := http.Get(srv.URL + "/music/" + id)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Audio: PUT /music/{id}
// ---------------------------------------------------------------------------

func TestSaveAudio_Success(t *testing.T) {
	srv, _ := newTestServer(t)
	id := newID()

	body, ct := multipartAudio(t, minimalMP3(), "audio")
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/music/"+id, body)
	req.Header.Set("Content-Type", ct)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 201, got %d: %s", resp.StatusCode, b)
	}
}

func TestSaveAudio_InvalidUUID(t *testing.T) {
	srv, _ := newTestServer(t)

	body, ct := multipartAudio(t, minimalMP3(), "audio")
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/music/bad-id", body)
	req.Header.Set("Content-Type", ct)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestSaveAudio_NotMP3(t *testing.T) {
	srv, _ := newTestServer(t)
	id := newID()

	// Plain text is not a valid MP3.
	body, ct := multipartAudio(t, []byte("this is not an mp3 file at all"), "audio")
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/music/"+id, body)
	req.Header.Set("Content-Type", ct)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestSaveAudio_WrongFieldName(t *testing.T) {
	srv, _ := newTestServer(t)
	id := newID()

	body, ct := multipartAudio(t, minimalMP3(), "file") // wrong field name
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/music/"+id, body)
	req.Header.Set("Content-Type", ct)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Audio: POST /music/{id} (update)
// ---------------------------------------------------------------------------

func TestUpdateAudio_Success(t *testing.T) {
	srv, tmpDir := newTestServer(t)
	id := newID()
	seedAudioFile(t, tmpDir, id)

	body, ct := multipartAudio(t, minimalMP3(), "audio")
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/music/"+id, body)
	req.Header.Set("Content-Type", ct)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 200, got %d: %s", resp.StatusCode, b)
	}
}

func TestUpdateAudio_NotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	id := newID()

	body, ct := multipartAudio(t, minimalMP3(), "audio")
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/music/"+id, body)
	req.Header.Set("Content-Type", ct)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Audio: DELETE /music/{id}
// ---------------------------------------------------------------------------

func TestDeleteAudio_Success(t *testing.T) {
	srv, tmpDir := newTestServer(t)
	id := newID()
	seedAudioFile(t, tmpDir, id)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/music/"+id, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// File should no longer exist.
	path := filepath.Join(tmpDir, "music", id+".mp3")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected file to be deleted, but it still exists")
	}
}

func TestDeleteAudio_NotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	id := newID()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/music/"+id, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeleteAudio_InvalidUUID(t *testing.T) {
	srv, _ := newTestServer(t)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/music/not-valid", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Image: GET /image/{folder}/{id}
// ---------------------------------------------------------------------------

func TestGetImage_Success(t *testing.T) {
	srv, tmpDir := newTestServer(t)
	id := newID()
	seedImageFile(t, tmpDir, "music_pictures", id)

	resp, err := http.Get(srv.URL + "/image/music_pictures/" + id)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("expected Content-Type image/jpeg, got %s", ct)
	}
}

func TestGetImage_InvalidBucket(t *testing.T) {
	srv, _ := newTestServer(t)
	id := newID()

	resp, err := http.Get(srv.URL + "/image/invalid_bucket/" + id)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetImage_InvalidUUID(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, err := http.Get(srv.URL + "/image/music_pictures/not-a-uuid")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetImage_NotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	id := newID()

	resp, err := http.Get(srv.URL + "/image/music_pictures/" + id)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Image: POST /image/{folder}/{id} (upsert)
// ---------------------------------------------------------------------------

func TestUpdateImage_Success(t *testing.T) {
	srv, _ := newTestServer(t)
	id := newID()

	body, ct := multipartImage(t, squareJPEG(t), "image")
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/image/music_pictures/"+id, body)
	req.Header.Set("Content-Type", ct)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 200, got %d: %s", resp.StatusCode, b)
	}
}

func TestUpdateImage_InvalidBucket(t *testing.T) {
	srv, _ := newTestServer(t)
	id := newID()

	body, ct := multipartImage(t, squareJPEG(t), "image")
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/image/bad_bucket/"+id, body)
	req.Header.Set("Content-Type", ct)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdateImage_NonSquare(t *testing.T) {
	srv, _ := newTestServer(t)
	id := newID()

	body, ct := multipartImage(t, nonSquareJPEG(t), "image")
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/image/music_pictures/"+id, body)
	req.Header.Set("Content-Type", ct)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdateImage_WrongFieldName(t *testing.T) {
	srv, _ := newTestServer(t)
	id := newID()

	body, ct := multipartImage(t, squareJPEG(t), "file") // wrong field name
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/image/music_pictures/"+id, body)
	req.Header.Set("Content-Type", ct)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdateImage_NotAnImage(t *testing.T) {
	srv, _ := newTestServer(t)
	id := newID()

	body, ct := multipartImage(t, []byte("this is definitely not an image"), "image")
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/image/music_pictures/"+id, body)
	req.Header.Set("Content-Type", ct)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Image: GET /image/{folder}/ (default image)
// ---------------------------------------------------------------------------

func TestGetDefaultImage_MusicPictures(t *testing.T) {
	srv, _ := newTestServer(t)

	// The default images are bundled under data/default/ in the repo,
	// but in tests the storage is a fresh temp dir with no defaults seeded.
	// This test only checks that the server responds (200 if defaults exist,
	// 500 if the default file is missing from the temp dir — both are valid
	// server behaviours; we just confirm no panic and no routing error).
	resp, err := http.Get(srv.URL + "/image/music_pictures/")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 200 (defaults seeded) or 500 (default file not found in temp dir).
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 200 or 500, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Request-ID header propagation
// ---------------------------------------------------------------------------

func TestRequestIDHeader_Propagated(t *testing.T) {
	srv, tmpDir := newTestServer(t)
	id := newID()
	seedAudioFile(t, tmpDir, id)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/music/"+id, nil)
	const testID = "test-request-id-1234"
	req.Header.Set("X-Request-ID", testID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("X-Request-ID"); got != testID {
		t.Errorf("expected X-Request-ID %q, got %q", testID, got)
	}
}

func TestRequestIDHeader_GeneratedWhenAbsent(t *testing.T) {
	srv, tmpDir := newTestServer(t)
	id := newID()
	seedAudioFile(t, tmpDir, id)

	resp, err := http.Get(srv.URL + "/music/" + id)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("X-Request-ID"); got == "" {
		t.Error("expected X-Request-ID to be set in response when not provided in request")
	}
}
