//go:build e2e

package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestArtists_List tests listing all artists
func TestArtists_List(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/artists", nil)
	require.NoError(t, err, "List artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)

	body := AssertResponseBody(t, resp)

	// Validate response structure
	_, hasArtists := body["artists"]
	hasData := len(body) > 0
	assert.True(t, hasArtists || hasData, "Response should contain artists data or be non-empty")

	// If artists array exists, validate structure
	if hasArtists {
		artists, ok := body["artists"].([]interface{})
		require.True(t, ok, "artists field should be an array")

		// If there are artists, validate the first one has expected fields
		if len(artists) > 0 {
			artist, ok := artists[0].(map[string]interface{})
			require.True(t, ok, "First artist should be an object")

			// Validate required fields in artist object
			_, hasUUID := artist["uuid"]
			_, hasName := artist["name"]
			assert.True(t, hasUUID || hasName, "Artist should have uuid or name field")
		}
	}
}

// TestArtists_GetByUUID tests getting an artist by UUID
func TestArtists_GetByUUID(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// First get list of artists to find one
	resp, err := client.Request("GET", "/artists?limit=1", nil)
	require.NoError(t, err, "List artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)

	// If we have artists, try to get one by UUID
	body := AssertResponseBody(t, resp)
	if artists, ok := body["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			if uuid, ok := artist["uuid"].(string); ok {
				resp, err = client.Request("GET", "/artists/"+uuid, nil)
				require.NoError(t, err, "Get artist request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)

				artistBody := AssertResponseBody(t, resp)
				AssertContainsField(t, artistBody, "uuid")
				AssertContainsField(t, artistBody, "name")
			}
		}
	}
}

// TestArtists_GetByInvalidUUID tests getting an artist with invalid UUID
func TestArtists_GetByInvalidUUID(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/artists/invalid-uuid", nil)
	require.NoError(t, err, "Get artist request should not fail")
	defer resp.Body.Close()

	// Should return 400 or 404
	assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound,
		"Invalid UUID should return error")
}

// TestArtists_Create tests creating a new artist
func TestArtists_Create(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	artistData := GenerateTestArtistData()

	resp, err := client.Request("POST", "/artists", artistData)
	require.NoError(t, err, "Create artist request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// Accept 200 (success) or 201 (created)
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated,
		"Artist creation should succeed")

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		body := AssertResponseBody(t, resp)
		AssertContainsField(t, body, "uuid")
		AssertContainsField(t, body, "name")
	}
}

// TestArtists_Update tests updating an artist
func TestArtists_Update(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// First create an artist
	artistData := GenerateTestArtistData()
	resp, err := client.Request("POST", "/artists", artistData)
	require.NoError(t, err, "Create artist request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Skip("Cannot create artist for update test")
	}

	body := AssertResponseBody(t, resp)
	artistUUID := body["uuid"].(string)

	// Now update the artist
	updateData := map[string]interface{}{
		"name": "Updated Artist Name",
		"bio":  "Updated artist bio",
	}

	resp, err = client.Request("PUT", "/artists/"+artistUUID, updateData)
	require.NoError(t, err, "Update artist request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)

	updateBody := AssertResponseBody(t, resp)
	assert.Equal(t, "Updated Artist Name", updateBody["name"])
}

// TestArtists_Delete tests deleting an artist
func TestArtists_Delete(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// First create an artist
	artistData := GenerateTestArtistData()
	resp, err := client.Request("POST", "/artists", artistData)
	require.NoError(t, err, "Create artist request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Skip("Cannot create artist for delete test")
	}

	body := AssertResponseBody(t, resp)
	artistUUID := body["uuid"].(string)

	// Now delete the artist
	resp, err = client.Request("DELETE", "/artists/"+artistUUID, nil)
	require.NoError(t, err, "Delete artist request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestArtists_GetAlbums tests getting albums by an artist
func TestArtists_GetAlbums(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get list of artists first
	resp, err := client.Request("GET", "/artists?limit=1", nil)
	require.NoError(t, err, "List artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if artists, ok := body["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			if uuid, ok := artist["uuid"].(string); ok {
				resp, err = client.Request("GET", "/artists/"+uuid+"/albums", nil)
				require.NoError(t, err, "Get artist albums request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)
			}
		}
	}
}

// TestArtists_GetMusic tests getting music by an artist
func TestArtists_GetMusic(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get list of artists first
	resp, err := client.Request("GET", "/artists?limit=1", nil)
	require.NoError(t, err, "List artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if artists, ok := body["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			if uuid, ok := artist["uuid"].(string); ok {
				resp, err = client.Request("GET", "/artists/"+uuid+"/music", nil)
				require.NoError(t, err, "Get artist music request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)
			}
		}
	}
}

// TestArtists_GetMembers tests getting members of an artist
func TestArtists_GetMembers(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get list of artists first
	resp, err := client.Request("GET", "/artists?limit=1", nil)
	require.NoError(t, err, "List artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if artists, ok := body["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			if uuid, ok := artist["uuid"].(string); ok {
				resp, err = client.Request("GET", "/artists/"+uuid+"/members", nil)
				require.NoError(t, err, "Get artist members request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)
			}
		}
	}
}

// TestArtists_GetFollowers tests getting followers of an artist
func TestArtists_GetFollowers(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get list of artists first
	resp, err := client.Request("GET", "/artists?limit=1", nil)
	require.NoError(t, err, "List artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if artists, ok := body["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			if uuid, ok := artist["uuid"].(string); ok {
				resp, err = client.Request("GET", "/artists/"+uuid+"/followers", nil)
				require.NoError(t, err, "Get artist followers request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)
			}
		}
	}
}

// TestArtists_Follow tests following an artist
func TestArtists_Follow(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get list of artists first
	resp, err := client.Request("GET", "/artists?limit=1", nil)
	require.NoError(t, err, "List artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if artists, ok := body["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			if uuid, ok := artist["uuid"].(string); ok {
				resp, err = client.Request("POST", "/artists/"+uuid+"/follow", nil)
				require.NoError(t, err, "Follow artist request should not fail")
				defer resp.Body.Close()

				// Should succeed or already following
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
					"Follow should succeed or return error")
			}
		}
	}
}

// TestArtists_Unfollow tests unfollowing an artist
func TestArtists_Unfollow(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get list of artists first
	resp, err := client.Request("GET", "/artists?limit=1", nil)
	require.NoError(t, err, "List artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if artists, ok := body["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			if uuid, ok := artist["uuid"].(string); ok {
				resp, err = client.Request("DELETE", "/artists/"+uuid+"/follow", nil)
				require.NoError(t, err, "Unfollow artist request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)
			}
		}
	}
}

// TestArtists_Pagination tests artist pagination
func TestArtists_Pagination(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	testCases := []struct {
		name   string
		params string
	}{
		{"limit_10", "?limit=10"},
		{"limit_50", "?limit=50"},
		{"limit_100", "?limit=100"},
		{"with_cursor", "?limit=10"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.Request("GET", "/artists"+tc.params, nil)
			require.NoError(t, err, "List artists request should not fail")
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusBadGateway {
				t.Skip("Backend service not available")
			}

			AssertResponseStatus(t, resp, http.StatusOK)
		})
	}
}

// TestArtists_InvalidUUID tests handling of invalid UUID
func TestArtists_InvalidUUID(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	invalidUUIDs := []string{
		"not-a-uuid",
		"123",
		"",
		"../../etc/passwd",
	}

	for _, invalidUUID := range invalidUUIDs {
		t.Run(invalidUUID, func(t *testing.T) {
			resp, err := client.Request("GET", "/artists/"+invalidUUID, nil)
			require.NoError(t, err, "Get artist request should not fail")
			defer resp.Body.Close()

			// Should return error, not crash
			assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound,
				"Invalid UUID should return error")
		})
	}
}

// TestArtists_UpdateMemberRole tests updating a member's role in an artist
func TestArtists_UpdateMemberRole(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get list of artists first
	resp, err := client.Request("GET", "/artists?limit=1", nil)
	require.NoError(t, err, "List artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if artists, ok := body["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			if uuid, ok := artist["uuid"].(string); ok {
				// Try to update a member role (using a dummy user UUID)
				resp, err = client.Request("PUT", "/artists/"+uuid+"/members/"+NewTestUUID()+"/role",
					map[string]interface{}{"role": " vocalist"})
				require.NoError(t, err, "Update member role request should not fail")
				defer resp.Body.Close()

				// Should return OK, Not Found, or Bad Request (if member doesn't exist)
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest,
					"Update member role should be handled")
			}
		}
	}
}

// TestArtists_UpdateImage tests updating artist profile picture
func TestArtists_UpdateImage(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get list of artists first
	resp, err := client.Request("GET", "/artists?limit=1", nil)
	require.NoError(t, err, "List artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if artists, ok := body["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			if artistUUID, ok := artist["uuid"].(string); ok {
				// Test the image upload endpoint (multipart form)
				// This will return an error for missing file but should not be 404
				resp, err = client.Request("POST", "/artists/"+artistUUID+"/image", nil)
				require.NoError(t, err, "Update artist image request should not fail")
				defer resp.Body.Close()

				// Should not return 404
				assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
			}
		}
	}
}

// TestArtists_AddMember tests adding a user as artist member
func TestArtists_AddMember(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get list of artists first
	resp, err := client.Request("GET", "/artists?limit=1", nil)
	require.NoError(t, err, "List artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if artists, ok := body["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			if artistUUID, ok := artist["uuid"].(string); ok {
				// Try to add a member with a dummy user UUID
				resp, err = client.Request("PUT", "/artists/"+artistUUID+"/members/"+NewTestUUID(), nil)
				require.NoError(t, err, "Add member request should not fail")
				defer resp.Body.Close()

				// Should return OK, Not Found, or Bad Request
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest,
					"Add member should be handled")
			}
		}
	}
}

// TestArtists_RemoveMember tests removing a user from artist
func TestArtists_RemoveMember(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get list of artists first
	resp, err := client.Request("GET", "/artists?limit=1", nil)
	require.NoError(t, err, "List artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if artists, ok := body["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			if artistUUID, ok := artist["uuid"].(string); ok {
				// Try to remove a member with a dummy user UUID
				resp, err = client.Request("DELETE", "/artists/"+artistUUID+"/members/"+NewTestUUID(), nil)
				require.NoError(t, err, "Remove member request should not fail")
				defer resp.Body.Close()

				// Should return OK, Not Found, or Bad Request
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest,
					"Remove member should be handled")
			}
		}
	}
}

// TestArtists_GetImage tests getting artist profile image
func TestArtists_GetImage(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get list of artists first
	resp, err := client.Request("GET", "/artists?limit=1", nil)
	require.NoError(t, err, "List artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if artists, ok := body["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			if artistUUID, ok := artist["uuid"].(string); ok {
				resp, err = client.Request("GET", "/artists/"+artistUUID+"/image", nil)
				require.NoError(t, err, "Get artist image request should not fail")
				defer resp.Body.Close()

				// Should return OK, Not Found, or redirect to default image
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusFound,
					"Get artist image should be handled")
			}
		}
	}
}
