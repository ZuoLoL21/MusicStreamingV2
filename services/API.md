# MusicStreamingV2 - Unified API Documentation

**Last Updated:** 2026-03-16

This document provides complete API reference for all services in the MusicStreamingV2 platform. For architecture and design decisions, see [ARCHITECTURE.md](ARCHITECTURE.md).

---

## Table of Contents

1. [Quick Reference](#quick-reference)
2. [Authentication Overview](#authentication-overview)
3. [Public Layer: gateway_api](#public-layer-gateway_api)
4. [Data Backend: service_user_database](#data-backend-service_user_database)
5. [Recommendation Backend: gateway_recommendation](#recommendation-backend-gateway_recommendation)
6. [Popularity Backend: service_popularity_system](#popularity-backend-service_popularity_system)
7. [Bandit ML Backend: service_bandit_system](#bandit-ml-backend-service_bandit_system)
8. [Common Patterns](#common-patterns)
9. [Error Responses](#error-responses)

---

## Quick Reference

### Service Access Patterns

| Service | Port | Public Access | Client Access Pattern |
|---------|------|---------------|----------------------|
| gateway_api | 8080 | **Yes** | Direct access with Normal/Refresh JWT |
| service_user_database | 8001 | Internal only | Via gateway_api (automatic Service JWT) |
| gateway_recommendation | 8002 | Internal only | Via gateway_api at `/recommend/*` and `/popular/*` |
| service_popularity_system | 8003 | Internal only | Via gateway_recommendation (automatic routing) |
| service_bandit_system | 8004 | Internal only | Via gateway_recommendation (automatic routing) |

### JWT Token Types

| Type | Subject | Lifetime | Use Case | Header Format |
|------|---------|----------|----------|---------------|
| Normal JWT | `"normal"` | ~10 min | User access token | `Authorization: Bearer <token>` |
| Refresh JWT | `"refresh"` | ~10 days | Token renewal | `Authorization: Bearer <token>` |
| Service JWT | `"service"` | ~2 min | Inter-service auth | `Authorization: Bearer <token>` |

**Client developers:** Only use Normal and Refresh JWTs. Service JWTs are generated automatically by gateways.

---

## Authentication Overview

### Login Flow
```
1. POST /login with credentials
   → gateway_api validates and forwards to service_user_database
   → Returns Normal JWT + Refresh JWT + User UUID
2. Store tokens securely (httpOnly cookies recommended)
3. Use Normal JWT for all protected requests
```

### Protected Request Flow
```
1. Client sends request with Normal JWT to gateway_api
2. Gateway validates JWT and generates Service JWT
3. Gateway forwards to backend with Service JWT
4. Backend validates Service JWT and processes request
5. Response returns through gateway to client
```

### Token Renewal Flow
```
1. POST /renew with Refresh JWT
   → gateway_api validates with service_user_database
   → Returns new Normal JWT
2. Update stored Normal JWT
```

---

## Public Layer: gateway_api

**Base URL:** `http://localhost:8080`
**Role:** Public-facing unified API gateway
**Access:** Direct client access

### Public Endpoints

#### Health Check
```http
GET /health
```

**Authentication:** None

**Response:**
```json
{
  "status": "healthy",
  "service": "gateway-api"
}
```

**Status Codes:**
- `200 OK` - Service is healthy

---

#### Login
```http
POST /login
OPTIONS /login
```

**Description:** Authenticate user and receive JWT tokens. The OPTIONS method is available for CORS preflight requests.

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
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Authentication successful
- `401 Unauthorized` - Invalid credentials
- `502 Bad Gateway` - Backend service unavailable

---

#### Register
```http
PUT /login
```

**Description:** Register a new user account with optional profile image.

**Authentication:** None

**Content-Type:** `multipart/form-data`

**Form Fields:**
- `username` (required, string, min 5 chars) - Username
- `email` (required, string, valid email format) - Email address
- `password` (required, string, min 8 chars) - Password
- `country` (required, string, 2 chars) - ISO 3166-1 alpha-2 country code (e.g., "US", "GB")
- `bio` (optional, string) - User biography
- `image` (optional, file) - Profile image (JPEG, PNG, or WebP, square, max 10MB)

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `201 Created` - Registration successful
- `400 Bad Request` - Invalid input (missing required fields, invalid format, invalid image)
- `409 Conflict` - Email or username already exists
- `415 Unsupported Media Type` - Wrong Content-Type (must be multipart/form-data)
- `500 Internal Server Error` - Server error
- `502 Bad Gateway` - Backend service unavailable

---

### Refresh Token Endpoints

#### Renew Token
```http
POST /renew
```

**Description:** Refresh access token using refresh token.

**Authentication:** Refresh JWT required

**Headers:**
```
Authorization: Bearer <refresh-token>
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Token refreshed successfully
- `401 Unauthorized` - Invalid or expired refresh token
- `502 Bad Gateway` - Backend service unavailable

---

### Protected Endpoints (Require Normal JWT)

All endpoints below require a valid Normal JWT in the `Authorization` header.

#### User Routes
```
GET    /users/*
POST   /users/*
PUT    /users/*
DELETE /users/*
```

**Description:** Proxies all user-related requests to service_user_database.

**Authentication:** Normal JWT required

**Proxied Routes Include:**
- User profile management (`/users/me`, `/users/{uuid}`)
- User profile updates (`/users/me/email`, `/users/me/password`, `/users/me/image`)
- User relationships (`/users/{uuid}/followers`, `/users/{uuid}/following/users`, `/users/{uuid}/following/artists`, `/users/{uuid}/following/check`)
- User content (`/users/{uuid}/music`, `/users/{uuid}/playlists`, `/users/{uuid}/likes`, `/users/{uuid}/artists`)
- User actions (`/users/{uuid}/follow`)

**See:** [service_user_database User Endpoints](#user-endpoints) for detailed documentation.

---

#### Artist Routes
```
GET    /artists/*
POST   /artists/*
PUT    /artists/*
DELETE /artists/*
```

**Description:** Proxies all artist-related requests to service_user_database.

**Authentication:** Normal JWT required

**Proxied Routes Include:**
- Artist profiles (`/artists`, `/artists/{uuid}`)
- Artist images (`/artists/{uuid}/image`)
- Artist members (`/artists/{uuid}/members/*`, `/artists/{uuid}/members/{userUuid}/role`)
- Artist content (`/artists/{uuid}/albums`, `/artists/{uuid}/music`)
- Artist followers (`/artists/{uuid}/followers`, `/artists/{uuid}/follow`)

**See:** [service_user_database Artist Endpoints](#artist-endpoints) for detailed documentation.

---

#### Album Routes
```
GET    /albums/*
POST   /albums/*
PUT    /albums/*
DELETE /albums/*
```

**Description:** Proxies all album-related requests to service_user_database.

**Authentication:** Normal JWT required

**Proxied Routes Include:**
- Album creation and management (`/albums`, `/albums/{uuid}`)
- Album image (`/albums/{uuid}/image`)
- Album music tracks (`/albums/{uuid}/music`)
- Album metadata and images

**See:** [service_user_database Album Endpoints](#album-endpoints) for detailed documentation.

---

#### Music Routes
```
GET    /music/*
POST   /music/*
PUT    /music/*
DELETE /music/*
```

**Description:** Proxies all music track-related requests to service_user_database.

**Authentication:** Normal JWT required

**Proxied Routes Include:**
- Music track management (`/music`, `/music/{uuid}`)
- Music metadata (`/music/{uuid}/storage`, `/music/{uuid}/image`)
- Play tracking (`/music/{uuid}/play`, `/music/{uuid}/listen`)
- Music likes (`/music/{uuid}/like`, `/music/{uuid}/liked`)
- Music tags (`/music/{uuid}/tags`, `/music/{uuid}/tags/{name}`)

**See:** [service_user_database Music Endpoints](#music-endpoints) for detailed documentation.

---

#### Tag Routes
```
GET    /tags/*
POST   /tags/*
DELETE /tags/*
```

**Description:** Proxies all tag-related requests to service_user_database.

**Authentication:** Normal JWT required

**Proxied Routes Include:**
- Tag browsing and management (`/tags`, `/tags/{name}`)
- Music tagging (`/music/{uuid}/tags`, `/tags/{name}/music`)
- Tag-based music discovery

**See:** [service_user_database Tag Endpoints](#tag-endpoints) for detailed documentation.

---

#### Playlist Routes
```
GET    /playlists/*
POST   /playlists/*
PUT    /playlists/*
DELETE /playlists/*
```

**Description:** Proxies all playlist-related requests to service_user_database.

**Authentication:** Normal JWT required

**Proxied Routes Include:**
- Playlist creation and management (`/playlists`, `/playlists/{uuid}`)
- Playlist image (`/playlists/{uuid}/image`)
- Playlist tracks (`/playlists/{uuid}/tracks`, `/playlists/{uuid}/tracks/{trackUuid}/position`)
- Track ordering

**See:** [service_user_database Playlist Endpoints](#playlist-endpoints) for detailed documentation.

---

#### History Routes
```
GET /history
GET /history/top
```

**Description:** Proxies listening history requests to service_user_database.

**Authentication:** Normal JWT required

**Proxied Routes Include:**
- User listening history (`/history`)
- Top played music (`/history/top`)

**See:** [service_user_database History Endpoints](#history-endpoints) for detailed documentation.

---

#### Search Routes
```
GET /search/users
GET /search/artists
GET /search/albums
GET /search/music
GET /search/playlists
```

**Description:** Proxies search requests to service_user_database. Search across users, artists, albums, music, and playlists using fuzzy text matching.

**Authentication:** Normal JWT required

**See:** [service_user_database Search Endpoints](#search-endpoints) for detailed documentation.

---

#### Recommendation Routes
```
POST /recommend/theme
```

**Description:** Get personalized theme recommendation using multi-armed bandit algorithm.

**Authentication:** Normal JWT required

**Proxied to:** gateway_recommendation → service_bandit_system

**See:** [gateway_recommendation Recommend Theme](#recommend-theme) for detailed documentation.

---

#### Popularity Routes
```
GET /popular/songs/all-time
GET /popular/artists/all-time
GET /popular/themes/all-time
GET /popular/songs/theme/{theme}
GET /popular/songs/timeframe
GET /popular/artists/timeframe
GET /popular/themes/timeframe
GET /popular/songs/theme/{theme}/timeframe
```

**Description:** Get popularity rankings and trending content.

**Authentication:** Normal JWT required

**Proxied to:** gateway_recommendation → service_popularity_system

**See:** [gateway_recommendation Popularity Endpoints](#all-time-popularity-endpoints) for detailed documentation.

---

#### File Routes
```
GET /files/public/*
GET /files/private/*
```

**Description:** Serve static files (images, audio).

**Authentication:**
- `/files/public/*` - None (public access)
- `/files/private/*` - Normal JWT required

**Proxied to:** service_user_database file handler

---

#### Event Routes
```
POST /events/*
```

**Description:** Track user events for analytics.

**Authentication:** Normal JWT required

**Proxied to:** service_event_ingestion

---

## Data Backend: service_user_database

**Base URL:** `http://localhost:8001` (internal)
**Client Access:** Via `gateway_api` at `http://localhost:8080`
**Authentication:** Service JWT (auto-generated by gateway_api)

**Note for developers:** All endpoints below are accessed through gateway_api. The gateway automatically handles Service JWT generation.

---

### User Endpoints

#### Get Current User
```http
GET /users/me
```

**Client Access:** `GET http://localhost:8080/users/me`

**Description:** Get authenticated user's profile.

**Authentication:** Normal JWT (via gateway_api)

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "username": "user123",
  "display_name": "User Name",
  "email": "user@example.com",
  "bio": "Music lover",
  "image_url": "http://localhost:8001/files/public/pictures-profile/user.jpg",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `401 Unauthorized` - Invalid JWT
- `500 Internal Server Error` - Server error

---

#### Update Profile
```http
POST /users/me
```

**Client Access:** `POST http://localhost:8080/users/me`

**Description:** Update user profile information.

**Authentication:** Normal JWT (via gateway_api)

**Request Body:**
```json
{
  "username": "newusername",
  "display_name": "New Display Name",
  "bio": "Updated bio"
}
```

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "username": "newusername",
  "display_name": "New Display Name",
  "bio": "Updated bio"
}
```

**Status Codes:**
- `200 OK` - Profile updated
- `400 Bad Request` - Invalid input
- `409 Conflict` - Username already taken
- `500 Internal Server Error` - Server error

---

#### Update Email
```http
POST /users/me/email
```

**Client Access:** `POST http://localhost:8080/users/me/email`

**Description:** Update user email address.

**Authentication:** Normal JWT (via gateway_api)

**Request Body:**
```json
{
  "email": "newemail@example.com"
}
```

**Status Codes:**
- `200 OK` - Email updated
- `400 Bad Request` - Invalid email format
- `409 Conflict` - Email already in use
- `500 Internal Server Error` - Server error

---

#### Update Password
```http
POST /users/me/password
```

**Client Access:** `POST http://localhost:8080/users/me/password`

**Description:** Change user password.

**Authentication:** Normal JWT (via gateway_api)

**Request Body:**
```json
{
  "old_password": "currentpassword",
  "new_password": "newsecurepassword"
}
```

**Status Codes:**
- `200 OK` - Password updated
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Incorrect old password
- `500 Internal Server Error` - Server error

---

#### Update Profile Image
```http
POST /users/me/image
```

**Client Access:** `POST http://localhost:8080/users/me/image`

**Description:** Update user profile image.

**Authentication:** Normal JWT (via gateway_api)

**Content-Type:** `multipart/form-data`

**Request:**
```
image: [file upload]
```

**Response:**
```json
{
  "image_url": "http://localhost:8001/files/public/pictures-profile/abc123.jpg"
}
```

**Status Codes:**
- `200 OK` - Image uploaded
- `400 Bad Request` - Invalid file format or size
- `500 Internal Server Error` - Upload failed

---

#### Get Public User Profile
```http
GET /users/{uuid}
```

**Client Access:** `GET http://localhost:8080/users/{uuid}`

**Description:** Get public profile of any user.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - User UUID

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "username": "user123",
  "display_name": "User Name",
  "bio": "Music lover",
  "image_url": "http://localhost:8001/files/public/pictures-profile/user.jpg",
  "follower_count": 150,
  "following_count": 200,
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - User not found
- `500 Internal Server Error` - Server error

---

#### Get User's Artists
```http
GET /users/{uuid}/artists
```

**Client Access:** `GET http://localhost:8080/users/{uuid}/artists`

**Description:** Get artists created/managed by user.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - User UUID

**Query Parameters:**
- `limit` (optional, default: 20, max: 100) - Number of results
- `cursor` (optional) - Timestamp cursor from previous response
- `cursor_id` (optional) - UUID cursor from previous response

**Response:**
```json
{
  "artists": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Artist Name",
      "bio": "Artist bio",
      "image_url": "http://localhost:8001/files/public/pictures-artist/artist.jpg",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - User not found
- `500 Internal Server Error` - Server error

---

#### Get User's Likes
```http
GET /users/{uuid}/likes
```

**Client Access:** `GET http://localhost:8080/users/{uuid}/likes`

**Description:** Get music liked by user.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - User UUID

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
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

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - User not found
- `500 Internal Server Error` - Server error

---

#### Get User Followers
```http
GET /users/{uuid}/followers
```

**Client Access:** `GET http://localhost:8080/users/{uuid}/followers`

**Description:** Get users following this user.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - User UUID

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `cursor` (optional) - Timestamp cursor
- `cursor_id` (optional) - UUID cursor

**Response:**
```json
{
  "followers": [
    {
      "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "username": "follower123",
      "display_name": "Follower Name",
      "image_url": "http://localhost:8001/files/public/pictures-profile/follower.jpg",
      "followed_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - User not found
- `500 Internal Server Error` - Server error

---

#### Get Following Users
```http
GET /users/{uuid}/following/users
```

**Client Access:** `GET http://localhost:8080/users/{uuid}/following/users`

**Description:** Get users that this user follows.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - User UUID

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `cursor` (optional) - Timestamp cursor
- `cursor_id` (optional) - UUID cursor

**Response:**
```json
{
  "following": [
    {
      "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "username": "followed123",
      "display_name": "Followed User",
      "image_url": "http://localhost:8001/files/public/pictures-profile/followed.jpg",
      "followed_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - User not found
- `500 Internal Server Error` - Server error

---

#### Get Followed Artists
```http
GET /users/{uuid}/following/artists
```

**Client Access:** `GET http://localhost:8080/users/{uuid}/following/artists`

**Description:** Get artists followed by user.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - User UUID

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `cursor` (optional) - Timestamp cursor
- `cursor_id` (optional) - UUID cursor

**Response:**
```json
{
  "artists": [
    {
      "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Artist Name",
      "image_url": "http://localhost:8001/files/public/pictures-artist/artist.jpg",
      "followed_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - User not found
- `500 Internal Server Error` - Server error

---

#### Check If Following User
```http
GET /users/{uuid}/following/check
```

**Client Access:** `GET http://localhost:8080/users/{uuid}/following/check`

**Description:** Check if the authenticated user is following the specified user.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - User UUID to check

**Response:**
```json
{
  "is_following": true
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - User not found
- `500 Internal Server Error` - Server error

---

#### Follow User
```http
POST /users/{uuid}/follow
```

**Client Access:** `POST http://localhost:8080/users/{uuid}/follow`

**Description:** Follow a user.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - User UUID to follow

**Status Codes:**
- `201 Created` - Follow successful
- `400 Bad Request` - Cannot follow yourself
- `404 Not Found` - User not found
- `409 Conflict` - Already following
- `500 Internal Server Error` - Server error

---

#### Unfollow User
```http
DELETE /users/{uuid}/follow
```

**Client Access:** `DELETE http://localhost:8080/users/{uuid}/follow`

**Description:** Unfollow a user.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - User UUID to unfollow

**Status Codes:**
- `200 OK` - Unfollow successful
- `404 Not Found` - User not found or not following
- `500 Internal Server Error` - Server error

---

#### Get User's Music
```http
GET /users/{uuid}/music
```

**Client Access:** `GET http://localhost:8080/users/{uuid}/music`

**Description:** Get music uploaded by user's artists.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - User UUID

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `cursor` (optional) - Timestamp cursor
- `cursor_id` (optional) - UUID cursor

**Response:**
```json
{
  "music": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Song Title",
      "artist_name": "Artist Name",
      "duration_seconds": 180,
      "play_count": 1500,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - User not found
- `500 Internal Server Error` - Server error

---

#### Get User's Playlists
```http
GET /users/{uuid}/playlists
```

**Client Access:** `GET http://localhost:8080/users/{uuid}/playlists`

**Description:** Get playlists created by user.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - User UUID

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `cursor` (optional) - Timestamp cursor
- `cursor_id` (optional) - UUID cursor

**Response:**
```json
{
  "playlists": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174000",
      "name": "My Playlist",
      "description": "Favorite songs",
      "is_public": true,
      "track_count": 25,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - User not found
- `500 Internal Server Error` - Server error

---

### Artist Endpoints

#### Get Artists (Alphabetically)
```http
GET /artists
```

**Client Access:** `GET http://localhost:8080/artists`

**Description:** Browse all artists alphabetically.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `cursor_name` (optional) - Artist name cursor for pagination
- `cursor` (optional) - Timestamp cursor

**Response:**
```json
{
  "artists": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Artist Name",
      "bio": "Artist biography",
      "image_url": "http://localhost:8001/files/public/pictures-artist/artist.jpg",
      "follower_count": 5000,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor_name": "Artist Name",
  "next_cursor": "2024-01-01T00:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `500 Internal Server Error` - Server error

---

#### Create Artist
```http
PUT /artists
```

**Client Access:** `PUT http://localhost:8080/artists`

**Description:** Create a new artist profile with optional profile image.

**Authentication:** Normal JWT (via gateway_api)

**Content-Type:** `multipart/form-data`

**Form Fields:**
- `artist_name` (required, string, 1-200 chars) - Artist name
- `bio` (optional, string) - Artist biography
- `image` (optional, file) - Profile image (JPEG, PNG, or WebP, square, max 10MB)

**Response:**
```text
artist created
```

**Status Codes:**
- `201 Created` - Artist created
- `400 Bad Request` - Invalid input (missing required fields, invalid image)
- `415 Unsupported Media Type` - Wrong Content-Type (must be multipart/form-data)
- `500 Internal Server Error` - Server error
- `400 Bad Request` - Invalid input
- `409 Conflict` - Artist name already exists
- `500 Internal Server Error` - Server error

---

#### Get Artist
```http
GET /artists/{uuid}
```

**Client Access:** `GET http://localhost:8080/artists/{uuid}`

**Description:** Get artist profile details.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Artist UUID

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Artist Name",
  "bio": "Artist biography",
  "image_url": "http://localhost:8001/files/public/pictures-artist/artist.jpg",
  "follower_count": 5000,
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Artist not found
- `500 Internal Server Error` - Server error

---

#### Update Artist Profile
```http
POST /artists/{uuid}
```

**Client Access:** `POST http://localhost:8080/artists/{uuid}`

**Description:** Update artist information.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Artist UUID

**Request Body:**
```json
{
  "name": "Updated Artist Name",
  "bio": "Updated biography"
}
```

**Status Codes:**
- `200 OK` - Artist updated
- `400 Bad Request` - Invalid input
- `403 Forbidden` - Not authorized (must be artist member)
- `404 Not Found` - Artist not found
- `500 Internal Server Error` - Server error

---

#### Update Artist Image
```http
POST /artists/{uuid}/image
```

**Client Access:** `POST http://localhost:8080/artists/{uuid}/image`

**Description:** Update artist profile picture.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Artist UUID

**Content-Type:** `multipart/form-data`

**Request:**
```
image: [file upload]
```

**Response:**
```json
{
  "image_url": "http://localhost:8001/files/public/pictures-artist/abc123.jpg"
}
```

**Status Codes:**
- `200 OK` - Image uploaded
- `400 Bad Request` - Invalid file
- `403 Forbidden` - Not authorized
- `404 Not Found` - Artist not found
- `500 Internal Server Error` - Server error

---

#### Get Artist Members
```http
GET /artists/{uuid}/members
```

**Client Access:** `GET http://localhost:8080/artists/{uuid}/members`

**Description:** Get users who manage this artist.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Artist UUID

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

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Artist not found
- `500 Internal Server Error` - Server error

---

#### Add User to Artist
```http
PUT /artists/{uuid}/members/{userUuid}
```

**Client Access:** `PUT http://localhost:8080/artists/{uuid}/members/{userUuid}`

**Description:** Add a user as artist member.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Artist UUID
- `userUuid` - User UUID to add

**Request Body:**
```json
{
  "role": "member"
}
```

**Status Codes:**
- `201 Created` - Member added
- `400 Bad Request` - Invalid role
- `403 Forbidden` - Not authorized (must be owner/admin)
- `404 Not Found` - Artist or user not found
- `409 Conflict` - User already a member
- `500 Internal Server Error` - Server error

---

#### Remove User from Artist
```http
DELETE /artists/{uuid}/members/{userUuid}
```

**Client Access:** `DELETE http://localhost:8080/artists/{uuid}/members/{userUuid}`

**Description:** Remove user from artist.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Artist UUID
- `userUuid` - User UUID to remove

**Status Codes:**
- `200 OK` - Member removed
- `403 Forbidden` - Not authorized
- `404 Not Found` - Artist, user, or membership not found
- `500 Internal Server Error` - Server error

---

#### Change User Role
```http
POST /artists/{uuid}/members/{userUuid}/role
```

**Client Access:** `POST http://localhost:8080/artists/{uuid}/members/{userUuid}/role`

**Description:** Update member's role in artist.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Artist UUID
- `userUuid` - User UUID

**Request Body:**
```json
{
  "role": "admin"
}
```

**Valid Roles:**
- `owner` - Full control
- `admin` - Manage members and content
- `member` - Upload content only

**Status Codes:**
- `200 OK` - Role updated
- `400 Bad Request` - Invalid role
- `403 Forbidden` - Not authorized
- `404 Not Found` - Artist or membership not found
- `500 Internal Server Error` - Server error

---

#### Get Artist Albums
```http
GET /artists/{uuid}/albums
```

**Client Access:** `GET http://localhost:8080/artists/{uuid}/albums`

**Description:** Get albums by artist.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Artist UUID

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `cursor` (optional) - Timestamp cursor
- `cursor_id` (optional) - UUID cursor

**Response:**
```json
{
  "albums": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Album Title",
      "description": "Album description",
      "image_url": "http://localhost:8001/files/public/pictures-album/album.jpg",
      "release_date": "2024-01-01",
      "track_count": 12,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Artist not found
- `500 Internal Server Error` - Server error

---

#### Get Artist Music
```http
GET /artists/{uuid}/music
```

**Client Access:** `GET http://localhost:8080/artists/{uuid}/music`

**Description:** Get music tracks by artist.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Artist UUID

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `cursor` (optional) - Timestamp cursor
- `cursor_id` (optional) - UUID cursor

**Response:**
```json
{
  "music": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Song Title",
      "album_title": "Album Title",
      "duration_seconds": 180,
      "play_count": 1500,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Artist not found
- `500 Internal Server Error` - Server error

---

#### Get Artist Followers
```http
GET /artists/{uuid}/followers
```

**Client Access:** `GET http://localhost:8080/artists/{uuid}/followers`

**Description:** Get users following this artist.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Artist UUID

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `cursor` (optional) - Timestamp cursor
- `cursor_id` (optional) - UUID cursor

**Response:**
```json
{
  "followers": [
    {
      "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "username": "user123",
      "display_name": "User Name",
      "followed_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Artist not found
- `500 Internal Server Error` - Server error

---

#### Follow Artist
```http
POST /artists/{uuid}/follow
```

**Client Access:** `POST http://localhost:8080/artists/{uuid}/follow`

**Description:** Follow an artist.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Artist UUID

**Status Codes:**
- `201 Created` - Follow successful
- `404 Not Found` - Artist not found
- `409 Conflict` - Already following
- `500 Internal Server Error` - Server error

---

#### Unfollow Artist
```http
DELETE /artists/{uuid}/follow
```

**Client Access:** `DELETE http://localhost:8080/artists/{uuid}/follow`

**Description:** Unfollow an artist.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Artist UUID

**Status Codes:**
- `200 OK` - Unfollow successful
- `404 Not Found` - Artist not found or not following
- `500 Internal Server Error` - Server error

---

### Album Endpoints

#### Create Album
```http
PUT /albums
```

**Client Access:** `PUT http://localhost:8080/albums`

**Description:** Create a new album with optional cover image.

**Authentication:** Normal JWT (via gateway_api)

**Content-Type:** `multipart/form-data`

**Form Fields:**
- `artist_uuid` (required, string) - Artist UUID
- `original_name` (required, string, 1-200 chars) - Album title
- `description` (optional, string) - Album description
- `image` (optional, file) - Album cover image (JPEG, PNG, or WebP, square, max 10MB)

**Response:**
```text
album created
```

**Status Codes:**
- `201 Created` - Album created
- `400 Bad Request` - Invalid input (missing required fields, invalid UUID, invalid image)
- `403 Forbidden` - Not authorized (must be artist member)
- `415 Unsupported Media Type` - Wrong Content-Type (must be multipart/form-data)
- `500 Internal Server Error` - Server error
```

**Status Codes:**
- `201 Created` - Album created
- `400 Bad Request` - Invalid input
- `403 Forbidden` - Not authorized (must be artist member)
- `404 Not Found` - Artist not found
- `500 Internal Server Error` - Server error

---

#### Get Album
```http
GET /albums/{uuid}
```

**Client Access:** `GET http://localhost:8080/albums/{uuid}`

**Description:** Get album details.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Album UUID

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Album Title",
  "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "artist_name": "Artist Name",
  "description": "Album description",
  "image_url": "http://localhost:8001/files/public/pictures-album/album.jpg",
  "release_date": "2024-01-01",
  "track_count": 12,
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Album not found
- `500 Internal Server Error` - Server error

---

#### Update Album
```http
POST /albums/{uuid}
```

**Client Access:** `POST http://localhost:8080/albums/{uuid}`

**Description:** Update album information.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Album UUID

**Request Body:**
```json
{
  "title": "Updated Album Title",
  "description": "Updated description"
}
```

**Status Codes:**
- `200 OK` - Album updated
- `400 Bad Request` - Invalid input
- `403 Forbidden` - Not authorized
- `404 Not Found` - Album not found
- `500 Internal Server Error` - Server error

---

#### Update Album Image
```http
POST /albums/{uuid}/image
```

**Client Access:** `POST http://localhost:8080/albums/{uuid}/image`

**Description:** Update album cover art.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Album UUID

**Content-Type:** `multipart/form-data`

**Request:**
```
image: [file upload]
```

**Response:**
```json
{
  "image_url": "http://localhost:8001/files/public/pictures-album/abc123.jpg"
}
```

**Status Codes:**
- `200 OK` - Image uploaded
- `400 Bad Request` - Invalid file
- `403 Forbidden` - Not authorized
- `404 Not Found` - Album not found
- `500 Internal Server Error` - Server error

---

#### Delete Album
```http
DELETE /albums/{uuid}
```

**Client Access:** `DELETE http://localhost:8080/albums/{uuid}`

**Description:** Delete an album.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Album UUID

**Status Codes:**
- `200 OK` - Album deleted
- `403 Forbidden` - Not authorized
- `404 Not Found` - Album not found
- `500 Internal Server Error` - Server error

---

#### Get Album Music
```http
GET /albums/{uuid}/music
```

**Client Access:** `GET http://localhost:8080/albums/{uuid}/music`

**Description:** Get tracks in album.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Album UUID

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `cursor` (optional) - Timestamp cursor
- `cursor_id` (optional) - UUID cursor

**Response:**
```json
{
  "music": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Song Title",
      "artist_name": "Artist Name",
      "duration_seconds": 180,
      "play_count": 1500,
      "track_number": 1,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Album not found
- `500 Internal Server Error` - Server error

---

### Music Endpoints

#### Create Music
```http
PUT /music
```

**Client Access:** `PUT http://localhost:8080/music`

**Description:** Upload a new music track with audio file and optional cover image.

**Authentication:** Normal JWT (via gateway_api)

**Content-Type:** `multipart/form-data`

**Form Fields:**
- `artist_uuid` (required, string) - Artist UUID
- `song_name` (required, string, 1-500 chars) - Song title
- `duration_seconds` (required, integer) - Duration in seconds (must be > 0)
- `audio` (required, file) - Audio file (MP3, WAV, FLAC, OGG, AAC, M4A)
- `in_album` (optional, string) - Album UUID to add this track to
- `image` (optional, file) - Track cover image (JPEG, PNG, or WebP, square, max 10MB)

**Response:**
```text
music created
```

**Status Codes:**
- `201 Created` - Music track created
- `400 Bad Request` - Invalid input (missing required fields, invalid UUID, invalid audio/image file)
- `403 Forbidden` - Not authorized (must be artist member)
- `415 Unsupported Media Type` - Wrong Content-Type (must be multipart/form-data)
- `500 Internal Server Error` - Server error
}
```

**Status Codes:**
- `201 Created` - Music created
- `400 Bad Request` - Invalid input
- `403 Forbidden` - Not authorized
- `404 Not Found` - Artist or album not found
- `500 Internal Server Error` - Server error

---

#### Get Music
```http
GET /music/{uuid}
```

**Client Access:** `GET http://localhost:8080/music/{uuid}`

**Description:** Get music track details.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Music UUID

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
  "image_url": "http://localhost:8001/files/public/pictures-music/music.jpg",
  "audio_url": "http://localhost:8001/files/public/audio/abc123.mp3",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Music not found
- `500 Internal Server Error` - Server error

---

#### Update Music Details
```http
POST /music/{uuid}
```

**Client Access:** `POST http://localhost:8080/music/{uuid}`

**Description:** Update music metadata.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Music UUID

**Request Body:**
```json
{
  "title": "Updated Song Title",
  "duration_seconds": 185
}
```

**Status Codes:**
- `200 OK` - Music updated
- `400 Bad Request` - Invalid input
- `403 Forbidden` - Not authorized
- `404 Not Found` - Music not found
- `500 Internal Server Error` - Server error

---

#### Update Music Storage
```http
POST /music/{uuid}/storage
```

**Client Access:** `POST http://localhost:8080/music/{uuid}/storage`

**Description:** Update music file storage location.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Music UUID

**Request Body:**
```json
{
  "storage_path": "audio/new-path.mp3"
}
```

**Status Codes:**
- `200 OK` - Storage path updated
- `400 Bad Request` - Invalid path
- `403 Forbidden` - Not authorized
- `404 Not Found` - Music not found
- `500 Internal Server Error` - Server error

---

#### Update Music Image
```http
POST /music/{uuid}/image
```

**Client Access:** `POST http://localhost:8080/music/{uuid}/image`

**Description:** Update music cover art.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Music UUID

**Content-Type:** `multipart/form-data`

**Request:**
```
image: [file upload]
```

**Response:**
```json
{
  "image_url": "http://localhost:8001/files/public/pictures-music/abc123.jpg"
}
```

**Status Codes:**
- `200 OK` - Image uploaded
- `400 Bad Request` - Invalid file
- `403 Forbidden` - Not authorized
- `404 Not Found` - Music not found
- `500 Internal Server Error` - Server error

---

#### Delete Music
```http
DELETE /music/{uuid}
```

**Client Access:** `DELETE http://localhost:8080/music/{uuid}`

**Description:** Delete a music track.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Music UUID

**Status Codes:**
- `200 OK` - Music deleted
- `403 Forbidden` - Not authorized
- `404 Not Found` - Music not found
- `500 Internal Server Error` - Server error

---

#### Increment Play Count
```http
POST /music/{uuid}/play
```

**Client Access:** `POST http://localhost:8080/music/{uuid}/play`

**Description:** Increment play count for a track.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Music UUID

**Status Codes:**
- `200 OK` - Play count incremented
- `404 Not Found` - Music not found
- `500 Internal Server Error` - Server error

---

#### Add Listening History Entry
```http
POST /music/{uuid}/listen
```

**Client Access:** `POST http://localhost:8080/music/{uuid}/listen`

**Description:** Record a listening event for analytics.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Music UUID

**Request Body:**
```json
{
  "played_at": "2024-01-01T12:00:00Z",
  "duration_listened_seconds": 180
}
```

**Status Codes:**
- `201 Created` - History entry created
- `400 Bad Request` - Invalid input
- `404 Not Found` - Music not found
- `500 Internal Server Error` - Server error

---

#### Check if Music is Liked
```http
GET /music/{uuid}/liked
```

**Client Access:** `GET http://localhost:8080/music/{uuid}/liked`

**Description:** Check if current user has liked this music.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Music UUID

**Response:**
```json
{
  "liked": true
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Music not found
- `500 Internal Server Error` - Server error

---

#### Like Music
```http
POST /music/{uuid}/like
```

**Client Access:** `POST http://localhost:8080/music/{uuid}/like`

**Description:** Like a music track.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Music UUID

**Status Codes:**
- `201 Created` - Like successful
- `404 Not Found` - Music not found
- `409 Conflict` - Already liked
- `500 Internal Server Error` - Server error

---

#### Unlike Music
```http
DELETE /music/{uuid}/like
```

**Client Access:** `DELETE http://localhost:8080/music/{uuid}/like`

**Description:** Remove like from music track.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Music UUID

**Status Codes:**
- `200 OK` - Unlike successful
- `404 Not Found` - Music not found or not liked
- `500 Internal Server Error` - Server error

---

#### Get Music Tags
```http
GET /music/{uuid}/tags
```

**Client Access:** `GET http://localhost:8080/music/{uuid}/tags`

**Description:** Get all tags assigned to music.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Music UUID

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

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Music not found
- `500 Internal Server Error` - Server error

---

#### Assign Tag to Music
```http
POST /music/{uuid}/tags/{name}
```

**Client Access:** `POST http://localhost:8080/music/{uuid}/tags/{name}`

**Description:** Add a tag to music track.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Music UUID
- `name` - Tag name

**Status Codes:**
- `201 Created` - Tag assigned
- `404 Not Found` - Music not found
- `409 Conflict` - Tag already assigned
- `500 Internal Server Error` - Server error

---

#### Remove Tag from Music
```http
DELETE /music/{uuid}/tags/{name}
```

**Client Access:** `DELETE http://localhost:8080/music/{uuid}/tags/{name}`

**Description:** Remove a tag from music track.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Music UUID
- `name` - Tag name

**Status Codes:**
- `200 OK` - Tag removed
- `404 Not Found` - Music or tag not found
- `500 Internal Server Error` - Server error

---

### Tag Endpoints

#### Get All Tags
```http
GET /tags
```

**Client Access:** `GET http://localhost:8080/tags`

**Description:** Browse all available tags.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
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

**Status Codes:**
- `200 OK` - Success
- `500 Internal Server Error` - Server error

---

#### Get Tag
```http
GET /tags/{name}
```

**Client Access:** `GET http://localhost:8080/tags/{name}`

**Description:** Get tag details.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `name` - Tag name

**Response:**
```json
{
  "name": "rock",
  "description": "Rock music",
  "music_count": 1500,
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Tag not found
- `500 Internal Server Error` - Server error

---

#### Get Music for Tag
```http
GET /tags/{name}/music
```

**Client Access:** `GET http://localhost:8080/tags/{name}/music`

**Description:** Get all music with this tag.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `name` - Tag name

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `cursor` (optional) - Timestamp cursor
- `cursor_id` (optional) - UUID cursor

**Response:**
```json
{
  "music": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Song Title",
      "artist_name": "Artist Name",
      "duration_seconds": 180,
      "play_count": 1500,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Tag not found
- `500 Internal Server Error` - Server error

---

### Playlist Endpoints

#### Create Playlist
```http
PUT /playlists
```

**Client Access:** `PUT http://localhost:8080/playlists`

**Description:** Create a new playlist with optional cover image.

**Authentication:** Normal JWT (via gateway_api)

**Content-Type:** `multipart/form-data`

**Form Fields:**
- `original_name` (required, string, 1-200 chars) - Playlist name
- `description` (optional, string) - Playlist description
- `is_public` (optional, boolean) - Whether playlist is public (default: false)
- `image` (optional, file) - Playlist cover image (JPEG, PNG, or WebP, square, max 10MB)

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "name": "My Playlist",
  "description": "Collection of favorite songs",
  "is_public": true,
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Status Codes:**
- `201 Created` - Playlist created
- `400 Bad Request` - Invalid input (missing required fields, invalid image)
- `415 Unsupported Media Type` - Wrong Content-Type (must be multipart/form-data)
- `500 Internal Server Error` - Server error

---

#### Get Playlist
```http
GET /playlists/{uuid}
```

**Client Access:** `GET http://localhost:8080/playlists/{uuid}`

**Description:** Get playlist details.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Playlist UUID

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
  "image_url": "http://localhost:8001/files/public/pictures-playlist/playlist.jpg",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `403 Forbidden` - Private playlist, not owner
- `404 Not Found` - Playlist not found
- `500 Internal Server Error` - Server error

---

#### Update Playlist
```http
POST /playlists/{uuid}
```

**Client Access:** `POST http://localhost:8080/playlists/{uuid}`

**Description:** Update playlist information.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Playlist UUID

**Request Body:**
```json
{
  "name": "Updated Playlist Name",
  "description": "Updated description",
  "is_public": false
}
```

**Status Codes:**
- `200 OK` - Playlist updated
- `400 Bad Request` - Invalid input
- `403 Forbidden` - Not owner
- `404 Not Found` - Playlist not found
- `500 Internal Server Error` - Server error

---

#### Update Playlist Image
```http
POST /playlists/{uuid}/image
```

**Client Access:** `POST http://localhost:8080/playlists/{uuid}/image`

**Description:** Update playlist cover image.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Playlist UUID

**Content-Type:** `multipart/form-data`

**Request:**
```
image: [file upload]
```

**Response:**
```json
{
  "image_url": "http://localhost:8001/files/public/pictures-playlist/abc123.jpg"
}
```

**Status Codes:**
- `200 OK` - Image uploaded
- `400 Bad Request` - Invalid file
- `403 Forbidden` - Not owner
- `404 Not Found` - Playlist not found
- `500 Internal Server Error` - Server error

---

#### Delete Playlist
```http
DELETE /playlists/{uuid}
```

**Client Access:** `DELETE http://localhost:8080/playlists/{uuid}`

**Description:** Delete a playlist.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Playlist UUID

**Status Codes:**
- `200 OK` - Playlist deleted
- `403 Forbidden` - Not owner
- `404 Not Found` - Playlist not found
- `500 Internal Server Error` - Server error

---

#### Get Playlist Tracks
```http
GET /playlists/{uuid}/tracks
```

**Client Access:** `GET http://localhost:8080/playlists/{uuid}/tracks`

**Description:** Get tracks in playlist with ordering.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Playlist UUID

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
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

**Status Codes:**
- `200 OK` - Success
- `403 Forbidden` - Private playlist, not owner
- `404 Not Found` - Playlist not found
- `500 Internal Server Error` - Server error

---

#### Add Track to Playlist
```http
PUT /playlists/{uuid}/tracks/{musicUuid}
```

**Client Access:** `PUT http://localhost:8080/playlists/{uuid}/tracks/{musicUuid}`

**Description:** Add a music track to playlist.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Playlist UUID
- `musicUuid` - Music UUID to add

**Status Codes:**
- `201 Created` - Track added
- `403 Forbidden` - Not owner
- `404 Not Found` - Playlist or music not found
- `409 Conflict` - Track already in playlist
- `500 Internal Server Error` - Server error

---

#### Remove Track from Playlist
```http
DELETE /playlists/{uuid}/tracks/{musicUuid}
```

**Client Access:** `DELETE http://localhost:8080/playlists/{uuid}/tracks/{musicUuid}`

**Description:** Remove a track from playlist.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Playlist UUID
- `musicUuid` - Music UUID to remove

**Status Codes:**
- `200 OK` - Track removed
- `403 Forbidden` - Not owner
- `404 Not Found` - Playlist, music, or track not found
- `500 Internal Server Error` - Server error

---

#### Update Track Position
```http
POST /playlists/{uuid}/tracks/{trackUuid}/position
```

**Client Access:** `POST http://localhost:8080/playlists/{uuid}/tracks/{trackUuid}/position`

**Description:** Reorder track in playlist.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `uuid` - Playlist UUID
- `trackUuid` - Track UUID (not music UUID)

**Request Body:**
```json
{
  "position": 5
}
```

**Status Codes:**
- `200 OK` - Position updated
- `400 Bad Request` - Invalid position
- `403 Forbidden` - Not owner
- `404 Not Found` - Playlist or track not found
- `500 Internal Server Error` - Server error

---

### History Endpoints

#### Get Listening History
```http
GET /history
```

**Client Access:** `GET http://localhost:8080/history`

**Description:** Get user's listening history.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `cursor` (optional) - Timestamp cursor
- `cursor_id` (optional) - UUID cursor

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

**Status Codes:**
- `200 OK` - Success
- `500 Internal Server Error` - Server error

---

#### Get Top Music
```http
GET /history/top
```

**Client Access:** `GET http://localhost:8080/history/top`

**Description:** Get user's most played music.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
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

**Status Codes:**
- `200 OK` - Success
- `500 Internal Server Error` - Server error

---

### Search Endpoints

#### Search Users
```http
GET /search/users
```

**Client Access:** `GET http://localhost:8080/search/users`

**Description:** Search for users by username or email using fuzzy text matching.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `q` (required) - Search query string (min 2 characters)
- `limit` (optional, default: 20, max: 100)
- `cursor_score` (optional) - Similarity score cursor
- `cursor_ts` (optional) - Timestamp cursor (ISO 8601 format)

**Response:**
```json
{
  "users": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174000",
      "username": "johndoe",
      "email": "john@example.com",
      "bio": "Music lover",
      "profile_image_path": "http://localhost:8001/files/public/pictures-profile/user.jpg",
      "similarity_score": 0.85
    }
  ]
}
```

**Status Codes:**
- `200 OK` - Success (empty array if no results)
- `400 Bad Request` - Invalid query (too short)
- `500 Internal Server Error` - Server error

---

#### Search Artists
```http
GET /search/artists
```

**Client Access:** `GET http://localhost:8080/search/artists`

**Description:** Search for artists by name using fuzzy text matching.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `q` (required) - Search query string (min 2 characters)
- `limit` (optional, default: 20, max: 100)
- `cursor_score` (optional) - Similarity score cursor
- `cursor_ts` (optional) - Timestamp cursor

**Response:**
```json
{
  "artists": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174001",
      "artist_name": "The Beatles",
      "bio": "Legendary rock band",
      "profile_image_path": "http://localhost:8001/files/public/pictures-artist/artist.jpg",
      "similarity_score": 0.92
    }
  ]
}
```

**Status Codes:**
- `200 OK` - Success (empty array if no results)
- `400 Bad Request` - Invalid query
- `500 Internal Server Error` - Server error

---

#### Search Albums
```http
GET /search/albums
```

**Client Access:** `GET http://localhost:8080/search/albums`

**Description:** Search for albums by name using fuzzy text matching.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `q` (required) - Search query string (min 2 characters)
- `limit` (optional, default: 20, max: 100)
- `cursor_score` (optional) - Similarity score cursor
- `cursor_ts` (optional) - Timestamp cursor

**Response:**
```json
{
  "albums": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174002",
      "from_artist": "123e4567-e89b-12d3-a456-426614174001",
      "original_name": "Abbey Road",
      "description": "Classic album",
      "image_path": "http://localhost:8001/files/public/pictures-album/album.jpg",
      "similarity_score": 0.78
    }
  ]
}
```

**Status Codes:**
- `200 OK` - Success (empty array if no results)
- `400 Bad Request` - Invalid query
- `500 Internal Server Error` - Server error

---

#### Search Music
```http
GET /search/music
```

**Client Access:** `GET http://localhost:8080/search/music`

**Description:** Search for music tracks by name using fuzzy text matching.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `q` (required) - Search query string (min 2 characters)
- `limit` (optional, default: 20, max: 100)
- `cursor_score` (optional) - Similarity score cursor
- `cursor_ts` (optional) - Timestamp cursor

**Response:**
```json
{
  "music": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174003",
      "from_artist": "123e4567-e89b-12d3-a456-426614174001",
      "uploaded_by": "123e4567-e89b-12d3-a456-426614174000",
      "in_album": "123e4567-e89b-12d3-a456-426614174002",
      "song_name": "Come Together",
      "path_in_file_storage": "audio/abc123.mp3",
      "image_path": "http://localhost:8001/files/public/pictures-music/music.jpg",
      "play_count": 1500,
      "duration_seconds": 259,
      "similarity_score": 0.88
    }
  ]
}
```

**Status Codes:**
- `200 OK` - Success (empty array if no results)
- `400 Bad Request` - Invalid query
- `500 Internal Server Error` - Server error

---

#### Search Playlists
```http
GET /search/playlists
```

**Client Access:** `GET http://localhost:8080/search/playlists`

**Description:** Search for playlists by name using fuzzy text matching. Only returns public playlists or playlists owned by the requesting user.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `q` (required) - Search query string (min 2 characters)
- `limit` (optional, default: 20, max: 100)
- `cursor_score` (optional) - Similarity score cursor
- `cursor_ts` (optional) - Timestamp cursor

**Response:**
```json
{
  "playlists": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174004",
      "from_user": "123e4567-e89b-12d3-a456-426614174000",
      "original_name": "My Favorites",
      "description": "Best songs",
      "is_public": true,
      "image_path": "http://localhost:8001/files/public/pictures-playlist/playlist.jpg",
      "similarity_score": 0.65
    }
  ]
}
```

**Status Codes:**
- `200 OK` - Success (empty array if no results)
- `400 Bad Request` - Invalid query
- `500 Internal Server Error` - Server error

**Note:** Search uses PostgreSQL's `pg_trgm` extension for fuzzy text matching. Results are ordered by similarity score (descending), then by creation date (descending).

---

## Recommendation Backend: gateway_recommendation

**Base URL:** `http://localhost:8002` (internal)
**Client Access:** Via `gateway_api` at `http://localhost:8080/recommend/*` and `http://localhost:8080/popular/*`
**Authentication:** Service JWT (auto-generated by gateway_api)

### Health Check

```http
GET /
```

**Direct Access:** `GET http://localhost:8002/`

**Description:** Health check endpoint.

**Authentication:** None

**Response:**
```json
{
  "status": "healthy",
  "service": "gateway-recommendation"
}
```

**Status Codes:**
- `200 OK` - Service healthy

---

### Recommend Theme

```http
POST /recommend/theme
```

**Client Access:** `POST http://localhost:8080/recommend/theme`

**Description:** Get personalized theme recommendation using multi-armed bandit algorithm (LinUCB).

**Authentication:** Normal JWT (via gateway_api)

**Request Body:**
```json
{
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Note:** The `user_uuid` is automatically extracted from the Service JWT by the backend. Clients don't need to provide it.

**Response:**
```json
{
  "recommended_theme": "rock",
  "theme_features": [0.75, 0.42, 0.88, 0.31, 0.65, 0.53, 0.91, 0.22, 0.68, 0.44, 0.77, 0.59],
  "popularity_data": [
    {
      "theme": "rock",
      "decay_plays": 1000.5
    }
  ]
}
```

**Response Fields:**
- `recommended_theme` - Recommended music theme/genre
- `theme_features` - Feature vector (12 floats) used for prediction (save for feedback)
- `popularity_data` - Array of theme popularity data for fallback display

**Note:** The `confidence` field is NOT returned by this endpoint. Only the theme, features, and popularity data are returned.

**Status Codes:**
- `200 OK` - Recommendation generated
- `401 Unauthorized` - Invalid JWT
- `500 Internal Server Error` - Bandit service error
- `502 Bad Gateway` - Backend unavailable

**Backend:** Routes to service_bandit_system

---

### All-Time Popularity Endpoints

#### Get Popular Songs (All-Time)
```http
GET /popular/songs/all-time
```

**Client Access:** `GET http://localhost:8080/popular/songs/all-time`

**Description:** Get the most popular songs of all time based on play count and likes.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `limit` (optional, default: 50, max: 100)
- `cursor` (optional) - Decay score cursor for pagination
- `cursor_id` (optional) - UUID cursor for pagination

**Response:**
Returns an array of song popularity objects (no wrapper):
```json
[
  {
    "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
    "decay_plays": 15000.5,
    "decay_listen_seconds": 500000.0
  }
]
```

**Response Fields:**
- `music_uuid` - Unique identifier for the music track
- `decay_plays` - Decay-weighted play count (0.0-1.0 normalized)
- `decay_listen_seconds` - Decay-weighted total listen time in seconds

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid limit or cursor
- `500 Internal Server Error` - Server error

**Backend:** Routes to service_popularity_system

---

#### Get Popular Artists (All-Time)
```http
GET /popular/artists/all-time
```

**Client Access:** `GET http://localhost:8080/popular/artists/all-time`

**Description:** Get the most popular artists of all time.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `limit` (optional, default: 50, max: 100)
- `cursor` (optional) - Decay score cursor for pagination
- `cursor_id` (optional) - UUID cursor for pagination

**Response:**
Returns an array of artist popularity objects (no wrapper):
```json
[
  {
    "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
    "decay_plays": 50000.5,
    "decay_listen_seconds": 2000000.0
  }
]
```

**Response Fields:**
- `artist_uuid` - Unique identifier for the artist
- `decay_plays` - Decay-weighted play count (0.0-1.0 normalized)
- `decay_listen_seconds` - Decay-weighted total listen time in seconds

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid limit or cursor
- `500 Internal Server Error` - Server error

**Backend:** Routes to service_popularity_system

---

#### Get Popular Themes (All-Time)
```http
GET /popular/themes/all-time
```

**Client Access:** `GET http://localhost:8080/popular/themes/all-time`

**Description:** Get the most popular music themes/genres of all time.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `limit` (optional, default: 50, max: 100)

**Response:**
Returns an array of theme popularity objects (no wrapper):
```json
[
  {
    "theme": "rock",
    "decay_plays": 1000000.5,
    "decay_listen_seconds": 50000000.0
  }
]
```

**Response Fields:**
- `theme` - Music theme/genre name
- `decay_plays` - Decay-weighted play count (0.0-1.0 normalized)
- `decay_listen_seconds` - Decay-weighted total listen time in seconds

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid limit
- `500 Internal Server Error` - Server error

**Backend:** Routes to service_popularity_system

---

#### Get Popular Songs by Theme (All-Time)
```http
GET /popular/songs/theme/{theme}
```

**Client Access:** `GET http://localhost:8080/popular/songs/theme/{theme}`

**Description:** Get the most popular songs for a specific theme/genre of all time.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `theme` - Theme/genre name (e.g., "rock", "jazz", "electronic")

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)

**Response:**
```json
{
  "theme": "rock",
  "songs": [
    {
      "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Rock Song",
      "artist_name": "Rock Artist",
      "play_count": 25000,
      "like_count": 3000,
      "popularity_score": 0.91
    }
  ]
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Theme does not exist
- `500 Internal Server Error` - Server error

**Backend:** Routes to service_popularity_system

---

### Timeframe Popularity Endpoints

#### Get Popular Songs (Timeframe)
```http
GET /popular/songs/timeframe
```

**Client Access:** `GET http://localhost:8080/popular/songs/timeframe`

**Description:** Get trending songs within a specific date range.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `limit` (optional, default: 50, max: 100)
- `start_date` (required) - Start date in YYYY-MM-DD format
- `end_date` (required) - End date in YYYY-MM-DD format
- `cursor` (optional) - Play count cursor for pagination
- `cursor_id` (optional) - UUID cursor for pagination

**Example:**
```
GET http://localhost:8080/popular/songs/timeframe?start_date=2024-01-01&end_date=2024-01-31&limit=50
```

**Response:**
```json
[
  {
    "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
    "plays": 5000,
    "listen_seconds": 150000
  }
]
```

**Response Fields:**
- `music_uuid` - Unique identifier for the song
- `plays` - Total plays within the date range (uint64)
- `listen_seconds` - Total listen seconds within the date range (uint64)

**Note:** This endpoint returns a raw array, NOT wrapped in `{"songs": [...]}`. The fields returned are raw metrics, not enriched with `title`, `artist_name`, `like_count`, `popularity_score`, or `trend_direction`. The `timeframe` parameter is NOT supported.

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid timeframe or limit
- `500 Internal Server Error` - Server error

**Backend:** Routes to service_popularity_system

---

#### Get Popular Artists (Timeframe)
```http
GET /popular/artists/timeframe
```

**Client Access:** `GET http://localhost:8080/popular/artists/timeframe`

**Description:** Get trending artists within a specific date range.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `limit` (optional, default: 50, max: 100)
- `start_date` (required) - Start date in YYYY-MM-DD format
- `end_date` (required) - End date in YYYY-MM-DD format
- `cursor` (optional) - Play count cursor for pagination
- `cursor_id` (optional) - UUID cursor for pagination

**Example:**
```
GET http://localhost:8080/popular/artists/timeframe?start_date=2024-01-01&end_date=2024-01-31&limit=50
```

**Response:**
```json
[
  {
    "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
    "plays": 25000,
    "listen_seconds": 750000
  }
]
```

**Response Fields:**
- `artist_uuid` - Unique identifier for the artist
- `plays` - Total plays within the date range (uint64)
- `listen_seconds` - Total listen seconds within the date range (uint64)

**Note:** This endpoint returns a raw array, NOT wrapped in `{"artists": [...]}`. The fields returned are raw metrics, not enriched with `name`, `new_followers`, `popularity_score`, or `trend_direction`. The `timeframe` parameter is NOT supported.

---

#### Get Popular Themes (Timeframe)
```http
GET /popular/themes/timeframe
```

**Client Access:** `GET http://localhost:8080/popular/themes/timeframe`

**Description:** Get trending themes/genres within a specific date range.

**Authentication:** Normal JWT (via gateway_api)

**Query Parameters:**
- `limit` (optional, default: 50, max: 100)
- `start_date` (required) - Start date in YYYY-MM-DD format
- `end_date` (required) - End date in YYYY-MM-DD format
- `cursor` (optional) - Play count cursor for pagination
- `cursor_id` (optional) - UUID cursor for pagination

**Example:**
```
GET http://localhost:8080/popular/themes/timeframe?start_date=2024-01-01&end_date=2024-01-31&limit=50
```

**Response:**
```json
[
  {
    "theme": "rock",
    "plays": 50000,
    "listen_seconds": 1500000
  }
]
```

**Response Fields:**
- `theme` - Theme/genre name
- `plays` - Total plays within the date range (uint64)
- `listen_seconds` - Total listen seconds within the date range (uint64)

**Note:** This endpoint returns a raw array, NOT wrapped in `{"themes": [...]}`. The fields returned are raw metrics, not enriched with `popularity_score` or `trend_direction`. The `timeframe` parameter is NOT supported.

**Response:**
```json
{
  "timeframe": "week",
  "themes": [
    {
      "theme": "electronic",
      "play_count": 50000,
      "song_count": 200,
      "popularity_score": 0.88,
      "trend_direction": "up"
    }
  ]
}
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid timeframe or limit
- `500 Internal Server Error` - Server error

**Backend:** Routes to service_popularity_system

---

#### Get Popular Songs by Theme (Timeframe)
```http
GET /popular/songs/theme/{theme}/timeframe
```

**Client Access:** `GET http://localhost:8080/popular/songs/theme/{theme}/timeframe`

**Description:** Get trending songs for a specific theme within a date range.

**Authentication:** Normal JWT (via gateway_api)

**Path Parameters:**
- `theme` - Theme/genre name

**Query Parameters:**
- `limit` (optional, default: 50, max: 100)
- `start_date` (required) - Start date in YYYY-MM-DD format
- `end_date` (required) - End date in YYYY-MM-DD format
- `cursor` (optional) - Play count cursor for pagination
- `cursor_id` (optional) - UUID cursor for pagination

**Example:**
```
GET http://localhost:8080/popular/songs/theme/rock/timeframe?start_date=2024-01-01&end_date=2024-01-31&limit=30
```

**Response:**
```json
[
  {
    "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
    "theme": "rock",
    "plays": 10000,
    "listen_seconds": 300000
  }
]
```

**Response Fields:**
- `music_uuid` - Unique identifier for the song
- `theme` - Theme/genre name
- `plays` - Total plays within the date range (uint64)
- `listen_seconds` - Total listen seconds within the date range (uint64)

**Note:** This endpoint returns a raw array, NOT wrapped in `{"songs": [...]}`. The fields returned are raw metrics, not enriched with `title`, `artist_name`, `like_count`, `popularity_score`, or `trend_direction`. The `timeframe` parameter is NOT supported.

| Value | Duration | Description |
|-------|----------|-------------|
| `"day"` | 24 hours | Last 24 hours |
| `"week"` | 7 days | Last 7 days |
| `"month"` | 30 days | Last 30 days |
| `"year"` | 365 days | Last 365 days |

---

## Popularity Backend: service_popularity_system

**Base URL:** `http://localhost:8003` (internal)
**Client Access:** Via `gateway_recommendation` → `gateway_api` at `http://localhost:8080/popular/*`
**Authentication:** Service JWT (auto-generated)

**Note:** This service is accessed through gateway_recommendation. All endpoints listed above in the [Recommendation Backend](#recommendation-backend-gateway_recommendation) section are implemented by this service.

### Implementation Details

**Popularity Score Calculation:**

Songs (All-Time):
- 70% total play count
- 30% like count
- Normalized to 0.0-1.0

Songs (Timeframe):
- 60% plays in period
- 20% likes
- 20% velocity (rate of change)

Artists (All-Time):
- 50% follower count
- 40% total plays
- 10% music count

Artists (Timeframe):
- 40% new followers in period
- 40% plays in period
- 20% velocity

Themes:
- 50% total plays
- 30% song count
- 20% unique listeners

**Trend Direction:**
- `"up"` - velocity > 0.1
- `"down"` - velocity < -0.1
- `"stable"` - -0.1 ≤ velocity ≤ 0.1

**Data Source:** ClickHouse data warehouse with decay-weighted metrics

---

## Bandit ML Backend: service_bandit_system

**Base URL:** `http://localhost:8004/api/v1` (internal)
**Client Access:** Via `gateway_recommendation` → `gateway_api` at `http://localhost:8080/recommend/theme`
**Authentication:** Service JWT (auto-generated)
**Framework:** FastAPI (Python)

### Health Check

```http
GET /api/v1/health
```

**Direct Access:** `GET http://localhost:8004/api/v1/health`

**Authentication:** None

**Response:**
```json
{
  "status": "healthy",
  "service": "bandit-system"
}
```

**Status Codes:**
- `200 OK` - Service healthy

---

### Predict Theme

```http
POST /api/v1/predict
```

**Description:** Get personalized theme recommendation using LinUCB algorithm.

**Note:** This endpoint is accessed via gateway_recommendation at `POST /recommend/theme`.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Response:**
```json
{
  "theme": "rock",
  "features": [0.75, 0.42, 0.88, 0.31, 0.65, 0.53, 0.91, 0.22, 0.68, 0.44, 0.77, 0.59]
}
```

**Response Fields:**
- `theme` (string) - Recommended music theme/genre
- `features` (array of 12 floats) - Feature vector for feedback loop

**Status Codes:**
- `200 OK` - Prediction successful
- `401 Unauthorized` - Invalid service JWT
- `422 Unprocessable Entity` - Invalid request body
- `500 Internal Server Error` - Model error

---

### Update Model

```http
POST /api/v1/update
```

**Description:** Update the bandit model with user feedback (reward).

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "theme": "rock",
  "reward": 0.85,
  "features": [0.75, 0.42, 0.88, 0.31, 0.65, 0.53, 0.91, 0.22, 0.68, 0.44, 0.77, 0.59]
}
```

**Request Fields:**
- `user_uuid` (UUID4, required) - User UUID
- `theme` (string, required) - Theme that was shown
- `reward` (float, required) - Reward value 0.0-1.0
- `features` (array, required) - Feature vector from prediction (must be exactly 12 floats)

**Reward Guidelines:**

| Reward | User Action |
|--------|-------------|
| `1.0` | Listened fully + liked + added to playlist |
| `0.8-0.9` | Listened fully + liked |
| `0.6-0.7` | Listened 75%+ OR liked |
| `0.4-0.5` | Listened 50-75% (neutral) |
| `0.2-0.3` | Listened <50% (disliked) |
| `0.0-0.1` | Skipped immediately |

**Response:**
```json
{
  "success": true
}
```

**Status Codes:**
- `202 Accepted` - Update accepted
- `400 Bad Request` - Invalid features length (must be 12 elements)
- `401 Unauthorized` - Invalid service JWT
- `422 Unprocessable Entity` - Invalid reward value
- `500 Internal Server Error` - Update failed

---

### Algorithm: LinUCB

**Overview:**
LinUCB (Linear Upper Confidence Bound) is a contextual multi-armed bandit algorithm that balances exploration (trying new themes) and exploitation (recommending known preferences).

**How It Works:**

1. **Feature Extraction** - User context is converted to a feature vector (5-20 dimensions)
2. **Prediction** - Calculate upper confidence bound for each theme:
   ```
   UCB = θ^T * x + α * sqrt(x^T * A^-1 * x)
   ```
   - `θ` = learned weight vector
   - `x` = user feature vector
   - `A` = covariance matrix
   - `α` = exploration parameter (default: 1.0)
3. **Update** - After receiving reward, update model parameters:
   ```
   A = A + x * x^T
   b = b + reward * x
   θ = A^-1 * b
   ```

**Feature Vector:**
Typically includes:
- User listening history
- Recent themes and play counts
- Time of day, day of week
- User preferences (likes, playlist adds)
- Session context

**Model Persistence:**
Models are stored per-user in PostgreSQL and loaded on service startup.

**Cold Start:**
Falls back to popularity-based recommendations for new users.

---

## Common Patterns

### Pagination

Most list endpoints use cursor-based pagination:

**Request Parameters:**
```
?limit=20&cursor=2024-01-01T00:00:00Z&cursor_id=123e4567-e89b-12d3-a456-426614174000
```

**Parameter Types:**
- `limit` - Number of results (default: 20, max: 100)
- `cursor` - Timestamp cursor (ISO 8601 format)
- `cursor_id` - UUID cursor
- `cursor_name` - Name cursor (for tags/artists)
- `cursor_pos` - Position cursor (for playlists)
- `cursor_count` - Count cursor (for top music)
- `cursor_score` - Similarity score cursor (for search)

**Response Format:**
```json
{
  "items": [...],
  "next_cursor": "2024-01-01T00:00:00Z",
  "next_cursor_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Usage:**
Use the `next_cursor` and `next_cursor_id` values from the response as `cursor` and `cursor_id` in the next request.

---

### File Uploads

Endpoints that accept file uploads use `multipart/form-data`:

**Endpoints:**
- `POST /users/me/image`
- `POST /artists/{uuid}/image`
- `POST /albums/{uuid}/image`
- `POST /music/{uuid}/image`
- `POST /playlists/{uuid}/image`

**Request Format:**
```
Content-Type: multipart/form-data

image: [binary file data]
```

**Supported Formats:**
- JPEG (`.jpg`, `.jpeg`)
- PNG (`.png`)
- WebP (`.webp`)

**Size Limits:**
- Profile images: 5 MB max
- Album/Music/Playlist images: 10 MB max

**Response:**
```json
{
  "image_url": "http://localhost:8001/files/public/pictures-profile/abc123.jpg"
}
```

---

### File Storage

**Storage Architecture:**
- Database stores **relative paths** (e.g., `audio/abc123.mp3`)
- Paths are converted to public URLs on read
- MinIO (S3-compatible) object storage backend

**URL Format:**
```
http://localhost:8001/files/public/{bucket}/{object-path}
```

**Buckets:**
- `pictures-profile` - User profile images
- `pictures-artist` - Artist images
- `pictures-album` - Album cover art
- `pictures-music` - Music track cover art
- `pictures-playlist` - Playlist cover images
- `audio` - Music audio files

**Default Images:**
Applied automatically on read if no image is set (not stored in DB).

---

### Search

All search endpoints use PostgreSQL's `pg_trgm` extension for fuzzy text matching.

**Common Parameters:**
- `q` (required) - Search query (min 2 characters)
- `limit` (optional) - Max results (default: 20, max: 100)
- `cursor_score` (optional) - Similarity score cursor
- `cursor_ts` (optional) - Timestamp cursor

**Ranking:**
Results are ordered by:
1. Similarity score (descending)
2. Creation date (descending)

**Empty Results:**
Returns empty array if no matches found (not 404).

---

### Timeframes

Popularity endpoints accept a `timeframe` query parameter:

**Valid Values:**
- `"day"` - Last 24 hours
- `"week"` - Last 7 days
- `"month"` - Last 30 days
- `"year"` - Last 365 days

**Usage:**
```
GET /popular/songs/timeframe?timeframe=week
```

**Invalid Timeframe:**
Returns `400 Bad Request` with error message.

---

## Error Responses

### Standard Error Format

```json
{
  "error": "Error message describing what went wrong",
  "detail": "Additional details (optional)"
}
```

### HTTP Status Codes

| Code | Name | When Used |
|------|------|-----------|
| `200 OK` | Success | Successful GET, POST, DELETE |
| `201 Created` | Created | Successful resource creation (PUT, POST) |
| `202 Accepted` | Accepted | Request accepted for async processing |
| `400 Bad Request` | Bad Request | Invalid input, malformed JSON, validation failure |
| `401 Unauthorized` | Unauthorized | Invalid/missing/expired JWT token |
| `403 Forbidden` | Forbidden | Valid JWT but insufficient permissions |
| `404 Not Found` | Not Found | Resource does not exist |
| `409 Conflict` | Conflict | Resource already exists, constraint violation |
| `422 Unprocessable Entity` | Validation Error | FastAPI validation errors (Python services) |
| `500 Internal Server Error` | Server Error | Database error, file system error, unexpected failure |
| `502 Bad Gateway` | Bad Gateway | Backend service unavailable or returned error |

### Common Errors

**Invalid JWT:**
```json
{
  "error": "Unauthorized",
  "detail": "Invalid or expired token"
}
```
Status: `401 Unauthorized`

**Resource Not Found:**
```json
{
  "error": "Music not found"
}
```
Status: `404 Not Found`

**Validation Error:**
```json
{
  "error": "Invalid input",
  "detail": "Email format is invalid"
}
```
Status: `400 Bad Request`

**Permission Denied:**
```json
{
  "error": "Forbidden",
  "detail": "You are not authorized to modify this resource"
}
```
Status: `403 Forbidden`

**Service Unavailable:**
```json
{
  "error": "Backend service unavailable"
}
```
Status: `502 Bad Gateway`

---

## Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture, design decisions, and request flows
- [.thoughts/DOCUMENTATION_STRATEGY.md](../.thoughts/DOCUMENTATION_STRATEGY.md) - Documentation maintenance rules
- [.thoughts/LOGGING_STRATEGY.md](../.thoughts/LOGGING_STRATEGY.md) - Logging guidelines

---

**Document Version:** 1.0
**Last Updated:** 2026-03-06
**Maintained By:** Development Team
