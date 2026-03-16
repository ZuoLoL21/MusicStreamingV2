//go:build e2e

package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAlbums_List tests listing all albums
func TestAlbums_List(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/albums", nil)
	require.NoError(t, err, "List albums request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestAlbums_GetByUUID tests getting an album by UUID
func TestAlbums_GetByUUID(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// First get list of albums to find one
	resp, err := client.Request("GET", "/albums?limit=1", nil)
	require.NoError(t, err, "List albums request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)

	// If we have albums, try to get one by UUID
	body := AssertResponseBody(t, resp)
	if albums, ok := body["albums"].([]interface{}); ok && len(albums) > 0 {
		if album, ok := albums[0].(map[string]interface{}); ok {
			if uuid, ok := album["uuid"].(string); ok {
				resp, err = client.Request("GET", "/albums/"+uuid, nil)
				require.NoError(t, err, "Get album request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)

				albumBody := AssertResponseBody(t, resp)
				AssertContainsField(t, albumBody, "uuid")
				AssertContainsField(t, albumBody, "name")
			}
		}
	}
}

// TestAlbums_Create tests creating a new album
func TestAlbums_Create(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	albumData := GenerateTestAlbumData()

	resp, err := client.Request("POST", "/albums", albumData)
	require.NoError(t, err, "Create album request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// Accept 200 or 201
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated,
		"Album creation should succeed")

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		body := AssertResponseBody(t, resp)
		AssertContainsField(t, body, "uuid")
		AssertContainsField(t, body, "name")
	}
}

// TestAlbums_Update tests updating an album
func TestAlbums_Update(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// First create an album
	albumData := GenerateTestAlbumData()
	resp, err := client.Request("POST", "/albums", albumData)
	require.NoError(t, err, "Create album request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Skip("Cannot create album for update test")
	}

	body := AssertResponseBody(t, resp)
	albumUUID := body["uuid"].(string)

	// Now update the album
	updateData := map[string]interface{}{
		"name":        "Updated Album Name",
		"description": "Updated album description",
	}

	resp, err = client.Request("PUT", "/albums/"+albumUUID, updateData)
	require.NoError(t, err, "Update album request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestAlbums_Delete tests deleting an album
func TestAlbums_Delete(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// First create an album
	albumData := GenerateTestAlbumData()
	resp, err := client.Request("POST", "/albums", albumData)
	require.NoError(t, err, "Create album request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Skip("Cannot create album for delete test")
	}

	body := AssertResponseBody(t, resp)
	albumUUID := body["uuid"].(string)

	// Now delete the album
	resp, err = client.Request("DELETE", "/albums/"+albumUUID, nil)
	require.NoError(t, err, "Delete album request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestAlbums_GetMusic tests getting music tracks in an album
func TestAlbums_GetMusic(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get list of albums first
	resp, err := client.Request("GET", "/albums?limit=1", nil)
	require.NoError(t, err, "List albums request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	if albums, ok := body["albums"].([]interface{}); ok && len(albums) > 0 {
		if album, ok := albums[0].(map[string]interface{}); ok {
			if uuid, ok := album["uuid"].(string); ok {
				resp, err = client.Request("GET", "/albums/"+uuid+"/music", nil)
				require.NoError(t, err, "Get album music request should not fail")
				defer resp.Body.Close()

				AssertResponseStatus(t, resp, http.StatusOK)
			}
		}
	}
}

// TestAlbums_Pagination tests album pagination
func TestAlbums_Pagination(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	testCases := []struct {
		name   string
		params string
	}{
		{"limit_10", "?limit=10"},
		{"limit_50", "?limit=50"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.Request("GET", "/albums"+tc.params, nil)
			require.NoError(t, err, "List albums request should not fail")
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusBadGateway {
				t.Skip("Backend service not available")
			}

			AssertResponseStatus(t, resp, http.StatusOK)
		})
	}
}

// TestAlbums_InvalidUUID tests handling of invalid album UUID
func TestAlbums_InvalidUUID(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/albums/invalid-uuid", nil)
	require.NoError(t, err, "Get album request should not fail")
	defer resp.Body.Close()

	assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound,
		"Invalid UUID should return error")
}
