package tests

import (
	"context"

	"backend/internal/handlers"
	sqlhandler "backend/sql/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

// Compile-time check: mockDB must implement handlers.DB.
var _ handlers.DB = (*mockDB)(nil)

// mockDB is a configurable test double for handlers.DB.
// Set only the Fn fields that a specific test case exercises;
// all others return zero values with no error.
type mockDB struct {
	createUserFn                 func(context.Context, sqlhandler.CreateUserParams) (pgtype.UUID, error)
	getUserByEmailFn             func(context.Context, string) (sqlhandler.User, error)
	getPublicUserFn              func(context.Context, pgtype.UUID) (sqlhandler.PublicUser, error)
	getHashPasswordFn            func(context.Context, pgtype.UUID) (string, error)
	updateProfileFn              func(context.Context, sqlhandler.UpdateProfileParams) error
	updateEmailFn                func(context.Context, sqlhandler.UpdateEmailParams) error
	updatePasswordFn             func(context.Context, sqlhandler.UpdatePasswordParams) error
	updateImageFn                func(context.Context, sqlhandler.UpdateImageParams) error
	getArtistForUserFn           func(context.Context, pgtype.UUID) ([]sqlhandler.GetArtistForUserRow, error)
	getArtistsAlphabeticallyFn   func(context.Context, sqlhandler.GetArtistsAlphabeticallyParams) ([]sqlhandler.Artist, error)
	getArtistFn                  func(context.Context, pgtype.UUID) (sqlhandler.Artist, error)
	createArtistFn               func(context.Context, sqlhandler.CreateArtistParams) error
	updateArtistProfileFn        func(context.Context, sqlhandler.UpdateArtistProfileParams) error
	updateArtistPictureFn        func(context.Context, sqlhandler.UpdateArtistPictureParams) error
	getUsersRepresentingArtistFn func(context.Context, pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error)
	addUserToArtistFn            func(context.Context, sqlhandler.AddUserToArtistParams) error
	removeUserFromArtistFn       func(context.Context, sqlhandler.RemoveUserFromArtistParams) error
	changeUserRoleFn             func(context.Context, sqlhandler.ChangeUserRoleParams) error
	getAlbumFn                   func(context.Context, pgtype.UUID) (sqlhandler.Album, error)
	getAlbumsForArtistFn         func(context.Context, sqlhandler.GetAlbumsForArtistParams) ([]sqlhandler.Album, error)
	createAlbumFn                func(context.Context, sqlhandler.CreateAlbumParams) error
	updateAlbumFn                func(context.Context, sqlhandler.UpdateAlbumParams) error
	updateAlbumImageFn           func(context.Context, sqlhandler.UpdateAlbumImageParams) error
	deleteAlbumFn                func(context.Context, pgtype.UUID) error
	getMusicFn                   func(context.Context, pgtype.UUID) (sqlhandler.Music, error)
	getMusicForArtistFn          func(context.Context, sqlhandler.GetMusicForArtistParams) ([]sqlhandler.Music, error)
	getMusicForAlbumFn           func(context.Context, sqlhandler.GetMusicForAlbumParams) ([]sqlhandler.Music, error)
	getMusicForUserFn            func(context.Context, sqlhandler.GetMusicForUserParams) ([]sqlhandler.Music, error)
	createMusicFn                func(context.Context, sqlhandler.CreateMusicParams) error
	updateMusicDetailsFn         func(context.Context, sqlhandler.UpdateMusicDetailsParams) error
	updateMusicStorageFn         func(context.Context, sqlhandler.UpdateMusicStorageParams) error
	deleteMusicFn                func(context.Context, pgtype.UUID) error
	incrementPlayCountFn         func(context.Context, pgtype.UUID) error
	addListeningHistoryEntryFn   func(context.Context, sqlhandler.AddListeningHistoryEntryParams) error
	getLikesForUserFn            func(context.Context, sqlhandler.GetLikesForUserParams) ([]sqlhandler.Music, error)
	isLikedFn                    func(context.Context, sqlhandler.IsLikedParams) (bool, error)
	likeMusicFn                  func(context.Context, sqlhandler.LikeMusicParams) error
	unlikeMusicFn                func(context.Context, sqlhandler.UnlikeMusicParams) error
	getFollowersForUserFn        func(context.Context, sqlhandler.GetFollowersForUserParams) ([]sqlhandler.PublicUser, error)
	getFollowsForUserFn          func(context.Context, sqlhandler.GetFollowsForUserParams) ([]sqlhandler.PublicUser, error)
	getFollowersForArtistFn      func(context.Context, sqlhandler.GetFollowersForArtistParams) ([]sqlhandler.PublicUser, error)
	followUserFn                 func(context.Context, sqlhandler.FollowUserParams) error
	unfollowUserFn               func(context.Context, sqlhandler.UnfollowUserParams) error
	followArtistFn               func(context.Context, sqlhandler.FollowArtistParams) error
	unfollowArtistFn             func(context.Context, sqlhandler.UnfollowArtistParams) error
	getAllTagsFn                 func(context.Context, sqlhandler.GetAllTagsParams) ([]sqlhandler.MusicTag, error)
	getTagFn                     func(context.Context, string) (sqlhandler.MusicTag, error)
	getMusicForTagFn             func(context.Context, sqlhandler.GetMusicForTagParams) ([]sqlhandler.Music, error)
	getTagsForMusicFn            func(context.Context, sqlhandler.GetTagsForMusicParams) ([]sqlhandler.MusicTag, error)
	createTagFn                  func(context.Context, sqlhandler.CreateTagParams) error
	assignTagToMusicFn           func(context.Context, sqlhandler.AssignTagToMusicParams) error
	removeTagFromMusicFn         func(context.Context, sqlhandler.RemoveTagFromMusicParams) error
	getPlaylistFn                func(context.Context, pgtype.UUID) (sqlhandler.Playlist, error)
	getPlaylistsForUserFn        func(context.Context, sqlhandler.GetPlaylistsForUserParams) ([]sqlhandler.Playlist, error)
	getPlaylistTracksFn          func(context.Context, sqlhandler.GetPlaylistTracksParams) ([]sqlhandler.Music, error)
	createPlaylistFn             func(context.Context, sqlhandler.CreatePlaylistParams) error
	updatePlaylistFn             func(context.Context, sqlhandler.UpdatePlaylistParams) error
	updatePlaylistImageFn        func(context.Context, sqlhandler.UpdatePlaylistImageParams) error
	deletePlaylistFn             func(context.Context, pgtype.UUID) error
	addTrackToPlaylistFn         func(context.Context, sqlhandler.AddTrackToPlaylistParams) error
	removeTrackFromPlaylistFn    func(context.Context, sqlhandler.RemoveTrackFromPlaylistParams) error
	updateTrackPositionFn        func(context.Context, sqlhandler.UpdateTrackPositionParams) error
	getListeningHistoryForUserFn func(context.Context, sqlhandler.GetListeningHistoryForUserParams) ([]sqlhandler.ListeningHistory, error)
	getTopMusicForUserFn         func(context.Context, sqlhandler.GetTopMusicForUserParams) ([]sqlhandler.GetTopMusicForUserRow, error)
}

func (m *mockDB) CreateUser(ctx context.Context, arg sqlhandler.CreateUserParams) (pgtype.UUID, error) {
	if m.createUserFn != nil {
		return m.createUserFn(ctx, arg)
	}
	return pgtype.UUID{}, nil
}
func (m *mockDB) GetUserByEmail(ctx context.Context, email string) (sqlhandler.User, error) {
	if m.getUserByEmailFn != nil {
		return m.getUserByEmailFn(ctx, email)
	}
	return sqlhandler.User{}, nil
}
func (m *mockDB) GetPublicUser(ctx context.Context, uuid pgtype.UUID) (sqlhandler.PublicUser, error) {
	if m.getPublicUserFn != nil {
		return m.getPublicUserFn(ctx, uuid)
	}
	return sqlhandler.PublicUser{}, nil
}
func (m *mockDB) GetHashPassword(ctx context.Context, uuid pgtype.UUID) (string, error) {
	if m.getHashPasswordFn != nil {
		return m.getHashPasswordFn(ctx, uuid)
	}
	return "", nil
}
func (m *mockDB) UpdateProfile(ctx context.Context, arg sqlhandler.UpdateProfileParams) error {
	if m.updateProfileFn != nil {
		return m.updateProfileFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UpdateEmail(ctx context.Context, arg sqlhandler.UpdateEmailParams) error {
	if m.updateEmailFn != nil {
		return m.updateEmailFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UpdatePassword(ctx context.Context, arg sqlhandler.UpdatePasswordParams) error {
	if m.updatePasswordFn != nil {
		return m.updatePasswordFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UpdateImage(ctx context.Context, arg sqlhandler.UpdateImageParams) error {
	if m.updateImageFn != nil {
		return m.updateImageFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) GetArtistForUser(ctx context.Context, userUuid pgtype.UUID) ([]sqlhandler.GetArtistForUserRow, error) {
	if m.getArtistForUserFn != nil {
		return m.getArtistForUserFn(ctx, userUuid)
	}
	return nil, nil
}
func (m *mockDB) GetArtistsAlphabetically(ctx context.Context, arg sqlhandler.GetArtistsAlphabeticallyParams) ([]sqlhandler.Artist, error) {
	if m.getArtistsAlphabeticallyFn != nil {
		return m.getArtistsAlphabeticallyFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) GetArtist(ctx context.Context, uuid pgtype.UUID) (sqlhandler.Artist, error) {
	if m.getArtistFn != nil {
		return m.getArtistFn(ctx, uuid)
	}
	return sqlhandler.Artist{}, nil
}
func (m *mockDB) CreateArtist(ctx context.Context, arg sqlhandler.CreateArtistParams) error {
	if m.createArtistFn != nil {
		return m.createArtistFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UpdateArtistProfile(ctx context.Context, arg sqlhandler.UpdateArtistProfileParams) error {
	if m.updateArtistProfileFn != nil {
		return m.updateArtistProfileFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UpdateArtistPicture(ctx context.Context, arg sqlhandler.UpdateArtistPictureParams) error {
	if m.updateArtistPictureFn != nil {
		return m.updateArtistPictureFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) GetUsersRepresentingArtist(ctx context.Context, artistUuid pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
	if m.getUsersRepresentingArtistFn != nil {
		return m.getUsersRepresentingArtistFn(ctx, artistUuid)
	}
	return nil, nil
}
func (m *mockDB) AddUserToArtist(ctx context.Context, arg sqlhandler.AddUserToArtistParams) error {
	if m.addUserToArtistFn != nil {
		return m.addUserToArtistFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) RemoveUserFromArtist(ctx context.Context, arg sqlhandler.RemoveUserFromArtistParams) error {
	if m.removeUserFromArtistFn != nil {
		return m.removeUserFromArtistFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) ChangeUserRole(ctx context.Context, arg sqlhandler.ChangeUserRoleParams) error {
	if m.changeUserRoleFn != nil {
		return m.changeUserRoleFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) GetAlbum(ctx context.Context, uuid pgtype.UUID) (sqlhandler.Album, error) {
	if m.getAlbumFn != nil {
		return m.getAlbumFn(ctx, uuid)
	}
	return sqlhandler.Album{}, nil
}
func (m *mockDB) GetAlbumsForArtist(ctx context.Context, arg sqlhandler.GetAlbumsForArtistParams) ([]sqlhandler.Album, error) {
	if m.getAlbumsForArtistFn != nil {
		return m.getAlbumsForArtistFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) CreateAlbum(ctx context.Context, arg sqlhandler.CreateAlbumParams) error {
	if m.createAlbumFn != nil {
		return m.createAlbumFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UpdateAlbum(ctx context.Context, arg sqlhandler.UpdateAlbumParams) error {
	if m.updateAlbumFn != nil {
		return m.updateAlbumFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UpdateAlbumImage(ctx context.Context, arg sqlhandler.UpdateAlbumImageParams) error {
	if m.updateAlbumImageFn != nil {
		return m.updateAlbumImageFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) DeleteAlbum(ctx context.Context, uuid pgtype.UUID) error {
	if m.deleteAlbumFn != nil {
		return m.deleteAlbumFn(ctx, uuid)
	}
	return nil
}
func (m *mockDB) GetMusic(ctx context.Context, uuid pgtype.UUID) (sqlhandler.Music, error) {
	if m.getMusicFn != nil {
		return m.getMusicFn(ctx, uuid)
	}
	return sqlhandler.Music{}, nil
}
func (m *mockDB) GetMusicForArtist(ctx context.Context, arg sqlhandler.GetMusicForArtistParams) ([]sqlhandler.Music, error) {
	if m.getMusicForArtistFn != nil {
		return m.getMusicForArtistFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) GetMusicForAlbum(ctx context.Context, arg sqlhandler.GetMusicForAlbumParams) ([]sqlhandler.Music, error) {
	if m.getMusicForAlbumFn != nil {
		return m.getMusicForAlbumFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) GetMusicForUser(ctx context.Context, arg sqlhandler.GetMusicForUserParams) ([]sqlhandler.Music, error) {
	if m.getMusicForUserFn != nil {
		return m.getMusicForUserFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) CreateMusic(ctx context.Context, arg sqlhandler.CreateMusicParams) error {
	if m.createMusicFn != nil {
		return m.createMusicFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UpdateMusicDetails(ctx context.Context, arg sqlhandler.UpdateMusicDetailsParams) error {
	if m.updateMusicDetailsFn != nil {
		return m.updateMusicDetailsFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UpdateMusicStorage(ctx context.Context, arg sqlhandler.UpdateMusicStorageParams) error {
	if m.updateMusicStorageFn != nil {
		return m.updateMusicStorageFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) DeleteMusic(ctx context.Context, uuid pgtype.UUID) error {
	if m.deleteMusicFn != nil {
		return m.deleteMusicFn(ctx, uuid)
	}
	return nil
}
func (m *mockDB) IncrementPlayCount(ctx context.Context, uuid pgtype.UUID) error {
	if m.incrementPlayCountFn != nil {
		return m.incrementPlayCountFn(ctx, uuid)
	}
	return nil
}
func (m *mockDB) AddListeningHistoryEntry(ctx context.Context, arg sqlhandler.AddListeningHistoryEntryParams) error {
	if m.addListeningHistoryEntryFn != nil {
		return m.addListeningHistoryEntryFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) GetLikesForUser(ctx context.Context, arg sqlhandler.GetLikesForUserParams) ([]sqlhandler.Music, error) {
	if m.getLikesForUserFn != nil {
		return m.getLikesForUserFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) IsLiked(ctx context.Context, arg sqlhandler.IsLikedParams) (bool, error) {
	if m.isLikedFn != nil {
		return m.isLikedFn(ctx, arg)
	}
	return false, nil
}
func (m *mockDB) LikeMusic(ctx context.Context, arg sqlhandler.LikeMusicParams) error {
	if m.likeMusicFn != nil {
		return m.likeMusicFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UnlikeMusic(ctx context.Context, arg sqlhandler.UnlikeMusicParams) error {
	if m.unlikeMusicFn != nil {
		return m.unlikeMusicFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) GetFollowersForUser(ctx context.Context, arg sqlhandler.GetFollowersForUserParams) ([]sqlhandler.PublicUser, error) {
	if m.getFollowersForUserFn != nil {
		return m.getFollowersForUserFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) GetFollowsForUser(ctx context.Context, arg sqlhandler.GetFollowsForUserParams) ([]sqlhandler.PublicUser, error) {
	if m.getFollowsForUserFn != nil {
		return m.getFollowsForUserFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) GetFollowersForArtist(ctx context.Context, arg sqlhandler.GetFollowersForArtistParams) ([]sqlhandler.PublicUser, error) {
	if m.getFollowersForArtistFn != nil {
		return m.getFollowersForArtistFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) FollowUser(ctx context.Context, arg sqlhandler.FollowUserParams) error {
	if m.followUserFn != nil {
		return m.followUserFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UnfollowUser(ctx context.Context, arg sqlhandler.UnfollowUserParams) error {
	if m.unfollowUserFn != nil {
		return m.unfollowUserFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) FollowArtist(ctx context.Context, arg sqlhandler.FollowArtistParams) error {
	if m.followArtistFn != nil {
		return m.followArtistFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UnfollowArtist(ctx context.Context, arg sqlhandler.UnfollowArtistParams) error {
	if m.unfollowArtistFn != nil {
		return m.unfollowArtistFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) GetAllTags(ctx context.Context, arg sqlhandler.GetAllTagsParams) ([]sqlhandler.MusicTag, error) {
	if m.getAllTagsFn != nil {
		return m.getAllTagsFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) GetTag(ctx context.Context, tagName string) (sqlhandler.MusicTag, error) {
	if m.getTagFn != nil {
		return m.getTagFn(ctx, tagName)
	}
	return sqlhandler.MusicTag{}, nil
}
func (m *mockDB) GetMusicForTag(ctx context.Context, arg sqlhandler.GetMusicForTagParams) ([]sqlhandler.Music, error) {
	if m.getMusicForTagFn != nil {
		return m.getMusicForTagFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) GetTagsForMusic(ctx context.Context, arg sqlhandler.GetTagsForMusicParams) ([]sqlhandler.MusicTag, error) {
	if m.getTagsForMusicFn != nil {
		return m.getTagsForMusicFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) CreateTag(ctx context.Context, arg sqlhandler.CreateTagParams) error {
	if m.createTagFn != nil {
		return m.createTagFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) AssignTagToMusic(ctx context.Context, arg sqlhandler.AssignTagToMusicParams) error {
	if m.assignTagToMusicFn != nil {
		return m.assignTagToMusicFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) RemoveTagFromMusic(ctx context.Context, arg sqlhandler.RemoveTagFromMusicParams) error {
	if m.removeTagFromMusicFn != nil {
		return m.removeTagFromMusicFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) GetPlaylist(ctx context.Context, uuid pgtype.UUID) (sqlhandler.Playlist, error) {
	if m.getPlaylistFn != nil {
		return m.getPlaylistFn(ctx, uuid)
	}
	return sqlhandler.Playlist{}, nil
}
func (m *mockDB) GetPlaylistsForUser(ctx context.Context, arg sqlhandler.GetPlaylistsForUserParams) ([]sqlhandler.Playlist, error) {
	if m.getPlaylistsForUserFn != nil {
		return m.getPlaylistsForUserFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) GetPlaylistTracks(ctx context.Context, arg sqlhandler.GetPlaylistTracksParams) ([]sqlhandler.Music, error) {
	if m.getPlaylistTracksFn != nil {
		return m.getPlaylistTracksFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) CreatePlaylist(ctx context.Context, arg sqlhandler.CreatePlaylistParams) error {
	if m.createPlaylistFn != nil {
		return m.createPlaylistFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UpdatePlaylist(ctx context.Context, arg sqlhandler.UpdatePlaylistParams) error {
	if m.updatePlaylistFn != nil {
		return m.updatePlaylistFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UpdatePlaylistImage(ctx context.Context, arg sqlhandler.UpdatePlaylistImageParams) error {
	if m.updatePlaylistImageFn != nil {
		return m.updatePlaylistImageFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) DeletePlaylist(ctx context.Context, uuid pgtype.UUID) error {
	if m.deletePlaylistFn != nil {
		return m.deletePlaylistFn(ctx, uuid)
	}
	return nil
}
func (m *mockDB) AddTrackToPlaylist(ctx context.Context, arg sqlhandler.AddTrackToPlaylistParams) error {
	if m.addTrackToPlaylistFn != nil {
		return m.addTrackToPlaylistFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) RemoveTrackFromPlaylist(ctx context.Context, arg sqlhandler.RemoveTrackFromPlaylistParams) error {
	if m.removeTrackFromPlaylistFn != nil {
		return m.removeTrackFromPlaylistFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) UpdateTrackPosition(ctx context.Context, arg sqlhandler.UpdateTrackPositionParams) error {
	if m.updateTrackPositionFn != nil {
		return m.updateTrackPositionFn(ctx, arg)
	}
	return nil
}
func (m *mockDB) GetListeningHistoryForUser(ctx context.Context, arg sqlhandler.GetListeningHistoryForUserParams) ([]sqlhandler.ListeningHistory, error) {
	if m.getListeningHistoryForUserFn != nil {
		return m.getListeningHistoryForUserFn(ctx, arg)
	}
	return nil, nil
}
func (m *mockDB) GetTopMusicForUser(ctx context.Context, arg sqlhandler.GetTopMusicForUserParams) ([]sqlhandler.GetTopMusicForUserRow, error) {
	if m.getTopMusicForUserFn != nil {
		return m.getTopMusicForUserFn(ctx, arg)
	}
	return nil, nil
}
