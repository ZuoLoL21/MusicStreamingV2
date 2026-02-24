package tests

import (
	"file-storage/internal/di"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func newTestStorageDeps(t *testing.T) (*di.LocalStorageManager, *di.Config) {
	t.Helper()
	cfg := &di.Config{
		StorageLocation: t.TempDir(),
		RequestIDKey:    "request_id",
	}
	storage := di.GetLocalStorageManager(zap.NewNop(), cfg)
	storage.InitStorage()
	return storage, cfg
}

func TestGetDataFolder_ValidNames(t *testing.T) {
	storage, cfg := newTestStorageDeps(t)
	for _, name := range di.PossibleStorages {
		t.Run(name, func(t *testing.T) {
			got, err := storage.GetDataFolder(name)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", name, err)
			}
			expected := filepath.Join(cfg.StorageLocation, name)
			if got != expected {
				t.Errorf("expected %q, got %q", expected, got)
			}
		})
	}
}

func TestGetDataFolder_InvalidName(t *testing.T) {
	storage, _ := newTestStorageDeps(t)
	_, err := storage.GetDataFolder("not_a_valid_storage")
	if err == nil {
		t.Fatal("expected error for invalid storage name, got nil")
	}
}

func TestSaveToFile_Success(t *testing.T) {
	storage, cfg := newTestStorageDeps(t)
	dest := filepath.Join(cfg.StorageLocation, "test.txt")
	content := "hello world"

	n, errS := storage.SaveToFile(strings.NewReader(content), dest)
	if errS != nil {
		t.Fatalf("unexpected error: %v", errS)
	}
	if n != int64(len(content)) {
		t.Errorf("expected %d bytes written, got %d", len(content), n)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("expected content %q, got %q", content, string(data))
	}
}

func TestSaveToFile_InvalidDirectory(t *testing.T) {
	storage, cfg := newTestStorageDeps(t)
	nonExistentDir := filepath.Join(cfg.StorageLocation, "does_not_exist", "file.txt")

	_, errS := storage.SaveToFile(strings.NewReader("data"), nonExistentDir)
	if errS == nil {
		t.Fatal("expected error for non-existent directory, got nil")
	}
}

func TestInitStorage_CreatesDirectories(t *testing.T) {
	cfg := &di.Config{
		StorageLocation: t.TempDir(),
		RequestIDKey:    "request_id",
	}
	storage := di.GetLocalStorageManager(zap.NewNop(), cfg)
	storage.InitStorage()

	for _, name := range di.PossibleStorages {
		dir := filepath.Join(cfg.StorageLocation, name)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected directory %q to be created, but it doesn't exist", dir)
		}
	}
}
