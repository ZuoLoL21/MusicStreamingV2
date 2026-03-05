package client

import (
	"backend/internal/storage"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// MinIOFileStorageClient implements FileStorageClient using MinIO
type MinIOFileStorageClient struct {
	client         *minio.Client
	bucketName     string
	endpoint       string
	publicEndpoint string // Public-facing MinIO endpoint (e.g., localhost:9000)
	useSSL         bool
	logger         *zap.Logger
}

// NewMinIOFileStorageClient creates a new MinIO file storage client
func NewMinIOFileStorageClient(endpoint, accessKey, secretKey, bucketName, publicEndpoint string, useSSL bool, logger *zap.Logger) (*MinIOFileStorageClient, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	logger.Info("MinIO client created successfully",
		zap.String("endpoint", endpoint),
		zap.String("publicEndpoint", publicEndpoint),
		zap.String("bucket", bucketName),
		zap.Bool("useSSL", useSSL),
	)

	return &MinIOFileStorageClient{
		client:         minioClient,
		bucketName:     bucketName,
		endpoint:       endpoint,
		publicEndpoint: publicEndpoint,
		useSSL:         useSSL,
		logger:         logger,
	}, nil
}

// SaveAudio uploads audio file and returns the object path (not full URL)
func (m *MinIOFileStorageClient) SaveAudio(ctx context.Context, musicID string, audioData io.Reader) (string, error) {
	objectName := fmt.Sprintf("audio/%s.mp3", musicID)

	_, err := m.client.PutObject(ctx, m.bucketName, objectName, audioData, -1, minio.PutObjectOptions{
		ContentType: "audio/mpeg",
	})
	if err != nil {
		m.logger.Error("failed to upload audio",
			zap.String("musicID", musicID),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to upload audio: %w", err)
	}

	m.logger.Info("audio uploaded successfully",
		zap.String("musicID", musicID),
		zap.String("objectPath", objectName),
	)

	return objectName, nil
}

// UpdateAudio updates existing audio file and returns the object path
func (m *MinIOFileStorageClient) UpdateAudio(ctx context.Context, musicID string, audioData io.Reader) (string, error) {
	return m.SaveAudio(ctx, musicID, audioData)
}

// DeleteAudio removes audio file from storage
func (m *MinIOFileStorageClient) DeleteAudio(ctx context.Context, musicID string) error {
	objectName := fmt.Sprintf("audio/%s.mp3", musicID)

	err := m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		m.logger.Warn("failed to delete audio (file may not exist)",
			zap.String("musicID", musicID),
			zap.Error(err),
		)
		return nil
	}

	m.logger.Info("audio deleted successfully",
		zap.String("musicID", musicID),
	)

	return nil
}

// SaveImage uploads image to specified folder and returns the object path (not full URL)
func (m *MinIOFileStorageClient) SaveImage(ctx context.Context, folder, imageID string, imageData io.Reader) (string, error) {
	objectName := fmt.Sprintf("%s/%s.jpg", folder, imageID)

	_, err := m.client.PutObject(ctx, m.bucketName, objectName, imageData, -1, minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		m.logger.Error("failed to upload image",
			zap.String("folder", folder),
			zap.String("imageID", imageID),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	m.logger.Info("image uploaded successfully",
		zap.String("folder", folder),
		zap.String("imageID", imageID),
		zap.String("objectPath", objectName),
	)

	return objectName, nil
}

// DeleteImage removes image from storage (tries multiple extensions)
func (m *MinIOFileStorageClient) DeleteImage(ctx context.Context, folder, imageID string) error {
	extensions := []string{".jpg", ".jpeg", ".png", ".webp"}

	for _, ext := range extensions {
		objectName := fmt.Sprintf("%s/%s%s", folder, imageID, ext)
		err := m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			// Continue trying other extensions
			continue
		}

		m.logger.Info("image deleted successfully",
			zap.String("folder", folder),
			zap.String("imageID", imageID),
			zap.String("extension", ext),
		)
		return nil
	}

	m.logger.Warn("failed to delete image with any extension (file may not exist)",
		zap.String("folder", folder),
		zap.String("imageID", imageID),
	)

	return nil
}

// buildPublicURL constructs a public URL for accessing a file from MinIO (internal helper)
func (m *MinIOFileStorageClient) buildPublicURL(objectName string) string {
	protocol := "http"
	if m.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, m.publicEndpoint, m.bucketName, objectName)
}

// BuildPublicURL converts a storage path to a full public URL (implements FileStorageClient interface)
func (m *MinIOFileStorageClient) BuildPublicURL(objectPath string) string {
	return m.buildPublicURL(objectPath)
}

// GetDefaultProfileImageURL returns the full URL for the default user profile image
func (m *MinIOFileStorageClient) GetDefaultProfileImageURL() string {
	return storage.GetDefaultProfileImageURL(m.publicEndpoint, m.bucketName, m.useSSL)
}

// GetDefaultArtistImageURL returns the full URL for the default artist profile image
func (m *MinIOFileStorageClient) GetDefaultArtistImageURL() string {
	return storage.GetDefaultArtistImageURL(m.publicEndpoint, m.bucketName, m.useSSL)
}

// GetDefaultAlbumImageURL returns the full URL for the default album image
func (m *MinIOFileStorageClient) GetDefaultAlbumImageURL() string {
	return storage.GetDefaultAlbumImageURL(m.publicEndpoint, m.bucketName, m.useSSL)
}

// GetDefaultPlaylistImageURL returns the full URL for the default playlist image
func (m *MinIOFileStorageClient) GetDefaultPlaylistImageURL() string {
	return storage.GetDefaultPlaylistImageURL(m.publicEndpoint, m.bucketName, m.useSSL)
}

// GetDefaultMusicImageURL returns the full URL for the default music track image
func (m *MinIOFileStorageClient) GetDefaultMusicImageURL() string {
	return storage.GetDefaultMusicImageURL(m.publicEndpoint, m.bucketName, m.useSSL)
}

// GetDefaultImageURL returns the default image URL based on entity type (user, artist, album, playlist, music)
func (m *MinIOFileStorageClient) GetDefaultImageURL(entityType string) string {
	switch entityType {
	case "user":
		return m.GetDefaultProfileImageURL()
	case "artist":
		return m.GetDefaultArtistImageURL()
	case "album":
		return m.GetDefaultAlbumImageURL()
	case "playlist":
		return m.GetDefaultPlaylistImageURL()
	case "music":
		return m.GetDefaultMusicImageURL()
	default:
		return m.GetDefaultMusicImageURL()
	}
}

// GetObject retrieves an object from storage by its full path
func (m *MinIOFileStorageClient) GetObject(ctx context.Context, objectPath string) (io.ReadCloser, string, int64, error) {
	object, err := m.client.GetObject(ctx, m.bucketName, objectPath, minio.GetObjectOptions{})
	if err != nil {
		m.logger.Error("failed to get object from MinIO",
			zap.String("objectPath", objectPath),
			zap.Error(err))
		return nil, "", 0, fmt.Errorf("failed to get object: %w", err)
	}

	stat, err := object.Stat()
	if err != nil {
		object.Close()
		m.logger.Error("failed to stat object",
			zap.String("objectPath", objectPath),
			zap.Error(err))
		return nil, "", 0, fmt.Errorf("failed to stat object: %w", err)
	}

	return object, stat.ContentType, stat.Size, nil
}
