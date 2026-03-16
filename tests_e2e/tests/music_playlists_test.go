//go:build e2e

package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMusic_List tests listing all music
func TestMusic_List(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/music", nil)
	require.NoError(t, err, "List music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestMusic_Create tests creating a new music track
func TestMusic_Create(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	musicData := GenerateTestMusicData()

	resp, err := client.Request("POST", "/music", musicData)
	require.NoError(t, err, "Create music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// Accept 200 or 201
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated,
		"Music creation should succeed")

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		body := AssertResponseBody(t, resp)
		AssertContainsField(t, body, "uuid")
		AssertContainsField(t, body, "title")
	}
}

// TestMusic_Update tests updating a music track
func TestMusic_Update(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// First create a music track
	musicData := GenerateTestMusicData()
	resp, err := client.Request("POST", "/music", musicData)
	require.NoError(t, err, "Create music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Skip("Cannot create music for update test")
	}

	body := AssertResponseBody(t, resp)
	musicUUID := body["uuid"].(string)

	// Now update the music
	updateData := map[string]interface{}{
		"title":       "Updated Song Title",
		"description": "Updated description",
	}

	resp, err = client.Request("PUT", "/music/"+musicUUID, updateData)
	require.NoError(t, err, "Update music request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestMusic_Delete tests deleting a music track
func TestMusic_Delete(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// First create a music track
	musicData := GenerateTestMusicData()
	resp, err := client.Request("POST", "/music", musicData)
	require.NoError(t, err, "Create music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Skip("Cannot create music for delete test")
	}

	body := AssertResponseBody(t, resp)
	musicUUID := body["uuid"].(string)

	// Now delete the music
	resp, err = client.Request("DELETE", "/music/"+musicUUID, nil)
	require.NoError(t, err, "Delete music request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestMusic_GetStorage tests getting storage info for a music track
func TestMusic_GetStorage(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get a music track first
	resp, err := client.Request("GET", "/music?limit=1", nil)
	require.NoError(t, err, "List music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if music, ok := body["music"].([]interface{}); ok && len(music) > 0 {
		if track, ok := music[0].(map[string]interface{}); ok {
			if uuid, ok := track["uuid"].(string); ok {
				resp, err = client.Request("GET", "/music/"+uuid+"/storage", nil)
				require.NoError(t, err, "Get storage request should not fail")
				defer resp.Body.Close()

				// Should return OK or Not Found
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound,
					"Get storage should be handled")
			}
		}
	}
}

// TestMusic_Play tests play tracking for a music track
func TestMusic_Play(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get a music track first
	resp, err := client.Request("GET", "/music?limit=1", nil)
	require.NoError(t, err, "List music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if music, ok := body["music"].([]interface{}); ok && len(music) > 0 {
		if track, ok := music[0].(map[string]interface{}); ok {
			if uuid, ok := track["uuid"].(string); ok {
				resp, err = client.Request("POST", "/music/"+uuid+"/play", nil)
				require.NoError(t, err, "Play request should not fail")
				defer resp.Body.Close()

				// Should return OK or Bad Request
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
					"Play tracking should be handled")
			}
		}
	}
}

// TestMusic_GetTags tests getting tags for a music track
func TestMusic_GetTags(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get a music track first
	resp, err := client.Request("GET", "/music?limit=1", nil)
	require.NoError(t, err, "List music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if music, ok := body["music"].([]interface{}); ok && len(music) > 0 {
		if track, ok := music[0].(map[string]interface{}); ok {
			if uuid, ok := track["uuid"].(string); ok {
				resp, err = client.Request("GET", "/music/"+uuid+"/tags", nil)
				require.NoError(t, err, "Get tags request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)
			}
		}
	}
}

// TestMusic_AddTag tests adding a tag to a music track
func TestMusic_AddTag(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get a music track first
	resp, err := client.Request("GET", "/music?limit=1", nil)
	require.NoError(t, err, "List music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if music, ok := body["music"].([]interface{}); ok && len(music) > 0 {
		if track, ok := music[0].(map[string]interface{}); ok {
			if uuid, ok := track["uuid"].(string); ok {
				tagName := "test-tag-" + NewTestUUID()[:8]
				resp, err = client.Request("POST", "/music/"+uuid+"/tags",
					map[string]interface{}{"name": tagName})
				require.NoError(t, err, "Add tag request should not fail")
				defer resp.Body.Close()

				// Should return OK or Bad Request
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
					"Add tag should be handled")
			}
		}
	}
}

// TestMusic_RemoveTag tests removing a tag from a music track
func TestMusic_RemoveTag(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get a music track first
	resp, err := client.Request("GET", "/music?limit=1", nil)
	require.NoError(t, err, "List music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if music, ok := body["music"].([]interface{}); ok && len(music) > 0 {
		if track, ok := music[0].(map[string]interface{}); ok {
			if uuid, ok := track["uuid"].(string); ok {
				// Try to remove a non-existent tag
				resp, err = client.Request("DELETE", "/music/"+uuid+"/tags/test-tag-nonexistent", nil)
				require.NoError(t, err, "Remove tag request should not fail")
				defer resp.Body.Close()

				// Should return OK or Not Found
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound,
					"Remove tag should be handled")
			}
		}
	}
}

// TestMusic_GetLikedStatus tests checking if a music track is liked
func TestMusic_GetLikedStatus(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get a music track first
	resp, err := client.Request("GET", "/music?limit=1", nil)
	require.NoError(t, err, "List music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if music, ok := body["music"].([]interface{}); ok && len(music) > 0 {
		if track, ok := music[0].(map[string]interface{}); ok {
			if uuid, ok := track["uuid"].(string); ok {
				resp, err = client.Request("GET", "/music/"+uuid+"/liked", nil)
				require.NoError(t, err, "Get liked status request should not fail")
				defer resp.Body.Close()

				// Should return OK
				AssertResponseStatus(t, resp, http.StatusOK)
			}
		}
	}
}

// TestMusic_GetByUUID tests getting a music track by UUID
func TestMusic_GetByUUID(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/music?limit=1", nil)
	require.NoError(t, err, "List music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if music, ok := body["music"].([]interface{}); ok && len(music) > 0 {
		if track, ok := music[0].(map[string]interface{}); ok {
			if uuid, ok := track["uuid"].(string); ok {
				resp, err = client.Request("GET", "/music/"+uuid, nil)
				require.NoError(t, err, "Get music request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)
			}
		}
	}
}

// TestMusic_Like tests liking a music track
func TestMusic_Like(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get first music track
	resp, err := client.Request("GET", "/music?limit=1", nil)
	require.NoError(t, err, "List music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if music, ok := body["music"].([]interface{}); ok && len(music) > 0 {
		if track, ok := music[0].(map[string]interface{}); ok {
			if uuid, ok := track["uuid"].(string); ok {
				resp, err = client.Request("POST", "/music/"+uuid+"/like", nil)
				require.NoError(t, err, "Like music request should not fail")
				defer resp.Body.Close()

				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
					"Like should succeed or return error")
			}
		}
	}
}

// TestMusic_Unlike tests unliking a music track
func TestMusic_Unlike(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/music?limit=1", nil)
	require.NoError(t, err, "List music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if music, ok := body["music"].([]interface{}); ok && len(music) > 0 {
		if track, ok := music[0].(map[string]interface{}); ok {
			if uuid, ok := track["uuid"].(string); ok {
				resp, err = client.Request("DELETE", "/music/"+uuid+"/like", nil)
				require.NoError(t, err, "Unlike music request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)
			}
		}
	}
}

// TestMusic_GetLikes tests getting liked music
func TestMusic_GetLikes(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/music/liked", nil)
	require.NoError(t, err, "Get liked music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestPlaylists_List tests listing all playlists
func TestPlaylists_List(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/playlists", nil)
	require.NoError(t, err, "List playlists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestPlaylists_Create tests creating a new playlist
func TestPlaylists_Create(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	playlistData := GenerateTestPlaylistData()

	resp, err := client.Request("POST", "/playlists", playlistData)
	require.NoError(t, err, "Create playlist request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated,
		"Playlist creation should succeed")

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		body := AssertResponseBody(t, resp)
		AssertContainsField(t, body, "uuid")
		AssertContainsField(t, body, "name")
	}
}

// TestPlaylists_GetByUUID tests getting a playlist by UUID
func TestPlaylists_GetByUUID(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/playlists?limit=1", nil)
	require.NoError(t, err, "List playlists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if playlists, ok := body["playlists"].([]interface{}); ok && len(playlists) > 0 {
		if playlist, ok := playlists[0].(map[string]interface{}); ok {
			if uuid, ok := playlist["uuid"].(string); ok {
				resp, err = client.Request("GET", "/playlists/"+uuid, nil)
				require.NoError(t, err, "Get playlist request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)
			}
		}
	}
}

// TestPlaylists_Update tests updating a playlist
func TestPlaylists_Update(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Create a playlist first
	playlistData := GenerateTestPlaylistData()
	resp, err := client.Request("POST", "/playlists", playlistData)
	require.NoError(t, err, "Create playlist request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Skip("Cannot create playlist for update test")
	}

	body := AssertResponseBody(t, resp)
	playlistUUID := body["uuid"].(string)

	// Update the playlist
	updateData := map[string]interface{}{
		"name":        "Updated Playlist Name",
		"description": "Updated description",
		"is_public":   true,
	}

	resp, err = client.Request("PUT", "/playlists/"+playlistUUID, updateData)
	require.NoError(t, err, "Update playlist request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestPlaylists_Delete tests deleting a playlist
func TestPlaylists_Delete(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Create a playlist first
	playlistData := GenerateTestPlaylistData()
	resp, err := client.Request("POST", "/playlists", playlistData)
	require.NoError(t, err, "Create playlist request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Skip("Cannot create playlist for delete test")
	}

	body := AssertResponseBody(t, resp)
	playlistUUID := body["uuid"].(string)

	// Delete the playlist
	resp, err = client.Request("DELETE", "/playlists/"+playlistUUID, nil)
	require.NoError(t, err, "Delete playlist request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestPlaylists_AddTrack tests adding a track to a playlist
func TestPlaylists_AddTrack(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get a playlist
	resp, err := client.Request("GET", "/playlists?limit=1", nil)
	require.NoError(t, err, "List playlists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if playlists, ok := body["playlists"].([]interface{}); ok && len(playlists) > 0 {
		if playlist, ok := playlists[0].(map[string]interface{}); ok {
			if playlistUUID, ok := playlist["uuid"].(string); ok {
				// Get a music track
				resp, err = client.Request("GET", "/music?limit=1", nil)
				require.NoError(t, err, "List music request should not fail")
				defer resp.Body.Close()

				musicBody := AssertResponseBody(t, resp)
				if music, ok := musicBody["music"].([]interface{}); ok && len(music) > 0 {
					if track, ok := music[0].(map[string]interface{}); ok {
						if musicUUID, ok := track["uuid"].(string); ok {
							// Add track to playlist
							resp, err = client.Request("POST", "/playlists/"+playlistUUID+"/tracks",
								map[string]interface{}{"music_uuid": musicUUID})
							require.NoError(t, err, "Add track request should not fail")
							defer resp.Body.Close()

							assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
								"Add track should succeed or return error")
						}
					}
				}
			}
		}
	}
}

// TestPlaylists_RemoveTrack tests removing a track from a playlist
func TestPlaylists_RemoveTrack(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get a playlist
	resp, err := client.Request("GET", "/playlists?limit=1", nil)
	require.NoError(t, err, "List playlists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if playlists, ok := body["playlists"].([]interface{}); ok && len(playlists) > 0 {
		if playlist, ok := playlists[0].(map[string]interface{}); ok {
			if playlistUUID, ok := playlist["uuid"].(string); ok {
				// Try to remove a track (might not exist but shouldn't crash)
				resp, err = client.Request("DELETE", "/playlists/"+playlistUUID+"/tracks/test-music-uuid", nil)
				require.NoError(t, err, "Remove track request should not fail")
				defer resp.Body.Close()

				// Should not return 404
				assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
			}
		}
	}
}

// TestPlaylists_GetTracks tests getting tracks in a playlist
func TestPlaylists_GetTracks(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/playlists?limit=1", nil)
	require.NoError(t, err, "List playlists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if playlists, ok := body["playlists"].([]interface{}); ok && len(playlists) > 0 {
		if playlist, ok := playlists[0].(map[string]interface{}); ok {
			if playlistUUID, ok := playlist["uuid"].(string); ok {
				resp, err = client.Request("GET", "/playlists/"+playlistUUID+"/tracks", nil)
				require.NoError(t, err, "Get playlist tracks request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)
			}
		}
	}
}

// TestPlaylists_ReorderTrack tests reordering a track in a playlist
func TestPlaylists_ReorderTrack(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get a playlist
	resp, err := client.Request("GET", "/playlists?limit=1", nil)
	require.NoError(t, err, "List playlists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if playlists, ok := body["playlists"].([]interface{}); ok && len(playlists) > 0 {
		if playlist, ok := playlists[0].(map[string]interface{}); ok {
			if playlistUUID, ok := playlist["uuid"].(string); ok {
				// Try to reorder with a dummy track UUID
				resp, err = client.Request("PUT", "/playlists/"+playlistUUID+"/tracks/"+NewTestUUID()+"/position",
					map[string]interface{}{"position": 1})
				require.NoError(t, err, "Reorder track request should not fail")
				defer resp.Body.Close()

				// Should return OK, Not Found, or Bad Request
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest,
					"Reorder track should be handled")
			}
		}
	}
}

// TestPlaylists_UpdateImage tests updating playlist cover image
func TestPlaylists_UpdateImage(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get a playlist
	resp, err := client.Request("GET", "/playlists?limit=1", nil)
	require.NoError(t, err, "List playlists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if playlists, ok := body["playlists"].([]interface{}); ok && len(playlists) > 0 {
		if playlist, ok := playlists[0].(map[string]interface{}); ok {
			if playlistUUID, ok := playlist["uuid"].(string); ok {
				// Test the image upload endpoint (multipart form)
				// This will return an error for missing file but should not be 404
				resp, err = client.Request("POST", "/playlists/"+playlistUUID+"/image", nil)
				require.NoError(t, err, "Update playlist image request should not fail")
				defer resp.Body.Close()

				// Should not return 404
				assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
			}
		}
	}
}

// TestMusic_UpdateImage tests updating music cover art
func TestMusic_UpdateImage(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get a music track
	resp, err := client.Request("GET", "/music?limit=1", nil)
	require.NoError(t, err, "List music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if music, ok := body["music"].([]interface{}); ok && len(music) > 0 {
		if track, ok := music[0].(map[string]interface{}); ok {
			if musicUUID, ok := track["uuid"].(string); ok {
				// Test the image upload endpoint (multipart form)
				// This will return an error for missing file but should not be 404
				resp, err = client.Request("POST", "/music/"+musicUUID+"/image", nil)
				require.NoError(t, err, "Update music image request should not fail")
				defer resp.Body.Close()

				// Should not return 404
				assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
			}
		}
	}
}

// TestPlaylists_GetTracksWithOrdering tests getting tracks in playlist with specific ordering
func TestPlaylists_GetTracksWithOrdering(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get a playlist
	resp, err := client.Request("GET", "/playlists?limit=1", nil)
	require.NoError(t, err, "List playlists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if playlists, ok := body["playlists"].([]interface{}); ok && len(playlists) > 0 {
		if playlist, ok := playlists[0].(map[string]interface{}); ok {
			if playlistUUID, ok := playlist["uuid"].(string); ok {
				// Test with ordering parameter
				resp, err = client.Request("GET", "/playlists/"+playlistUUID+"/tracks?order=position&sort=asc", nil)
				require.NoError(t, err, "Get playlist tracks with ordering request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)
			}
		}
	}
}
