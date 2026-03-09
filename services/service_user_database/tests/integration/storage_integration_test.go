//go:build integration

package integration

import (
	"backend/internal/consts"
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_NewMinIOFileStorageClient(t *testing.T) {
	minioClient := SetupMinIOClient(t)

	config := GetTestConfig()
	assert.Equal(t, config.MinIOBucket, minioClient.GetBucketName())
	assert.Equal(t, config.MinIOEndpoint, minioClient.GetEndpoint())
}

func TestIntegration_SaveAudio_Success(t *testing.T) {
	minioClient := SetupMinIOClient(t)
	ctx := context.Background()

	musicID := "test-audio-001"
	audioContent := []byte("This is fake audio data for testing")
	audioReader := bytes.NewReader(audioContent)

	// Save audio
	objectPath, err := minioClient.SaveAudio(ctx, musicID, audioReader)

	assert.NoError(t, err)
	assert.Equal(t, "audio/test-audio-001.mp3", objectPath)

	// Verify the file exists by retrieving it
	reader, contentType, size, err := minioClient.GetObject(ctx, objectPath)
	require.NoError(t, err)
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

	assert.Equal(t, "audio/mpeg", contentType)
	assert.Equal(t, int64(len(audioContent)), size)

	// Read and verify content
	retrievedData, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, audioContent, retrievedData)

	// Cleanup
	err = minioClient.DeleteAudio(ctx, musicID)
	assert.NoError(t, err)
}

func TestIntegration_UpdateAudio_Success(t *testing.T) {
	minioClient := SetupMinIOClient(t)
	ctx := context.Background()

	musicID := "test-audio-002"

	// Initial save
	initialContent := []byte("Initial audio content")
	objectPath, err := minioClient.SaveAudio(ctx, musicID, bytes.NewReader(initialContent))
	require.NoError(t, err)

	// Update audio
	updatedContent := []byte("Updated audio content - much longer now")
	objectPath, err = minioClient.UpdateAudio(ctx, musicID, bytes.NewReader(updatedContent))

	assert.NoError(t, err)
	assert.Equal(t, "audio/test-audio-002.mp3", objectPath)

	// Verify updated content
	reader, _, size, err := minioClient.GetObject(ctx, objectPath)
	require.NoError(t, err)
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

	assert.Equal(t, int64(len(updatedContent)), size)

	retrievedData, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, updatedContent, retrievedData)

	// Cleanup
	err = minioClient.DeleteAudio(ctx, musicID)
	assert.NoError(t, err)
}

func TestIntegration_DeleteAudio_Success(t *testing.T) {
	minioClient := SetupMinIOClient(t)
	ctx := context.Background()

	musicID := "test-audio-003"

	// Create audio file
	audioContent := []byte("Audio to be deleted")
	_, err := minioClient.SaveAudio(ctx, musicID, bytes.NewReader(audioContent))
	require.NoError(t, err)

	// Delete audio
	err = minioClient.DeleteAudio(ctx, musicID)
	assert.NoError(t, err)

	// Verify deletion - GetObject should fail
	_, _, _, err = minioClient.GetObject(ctx, "audio/test-audio-003.mp3")
	assert.Error(t, err, "file should not exist after deletion")
}

func TestIntegration_DeleteAudio_NonExistent(t *testing.T) {
	minioClient := SetupMinIOClient(t)
	ctx := context.Background()

	// Try to delete non-existent file (should not error)
	err := minioClient.DeleteAudio(ctx, "non-existent-audio")
	assert.NoError(t, err, "deleting non-existent audio should not error")
}

func TestIntegration_SaveImage_Success(t *testing.T) {
	minioClient := SetupMinIOClient(t)
	ctx := context.Background()

	testCases := []struct {
		folder  string
		imageID string
	}{
		{consts.PicturesAlbumFolder, "album-001"},
		{consts.PicturesArtistFolder, "artist-001"},
		{consts.PicturesPlaylistFolder, "playlist-001"},
		{consts.PicturesMusicFolder, "music-001"},
		{consts.PicturesProfileFolder, "user-001"},
	}

	for _, tc := range testCases {
		t.Run(tc.folder, func(t *testing.T) {
			imageContent := createTestImage(512, 512)
			imageReader := bytes.NewReader(imageContent)

			// Save image
			objectPath, err := minioClient.SaveImage(ctx, tc.folder, tc.imageID, imageReader)

			expectedPath := tc.folder + "/" + tc.imageID + ".jpg"
			assert.NoError(t, err)
			assert.Equal(t, expectedPath, objectPath)

			// Verify the file exists
			reader, contentType, size, err := minioClient.GetObject(ctx, objectPath)
			require.NoError(t, err)
			defer func(reader io.ReadCloser) {
				_ = reader.Close()
			}(reader)

			assert.Equal(t, "image/jpeg", contentType)
			assert.Equal(t, int64(len(imageContent)), size)

			// Read and verify content
			retrievedData, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, imageContent, retrievedData)

			// Cleanup
			err = minioClient.DeleteImage(ctx, tc.folder, tc.imageID)
			assert.NoError(t, err)
		})
	}
}

func TestIntegration_DeleteImage_Success(t *testing.T) {
	minioClient := SetupMinIOClient(t)
	ctx := context.Background()

	folder := consts.PicturesAlbumFolder
	imageID := "album-delete-001"

	// Create image
	imageContent := []byte("Image to be deleted")
	_, err := minioClient.SaveImage(ctx, folder, imageID, bytes.NewReader(imageContent))
	require.NoError(t, err)

	// Delete image
	err = minioClient.DeleteImage(ctx, folder, imageID)
	assert.NoError(t, err)

	// Verify deletion
	_, _, _, err = minioClient.GetObject(ctx, folder+"/"+imageID+".jpg")
	assert.Error(t, err, "file should not exist after deletion")
}

func TestIntegration_DeleteImage_NonExistent(t *testing.T) {
	minioClient := SetupMinIOClient(t)
	ctx := context.Background()

	// Try to delete non-existent image (should not error)
	err := minioClient.DeleteImage(ctx, consts.PicturesAlbumFolder, "non-existent-image")
	assert.NoError(t, err, "deleting non-existent image should not error")
}

func TestIntegration_GetObject_NonExistent(t *testing.T) {
	minioClient := SetupMinIOClient(t)
	ctx := context.Background()

	_, _, _, err := minioClient.GetObject(ctx, "non-existent/file.mp3")
	assert.Error(t, err, "getting non-existent object should error")
}

func TestIntegration_FullWorkflow_Audio(t *testing.T) {
	minioClient := SetupMinIOClient(t)
	ctx := context.Background()

	musicID := "full-workflow-audio"
	originalContent := []byte("Original audio content")
	updatedContent := []byte("Updated audio content with more data")

	// Step 1: Save audio
	objectPath, err := minioClient.SaveAudio(ctx, musicID, bytes.NewReader(originalContent))
	require.NoError(t, err)
	assert.Equal(t, "audio/"+musicID+".mp3", objectPath)

	// Step 2: Verify save
	reader, contentType, size, err := minioClient.GetObject(ctx, objectPath)
	require.NoError(t, err)
	assert.Equal(t, "audio/mpeg", contentType)
	assert.Equal(t, int64(len(originalContent)), size)
	data, _ := io.ReadAll(reader)
	_ = reader.Close()
	assert.Equal(t, originalContent, data)

	// Step 3: Update audio
	objectPath, err = minioClient.UpdateAudio(ctx, musicID, bytes.NewReader(updatedContent))
	require.NoError(t, err)

	// Step 4: Verify update
	reader, _, size, err = minioClient.GetObject(ctx, objectPath)
	require.NoError(t, err)
	assert.Equal(t, int64(len(updatedContent)), size)
	data, _ = io.ReadAll(reader)
	_ = reader.Close()
	assert.Equal(t, updatedContent, data)

	// Step 5: Delete audio
	err = minioClient.DeleteAudio(ctx, musicID)
	require.NoError(t, err)

	// Step 6: Verify deletion
	_, _, _, err = minioClient.GetObject(ctx, objectPath)
	assert.Error(t, err)
}

func TestIntegration_FullWorkflow_Image(t *testing.T) {
	minioClient := SetupMinIOClient(t)
	ctx := context.Background()

	folder := consts.PicturesAlbumFolder
	imageID := "full-workflow-image"
	imageContent := createTestImage(1024, 1024)

	// Step 1: Save image
	objectPath, err := minioClient.SaveImage(ctx, folder, imageID, bytes.NewReader(imageContent))
	require.NoError(t, err)
	assert.Equal(t, folder+"/"+imageID+".jpg", objectPath)

	// Step 2: Verify save
	reader, contentType, size, err := minioClient.GetObject(ctx, objectPath)
	require.NoError(t, err)
	assert.Equal(t, "image/jpeg", contentType)
	assert.Equal(t, int64(len(imageContent)), size)
	data, _ := io.ReadAll(reader)
	_ = reader.Close()
	assert.Equal(t, imageContent, data)

	// Step 3: Delete image
	err = minioClient.DeleteImage(ctx, folder, imageID)
	require.NoError(t, err)

	// Step 4: Verify deletion
	_, _, _, err = minioClient.GetObject(ctx, objectPath)
	assert.Error(t, err)
}

func TestIntegration_ConcurrentOperations(t *testing.T) {
	minioClient := SetupMinIOClient(t)
	ctx := context.Background()

	// Test that multiple operations can be performed concurrently
	musicID1 := "concurrent-audio-1"
	musicID2 := "concurrent-audio-2"

	content1 := []byte("Audio content 1")
	content2 := []byte("Audio content 2")

	// Save both files
	_, err1 := minioClient.SaveAudio(ctx, musicID1, bytes.NewReader(content1))
	_, err2 := minioClient.SaveAudio(ctx, musicID2, bytes.NewReader(content2))

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Verify both exist
	reader1, _, _, err := minioClient.GetObject(ctx, "audio/"+musicID1+".mp3")
	require.NoError(t, err)
	_ = reader1.Close()

	reader2, _, _, err := minioClient.GetObject(ctx, "audio/"+musicID2+".mp3")
	require.NoError(t, err)
	_ = reader2.Close()

	// Cleanup
	_ = minioClient.DeleteAudio(ctx, musicID1)
	_ = minioClient.DeleteAudio(ctx, musicID2)
}
