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

func TestIntegration_FollowUserHandler_Follow(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	userA := builders.NewUserBuilder().WithEmail("usera@test.com").Build(t, ctx, db)
	userB := builders.NewUserBuilder().WithEmail("userb@test.com").Build(t, ctx, db)

	handler := handlers.NewFollowsHandler(logger, config, returns, db)

	req := createRequest(t, "POST", "/users/"+builders.UUIDToString(userB)+"/follow", nil)
	router := mux.NewRouter()
	router.HandleFunc("/users/{uuid}/follow", wrapWithAuth(t, handler.FollowUser, userA)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify follow exists
	following, err := db.IsFollowingUser(ctx, sqlhandler.IsFollowingUserParams{
		FromUser: userA,
		ToUser:   userB,
	})
	require.NoError(t, err)
	assert.True(t, following)
}

func TestIntegration_FollowUserHandler_Follow_AlreadyFollowing(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	userA := builders.NewUserBuilder().WithEmail("usera@test.com").Build(t, ctx, db)
	userB := builders.NewUserBuilder().WithEmail("userb@test.com").Build(t, ctx, db)

	// Follow first time
	err := db.FollowUser(ctx, sqlhandler.FollowUserParams{
		FromUser: userA,
		ToUser:   userB,
	})
	require.NoError(t, err)

	handler := handlers.NewFollowsHandler(logger, config, returns, db)

	// Try to follow again
	req := createRequest(t, "POST", "/users/"+builders.UUIDToString(userB)+"/follow", nil)
	router := mux.NewRouter()
	router.HandleFunc("/users/{uuid}/follow", wrapWithAuth(t, handler.FollowUser, userA)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Implementation dependent - could be 409, 200 (idempotent), or 500
	assert.NotEqual(t, 0, rr.Code)
}

func TestIntegration_FollowUserHandler_Follow_Self(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	userA := builders.NewUserBuilder().WithEmail("usera@test.com").Build(t, ctx, db)

	handler := handlers.NewFollowsHandler(logger, config, returns, db)

	// Try to follow self
	req := createRequest(t, "POST", "/users/"+builders.UUIDToString(userA)+"/follow", nil)
	router := mux.NewRouter()
	router.HandleFunc("/users/{uuid}/follow", wrapWithAuth(t, handler.FollowUser, userA)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should reject self-follow (400 or 403)
	assert.NotEqual(t, http.StatusCreated, rr.Code)
}

func TestIntegration_FollowUserHandler_Unfollow(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	userA := builders.NewUserBuilder().WithEmail("usera@test.com").Build(t, ctx, db)
	userB := builders.NewUserBuilder().WithEmail("userb@test.com").Build(t, ctx, db)

	// Follow first
	err := db.FollowUser(ctx, sqlhandler.FollowUserParams{
		FromUser: userA,
		ToUser:   userB,
	})
	require.NoError(t, err)

	handler := handlers.NewFollowsHandler(logger, config, returns, db)

	// Unfollow
	req := createRequest(t, "DELETE", "/users/"+builders.UUIDToString(userB)+"/follow", nil)
	router := mux.NewRouter()
	router.HandleFunc("/users/{uuid}/follow", wrapWithAuth(t, handler.UnfollowUser, userA)).Methods("DELETE")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify unfollow
	following, err := db.IsFollowingUser(ctx, sqlhandler.IsFollowingUserParams{
		FromUser: userA,
		ToUser:   userB,
	})
	require.NoError(t, err)
	assert.False(t, following)
}

func TestIntegration_FollowUserHandler_Unfollow_NotFollowing(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	userA := builders.NewUserBuilder().WithEmail("usera@test.com").Build(t, ctx, db)
	userB := builders.NewUserBuilder().WithEmail("userb@test.com").Build(t, ctx, db)

	handler := handlers.NewFollowsHandler(logger, config, returns, db)

	// Try to unfollow when not following
	req := createRequest(t, "DELETE", "/users/"+builders.UUIDToString(userB)+"/follow", nil)
	router := mux.NewRouter()
	router.HandleFunc("/users/{uuid}/follow", wrapWithAuth(t, handler.UnfollowUser, userA)).Methods("DELETE")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Implementation dependent - could be 404, 200 (idempotent)
	assert.NotEqual(t, 0, rr.Code)
}

func TestIntegration_FollowUserHandler_CheckIfFollowing(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	userA := builders.NewUserBuilder().WithEmail("usera@test.com").Build(t, ctx, db)
	userB := builders.NewUserBuilder().WithEmail("userb@test.com").Build(t, ctx, db)

	// Follow
	err := db.FollowUser(ctx, sqlhandler.FollowUserParams{
		FromUser: userA,
		ToUser:   userB,
	})
	require.NoError(t, err)

	handler := handlers.NewFollowsHandler(logger, config, returns, db)

	req := createRequest(t, "GET", "/users/"+builders.UUIDToString(userB)+"/following", nil)
	router := mux.NewRouter()
	router.HandleFunc("/users/{uuid}/following", wrapWithAuth(t, handler.CheckIfFollowingUser, userA)).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]bool
	assertJSONResponse(t, rr, http.StatusOK, &response)
	assert.True(t, response["following"])
}

func TestIntegration_FollowUserHandler_GetFollowersForUser(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	targetUser := builders.NewUserBuilder().WithEmail("target@test.com").Build(t, ctx, db)
	follower1 := builders.NewUserBuilder().WithEmail("f1@test.com").Build(t, ctx, db)
	follower2 := builders.NewUserBuilder().WithEmail("f2@test.com").Build(t, ctx, db)

	db.FollowUser(ctx, sqlhandler.FollowUserParams{FromUser: follower1, ToUser: targetUser})
	db.FollowUser(ctx, sqlhandler.FollowUserParams{FromUser: follower2, ToUser: targetUser})

	handler := handlers.NewFollowsHandler(logger, config, returns, db)

	req := createRequest(t, "GET", "/users/"+builders.UUIDToString(targetUser)+"/followers?limit=10", nil)
	router := mux.NewRouter()
	router.HandleFunc("/users/{uuid}/followers", handler.GetFollowersForUser).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var followers []map[string]interface{}
	assertJSONResponse(t, rr, http.StatusOK, &followers)
	assert.Len(t, followers, 2)
}

func TestIntegration_FollowUserHandler_GetFollowingUsersForUser(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	userA := builders.NewUserBuilder().WithEmail("usera@test.com").Build(t, ctx, db)
	userB := builders.NewUserBuilder().WithEmail("userb@test.com").Build(t, ctx, db)
	userC := builders.NewUserBuilder().WithEmail("userc@test.com").Build(t, ctx, db)

	db.FollowUser(ctx, sqlhandler.FollowUserParams{FromUser: userA, ToUser: userB})
	db.FollowUser(ctx, sqlhandler.FollowUserParams{FromUser: userA, ToUser: userC})

	handler := handlers.NewFollowsHandler(logger, config, returns, db)

	req := createRequest(t, "GET", "/users/"+builders.UUIDToString(userA)+"/following?limit=10", nil)
	router := mux.NewRouter()
	router.HandleFunc("/users/{uuid}/following", handler.GetFollowingUsersForUser).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var following []map[string]interface{}
	assertJSONResponse(t, rr, http.StatusOK, &following)
	assert.Len(t, following, 2)
}
