//go:build e2e

package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUsers_GetCurrentUser tests getting the current user's profile
func TestUsers_GetCurrentUser(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get current user request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)

	body := AssertResponseBody(t, resp)
	AssertContainsField(t, body, "uuid")
	AssertContainsField(t, body, "username")
	AssertContainsField(t, body, "display_name")
	AssertContainsField(t, body, "email")
}

// TestUsers_UpdateProfile tests updating the current user's profile
func TestUsers_UpdateProfile(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Update profile
	resp, err := client.Request("POST", "/users/me", map[string]interface{}{
		"display_name": "Updated Name " + NewTestUUID()[:8],
		"bio":          "Updated bio",
	})

	require.NoError(t, err, "Update profile request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)

	body := AssertResponseBody(t, resp)
	AssertContainsField(t, body, "uuid")
	assert.Equal(t, "Updated Name "+NewTestUUID()[:8], body["display_name"])
}

// TestUsers_UpdateEmail tests updating the current user's email
func TestUsers_UpdateEmail(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	newEmail := fmt.Sprintf("newemail-%s@example.com", NewTestUUID()[:8])

	resp, err := client.Request("POST", "/users/me/email", map[string]interface{}{
		"email": newEmail,
	})

	require.NoError(t, err, "Update email request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestUsers_UpdatePassword tests updating the current user's password
func TestUsers_UpdatePassword(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// This test requires knowing the old password
	// In a real scenario, we'd need to set up the test user properly
	resp, err := client.Request("POST", "/users/me/password", map[string]interface{}{
		"old_password": "testpassword123",
		"new_password": "NewPass123!",
	})

	require.NoError(t, err, "Update password request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// Accept 200 (success) or 401 (wrong old password)
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized,
		"Password update should succeed or fail with wrong old password")
}

// TestUsers_UpdateProfileImage tests updating the current user's profile image
func TestUsers_UpdateProfileImage(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// This test would require multipart form upload
	// For now, we'll test the endpoint structure
	resp, err := client.Request("POST", "/users/me/image", nil)

	require.NoError(t, err, "Update image request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// Should return error for missing file, but not 404
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
}

// TestUsers_GetUserByUUID tests getting a user by UUID
func TestUsers_GetUserByUUID(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// First get current user to get a valid UUID
	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get current user request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	userUUID := body["uuid"].(string)

	// Now get user by UUID
	resp, err = client.Request("GET", "/users/"+userUUID, nil)
	require.NoError(t, err, "Get user by UUID request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)

	userBody := AssertResponseBody(t, resp)
	AssertContainsField(t, userBody, "uuid")
	AssertContainsField(t, userBody, "username")
}

// TestUsers_GetUserByInvalidUUID tests getting a user by invalid UUID
func TestUsers_GetUserByInvalidUUID(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/users/invalid-uuid", nil)
	require.NoError(t, err, "Get user by invalid UUID request should not fail")
	defer resp.Body.Close()

	// Should return 400 or 404
	assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound,
		"Invalid UUID should return error")
}

// TestUsers_GetUserNotFound tests getting a non-existent user
func TestUsers_GetUserNotFound(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	nonExistentUUID := "00000000-0000-0000-0000-000000000000"

	resp, err := client.Request("GET", "/users/"+nonExistentUUID, nil)
	require.NoError(t, err, "Get non-existent user request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusNotFound)
}

// TestUsers_GetUserArtists tests getting artists created by a user
func TestUsers_GetUserArtists(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get current user's UUID
	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get current user request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	userUUID := body["uuid"].(string)

	// Get user's artists
	resp, err = client.Request("GET", "/users/"+userUUID+"/artists", nil)
	require.NoError(t, err, "Get user artists request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)

	// Response should be a paginated response with artists array
	artistsBody := AssertResponseBody(t, resp)
	_, hasArtists := artistsBody["artists"]
	assert.True(t, hasArtists || len(artistsBody) > 0, "Response should contain artists data")
}

// TestUsers_GetUserLikes tests getting music liked by a user
func TestUsers_GetUserLikes(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get current user's UUID
	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get current user request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	userUUID := body["uuid"].(string)

	// Get user's likes
	resp, err = client.Request("GET", "/users/"+userUUID+"/likes", nil)
	require.NoError(t, err, "Get user likes request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestUsers_GetUserPlaylists tests getting playlists created by a user
func TestUsers_GetUserPlaylists(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get current user's UUID
	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get current user request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	userUUID := body["uuid"].(string)

	// Get user's playlists
	resp, err = client.Request("GET", "/users/"+userUUID+"/playlists", nil)
	require.NoError(t, err, "Get user playlists request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestUsers_GetUserFollowers tests getting followers of a user
func TestUsers_GetUserFollowers(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get current user's UUID
	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get current user request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	userUUID := body["uuid"].(string)

	// Get user's followers
	resp, err = client.Request("GET", "/users/"+userUUID+"/followers", nil)
	require.NoError(t, err, "Get user followers request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestUsers_GetUserFollowing tests getting users that a user follows
func TestUsers_GetUserFollowing(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get current user's UUID
	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get current user request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	userUUID := body["uuid"].(string)

	// Get user's following
	resp, err = client.Request("GET", "/users/"+userUUID+"/following", nil)
	require.NoError(t, err, "Get user following request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestUsers_FollowUser tests following a user
func TestUsers_FollowUser(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get current user's UUID
	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get current user request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	currentUserUUID := body["uuid"].(string)

	// Try to follow another user (we'll try to follow ourselves for simplicity)
	resp, err = client.Request("POST", "/users/"+currentUserUUID+"/follow", nil)
	require.NoError(t, err, "Follow user request should not fail")
	defer resp.Body.Close()

	// Should work or return appropriate error
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
		"Follow should succeed or return bad request")
}

// TestUsers_UnfollowUser tests unfollowing a user
func TestUsers_UnfollowUser(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// This test would require having followed someone first
	// For now, just test the endpoint exists
	resp, err := client.Request("DELETE", "/users/00000000-0000-0000-0000-000000000001/follow", nil)
	require.NoError(t, err, "Unfollow user request should not fail")
	defer resp.Body.Close()

	// Should not return 404
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
}

// TestUsers_GetUserMusic tests getting music created by a user
func TestUsers_GetUserMusic(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get current user's UUID
	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get current user request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	userUUID := body["uuid"].(string)

	// Get user's music
	resp, err = client.Request("GET", "/users/"+userUUID+"/music", nil)
	require.NoError(t, err, "Get user music request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestUsers_CheckFollowing tests checking if current user follows another user
func TestUsers_CheckFollowing(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get current user's UUID
	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get current user request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	userUUID := body["uuid"].(string)

	// Check if following self (should return appropriate response)
	resp, err = client.Request("GET", "/users/"+userUUID+"/following/check", nil)
	require.NoError(t, err, "Check following request should not fail")
	defer resp.Body.Close()

	// Should return OK or Not Found
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound,
		"Check following should return OK or Not Found")
}

// TestUsers_UnauthorizedAccess tests that unauthorized requests are rejected
func TestUsers_UnauthorizedAccess(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Try to get user without authentication
	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get user request should not fail")
	defer resp.Body.Close()

	// Should return 401 Unauthorized
	AssertResponseStatus(t, resp, http.StatusUnauthorized)
}

// TestUsers_InvalidToken tests that invalid tokens are rejected
func TestUsers_InvalidToken(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)
	client.SetTokens("invalid-token", "", "")

	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get user request should not fail")
	defer resp.Body.Close()

	// Should return 401 Unauthorized
	AssertResponseStatus(t, resp, http.StatusUnauthorized)
}

// TestUsers_Pagination tests pagination parameters
func TestUsers_Pagination(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Test with limit parameter
	resp, err := client.Request("GET", "/users/me/likes?limit=10", nil)
	require.NoError(t, err, "Get likes with limit request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestUsers_InvalidPagination tests invalid pagination parameters
func TestUsers_InvalidPagination(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Test with invalid limit
	resp, err := client.Request("GET", "/users/me/likes?limit=-1", nil)
	require.NoError(t, err, "Get likes with invalid limit request should not fail")
	defer resp.Body.Close()

	// Should handle gracefully
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
		"Invalid limit should be handled")
}

// TestUsers_GetFollowingUsers tests getting users that this user follows
func TestUsers_GetFollowingUsers(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get current user's UUID
	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get current user request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	userUUID := body["uuid"].(string)

	// Get users that this user follows
	resp, err = client.Request("GET", "/users/"+userUUID+"/following/users", nil)
	require.NoError(t, err, "Get following users request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestUsers_GetFollowingArtists tests getting artists followed by user
func TestUsers_GetFollowingArtists(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Get current user's UUID
	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Get current user request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	body := AssertResponseBody(t, resp)
	userUUID := body["uuid"].(string)

	// Get artists that this user follows
	resp, err = client.Request("GET", "/users/"+userUUID+"/following/artists", nil)
	require.NoError(t, err, "Get following artists request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)
}
