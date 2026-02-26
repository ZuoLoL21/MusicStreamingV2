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

	// SaveImage uploads image to specified folder (pictures-profile, pictures-artist, pictures-album, pictures-playlist, pictures-music) and returns the public URL
	SaveImage(ctx context.Context, folder, imageID string, imageData io.Reader) (url string, err error)

	// DeleteImage removes image from storage
	DeleteImage(ctx context.Context, folder, imageID string) error

	// GetDefaultProfileImageURL returns the full URL for default user profile images
	GetDefaultProfileImageURL() string

	// GetDefaultArtistImageURL returns the full URL for default artist profile images
	GetDefaultArtistImageURL() string

	// GetDefaultAlbumImageURL returns the full URL for default album images
	GetDefaultAlbumImageURL() string

	// GetDefaultPlaylistImageURL returns the full URL for default playlist images
	GetDefaultPlaylistImageURL() string

	// GetDefaultMusicImageURL returns the full URL for default music track images
	GetDefaultMusicImageURL() string

	// GetDefaultImageURL returns the default image URL based on entity type (user, artist, album, playlist, music)
	GetDefaultImageURL(entityType string) string
}
