package tests

import (
	"bytes"
	"file-storage/internal/service"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
)

func makeAudioRequest(t *testing.T, fieldName, filename string, data []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename))
	h.Set("Content-Type", "audio/mpeg")
	part, err := mw.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(data)
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPut, "/", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func audioWithID3Header() []byte {
	return append([]byte{0x49, 0x44, 0x33}, make([]byte, 100)...)
}

func audioWithMPEGHeader() []byte {
	return append([]byte{0xFF, 0xFB, 0x90}, make([]byte, 100)...)
}

func TestParseAudioFromRequest_ID3Header(t *testing.T) {
	req := makeAudioRequest(t, "audio", "track.mp3", audioWithID3Header())
	result, err := service.ParseAudioFromRequest(req, "test-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %q", result.ID)
	}
}

func TestParseAudioFromRequest_MPEGHeader(t *testing.T) {
	req := makeAudioRequest(t, "audio", "track.mp3", audioWithMPEGHeader())
	result, err := service.ParseAudioFromRequest(req, "test-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestParseAudioFromRequest_WrongFieldName(t *testing.T) {
	req := makeAudioRequest(t, "file", "track.mp3", audioWithID3Header())
	_, err := service.ParseAudioFromRequest(req, "test-id")
	if err == nil {
		t.Fatal("expected error for wrong field name, got nil")
	}
	if err.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.Status)
	}
}

func TestParseAudioFromRequest_InvalidFormat(t *testing.T) {
	req := makeAudioRequest(t, "audio", "track.mp3", []byte("not an mp3 file"))
	_, err := service.ParseAudioFromRequest(req, "test-id")
	if err == nil {
		t.Fatal("expected error for invalid MP3 format, got nil")
	}
	if err.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.Status)
	}
}

func TestParseAudioFromRequest_MissingFilename(t *testing.T) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="audio"`) // no filename
	part, _ := mw.CreatePart(h)
	_, _ = part.Write(audioWithID3Header())
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPut, "/", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	_, err := service.ParseAudioFromRequest(req, "test-id")
	if err == nil {
		t.Fatal("expected error for missing filename, got nil")
	}
	if err.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.Status)
	}
}

func TestParseAudioFromRequest_NotMultipart(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString("plain text"))
	req.Header.Set("Content-Type", "application/json")
	_, err := service.ParseAudioFromRequest(req, "test-id")
	if err == nil {
		t.Fatal("expected error for non-multipart request, got nil")
	}
	if err.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.Status)
	}
}
