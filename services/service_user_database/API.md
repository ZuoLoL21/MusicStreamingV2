# Service User Database - API Documentation

**Base URL:** `http://localhost:8001`
**Service:** User and music data backend
**Authentication:** Service JWT (from gateway)

## Overview

The Service User Database is the core backend service managing all user data, music content, artists, albums, playlists, tags, likes, follows, and listening history.

## Authentication

All endpoints (except `/health` and auth endpoints) require a valid Service JWT in the Authorization header.

### Headers
```
Authorization: Bearer <service-jwt-token>
```

---

## Public Endpoints

### Health Check
```http
GET /health
```

**Authentication:** None

**Response:**
```json
{
  "status": "healthy",
  "service": "service-user-database"
}
```

---

## Authentication Endpoints

### Login
```http
POST /login
```

**Description:** Authenticate user credentials.

**Authentication:** None

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:**
```json
{
  "normal_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000"
}
```

---

### Register
```http
PUT /login
```

**Description:** Create new user account.

**Authentication:** None

**Request Body:**
```json
{
  "email": "newuser@example.com",
  "password": "securepassword",
  "username": "newuser",
  "display_name": "New User"
}
```

**Response:**
```json
{
  "normal_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000"
}
```

---

### Renew Token
```http
POST /renew
```

**Description:** Refresh access token.

**Authentication:** Service JWT required

**Response:**
```json
{
  "normal_token": "eyJhbGciOiJIUzI1NiIs...",
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000"
}
```

---

## User Endpoints

### Get Current User
```http
GET /users/me
```

**Description:** Get authenticated user's profile.

**Authentication:** Service JWT required

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "username": "user123",
  "display_name": "User Name",
  "email": "user@example.com",
  "bio": "Music lover",
  "image_url": "https://storage.example.com/images/user.jpg",
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### Update Profile
```http
POST /users/me
```

**Description:** Update user profile information.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "username": "newusername",
  "display_name": "New Display Name",
  "bio": "Updated bio"
}
```

---

### Update Email
```http
POST /users/me/email
```

**Description:** Update user email address.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "email": "newemail@example.com"
}
```

---

### Update Password
```http
POST /users/me/password
```

**Description:** Change user password.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "old_password": "currentpassword",
  "new_password": "newsecurepassword"
}
```

---

### Update Profile Image
```http
POST /users/me/image
```

**Description:** Update user profile image.

**Authentication:** Service JWT required

**Content-Type:** `multipart/form-data`

**Request:**
```
image: [file upload]
```

---

### Get Public User
```http
GET /users/{uuid}
```

**Description:** Get public profile of any user.

**Authentication:** Service JWT required

**Path Parameters:**
- `uuid` - User UUID

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "username": "user123",
  "display_name": "User Name",
  "bio": "Music lover",
  "image_url": "https://storage.example.com/images/user.jpg",
  "follower_count": 150,
  "following_count": 200,
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### Get User's Artists
```http
GET /users/{uuid}/artists
```

**Description:** Get artists created/managed by user.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20) - Number of results
- `cursor` (optional) - Pagination cursor (timestamp)
- `cursor_id` (optional) - Pagination cursor (UUID)

---

### Get User's Likes
```http
GET /users/{uuid}/likes
```

**Description:** Get music liked by user.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor` (optional) - Timestamp cursor
- `cursor_id` (optional) - UUID cursor

**Response:**
```json
{
  "likes": [
    {
      "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "music_title": "Song Title",
      "artist_name": "Artist Name",
      "liked_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

---

### Get User Followers
```http
GET /users/{uuid}/followers
```

**Description:** Get users following this user.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor`, `cursor_id` (optional) - Pagination cursors

---

### Get Following Users
```http
GET /users/{uuid}/following/users
```

**Description:** Get users that this user follows.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor`, `cursor_id` (optional) - Pagination cursors

---

### Get Followed Artists
```http
GET /users/{uuid}/following/artists
```

**Description:** Get artists followed by user.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor`, `cursor_id` (optional) - Pagination cursors

---

### Follow User
```http
POST /users/{uuid}/follow
```

**Description:** Follow a user.

**Authentication:** Service JWT required

---

### Unfollow User
```http
DELETE /users/{uuid}/follow
```

**Description:** Unfollow a user.

**Authentication:** Service JWT required

---

### Get User's Music
```http
GET /users/{uuid}/music
```

**Description:** Get music uploaded by user's artists.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor`, `cursor_id` (optional) - Pagination cursors

---

### Get User's Playlists
```http
GET /users/{uuid}/playlists
```

**Description:** Get playlists created by user.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor`, `cursor_id` (optional) - Pagination cursors

---

## Artist Endpoints

### Get Artists (Alphabetically)
```http
GET /artists
```

**Description:** Browse all artists alphabetically.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor_name` (optional) - Artist name cursor for pagination
- `cursor` (optional) - Timestamp cursor

---

### Create Artist
```http
PUT /artists
```

**Description:** Create a new artist profile.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "name": "Artist Name",
  "bio": "Artist biography"
}
```

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Artist Name",
  "bio": "Artist biography",
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### Get Artist
```http
GET /artists/{uuid}
```

**Description:** Get artist profile details.

**Authentication:** Service JWT required

**Path Parameters:**
- `uuid` - Artist UUID

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Artist Name",
  "bio": "Artist biography",
  "image_url": "https://storage.example.com/images/artist.jpg",
  "follower_count": 5000,
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### Update Artist Profile
```http
POST /artists/{uuid}
```

**Description:** Update artist information.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "name": "Updated Artist Name",
  "bio": "Updated biography"
}
```

---

### Update Artist Image
```http
POST /artists/{uuid}/image
```

**Description:** Update artist profile picture.

**Authentication:** Service JWT required

**Content-Type:** `multipart/form-data`

---

### Get Artist Members
```http
GET /artists/{uuid}/members
```

**Description:** Get users who manage this artist.

**Authentication:** Service JWT required

**Response:**
```json
{
  "members": [
    {
      "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "username": "user123",
      "role": "owner",
      "joined_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### Add User to Artist
```http
PUT /artists/{uuid}/members/{userUuid}
```

**Description:** Add a user as artist member.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "role": "member"
}
```

---

### Remove User from Artist
```http
DELETE /artists/{uuid}/members/{userUuid}
```

**Description:** Remove user from artist.

**Authentication:** Service JWT required

---

### Change User Role
```http
POST /artists/{uuid}/members/{userUuid}/role
```

**Description:** Update member's role in artist.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "role": "admin"
}
```

---

### Get Artist Albums
```http
GET /artists/{uuid}/albums
```

**Description:** Get albums by artist.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor`, `cursor_id` (optional) - Pagination cursors

---

### Get Artist Music
```http
GET /artists/{uuid}/music
```

**Description:** Get music tracks by artist.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor`, `cursor_id` (optional) - Pagination cursors

---

### Get Artist Followers
```http
GET /artists/{uuid}/followers
```

**Description:** Get users following this artist.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor`, `cursor_id` (optional) - Pagination cursors

---

### Follow Artist
```http
POST /artists/{uuid}/follow
```

**Description:** Follow an artist.

**Authentication:** Service JWT required

---

### Unfollow Artist
```http
DELETE /artists/{uuid}/follow
```

**Description:** Unfollow an artist.

**Authentication:** Service JWT required

---

## Album Endpoints

### Create Album
```http
PUT /albums
```

**Description:** Create a new album.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Album Title",
  "description": "Album description",
  "release_date": "2024-01-01"
}
```

---

### Get Album
```http
GET /albums/{uuid}
```

**Description:** Get album details.

**Authentication:** Service JWT required

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Album Title",
  "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "artist_name": "Artist Name",
  "description": "Album description",
  "image_url": "https://storage.example.com/images/album.jpg",
  "release_date": "2024-01-01",
  "track_count": 12,
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### Update Album
```http
POST /albums/{uuid}
```

**Description:** Update album information.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "title": "Updated Album Title",
  "description": "Updated description"
}
```

---

### Update Album Image
```http
POST /albums/{uuid}/image
```

**Description:** Update album cover art.

**Authentication:** Service JWT required

**Content-Type:** `multipart/form-data`

---

### Delete Album
```http
DELETE /albums/{uuid}
```

**Description:** Delete an album.

**Authentication:** Service JWT required

---

### Get Album Music
```http
GET /albums/{uuid}/music
```

**Description:** Get tracks in album.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor`, `cursor_id` (optional) - Pagination cursors

---

## Music Endpoints

### Create Music
```http
PUT /music
```

**Description:** Upload a new music track.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "album_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Song Title",
  "duration_seconds": 180,
  "storage_path": "s3://bucket/path/to/file.mp3"
}
```

---

### Get Music
```http
GET /music/{uuid}
```

**Description:** Get music track details.

**Authentication:** Service JWT required

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Song Title",
  "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "artist_name": "Artist Name",
  "album_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "album_title": "Album Title",
  "duration_seconds": 180,
  "play_count": 1500,
  "like_count": 200,
  "storage_path": "s3://bucket/path/to/file.mp3",
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### Update Music Details
```http
POST /music/{uuid}
```

**Description:** Update music metadata.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "title": "Updated Song Title",
  "duration_seconds": 185
}
```

---

### Update Music Storage
```http
POST /music/{uuid}/storage
```

**Description:** Update music file storage location.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "storage_path": "s3://bucket/new/path.mp3"
}
```

---

### Delete Music
```http
DELETE /music/{uuid}
```

**Description:** Delete a music track.

**Authentication:** Service JWT required

---

### Increment Play Count
```http
POST /music/{uuid}/play
```

**Description:** Increment play count for a track.

**Authentication:** Service JWT required

---

### Add Listening History Entry
```http
POST /music/{uuid}/listen
```

**Description:** Record a listening event for analytics.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "played_at": "2024-01-01T12:00:00Z",
  "duration_listened_seconds": 180
}
```

---

### Check if Music is Liked
```http
GET /music/{uuid}/liked
```

**Description:** Check if current user has liked this music.

**Authentication:** Service JWT required

**Response:**
```json
{
  "liked": true
}
```

---

### Like Music
```http
POST /music/{uuid}/like
```

**Description:** Like a music track.

**Authentication:** Service JWT required

---

### Unlike Music
```http
DELETE /music/{uuid}/like
```

**Description:** Remove like from music track.

**Authentication:** Service JWT required

---

### Get Music Tags
```http
GET /music/{uuid}/tags
```

**Description:** Get all tags assigned to music.

**Authentication:** Service JWT required

**Response:**
```json
{
  "tags": [
    {
      "name": "rock",
      "description": "Rock music"
    },
    {
      "name": "indie",
      "description": "Independent music"
    }
  ]
}
```

---

### Assign Tag to Music
```http
POST /music/{uuid}/tags/{name}
```

**Description:** Add a tag to music track.

**Authentication:** Service JWT required

---

### Remove Tag from Music
```http
DELETE /music/{uuid}/tags/{name}
```

**Description:** Remove a tag from music track.

**Authentication:** Service JWT required

---

## Tag Endpoints

### Get All Tags
```http
GET /tags
```

**Description:** Browse all available tags.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor_name` (optional) - Tag name cursor for pagination

**Response:**
```json
{
  "tags": [
    {
      "name": "rock",
      "description": "Rock music",
      "music_count": 1500
    }
  ],
  "next_cursor": "rock"
}
```

---

### Create Tag
```http
PUT /tags
```

**Description:** Create a new tag.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "name": "electronic",
  "description": "Electronic music"
}
```

---

### Get Tag
```http
GET /tags/{name}
```

**Description:** Get tag details.

**Authentication:** Service JWT required

**Response:**
```json
{
  "name": "rock",
  "description": "Rock music",
  "music_count": 1500,
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### Get Music for Tag
```http
GET /tags/{name}/music
```

**Description:** Get all music with this tag.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor`, `cursor_id` (optional) - Pagination cursors

---

## Playlist Endpoints

### Create Playlist
```http
PUT /playlists
```

**Description:** Create a new playlist.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "name": "My Playlist",
  "description": "Collection of favorite songs",
  "is_public": true
}
```

---

### Get Playlist
```http
GET /playlists/{uuid}
```

**Description:** Get playlist details.

**Authentication:** Service JWT required

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "name": "My Playlist",
  "description": "Collection of favorite songs",
  "owner_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "owner_username": "user123",
  "is_public": true,
  "track_count": 25,
  "image_url": "https://storage.example.com/images/playlist.jpg",
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### Update Playlist
```http
POST /playlists/{uuid}
```

**Description:** Update playlist information.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "name": "Updated Playlist Name",
  "description": "Updated description",
  "is_public": false
}
```

---

### Update Playlist Image
```http
POST /playlists/{uuid}/image
```

**Description:** Update playlist cover image.

**Authentication:** Service JWT required

**Content-Type:** `multipart/form-data`

---

### Delete Playlist
```http
DELETE /playlists/{uuid}
```

**Description:** Delete a playlist.

**Authentication:** Service JWT required

---

### Get Playlist Tracks
```http
GET /playlists/{uuid}/tracks
```

**Description:** Get tracks in playlist with ordering.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor_pos` (optional) - Position cursor for pagination

**Response:**
```json
{
  "tracks": [
    {
      "track_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "position": 1,
      "title": "Song Title",
      "artist_name": "Artist Name",
      "duration_seconds": 180,
      "added_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor_pos": 21
}
```

---

### Add Track to Playlist
```http
PUT /playlists/{uuid}/tracks/{musicUuid}
```

**Description:** Add a music track to playlist.

**Authentication:** Service JWT required

---

### Remove Track from Playlist
```http
DELETE /playlists/{uuid}/tracks/{musicUuid}
```

**Description:** Remove a track from playlist.

**Authentication:** Service JWT required

---

### Update Track Position
```http
POST /playlists/{uuid}/tracks/{trackUuid}/position
```

**Description:** Reorder track in playlist.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "position": 5
}
```

---

## History Endpoints

### Get Listening History
```http
GET /history
```

**Description:** Get user's listening history.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor`, `cursor_id` (optional) - Pagination cursors

**Response:**
```json
{
  "history": [
    {
      "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Song Title",
      "artist_name": "Artist Name",
      "played_at": "2024-01-01T12:00:00Z",
      "duration_listened_seconds": 180
    }
  ],
  "next_cursor": "2024-01-01T12:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

---

### Get Top Music
```http
GET /history/top
```

**Description:** Get user's most played music.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20)
- `cursor_count` (optional) - Play count cursor for pagination

**Response:**
```json
{
  "top_music": [
    {
      "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Song Title",
      "artist_name": "Artist Name",
      "play_count": 45,
      "last_played": "2024-01-01T12:00:00Z"
    }
  ],
  "next_cursor_count": 20
}
```

---

## Pagination

Most list endpoints support cursor-based pagination:

**Query Parameters:**
- `limit` - Number of results (default: 20, max: 100)
- `cursor` - Timestamp cursor from previous response
- `cursor_id` - UUID cursor from previous response
- `cursor_name` - Name cursor (for tags/artists)
- `cursor_pos` - Position cursor (for playlists)
- `cursor_count` - Count cursor (for top music)

**Response includes:**
- `next_cursor`, `next_cursor_id`, `next_cursor_name`, etc. - Use these for next page

---

## Error Responses

- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Invalid or missing service JWT
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate email)
- `500 Internal Server Error` - Server error
