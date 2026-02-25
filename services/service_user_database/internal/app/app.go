package app

import (
	"backend/internal/di"
	"backend/internal/handlers"
	"backend/internal/storage"
	sqlhandler "backend/sql/sqlc"

	libsdi "libs/di"
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

	r.Use(
		libsmiddleware.RequestIDMiddleware(a.config),
		libsmiddleware.LoggingMiddleware(a.logger, a.config),
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

	// Auth
	r.HandleFunc("/login", userH.Login).Methods("POST")
	r.HandleFunc("/login", userH.Register).Methods("PUT")
	r.HandleFunc("/renew", userH.Renew).Methods("POST")

	// Users — static /me routes BEFORE /{uuid}
	r.HandleFunc("/users/me", userH.GetMe).Methods("GET")
	r.HandleFunc("/users/me", userH.UpdateProfile).Methods("POST")
	r.HandleFunc("/users/me/email", userH.UpdateEmail).Methods("POST")
	r.HandleFunc("/users/me/password", userH.UpdatePassword).Methods("POST")
	r.HandleFunc("/users/me/image", userH.UpdateImage).Methods("POST")
	r.HandleFunc("/users/{uuid}", userH.GetPublicUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/artists", userH.GetArtistForUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/likes", likesH.GetLikesForUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/followers", followsH.GetFollowersForUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/following/users", followsH.GetFollowingUsersForUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/following/artists", followsH.GetFollowedArtistsForUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/follow", followsH.FollowUser).Methods("POST")
	r.HandleFunc("/users/{uuid}/follow", followsH.UnfollowUser).Methods("DELETE")
	r.HandleFunc("/users/{uuid}/music", musicH.GetMusicForUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/playlists", playlistH.GetPlaylistsForUser).Methods("GET")

	// Artists
	r.HandleFunc("/artists", artistH.GetArtistsAlphabetically).Methods("GET")
	r.HandleFunc("/artists", artistH.CreateArtist).Methods("PUT")
	r.HandleFunc("/artists/{uuid}", artistH.GetArtist).Methods("GET")
	r.HandleFunc("/artists/{uuid}", artistH.UpdateArtistProfile).Methods("POST")
	r.HandleFunc("/artists/{uuid}/image", artistH.UpdateArtistPicture).Methods("POST")
	r.HandleFunc("/artists/{uuid}/members", artistH.GetUsersRepresentingArtist).Methods("GET")
	r.HandleFunc("/artists/{uuid}/members/{userUuid}", artistH.AddUserToArtist).Methods("PUT")
	r.HandleFunc("/artists/{uuid}/members/{userUuid}", artistH.RemoveUserFromArtist).Methods("DELETE")
	r.HandleFunc("/artists/{uuid}/members/{userUuid}/role", artistH.ChangeUserRole).Methods("POST")
	r.HandleFunc("/artists/{uuid}/albums", albumH.GetAlbumsForArtist).Methods("GET")
	r.HandleFunc("/artists/{uuid}/music", musicH.GetMusicForArtist).Methods("GET")
	r.HandleFunc("/artists/{uuid}/followers", followsH.GetFollowersForArtist).Methods("GET")
	r.HandleFunc("/artists/{uuid}/follow", followsH.FollowArtist).Methods("POST")
	r.HandleFunc("/artists/{uuid}/follow", followsH.UnfollowArtist).Methods("DELETE")

	// Albums
	r.HandleFunc("/albums", albumH.CreateAlbum).Methods("PUT")
	r.HandleFunc("/albums/{uuid}", albumH.GetAlbum).Methods("GET")
	r.HandleFunc("/albums/{uuid}", albumH.UpdateAlbum).Methods("POST")
	r.HandleFunc("/albums/{uuid}/image", albumH.UpdateAlbumImage).Methods("POST")
	r.HandleFunc("/albums/{uuid}", albumH.DeleteAlbum).Methods("DELETE")
	r.HandleFunc("/albums/{uuid}/music", musicH.GetMusicForAlbum).Methods("GET")

	// Music
	r.HandleFunc("/music", musicH.CreateMusic).Methods("PUT")
	r.HandleFunc("/music/{uuid}", musicH.GetMusic).Methods("GET")
	r.HandleFunc("/music/{uuid}", musicH.UpdateMusicDetails).Methods("POST")
	r.HandleFunc("/music/{uuid}/storage", musicH.UpdateMusicStorage).Methods("POST")
	r.HandleFunc("/music/{uuid}", musicH.DeleteMusic).Methods("DELETE")
	r.HandleFunc("/music/{uuid}/play", musicH.IncrementPlayCount).Methods("POST")
	r.HandleFunc("/music/{uuid}/listen", musicH.AddListeningHistoryEntry).Methods("POST")
	r.HandleFunc("/music/{uuid}/liked", likesH.IsLiked).Methods("GET")
	r.HandleFunc("/music/{uuid}/like", likesH.LikeMusic).Methods("POST")
	r.HandleFunc("/music/{uuid}/like", likesH.UnlikeMusic).Methods("DELETE")
	r.HandleFunc("/music/{uuid}/tags", tagsH.GetTagsForMusic).Methods("GET")
	r.HandleFunc("/music/{uuid}/tags/{name}", tagsH.AssignTagToMusic).Methods("POST")
	r.HandleFunc("/music/{uuid}/tags/{name}", tagsH.RemoveTagFromMusic).Methods("DELETE")

	// Tags
	r.HandleFunc("/tags", tagsH.GetAllTags).Methods("GET")
	r.HandleFunc("/tags", tagsH.CreateTag).Methods("PUT")
	r.HandleFunc("/tags/{name}", tagsH.GetTag).Methods("GET")
	r.HandleFunc("/tags/{name}/music", tagsH.GetMusicForTag).Methods("GET")

	// Playlists
	r.HandleFunc("/playlists", playlistH.CreatePlaylist).Methods("PUT")
	r.HandleFunc("/playlists/{uuid}", playlistH.GetPlaylist).Methods("GET")
	r.HandleFunc("/playlists/{uuid}", playlistH.UpdatePlaylist).Methods("POST")
	r.HandleFunc("/playlists/{uuid}/image", playlistH.UpdatePlaylistImage).Methods("POST")
	r.HandleFunc("/playlists/{uuid}", playlistH.DeletePlaylist).Methods("DELETE")
	r.HandleFunc("/playlists/{uuid}/tracks", playlistH.GetPlaylistTracks).Methods("GET")
	r.HandleFunc("/playlists/{uuid}/tracks/{musicUuid}", playlistH.AddTrackToPlaylist).Methods("PUT")
	r.HandleFunc("/playlists/{uuid}/tracks/{musicUuid}", playlistH.RemoveTrackFromPlaylist).Methods("DELETE")
	r.HandleFunc("/playlists/{uuid}/tracks/{trackUuid}/position", playlistH.UpdateTrackPosition).Methods("POST")

	// History
	r.HandleFunc("/history", historyH.GetListeningHistoryForUser).Methods("GET")
	r.HandleFunc("/history/top", historyH.GetTopMusicForUser).Methods("GET")

	return r
}
