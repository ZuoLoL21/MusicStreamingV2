package storage

import (
	"fmt"
)

// GetDefaultProfileImageURL returns the full MinIO URL for default user profile image
func GetDefaultProfileImageURL(endpoint, bucketName string, useSSL bool) string {
	protocol := "http"
	if useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/pictures-profile/default.jpg", protocol, endpoint, bucketName)
}

// GetDefaultArtistImageURL returns the full MinIO URL for default artist profile image
func GetDefaultArtistImageURL(endpoint, bucketName string, useSSL bool) string {
	protocol := "http"
	if useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/pictures-artist/default.jpg", protocol, endpoint, bucketName)
}

// GetDefaultAlbumImageURL returns the full MinIO URL for default album image
func GetDefaultAlbumImageURL(endpoint, bucketName string, useSSL bool) string {
	protocol := "http"
	if useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/pictures-album/default.jpg", protocol, endpoint, bucketName)
}

// GetDefaultPlaylistImageURL returns the full MinIO URL for default playlist image
func GetDefaultPlaylistImageURL(endpoint, bucketName string, useSSL bool) string {
	protocol := "http"
	if useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/pictures-playlist/default.jpg", protocol, endpoint, bucketName)
}

// GetDefaultMusicImageURL returns the full MinIO URL for default music track image
func GetDefaultMusicImageURL(endpoint, bucketName string, useSSL bool) string {
	protocol := "http"
	if useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/pictures-music/default.jpg", protocol, endpoint, bucketName)
}
