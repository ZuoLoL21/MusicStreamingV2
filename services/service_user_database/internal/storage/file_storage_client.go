package storage

import (
	"context"
	"io"
)

// FileStorageClient abstracts file storage operations, allowing swapping between local storage, S3, or other providers without changing handler code.
type FileStorageClient interface {
	// SaveAudio uploads audio file and returns the public URL for streaming
	SaveAudio(ctx context.Context, musicID string, audioData io.Reader) (url string, err error)

	// UpdateAudio updates existing audio file and returns the public URL
	UpdateAudio(ctx context.Context, musicID string, audioData io.Reader) (url string, err error)

	// DeleteAudio removes audio file from storage
	DeleteAudio(ctx context.Context, musicID string) error

	// SaveImage uploads image to specified folder (music_pictures, profile_pictures) and returns the public URL
	SaveImage(ctx context.Context, folder, imageID string, imageData io.Reader) (url string, err error)

	// DeleteImage removes image from storage
	DeleteImage(ctx context.Context, folder, imageID string) error

	// GetDefaultProfileImageURL returns the full URL for default profile images
	GetDefaultProfileImageURL() string

	// GetDefaultMusicImageURL returns the full URL for default music/album/playlist images
	GetDefaultMusicImageURL() string
}
