package di

import (
	"time"

	"github.com/google/uuid"
)

type ListenEventRequest struct {
	UserUUID              uuid.UUID  `json:"user_uuid"`
	MusicUUID             uuid.UUID  `json:"music_uuid"`
	ArtistUUID            uuid.UUID  `json:"artist_uuid"`
	AlbumUUID             *uuid.UUID `json:"album_uuid"`
	ListenDurationSeconds int        `json:"listen_duration_seconds"`
	TrackDurationSeconds  int        `json:"track_duration_seconds"`
	CompletionRatio       float64    `json:"completion_ratio"`
}

type LikeEventRequest struct {
	UserUUID   uuid.UUID `json:"user_uuid"`
	MusicUUID  uuid.UUID `json:"music_uuid"`
	ArtistUUID uuid.UUID `json:"artist_uuid"`
}

type ThemeEventRequest struct {
	MusicUUID uuid.UUID `json:"music_uuid"`
	Theme     string    `json:"theme"`
}

type UserDimRequest struct {
	UserUUID  uuid.UUID `json:"user_uuid"`
	CreatedAt time.Time `json:"created_at"`
	Country   string    `json:"country"`
}
