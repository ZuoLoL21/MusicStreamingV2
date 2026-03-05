package storage

import (
	"context"
	"io"
)

// FileStorageClient abstracts file storage operations, allowing swapping between local storage, S3, or other providers without changing handler code.
type FileStorageClient interface {
	// SaveAudio uploads audio file and returns the object path (e.g., "audio/123.mp3")
	SaveAudio(ctx context.Context, musicID string, audioData io.Reader) (objectPath string, err error)

	// UpdateAudio updates existing audio file and returns the object path
	UpdateAudio(ctx context.Context, musicID string, audioData io.Reader) (objectPath string, err error)

	// DeleteAudio removes audio file from storage
	DeleteAudio(ctx context.Context, musicID string) error

	// SaveImage uploads image to specified folder (pictures-profile, pictures-artist, pictures-album, pictures-playlist, pictures-music) and returns the object path (e.g., "pictures-album/456.jpg")
	SaveImage(ctx context.Context, folder, imageID string, imageData io.Reader) (objectPath string, err error)

	// DeleteImage removes image from storage
	DeleteImage(ctx context.Context, folder, imageID string) error

	// GetObject retrieves an object from storage by its full path
	GetObject(ctx context.Context, objectPath string) (data io.ReadCloser, contentType string, size int64, err error)
}
