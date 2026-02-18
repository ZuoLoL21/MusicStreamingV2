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

func TestStreamAudio_InvalidUUID(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewMusicHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/music/invalid", nil)
	r = mux.SetURLVars(r, map[string]string{"id": "not-a-uuid"})

	h.StreamAudio(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestStreamAudio_NotFound(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewMusicHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/music/"+testUUID, nil)
	r = mux.SetURLVars(r, map[string]string{"id": testUUID})

	h.StreamAudio(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestStreamAudio_Success(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewMusicHandler(logger, cfg, storage, returns)

	baseDir, _ := storage.GetDataFolder("music")
	_ = os.WriteFile(filepath.Join(baseDir, testUUID+".mp3"), id3Data(), 0644)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/music/"+testUUID, nil)
	r = mux.SetURLVars(r, map[string]string{"id": testUUID})

	h.StreamAudio(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestSaveAudio_InvalidUUID(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewMusicHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := makeMultipartReq(t, http.MethodPut, "/music/bad", "audio", "track.mp3", id3Data())
	r = mux.SetURLVars(r, map[string]string{"id": "not-a-uuid"})

	h.SaveAudio(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSaveAudio_WrongFieldName(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewMusicHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := makeMultipartReq(t, http.MethodPut, "/music/"+testUUID, "file", "track.mp3", id3Data())
	r = mux.SetURLVars(r, map[string]string{"id": testUUID})

	h.SaveAudio(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSaveAudio_Success(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewMusicHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := makeMultipartReq(t, http.MethodPut, "/music/"+testUUID, "audio", "track.mp3", id3Data())
	r = mux.SetURLVars(r, map[string]string{"id": testUUID})

	h.SaveAudio(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestUpdateAudio_InvalidUUID(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewMusicHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := makeMultipartReq(t, http.MethodPost, "/music/bad", "audio", "track.mp3", id3Data())
	r = mux.SetURLVars(r, map[string]string{"id": "not-a-uuid"})

	h.UpdateAudio(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateAudio_FileNotFound(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewMusicHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := makeMultipartReq(t, http.MethodPost, "/music/"+testUUID, "audio", "track.mp3", id3Data())
	r = mux.SetURLVars(r, map[string]string{"id": testUUID})

	h.UpdateAudio(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdateAudio_Success(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewMusicHandler(logger, cfg, storage, returns)

	baseDir, _ := storage.GetDataFolder("music")
	_ = os.WriteFile(filepath.Join(baseDir, testUUID+".mp3"), id3Data(), 0644)

	w := httptest.NewRecorder()
	r := makeMultipartReq(t, http.MethodPost, "/music/"+testUUID, "audio", "track.mp3", id3Data())
	r = mux.SetURLVars(r, map[string]string{"id": testUUID})

	h.UpdateAudio(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestDeleteAudio_InvalidUUID(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewMusicHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/music/bad", nil)
	r = mux.SetURLVars(r, map[string]string{"id": "not-a-uuid"})

	h.DeleteAudio(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeleteAudio_NotFound(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewMusicHandler(logger, cfg, storage, returns)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/music/"+testUUID, nil)
	r = mux.SetURLVars(r, map[string]string{"id": testUUID})

	h.DeleteAudio(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDeleteAudio_Success(t *testing.T) {
	logger, cfg, storage, returns := newTestDeps(t)
	h := handlers.NewMusicHandler(logger, cfg, storage, returns)

	baseDir, _ := storage.GetDataFolder("music")
	filePath := filepath.Join(baseDir, testUUID+".mp3")
	_ = os.WriteFile(filePath, id3Data(), 0644)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/music/"+testUUID, nil)
	r = mux.SetURLVars(r, map[string]string{"id": testUUID})

	h.DeleteAudio(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("expected file to be deleted, but it still exists")
	}
}
