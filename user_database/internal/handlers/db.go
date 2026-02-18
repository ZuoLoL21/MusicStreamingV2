package handlers

import (
	"context"

	sqlhandler "backend/sql/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

// DB is the interface satisfied by *sqlhandler.Queries and used by all handlers.
// It exists so handlers can be unit-tested with a mock.
type DB interface {
	CreateUser(ctx context.Context, arg sqlhandler.CreateUserParams) (pgtype.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (sqlhandler.User, error)
	GetPublicUser(ctx context.Context, uuid pgtype.UUID) (sqlhandler.PublicUser, error)
	GetHashPassword(ctx context.Context, uuid pgtype.UUID) (string, error)
	UpdateProfile(ctx context.Context, arg sqlhandler.UpdateProfileParams) error
	UpdateEmail(ctx context.Context, arg sqlhandler.UpdateEmailParams) error
	UpdatePassword(ctx context.Context, arg sqlhandler.UpdatePasswordParams) error
	UpdateImage(ctx context.Context, arg sqlhandler.UpdateImageParams) error
	GetArtistForUser(ctx context.Context, userUuid pgtype.UUID) ([]sqlhandler.GetArtistForUserRow, error)

	GetArtistsAlphabetically(ctx context.Context) ([]sqlhandler.Artist, error)
	GetArtist(ctx context.Context, uuid pgtype.UUID) (sqlhandler.Artist, error)
	CreateArtist(ctx context.Context, arg sqlhandler.CreateArtistParams) error
	UpdateArtistProfile(ctx context.Context, arg sqlhandler.UpdateArtistProfileParams) error
	UpdateArtistPicture(ctx context.Context, arg sqlhandler.UpdateArtistPictureParams) error
	GetUsersRepresentingArtist(ctx context.Context, artistUuid pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error)
	AddUserToArtist(ctx context.Context, arg sqlhandler.AddUserToArtistParams) error
	RemoveUserFromArtist(ctx context.Context, arg sqlhandler.RemoveUserFromArtistParams) error
	ChangeUserRole(ctx context.Context, arg sqlhandler.ChangeUserRoleParams) error

	GetAlbum(ctx context.Context, uuid pgtype.UUID) (sqlhandler.Album, error)
	GetAlbumsForArtist(ctx context.Context, fromArtist pgtype.UUID) ([]sqlhandler.Album, error)
	CreateAlbum(ctx context.Context, arg sqlhandler.CreateAlbumParams) error
	UpdateAlbum(ctx context.Context, arg sqlhandler.UpdateAlbumParams) error
	UpdateAlbumImage(ctx context.Context, arg sqlhandler.UpdateAlbumImageParams) error
	DeleteAlbum(ctx context.Context, uuid pgtype.UUID) error

	GetMusic(ctx context.Context, uuid pgtype.UUID) (sqlhandler.Music, error)
	GetMusicForArtist(ctx context.Context, fromArtist pgtype.UUID) ([]sqlhandler.Music, error)
	GetMusicForAlbum(ctx context.Context, inAlbum pgtype.UUID) ([]sqlhandler.Music, error)
	GetMusicForUser(ctx context.Context, uploadedBy pgtype.UUID) ([]sqlhandler.Music, error)
	CreateMusic(ctx context.Context, arg sqlhandler.CreateMusicParams) error
	UpdateMusicDetails(ctx context.Context, arg sqlhandler.UpdateMusicDetailsParams) error
	UpdateMusicStorage(ctx context.Context, arg sqlhandler.UpdateMusicStorageParams) error
	DeleteMusic(ctx context.Context, uuid pgtype.UUID) error
	IncrementPlayCount(ctx context.Context, uuid pgtype.UUID) error
	AddListeningHistoryEntry(ctx context.Context, arg sqlhandler.AddListeningHistoryEntryParams) error

	GetLikesForMusic(ctx context.Context, toMusic pgtype.UUID) ([]sqlhandler.PublicUser, error)
	GetLikesForUser(ctx context.Context, fromUser pgtype.UUID) ([]sqlhandler.Music, error)
	IsLiked(ctx context.Context, arg sqlhandler.IsLikedParams) (bool, error)
	LikeMusic(ctx context.Context, arg sqlhandler.LikeMusicParams) error
	UnlikeMusic(ctx context.Context, arg sqlhandler.UnlikeMusicParams) error

	GetFollowersForUser(ctx context.Context, toUser pgtype.UUID) ([]sqlhandler.PublicUser, error)
	GetFollowingUsersForUser(ctx context.Context, fromUser pgtype.UUID) ([]sqlhandler.PublicUser, error)
	GetFollowingArtistsForUser(ctx context.Context, fromUser pgtype.UUID) ([]sqlhandler.Artist, error)
	GetFollowersForArtist(ctx context.Context, toArtist pgtype.UUID) ([]sqlhandler.PublicUser, error)
	FollowUser(ctx context.Context, arg sqlhandler.FollowUserParams) error
	UnfollowUser(ctx context.Context, arg sqlhandler.UnfollowUserParams) error
	FollowArtist(ctx context.Context, arg sqlhandler.FollowArtistParams) error
	UnfollowArtist(ctx context.Context, arg sqlhandler.UnfollowArtistParams) error

	GetAllTags(ctx context.Context) ([]sqlhandler.MusicTag, error)
	GetTag(ctx context.Context, tagName string) (sqlhandler.MusicTag, error)
	GetMusicForTag(ctx context.Context, tagName string) ([]sqlhandler.Music, error)
	GetTagsForMusic(ctx context.Context, musicUuid pgtype.UUID) ([]sqlhandler.MusicTag, error)
	CreateTag(ctx context.Context, arg sqlhandler.CreateTagParams) error
	AssignTagToMusic(ctx context.Context, arg sqlhandler.AssignTagToMusicParams) error
	RemoveTagFromMusic(ctx context.Context, arg sqlhandler.RemoveTagFromMusicParams) error

	GetPlaylist(ctx context.Context, uuid pgtype.UUID) (sqlhandler.Playlist, error)
	GetPlaylistsForUser(ctx context.Context, fromUser pgtype.UUID) ([]sqlhandler.Playlist, error)
	GetPlaylistTracks(ctx context.Context, playlistUuid pgtype.UUID) ([]sqlhandler.Music, error)
	CreatePlaylist(ctx context.Context, arg sqlhandler.CreatePlaylistParams) error
	UpdatePlaylist(ctx context.Context, arg sqlhandler.UpdatePlaylistParams) error
	DeletePlaylist(ctx context.Context, uuid pgtype.UUID) error
	AddTrackToPlaylist(ctx context.Context, arg sqlhandler.AddTrackToPlaylistParams) error
	RemoveTrackFromPlaylist(ctx context.Context, arg sqlhandler.RemoveTrackFromPlaylistParams) error
	UpdateTrackPosition(ctx context.Context, arg sqlhandler.UpdateTrackPositionParams) error

	GetListeningHistoryForUser(ctx context.Context, userUuid pgtype.UUID) ([]sqlhandler.ListeningHistory, error)
	GetRecentlyPlayedForUser(ctx context.Context, arg sqlhandler.GetRecentlyPlayedForUserParams) ([]sqlhandler.ListeningHistory, error)
	GetTopMusicForUser(ctx context.Context, arg sqlhandler.GetTopMusicForUserParams) ([]sqlhandler.GetTopMusicForUserRow, error)
}
