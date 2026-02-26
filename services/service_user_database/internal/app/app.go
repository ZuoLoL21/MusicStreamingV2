package app

import (
	"backend/internal/di"
	"backend/internal/handlers"
	"backend/internal/storage"
	sqlhandler "backend/sql/sqlc"
	libsdi "libs/di"
	libshelpers "libs/helpers"
	libsmiddleware "libs/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	logger      *zap.Logger
	config      *di.Config
	secrets     *libsdi.SecretsManager
	returns     *libsdi.ReturnManager
	db          *sqlhandler.Queries
	fileStorage storage.FileStorageClient
}

func New(logger *zap.Logger, config *di.Config, secrets *libsdi.SecretsManager, returns *libsdi.ReturnManager, db *sqlhandler.Queries, fileStorage storage.FileStorageClient) *App {
	return &App{
		logger:      logger,
		config:      config,
		secrets:     secrets,
		returns:     returns,
		db:          db,
		fileStorage: fileStorage,
	}
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()

	// Create service JWT auth handler
	serviceAuthHandler := libsmiddleware.NewAuthHandler(
		a.logger,
		a.config,
		a.secrets,
		a.returns,
		libshelpers.JWTSubjectService,
	)
	userH := handlers.NewUserHandler(a.logger, a.config, a.secrets, a.returns, a.db, a.fileStorage)
	artistH := handlers.NewArtistHandler(a.logger, a.config, a.returns, a.db, a.fileStorage)
	albumH := handlers.NewAlbumHandler(a.logger, a.config, a.returns, a.db, a.fileStorage)
	musicH := handlers.NewMusicHandler(a.logger, a.config, a.returns, a.db, a.fileStorage)
	likesH := handlers.NewLikesHandler(a.logger, a.config, a.returns, a.db)
	followsH := handlers.NewFollowsHandler(a.logger, a.config, a.returns, a.db)
	tagsH := handlers.NewTagsHandler(a.logger, a.config, a.returns, a.db)
	playlistH := handlers.NewPlaylistHandler(a.logger, a.config, a.returns, a.db, a.fileStorage)
	historyH := handlers.NewHistoryHandler(a.logger, a.config, a.returns, a.db)

	protectedRouter := r.PathPrefix("").Subrouter()
	r.Use(
		libsmiddleware.RequestIDMiddleware(a.config),
		libsmiddleware.LoggingMiddleware(a.logger, a.config),
	)
	protectedRouter.Use(
		serviceAuthHandler.GetAuthMiddleware(),
	)

	// Health
	r.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// Auth
	r.HandleFunc("/login", userH.Login).Methods("POST")
	r.HandleFunc("/login", userH.Register).Methods("PUT")
	r.HandleFunc("/renew", userH.Renew).Methods("POST")

	// Users — static /me routes BEFORE /{uuid}
	protectedRouter.HandleFunc("/users/me", userH.GetMe).Methods("GET")
	protectedRouter.HandleFunc("/users/me", userH.UpdateProfile).Methods("POST")
	protectedRouter.HandleFunc("/users/me/email", userH.UpdateEmail).Methods("POST")
	protectedRouter.HandleFunc("/users/me/password", userH.UpdatePassword).Methods("POST")
	protectedRouter.HandleFunc("/users/me/image", userH.UpdateImage).Methods("POST")
	protectedRouter.HandleFunc("/users/{uuid}", userH.GetPublicUser).Methods("GET")
	protectedRouter.HandleFunc("/users/{uuid}/artists", userH.GetArtistForUser).Methods("GET")
	protectedRouter.HandleFunc("/users/{uuid}/likes", likesH.GetLikesForUser).Methods("GET")
	protectedRouter.HandleFunc("/users/{uuid}/followers", followsH.GetFollowersForUser).Methods("GET")
	protectedRouter.HandleFunc("/users/{uuid}/following/users", followsH.GetFollowingUsersForUser).Methods("GET")
	protectedRouter.HandleFunc("/users/{uuid}/following/artists", followsH.GetFollowedArtistsForUser).Methods("GET")
	protectedRouter.HandleFunc("/users/{uuid}/follow", followsH.FollowUser).Methods("POST")
	protectedRouter.HandleFunc("/users/{uuid}/follow", followsH.UnfollowUser).Methods("DELETE")
	protectedRouter.HandleFunc("/users/{uuid}/music", musicH.GetMusicForUser).Methods("GET")
	protectedRouter.HandleFunc("/users/{uuid}/playlists", playlistH.GetPlaylistsForUser).Methods("GET")

	// Artists
	protectedRouter.HandleFunc("/artists", artistH.GetArtistsAlphabetically).Methods("GET")
	protectedRouter.HandleFunc("/artists", artistH.CreateArtist).Methods("PUT")
	protectedRouter.HandleFunc("/artists/{uuid}", artistH.GetArtist).Methods("GET")
	protectedRouter.HandleFunc("/artists/{uuid}", artistH.UpdateArtistProfile).Methods("POST")
	protectedRouter.HandleFunc("/artists/{uuid}/image", artistH.UpdateArtistPicture).Methods("POST")
	protectedRouter.HandleFunc("/artists/{uuid}/members", artistH.GetUsersRepresentingArtist).Methods("GET")
	protectedRouter.HandleFunc("/artists/{uuid}/members/{userUuid}", artistH.AddUserToArtist).Methods("PUT")
	protectedRouter.HandleFunc("/artists/{uuid}/members/{userUuid}", artistH.RemoveUserFromArtist).Methods("DELETE")
	protectedRouter.HandleFunc("/artists/{uuid}/members/{userUuid}/role", artistH.ChangeUserRole).Methods("POST")
	protectedRouter.HandleFunc("/artists/{uuid}/albums", albumH.GetAlbumsForArtist).Methods("GET")
	protectedRouter.HandleFunc("/artists/{uuid}/music", musicH.GetMusicForArtist).Methods("GET")
	protectedRouter.HandleFunc("/artists/{uuid}/followers", followsH.GetFollowersForArtist).Methods("GET")
	protectedRouter.HandleFunc("/artists/{uuid}/follow", followsH.FollowArtist).Methods("POST")
	protectedRouter.HandleFunc("/artists/{uuid}/follow", followsH.UnfollowArtist).Methods("DELETE")

	// Albums
	protectedRouter.HandleFunc("/albums", albumH.CreateAlbum).Methods("PUT")
	protectedRouter.HandleFunc("/albums/{uuid}", albumH.GetAlbum).Methods("GET")
	protectedRouter.HandleFunc("/albums/{uuid}", albumH.UpdateAlbum).Methods("POST")
	protectedRouter.HandleFunc("/albums/{uuid}/image", albumH.UpdateAlbumImage).Methods("POST")
	protectedRouter.HandleFunc("/albums/{uuid}", albumH.DeleteAlbum).Methods("DELETE")
	protectedRouter.HandleFunc("/albums/{uuid}/music", musicH.GetMusicForAlbum).Methods("GET")

	// Music
	protectedRouter.HandleFunc("/music", musicH.CreateMusic).Methods("PUT")
	protectedRouter.HandleFunc("/music/{uuid}", musicH.GetMusic).Methods("GET")
	protectedRouter.HandleFunc("/music/{uuid}", musicH.UpdateMusicDetails).Methods("POST")
	protectedRouter.HandleFunc("/music/{uuid}/storage", musicH.UpdateMusicStorage).Methods("POST")
	protectedRouter.HandleFunc("/music/{uuid}", musicH.DeleteMusic).Methods("DELETE")
	protectedRouter.HandleFunc("/music/{uuid}/play", musicH.IncrementPlayCount).Methods("POST")
	protectedRouter.HandleFunc("/music/{uuid}/listen", musicH.AddListeningHistoryEntry).Methods("POST")
	protectedRouter.HandleFunc("/music/{uuid}/liked", likesH.IsLiked).Methods("GET")
	protectedRouter.HandleFunc("/music/{uuid}/like", likesH.LikeMusic).Methods("POST")
	protectedRouter.HandleFunc("/music/{uuid}/like", likesH.UnlikeMusic).Methods("DELETE")
	protectedRouter.HandleFunc("/music/{uuid}/tags", tagsH.GetTagsForMusic).Methods("GET")
	protectedRouter.HandleFunc("/music/{uuid}/tags/{name}", tagsH.AssignTagToMusic).Methods("POST")
	protectedRouter.HandleFunc("/music/{uuid}/tags/{name}", tagsH.RemoveTagFromMusic).Methods("DELETE")

	// Tags
	protectedRouter.HandleFunc("/tags", tagsH.GetAllTags).Methods("GET")
	protectedRouter.HandleFunc("/tags", tagsH.CreateTag).Methods("PUT")
	protectedRouter.HandleFunc("/tags/{name}", tagsH.GetTag).Methods("GET")
	protectedRouter.HandleFunc("/tags/{name}/music", tagsH.GetMusicForTag).Methods("GET")

	// Playlists
	protectedRouter.HandleFunc("/playlists", playlistH.CreatePlaylist).Methods("PUT")
	protectedRouter.HandleFunc("/playlists/{uuid}", playlistH.GetPlaylist).Methods("GET")
	protectedRouter.HandleFunc("/playlists/{uuid}", playlistH.UpdatePlaylist).Methods("POST")
	protectedRouter.HandleFunc("/playlists/{uuid}/image", playlistH.UpdatePlaylistImage).Methods("POST")
	protectedRouter.HandleFunc("/playlists/{uuid}", playlistH.DeletePlaylist).Methods("DELETE")
	protectedRouter.HandleFunc("/playlists/{uuid}/tracks", playlistH.GetPlaylistTracks).Methods("GET")
	protectedRouter.HandleFunc("/playlists/{uuid}/tracks/{musicUuid}", playlistH.AddTrackToPlaylist).Methods("PUT")
	protectedRouter.HandleFunc("/playlists/{uuid}/tracks/{musicUuid}", playlistH.RemoveTrackFromPlaylist).Methods("DELETE")
	protectedRouter.HandleFunc("/playlists/{uuid}/tracks/{trackUuid}/position", playlistH.UpdateTrackPosition).Methods("POST")

	// History
	protectedRouter.HandleFunc("/history", historyH.GetListeningHistoryForUser).Methods("GET")
	protectedRouter.HandleFunc("/history/top", historyH.GetTopMusicForUser).Methods("GET")

	return r
}
