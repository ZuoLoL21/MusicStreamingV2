package storage

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ExtractImageID parses MinIO URL to get folder and imageID
// URL format: http://endpoint/bucket/folder/imageID.ext
func ExtractImageID(url string) (folder string, imageID string, err error) {
	if url == "" {
		return "", "", fmt.Errorf("empty URL provided")
	}

	parts := strings.Split(url, "/")
	if len(parts) < 5 {
		return "", "", fmt.Errorf("malformed URL: expected format http://endpoint/bucket/folder/imageID.ext, got %s", url)
	}

	folder = parts[len(parts)-2]

	filenameWithExt := parts[len(parts)-1]
	ext := filepath.Ext(filenameWithExt)
	imageID = strings.TrimSuffix(filenameWithExt, ext)

	if folder == "" || imageID == "" {
		return "", "", fmt.Errorf("could not extract folder or imageID from URL: %s", url)
	}

	return folder, imageID, nil
}

// ExtractMusicID parses MinIO URL to get musicID
// URL format: http://endpoint/bucket/audio/musicID.mp3
func ExtractMusicID(url string) (musicID string, err error) {
	if url == "" {
		return "", fmt.Errorf("empty URL provided")
	}

	parts := strings.Split(url, "/")
	if len(parts) < 5 {
		return "", fmt.Errorf("malformed URL: expected format http://endpoint/bucket/audio/musicID.mp3, got %s", url)
	}

	filenameWithExt := parts[len(parts)-1]
	ext := filepath.Ext(filenameWithExt)
	musicID = strings.TrimSuffix(filenameWithExt, ext)

	if musicID == "" {
		return "", fmt.Errorf("could not extract musicID from URL: %s", url)
	}

	return musicID, nil
}

// GetDefaultProfileImageURL returns the full MinIO URL for default profile image
func GetDefaultProfileImageURL(endpoint, bucketName string, useSSL bool) string {
	protocol := "http"
	if useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/profile_pictures/default.jpg", protocol, endpoint, bucketName)
}

// GetDefaultMusicImageURL returns the full MinIO URL for default music/album/playlist image
func GetDefaultMusicImageURL(endpoint, bucketName string, useSSL bool) string {
	protocol := "http"
	if useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/music_pictures/default.jpg", protocol, endpoint, bucketName)
}
