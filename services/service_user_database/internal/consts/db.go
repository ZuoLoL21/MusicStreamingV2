package consts

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

	GetArtistsAlphabetically(ctx context.Context, arg sqlhandler.GetArtistsAlphabeticallyParams) ([]sqlhandler.Artist, error)
	GetArtist(ctx context.Context, uuid pgtype.UUID) (sqlhandler.Artist, error)
	CreateArtist(ctx context.Context, arg sqlhandler.CreateArtistParams) error
	UpdateArtistProfile(ctx context.Context, arg sqlhandler.UpdateArtistProfileParams) error
	UpdateArtistPicture(ctx context.Context, arg sqlhandler.UpdateArtistPictureParams) error
	GetUsersRepresentingArtist(ctx context.Context, artistUuid pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error)
	AddUserToArtist(ctx context.Context, arg sqlhandler.AddUserToArtistParams) error
	RemoveUserFromArtist(ctx context.Context, arg sqlhandler.RemoveUserFromArtistParams) error
	ChangeUserRole(ctx context.Context, arg sqlhandler.ChangeUserRoleParams) error

	GetAlbum(ctx context.Context, uuid pgtype.UUID) (sqlhandler.Album, error)
	GetAlbumsForArtist(ctx context.Context, arg sqlhandler.GetAlbumsForArtistParams) ([]sqlhandler.Album, error)
	CreateAlbum(ctx context.Context, arg sqlhandler.CreateAlbumParams) error
	UpdateAlbum(ctx context.Context, arg sqlhandler.UpdateAlbumParams) error
	UpdateAlbumImage(ctx context.Context, arg sqlhandler.UpdateAlbumImageParams) error
	DeleteAlbum(ctx context.Context, uuid pgtype.UUID) error

	GetMusic(ctx context.Context, uuid pgtype.UUID) (sqlhandler.Music, error)
	GetMusicForArtist(ctx context.Context, arg sqlhandler.GetMusicForArtistParams) ([]sqlhandler.Music, error)
	GetMusicForAlbum(ctx context.Context, arg sqlhandler.GetMusicForAlbumParams) ([]sqlhandler.Music, error)
	GetMusicForUser(ctx context.Context, arg sqlhandler.GetMusicForUserParams) ([]sqlhandler.Music, error)
	CreateMusic(ctx context.Context, arg sqlhandler.CreateMusicParams) error
	UpdateMusicDetails(ctx context.Context, arg sqlhandler.UpdateMusicDetailsParams) error
	UpdateMusicImage(ctx context.Context, arg sqlhandler.UpdateMusicImageParams) error
	UpdateMusicStorage(ctx context.Context, arg sqlhandler.UpdateMusicStorageParams) error
	DeleteMusic(ctx context.Context, uuid pgtype.UUID) error
	IncrementPlayCount(ctx context.Context, uuid pgtype.UUID) error
	AddListeningHistoryEntry(ctx context.Context, arg sqlhandler.AddListeningHistoryEntryParams) error

	GetLikesForUser(ctx context.Context, arg sqlhandler.GetLikesForUserParams) ([]sqlhandler.Music, error)
	IsLiked(ctx context.Context, arg sqlhandler.IsLikedParams) (bool, error)
	LikeMusic(ctx context.Context, arg sqlhandler.LikeMusicParams) error
	UnlikeMusic(ctx context.Context, arg sqlhandler.UnlikeMusicParams) error

	GetFollowersForUser(ctx context.Context, arg sqlhandler.GetFollowersForUserParams) ([]sqlhandler.PublicUser, error)
	GetFollowedUsersForUser(ctx context.Context, arg sqlhandler.GetFollowedUsersForUserParams) ([]sqlhandler.PublicUser, error)
	GetFollowedArtistsForUser(ctx context.Context, arg sqlhandler.GetFollowedArtistsForUserParams) ([]sqlhandler.Artist, error)
	GetFollowersForArtist(ctx context.Context, arg sqlhandler.GetFollowersForArtistParams) ([]sqlhandler.PublicUser, error)
	IsFollowingUser(ctx context.Context, arg sqlhandler.IsFollowingUserParams) (bool, error)
	FollowUser(ctx context.Context, arg sqlhandler.FollowUserParams) error
	UnfollowUser(ctx context.Context, arg sqlhandler.UnfollowUserParams) error
	FollowArtist(ctx context.Context, arg sqlhandler.FollowArtistParams) error
	UnfollowArtist(ctx context.Context, arg sqlhandler.UnfollowArtistParams) error
	IsFollowingArtist(ctx context.Context, arg sqlhandler.IsFollowingArtistParams) (bool, error)

	GetAllTags(ctx context.Context, arg sqlhandler.GetAllTagsParams) ([]sqlhandler.MusicTag, error)
	GetTag(ctx context.Context, tagName string) (sqlhandler.MusicTag, error)
	GetMusicForTag(ctx context.Context, arg sqlhandler.GetMusicForTagParams) ([]sqlhandler.Music, error)
	GetTagsForMusic(ctx context.Context, arg sqlhandler.GetTagsForMusicParams) ([]sqlhandler.MusicTag, error)
	CreateTag(ctx context.Context, arg sqlhandler.CreateTagParams) error
	AssignTagToMusic(ctx context.Context, arg sqlhandler.AssignTagToMusicParams) error
	RemoveTagFromMusic(ctx context.Context, arg sqlhandler.RemoveTagFromMusicParams) error

	GetPlaylist(ctx context.Context, uuid pgtype.UUID) (sqlhandler.Playlist, error)
	GetPlaylists(ctx context.Context, arg sqlhandler.GetPlaylistsParams) ([]sqlhandler.Playlist, error)
	GetPlaylistsForUser(ctx context.Context, arg sqlhandler.GetPlaylistsForUserParams) ([]sqlhandler.Playlist, error)
	IsPlaylistPublicOrOwnedByUser(ctx context.Context, arg sqlhandler.IsPlaylistPublicOrOwnedByUserParams) (bool, error)
	GetPlaylistTracks(ctx context.Context, arg sqlhandler.GetPlaylistTracksParams) ([]sqlhandler.Music, error)
	CreatePlaylist(ctx context.Context, arg sqlhandler.CreatePlaylistParams) error
	UpdatePlaylist(ctx context.Context, arg sqlhandler.UpdatePlaylistParams) error
	UpdatePlaylistImage(ctx context.Context, arg sqlhandler.UpdatePlaylistImageParams) error
	DeletePlaylist(ctx context.Context, uuid sqlhandler.DeletePlaylistParams) error
	AddTrackToPlaylist(ctx context.Context, arg sqlhandler.AddTrackToPlaylistParams) error
	RemoveTrackFromPlaylist(ctx context.Context, arg sqlhandler.RemoveTrackFromPlaylistParams) error
	UpdateTrackPosition(ctx context.Context, arg sqlhandler.UpdateTrackPositionParams) error

	GetListeningHistoryForUser(ctx context.Context, arg sqlhandler.GetListeningHistoryForUserParams) ([]sqlhandler.ListeningHistory, error)
	GetTopMusicForUser(ctx context.Context, arg sqlhandler.GetTopMusicForUserParams) ([]sqlhandler.GetTopMusicForUserRow, error)

	SearchForMusic(ctx context.Context, arg sqlhandler.SearchForMusicParams) ([]sqlhandler.SearchForMusicRow, error)
	SearchForAlbum(ctx context.Context, arg sqlhandler.SearchForAlbumParams) ([]sqlhandler.SearchForAlbumRow, error)
	SearchForArtist(ctx context.Context, arg sqlhandler.SearchForArtistParams) ([]sqlhandler.SearchForArtistRow, error)
	SearchForUser(ctx context.Context, arg sqlhandler.SearchForUserParams) ([]sqlhandler.SearchForUserRow, error)
	SearchForPlaylist(ctx context.Context, arg sqlhandler.SearchForPlaylistParams) ([]sqlhandler.SearchForPlaylistRow, error)
}
