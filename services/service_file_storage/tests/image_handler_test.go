package tests

import (
	"file-storage/internal/handlers"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
)

const testImageUUID = "660e8400-e29b-41d4-a716-446655440001"

func TestGetImage_InvalidUUID(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewImageHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/image/profile_pictures/bad", nil)
	r = mux.SetURLVars(r, map[string]string{"folder": "profile_pictures", "id": "not-a-uuid"})

	h.GetImage(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetImage_InvalidBucket(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewImageHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/image/bad_bucket/"+testImageUUID, nil)
	r = mux.SetURLVars(r, map[string]string{"folder": "bad_bucket", "id": testImageUUID})

	h.GetImage(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetImage_NotFound(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewImageHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/image/profile_pictures/"+testImageUUID, nil)
	r = mux.SetURLVars(r, map[string]string{"folder": "profile_pictures", "id": testImageUUID})

	h.GetImage(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetImage_Success(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewImageHandler(logger, cfg, storage, returns)

	baseDir, _ := storage.GetDataFolder("profile_pictures")
	_ = os.WriteFile(filepath.Join(baseDir, testImageUUID+".jpeg"), squareJPEGData(t, 640), 0644)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/image/profile_pictures/"+testImageUUID, nil)
	r = mux.SetURLVars(r, map[string]string{"folder": "profile_pictures", "id": testImageUUID})

	h.GetImage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetDefaultImage_InvalidBucket(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewImageHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/image/invalid/", nil)
	r = mux.SetURLVars(r, map[string]string{"folder": "invalid"})

	h.GetDefaultImage(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateImage_InvalidUUID(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewImageHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := makeMultipartReq(t, http.MethodPost, "/image/profile_pictures/bad", "image", "photo.jpg", squareJPEGData(t, 640))
	r = mux.SetURLVars(r, map[string]string{"folder": "profile_pictures", "id": "not-a-uuid"})

	h.UpdateImage(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateImage_InvalidBucket(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewImageHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := makeMultipartReq(t, http.MethodPost, "/image/bad_bucket/"+testImageUUID, "image", "photo.jpg", squareJPEGData(t, 640))
	r = mux.SetURLVars(r, map[string]string{"folder": "bad_bucket", "id": testImageUUID})

	h.UpdateImage(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateImage_Success(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewImageHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := makeMultipartReq(t, http.MethodPost, "/image/profile_pictures/"+testImageUUID, "image", "photo.jpg", squareJPEGData(t, 640))
	r = mux.SetURLVars(r, map[string]string{"folder": "profile_pictures", "id": testImageUUID})

	h.UpdateImage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestDeleteImage_InvalidUUID(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewImageHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/image/profile_pictures/bad", nil)
	r = mux.SetURLVars(r, map[string]string{"folder": "profile_pictures", "id": "not-a-uuid"})

	h.DeleteImage(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeleteImage_InvalidBucket(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewImageHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/image/bad_bucket/"+testImageUUID, nil)
	r = mux.SetURLVars(r, map[string]string{"folder": "bad_bucket", "id": testImageUUID})

	h.DeleteImage(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeleteImage_NotFound(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewImageHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/image/profile_pictures/"+testImageUUID, nil)
	r = mux.SetURLVars(r, map[string]string{"folder": "profile_pictures", "id": testImageUUID})

	h.DeleteImage(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDeleteImage_Success(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewImageHandler(logger, cfg, storage, returns)

	baseDir, _ := storage.GetDataFolder("profile_pictures")
	filePath := filepath.Join(baseDir, testImageUUID+".jpeg")
	_ = os.WriteFile(filePath, squareJPEGData(t, 640), 0644)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/image/profile_pictures/"+testImageUUID, nil)
	r = mux.SetURLVars(r, map[string]string{"folder": "profile_pictures", "id": testImageUUID})

	h.DeleteImage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("expected image file to be deleted, but it still exists")
	}
}
