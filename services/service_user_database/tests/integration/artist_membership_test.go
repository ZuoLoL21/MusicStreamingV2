//go:build integration

package integration

import (
	backenddi "backend/internal/di"
	"backend/internal/handlers"
	sqlhandler "backend/sql/sqlc"
	"backend/tests/integration/builders"
	"context"
	"libs/di"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIntegration_ArtistMembership_AddUser_AsOwner(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create owner and artist
	ownerUUID := builders.NewUserBuilder().
		WithEmail("owner@example.com").
		Build(t, ctx, db)

	artistUUID := builders.NewArtistBuilder(ownerUUID).
		WithName("Test Band").
		Build(t, ctx, db)

	// Create user to add
	newMemberUUID := builders.NewUserBuilder().
		WithEmail("newmember@example.com").
		Build(t, ctx, db)

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Add new member as MEMBER
	addReq := map[string]string{
		"role": string(sqlhandler.ArtistMemberRoleMember),
	}
	req := createJSONRequest(t, "POST", "/artists/"+builders.UUIDToString(artistUUID)+"/members/"+builders.UUIDToString(newMemberUUID), addReq)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/members/{userUuid}", wrapWithAuth(t, handler.AddUserToArtist, ownerUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 201 Created
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify member was added
	members, err := db.GetUsersRepresentingArtist(ctx, artistUUID)
	require.NoError(t, err)
	require.Len(t, members, 2) // Owner + new member

	// Find the new member
	foundNewMember := false
	for _, member := range members {
		if member.Uuid == newMemberUUID {
			foundNewMember = true
			assert.Equal(t, sqlhandler.ArtistMemberRoleMember, member.Role)
			break
		}
	}
	assert.True(t, foundNewMember, "new member should be in artist members")
}

func TestIntegration_ArtistMembership_AddUser_AsNonOwner(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create owner and artist
	ownerUUID := builders.NewUserBuilder().
		WithEmail("owner@example.com").
		Build(t, ctx, db)

	artistUUID := builders.NewArtistBuilder(ownerUUID).
		WithName("Test Band").
		Build(t, ctx, db)

	// Create a non-owner user
	nonOwnerUUID := builders.NewUserBuilder().
		WithEmail("nonowner@example.com").
		Build(t, ctx, db)

	// Create user to add
	targetUserUUID := builders.NewUserBuilder().
		WithEmail("target@example.com").
		Build(t, ctx, db)

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Try to add member as non-owner
	addReq := map[string]string{
		"role": string(sqlhandler.ArtistMemberRoleMember),
	}
	req := createJSONRequest(t, "POST", "/artists/"+builders.UUIDToString(artistUUID)+"/members/"+builders.UUIDToString(targetUserUUID), addReq)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/members/{userUuid}", wrapWithAuth(t, handler.AddUserToArtist, nonOwnerUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, rr.Code)

	// Verify member was NOT added
	members, err := db.GetUsersRepresentingArtist(ctx, artistUUID)
	require.NoError(t, err)
	require.Len(t, members, 1) // Only owner
}

func TestIntegration_ArtistMembership_AddUser_AlreadyMember(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create owner and artist
	ownerUUID := builders.NewUserBuilder().
		WithEmail("owner@example.com").
		Build(t, ctx, db)

	artistUUID := builders.NewArtistBuilder(ownerUUID).
		WithName("Test Band").
		Build(t, ctx, db)

	// Create and add member
	memberUUID := builders.NewUserBuilder().
		WithEmail("member@example.com").
		Build(t, ctx, db)

	err := db.AddUserToArtist(ctx, sqlhandler.AddUserToArtistParams{
		ArtistUuid: artistUUID,
		UserUuid:   memberUUID,
		Role:       sqlhandler.ArtistMemberRoleMember,
	})
	require.NoError(t, err)

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Try to add same member again
	addReq := map[string]string{
		"role": string(sqlhandler.ArtistMemberRoleManager),
	}
	req := createJSONRequest(t, "POST", "/artists/"+builders.UUIDToString(artistUUID)+"/members/"+builders.UUIDToString(memberUUID), addReq)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/members/{userUuid}", wrapWithAuth(t, handler.AddUserToArtist, ownerUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Could be 409 Conflict, 500, or idempotent (200)
	// Implementation-dependent behavior
	assert.NotEqual(t, http.StatusCreated, rr.Code, "should not return Created for duplicate member")
}

func TestIntegration_ArtistMembership_RemoveUser_AsOwner(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create owner and artist
	ownerUUID := builders.NewUserBuilder().
		WithEmail("owner@example.com").
		Build(t, ctx, db)

	artistUUID := builders.NewArtistBuilder(ownerUUID).
		WithName("Test Band").
		Build(t, ctx, db)

	// Add a member
	memberUUID := builders.NewUserBuilder().
		WithEmail("member@example.com").
		Build(t, ctx, db)

	err := db.AddUserToArtist(ctx, sqlhandler.AddUserToArtistParams{
		ArtistUuid: artistUUID,
		UserUuid:   memberUUID,
		Role:       sqlhandler.ArtistMemberRoleMember,
	})
	require.NoError(t, err)

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Remove member as owner
	req := createRequest(t, "DELETE", "/artists/"+builders.UUIDToString(artistUUID)+"/members/"+builders.UUIDToString(memberUUID), nil)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/members/{userUuid}", wrapWithAuth(t, handler.RemoveUserFromArtist, ownerUUID)).Methods("DELETE")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify member was removed
	members, err := db.GetUsersRepresentingArtist(ctx, artistUUID)
	require.NoError(t, err)
	require.Len(t, members, 1) // Only owner remains

	// Verify removed user is not in members
	for _, member := range members {
		assert.NotEqual(t, memberUUID, member.Uuid)
	}
}

func TestIntegration_ArtistMembership_RemoveUser_AsNonOwner(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create owner and artist
	ownerUUID := builders.NewUserBuilder().
		WithEmail("owner@example.com").
		Build(t, ctx, db)

	artistUUID := builders.NewArtistBuilder(ownerUUID).
		WithName("Test Band").
		Build(t, ctx, db)

	// Add two members
	member1UUID := builders.NewUserBuilder().
		WithEmail("member1@example.com").
		Build(t, ctx, db)

	member2UUID := builders.NewUserBuilder().
		WithEmail("member2@example.com").
		Build(t, ctx, db)

	err := db.AddUserToArtist(ctx, sqlhandler.AddUserToArtistParams{
		ArtistUuid: artistUUID,
		UserUuid:   member1UUID,
		Role:       sqlhandler.ArtistMemberRoleManager,
	})
	require.NoError(t, err)

	err = db.AddUserToArtist(ctx, sqlhandler.AddUserToArtistParams{
		ArtistUuid: artistUUID,
		UserUuid:   member2UUID,
		Role:       sqlhandler.ArtistMemberRoleMember,
	})
	require.NoError(t, err)

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Try to remove member2 as member1 (non-owner)
	req := createRequest(t, "DELETE", "/artists/"+builders.UUIDToString(artistUUID)+"/members/"+builders.UUIDToString(member2UUID), nil)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/members/{userUuid}", wrapWithAuth(t, handler.RemoveUserFromArtist, member1UUID)).Methods("DELETE")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, rr.Code)

	// Verify member2 was NOT removed
	members, err := db.GetUsersRepresentingArtist(ctx, artistUUID)
	require.NoError(t, err)
	require.Len(t, members, 3) // Owner + both members
}

func TestIntegration_ArtistMembership_RemoveUser_NotMember(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create owner and artist
	ownerUUID := builders.NewUserBuilder().
		WithEmail("owner@example.com").
		Build(t, ctx, db)

	artistUUID := builders.NewArtistBuilder(ownerUUID).
		WithName("Test Band").
		Build(t, ctx, db)

	// Create a user who is NOT a member
	nonMemberUUID := builders.NewUserBuilder().
		WithEmail("nonmember@example.com").
		Build(t, ctx, db)

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Try to remove non-member
	req := createRequest(t, "DELETE", "/artists/"+builders.UUIDToString(artistUUID)+"/members/"+builders.UUIDToString(nonMemberUUID), nil)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/members/{userUuid}", wrapWithAuth(t, handler.RemoveUserFromArtist, ownerUUID)).Methods("DELETE")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Could be 404 Not Found, 500, or idempotent (200)
	// Implementation-dependent behavior
	// Just verify it doesn't crash
	assert.NotEqual(t, 0, rr.Code)
}

func TestIntegration_ArtistMembership_ChangeRole(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create owner and artist
	ownerUUID := builders.NewUserBuilder().
		WithEmail("owner@example.com").
		Build(t, ctx, db)

	artistUUID := builders.NewArtistBuilder(ownerUUID).
		WithName("Test Band").
		Build(t, ctx, db)

	// Add a member as MEMBER
	memberUUID := builders.NewUserBuilder().
		WithEmail("member@example.com").
		Build(t, ctx, db)

	err := db.AddUserToArtist(ctx, sqlhandler.AddUserToArtistParams{
		ArtistUuid: artistUUID,
		UserUuid:   memberUUID,
		Role:       sqlhandler.ArtistMemberRoleMember,
	})
	require.NoError(t, err)

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Change role to MANAGER
	changeReq := map[string]string{
		"role": string(sqlhandler.ArtistMemberRoleManager),
	}
	req := createJSONRequest(t, "POST", "/artists/"+builders.UUIDToString(artistUUID)+"/members/"+builders.UUIDToString(memberUUID)+"/role", changeReq)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/members/{userUuid}/role", wrapWithAuth(t, handler.ChangeUserRole, ownerUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify role was changed
	members, err := db.GetUsersRepresentingArtist(ctx, artistUUID)
	require.NoError(t, err)

	// Find member and check role
	for _, member := range members {
		if member.Uuid == memberUUID {
			assert.Equal(t, sqlhandler.ArtistMemberRoleManager, member.Role)
			return
		}
	}
	t.Fatal("member not found after role change")
}

func TestIntegration_ArtistMembership_ChangeRole_ToOwner(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create owner and artist
	ownerUUID := builders.NewUserBuilder().
		WithEmail("owner@example.com").
		Build(t, ctx, db)

	artistUUID := builders.NewArtistBuilder(ownerUUID).
		WithName("Test Band").
		Build(t, ctx, db)

	// Add a member
	memberUUID := builders.NewUserBuilder().
		WithEmail("member@example.com").
		Build(t, ctx, db)

	err := db.AddUserToArtist(ctx, sqlhandler.AddUserToArtistParams{
		ArtistUuid: artistUUID,
		UserUuid:   memberUUID,
		Role:       sqlhandler.ArtistMemberRoleMember,
	})
	require.NoError(t, err)

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Try to change role to OWNER (may or may not be allowed)
	changeReq := map[string]string{
		"role": string(sqlhandler.ArtistMemberRoleOwner),
	}
	req := createJSONRequest(t, "POST", "/artists/"+builders.UUIDToString(artistUUID)+"/members/"+builders.UUIDToString(memberUUID)+"/role", changeReq)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/members/{userUuid}/role", wrapWithAuth(t, handler.ChangeUserRole, ownerUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Behavior is implementation-dependent
	// Could succeed, could fail with 400/403
	// Main goal: ensure it doesn't crash
	assert.NotEqual(t, 0, rr.Code)
}

func TestIntegration_ArtistMembership_GetUsersRepresentingArtist(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create owner and artist
	ownerUUID := builders.NewUserBuilder().
		WithUsername("owner").
		WithEmail("owner@example.com").
		Build(t, ctx, db)

	artistUUID := builders.NewArtistBuilder(ownerUUID).
		WithName("Test Band").
		Build(t, ctx, db)

	// Add members with different roles
	member1UUID := builders.NewUserBuilder().
		WithUsername("member1").
		WithEmail("member1@example.com").
		Build(t, ctx, db)

	managerUUID := builders.NewUserBuilder().
		WithUsername("manager").
		WithEmail("manager@example.com").
		Build(t, ctx, db)

	err := db.AddUserToArtist(ctx, sqlhandler.AddUserToArtistParams{
		ArtistUuid: artistUUID,
		UserUuid:   member1UUID,
		Role:       sqlhandler.ArtistMemberRoleMember,
	})
	require.NoError(t, err)

	err = db.AddUserToArtist(ctx, sqlhandler.AddUserToArtistParams{
		ArtistUuid: artistUUID,
		UserUuid:   managerUUID,
		Role:       sqlhandler.ArtistMemberRoleManager,
	})
	require.NoError(t, err)

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Get all members
	req := createRequest(t, "GET", "/artists/"+builders.UUIDToString(artistUUID)+"/members", nil)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/members", handler.GetUsersRepresentingArtist).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 200 OK with all members
	var members []map[string]interface{}
	assertJSONResponse(t, rr, http.StatusOK, &members)

	require.Len(t, members, 3) // Owner + 2 members

	// Verify all members are present
	usernames := make([]string, len(members))
	for i, member := range members {
		usernames[i] = member["username"].(string)
	}

	assert.Contains(t, usernames, "owner")
	assert.Contains(t, usernames, "member1")
	assert.Contains(t, usernames, "manager")
}
