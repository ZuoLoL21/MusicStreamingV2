package app

import (
	"backend/internal/di"
	"backend/internal/handlers"
	"backend/internal/storage"

	sqlhandler "backend/sql/sqlc"
	libsdi "libs/di"
	libshandlers "libs/handlers"
	libsmiddleware "libs/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	logger      *zap.Logger
	config      *di.Config
	jwtHandler  *libsdi.JWTHandler
	returns     *libsdi.ReturnManager
	db          *sqlhandler.Queries
	fileStorage storage.FileStorageClient
	handlers    *HandlerRegistry
}

type HandlerRegistry struct {
	User     *handlers.UserHandler
	Artist   *handlers.ArtistHandler
	Album    *handlers.AlbumHandler
	Music    *handlers.MusicHandler
	Likes    *handlers.LikesHandler
	Follows  *handlers.FollowsHandler
	Tags     *handlers.TagsHandler
	Playlist *handlers.PlaylistHandler
	History  *handlers.HistoryHandler
	Search   *handlers.SearchHandler
	File     *handlers.FileHandler
}

func New(logger *zap.Logger, config *di.Config, jwtHandler *libsdi.JWTHandler, returns *libsdi.ReturnManager, db *sqlhandler.Queries, fileStorage storage.FileStorageClient) *App {
	return &App{
		logger:      logger,
		config:      config,
		jwtHandler:  jwtHandler,
		returns:     returns,
		db:          db,
		fileStorage: fileStorage,
	}
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()

	// Initialize all handlers
	a.initHandlers()

	// Setup middleware
	publicRouter, protectedRouter := a.setupMiddleware(r)

	// Register all routes
	a.registerHealthRoutes(publicRouter)
	a.registerAuthRoutes(publicRouter, protectedRouter)
	a.registerFileRoutes(publicRouter)
	a.registerUserRoutes(protectedRouter)
	a.registerArtistRoutes(protectedRouter)
	a.registerAlbumRoutes(protectedRouter)
	a.registerMusicRoutes(protectedRouter)
	a.registerTagRoutes(protectedRouter)
	a.registerPlaylistRoutes(protectedRouter)
	a.registerHistoryRoutes(protectedRouter)
	a.registerSearchRoutes(protectedRouter)

	return r
}

func (a *App) initHandlers() {
	a.handlers = &HandlerRegistry{
		User:     handlers.NewUserHandler(a.logger, a.config, a.jwtHandler, a.returns, a.db, a.fileStorage),
		Artist:   handlers.NewArtistHandler(a.logger, a.config, a.returns, a.db, a.fileStorage),
		Album:    handlers.NewAlbumHandler(a.logger, a.config, a.returns, a.db, a.fileStorage),
		Music:    handlers.NewMusicHandler(a.logger, a.config, a.returns, a.db, a.fileStorage),
		Likes:    handlers.NewLikesHandler(a.logger, a.config, a.returns, a.db, a.fileStorage),
		Follows:  handlers.NewFollowsHandler(a.logger, a.config, a.returns, a.db),
		Tags:     handlers.NewTagsHandler(a.logger, a.config, a.returns, a.db, a.fileStorage),
		Playlist: handlers.NewPlaylistHandler(a.logger, a.config, a.returns, a.db, a.fileStorage),
		History:  handlers.NewHistoryHandler(a.logger, a.config, a.returns, a.db),
		Search:   handlers.NewSearchHandler(a.logger, a.config, a.returns, a.db, a.fileStorage),
		File:     handlers.NewFileHandler(a.fileStorage, a.logger, a.returns, a.config, a.db),
	}
}

func (a *App) setupMiddleware(r *mux.Router) (*mux.Router, *mux.Router) {
	serviceAuthHandler := libsmiddleware.NewAuthHandler(
		a.logger,
		a.config,
		a.jwtHandler,
		a.returns,
		libsdi.JWTSubjectService,
	)

	publicRouter := r.PathPrefix("").Subrouter()
	protectedRouter := r.PathPrefix("").Subrouter()

	publicRouter.Use(
		libsmiddleware.RequestIDMiddleware(a.config),
		libsmiddleware.LoggingMiddleware(a.logger, a.config),
		libsmiddleware.Logger(a.logger, libsmiddleware.LoggerConfig{
			RequestIDKey: a.config.RequestIDKey,
			UserUUIDKey:  a.config.UserUUIDKey,
		}),
	)
	protectedRouter.Use(
		libsmiddleware.RequestIDMiddleware(a.config),
		libsmiddleware.LoggingMiddleware(a.logger, a.config),
		serviceAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.logger, libsmiddleware.LoggerConfig{
			RequestIDKey: a.config.RequestIDKey,
			UserUUIDKey:  a.config.UserUUIDKey,
		}),
	)

	return publicRouter, protectedRouter
}

func (a *App) registerHealthRoutes(r *mux.Router) {
	r.HandleFunc("/health", libshandlers.NewHealthCheckHandler("service-user-database")).Methods("GET")
}

func (a *App) registerFileRoutes(r *mux.Router) {
	r.PathPrefix("/files/").HandlerFunc(a.handlers.File.ServeFile).Methods("GET")
}

func (a *App) registerAuthRoutes(r, protected *mux.Router) {
	r.HandleFunc("/login", a.handlers.User.Login).Methods("POST")
	r.HandleFunc("/login", a.handlers.User.Register).Methods("PUT")
	protected.HandleFunc("/renew", a.handlers.User.Renew).Methods("POST")
}

func (a *App) registerUserRoutes(r *mux.Router) {
	// Static /me routes BEFORE /{uuid}
	r.HandleFunc("/users/me", a.handlers.User.GetMe).Methods("GET")
	r.HandleFunc("/users/me", a.handlers.User.UpdateProfile).Methods("POST")
	r.HandleFunc("/users/me/email", a.handlers.User.UpdateEmail).Methods("POST")
	r.HandleFunc("/users/me/password", a.handlers.User.UpdatePassword).Methods("POST")
	r.HandleFunc("/users/me/image", a.handlers.User.UpdateImage).Methods("POST")

	// User-specific routes
	r.HandleFunc("/users/{uuid}", a.handlers.User.GetPublicUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/artists", a.handlers.User.GetArtistForUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/likes", a.handlers.Likes.GetLikesForUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/followers", a.handlers.Follows.GetFollowersForUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/following/users", a.handlers.Follows.GetFollowingUsersForUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/following/artists", a.handlers.Follows.GetFollowedArtistsForUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/follow", a.handlers.Follows.FollowUser).Methods("POST")
	r.HandleFunc("/users/{uuid}/follow", a.handlers.Follows.UnfollowUser).Methods("DELETE")
	r.HandleFunc("/users/{uuid}/music", a.handlers.Music.GetMusicForUser).Methods("GET")
	r.HandleFunc("/users/{uuid}/playlists", a.handlers.Playlist.GetPlaylistsForUser).Methods("GET")
}

func (a *App) registerArtistRoutes(r *mux.Router) {
	r.HandleFunc("/artists", a.handlers.Artist.GetArtistsAlphabetically).Methods("GET")
	r.HandleFunc("/artists", a.handlers.Artist.CreateArtist).Methods("PUT")
	r.HandleFunc("/artists/{uuid}", a.handlers.Artist.GetArtist).Methods("GET")
	r.HandleFunc("/artists/{uuid}", a.handlers.Artist.UpdateArtistProfile).Methods("POST")
	r.HandleFunc("/artists/{uuid}/image", a.handlers.Artist.UpdateArtistPicture).Methods("POST")
	r.HandleFunc("/artists/{uuid}/members", a.handlers.Artist.GetUsersRepresentingArtist).Methods("GET")
	r.HandleFunc("/artists/{uuid}/members/{userUuid}", a.handlers.Artist.AddUserToArtist).Methods("PUT")
	r.HandleFunc("/artists/{uuid}/members/{userUuid}", a.handlers.Artist.RemoveUserFromArtist).Methods("DELETE")
	r.HandleFunc("/artists/{uuid}/members/{userUuid}/role", a.handlers.Artist.ChangeUserRole).Methods("POST")
	r.HandleFunc("/artists/{uuid}/albums", a.handlers.Album.GetAlbumsForArtist).Methods("GET")
	r.HandleFunc("/artists/{uuid}/music", a.handlers.Music.GetMusicForArtist).Methods("GET")
	r.HandleFunc("/artists/{uuid}/followers", a.handlers.Follows.GetFollowersForArtist).Methods("GET")
	r.HandleFunc("/artists/{uuid}/follow", a.handlers.Follows.FollowArtist).Methods("POST")
	r.HandleFunc("/artists/{uuid}/follow", a.handlers.Follows.UnfollowArtist).Methods("DELETE")
}

func (a *App) registerAlbumRoutes(r *mux.Router) {
	r.HandleFunc("/albums", a.handlers.Album.CreateAlbum).Methods("PUT")
	r.HandleFunc("/albums/{uuid}", a.handlers.Album.GetAlbum).Methods("GET")
	r.HandleFunc("/albums/{uuid}", a.handlers.Album.UpdateAlbum).Methods("POST")
	r.HandleFunc("/albums/{uuid}/image", a.handlers.Album.UpdateAlbumImage).Methods("POST")
	r.HandleFunc("/albums/{uuid}", a.handlers.Album.DeleteAlbum).Methods("DELETE")
	r.HandleFunc("/albums/{uuid}/music", a.handlers.Music.GetMusicForAlbum).Methods("GET")
}

func (a *App) registerMusicRoutes(r *mux.Router) {
	r.HandleFunc("/music", a.handlers.Music.CreateMusic).Methods("PUT")
	r.HandleFunc("/music/{uuid}", a.handlers.Music.GetMusic).Methods("GET")
	r.HandleFunc("/music/{uuid}", a.handlers.Music.UpdateMusicDetails).Methods("POST")
	r.HandleFunc("/music/{uuid}/storage", a.handlers.Music.UpdateMusicStorage).Methods("POST")
	r.HandleFunc("/music/{uuid}", a.handlers.Music.DeleteMusic).Methods("DELETE")
	r.HandleFunc("/music/{uuid}/play", a.handlers.Music.IncrementPlayCount).Methods("POST")
	r.HandleFunc("/music/{uuid}/listen", a.handlers.Music.AddListeningHistoryEntry).Methods("POST")
	r.HandleFunc("/music/{uuid}/liked", a.handlers.Likes.IsLiked).Methods("GET")
	r.HandleFunc("/music/{uuid}/like", a.handlers.Likes.LikeMusic).Methods("POST")
	r.HandleFunc("/music/{uuid}/like", a.handlers.Likes.UnlikeMusic).Methods("DELETE")
	r.HandleFunc("/music/{uuid}/tags", a.handlers.Tags.GetTagsForMusic).Methods("GET")
	r.HandleFunc("/music/{uuid}/tags/{name}", a.handlers.Tags.AssignTagToMusic).Methods("POST")
	r.HandleFunc("/music/{uuid}/tags/{name}", a.handlers.Tags.RemoveTagFromMusic).Methods("DELETE")
}

func (a *App) registerTagRoutes(r *mux.Router) {
	r.HandleFunc("/tags", a.handlers.Tags.GetAllTags).Methods("GET")
	r.HandleFunc("/tags/{name}", a.handlers.Tags.GetTag).Methods("GET")
	r.HandleFunc("/tags/{name}/music", a.handlers.Tags.GetMusicForTag).Methods("GET")
}

func (a *App) registerPlaylistRoutes(r *mux.Router) {
	r.HandleFunc("/playlists", a.handlers.Playlist.CreatePlaylist).Methods("PUT")
	r.HandleFunc("/playlists/{uuid}", a.handlers.Playlist.GetPlaylist).Methods("GET")
	r.HandleFunc("/playlists/{uuid}", a.handlers.Playlist.UpdatePlaylist).Methods("POST")
	r.HandleFunc("/playlists/{uuid}/image", a.handlers.Playlist.UpdatePlaylistImage).Methods("POST")
	r.HandleFunc("/playlists/{uuid}", a.handlers.Playlist.DeletePlaylist).Methods("DELETE")
	r.HandleFunc("/playlists/{uuid}/tracks", a.handlers.Playlist.GetPlaylistTracks).Methods("GET")
	r.HandleFunc("/playlists/{uuid}/tracks/{musicUuid}", a.handlers.Playlist.AddTrackToPlaylist).Methods("PUT")
	r.HandleFunc("/playlists/{uuid}/tracks/{musicUuid}", a.handlers.Playlist.RemoveTrackFromPlaylist).Methods("DELETE")
	r.HandleFunc("/playlists/{uuid}/tracks/{trackUuid}/position", a.handlers.Playlist.UpdateTrackPosition).Methods("POST")
}

func (a *App) registerHistoryRoutes(r *mux.Router) {
	r.HandleFunc("/history", a.handlers.History.GetListeningHistoryForUser).Methods("GET")
	r.HandleFunc("/history/top", a.handlers.History.GetTopMusicForUser).Methods("GET")
}

func (a *App) registerSearchRoutes(r *mux.Router) {
	r.HandleFunc("/search/users", a.handlers.Search.SearchUsers).Methods("GET")
	r.HandleFunc("/search/artists", a.handlers.Search.SearchArtists).Methods("GET")
	r.HandleFunc("/search/albums", a.handlers.Search.SearchAlbums).Methods("GET")
	r.HandleFunc("/search/music", a.handlers.Search.SearchMusic).Methods("GET")
	r.HandleFunc("/search/playlists", a.handlers.Search.SearchPlaylists).Methods("GET")
}
