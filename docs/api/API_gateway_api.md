## Public Layer: gateway_api

**Base URL:** `http://localhost:8080`
**Role:** Public-facing unified API gateway
**Access:** Direct client access

---

## Authentication

The gateway API supports three types of JWT tokens for authentication:

### JWT Token Types

| Token Type | Purpose | Header Value | Description |
|------------|---------|--------------|-------------|
| **Normal JWT** | User authentication | `Authorization: Bearer <token>` | Standard access token for authenticated user operations |
| **Refresh JWT** | Token renewal | `Authorization: Bearer <token>` | Long-lived token used to obtain new access tokens |
| **Service JWT** | Service-to-service | `Authorization: Bearer <token>` | Internal token for inter-service communication |

### Authentication Flow

1. **Login** (`POST /login`) - Returns Normal JWT (short-lived, ~10 minutes) and Refresh JWT (long-lived, ~10 days)
2. **Token Renewal** (`POST /renew`) - Use Refresh JWT to obtain a new Normal JWT
3. **Protected Operations** - Use Normal JWT for all authenticated endpoints
4. **Service Communication** - Gateway adds Service JWT when proxying to backend services

### Public Endpoints

These endpoints do not require any authentication:

- `GET /health` - Health check

### Registration and Login

- `POST /login` - User login (returns JWT tokens)
- `PUT /login` - User registration (returns JWT tokens)
- `POST /renew` - Refresh access token (requires Refresh JWT)

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
  "service": "gateway-api"
}
```

**Status Codes:**
- `200 OK` - Service is healthy

---

### Login
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
- `400 Bad Request` - Invalid request body format
- `401 Unauthorized` - Invalid credentials
- `415 Unsupported Media Type` - Wrong Content-Type
- `500 Internal Server Error` - Server error
- `502 Bad Gateway` - Backend service unavailable
- `503 Service Unavailable` - Service temporarily unavailable

---

### Register
```http
PUT /login
OPTIONS /login
```

**Description:** Register a new user account with optional profile image. The PUT method is recommended; OPTIONS is available for CORS preflight requests.

**Authentication:** None

**Form Fields (multipart/form-data):**
- `username` (required, string, min 5 chars) - Username
- `email` (required, string, valid email format) - Email address
- `password` (required, string, min 8 chars, must contain uppercase, lowercase, number, and special character) - Password
- `country` (required, string, 2 chars) - ISO 3166-1 alpha-2 country code (e.g., "US", "GB")
- `bio` (optional, string) - User biography
- `image` (optional, file, max 10MB) - Profile image (JPEG, PNG, or WebP)

**Content-Type:** `multipart/form-data` (required)

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
- `413 Payload Too Large` - Image file too large
- `415 Unsupported Media Type` - Wrong Content-Type (must be multipart/form-data)
- `422 Unprocessable Entity` - Validation failed
- `500 Internal Server Error` - Server error
- `502 Bad Gateway` - Backend service unavailable
- `503 Service Unavailable` - Service temporarily unavailable

---

## Token Refresh Endpoints

### Renew Token
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
  "access_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Status Codes:**
- `200 OK` - Token refreshed successfully
- `400 Bad Request` - Invalid request
- `401 Unauthorized` - Invalid or expired refresh token
- `403 Forbidden` - Refresh token required
- `415 Unsupported Media Type` - Wrong Content-Type
- `502 Bad Gateway` - Backend service unavailable
- `503 Service Unavailable` - Service temporarily unavailable

---

## Protected Endpoints (Require Normal JWT)

All endpoints below require a valid Normal JWT in the `Authorization` header.

**Common Status Codes:**
- `200 OK` - Success
- `201 Created` - Resource created
- `204 No Content` - Success, no content to return
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Invalid or missing JWT
- `403 Forbidden` - Valid JWT but insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists
- `413 Payload Too Large` - File too large
- `415 Unsupported Media Type` - Wrong Content-Type
- `422 Unprocessable Entity` - Validation failed
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error
- `502 Bad Gateway` - Backend service unavailable
- `503 Service Unavailable` - Service temporarily unavailable

For detailed endpoint-specific status codes, see the linked backend documentation.

### User Routes
```
GET    /users/*
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

**File Upload Endpoints (multipart/form-data):**
- `PUT /users/me/image` - Upload user profile image

**See:** [service_user_database User Endpoints](API_service_user_database.md#user-endpoints) for detailed documentation.

---

### Artist Routes
```
GET    /artists/*
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

**File Upload Endpoints (multipart/form-data):**
- `PUT /artists` - Create new artist
- `PUT /artists/{uuid}/image` - Upload artist image

**See:** [service_user_database Artist Endpoints](API_service_user_database.md#artist-endpoints) for detailed documentation.

---

### Album Routes
```
GET    /albums/*
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

**File Upload Endpoints (multipart/form-data):**
- `PUT /albums` - Create new album
- `PUT /albums/{uuid}/image` - Upload album image

**See:** [service_user_database Album Endpoints](API_service_user_database.md#album-endpoints) for detailed documentation.

---

### Music Routes
```
GET    /music/*
PUT    /music/*
DELETE /music/*
```

**Description:** Proxies all music track-related requests to service_user_database.

**Authentication:** Normal JWT required (except `/music/{uuid}/play`)

**Proxied Routes Include:**
- Music track management (`/music`, `/music/{uuid}`)
- Music metadata (`/music/{uuid}/storage`, `/music/{uuid}/image`)
- Play tracking (`/music/{uuid}/play`) - **Requires Normal JWT**
- Listening history (`/music/{uuid}/listen`)
- Music likes (`/music/{uuid}/like`, `/music/{uuid}/liked`)
- Music tags (`/music/{uuid}/tags`, `/music/{uuid}/tags/{name}`)

**File Upload Endpoints (multipart/form-data):**
- `PUT /music` - Upload new music track
- `PUT /music/{uuid}/image` - Upload music track image

**See:** [service_user_database Music Endpoints](API_service_user_database.md#music-endpoints) for detailed documentation.

---

### Tag Routes
```
GET    /tags/*
PUT    /tags/*
DELETE /tags/*
```

**Description:** Proxies all tag-related requests to service_user_database.

**Authentication:** Normal JWT required

**Proxied Routes Include:**
- Tag browsing and management (`/tags`, `/tags/{name}`)
- Music tagging (`/music/{uuid}/tags`, `/tags/{name}/music`)
- Tag-based music discovery

**See:** [service_user_database Tag Endpoints](API_service_user_database.md#tag-endpoints) for detailed documentation.

---

### Playlist Routes
```
GET    /playlists/*
PUT    /playlists/*
DELETE /playlists/*
```

**Description:** Proxies all playlist-related requests to service_user_database.

**Authentication:** Normal JWT required

**Proxied Routes Include:**
- Playlist creation and management (`/playlists`, `/playlists/{uuid}`)
- Playlist image (`/playlists/{uuid}/image`)
- Playlist tracks (`/playlists/{uuid}/tracks`, `/playlists/{uuid}/reorder`)
- Track management and reordering

**File Upload Endpoints (multipart/form-data):**
- `PUT /playlists` - Create new playlist
- `PUT /playlists/{uuid}/image` - Upload playlist image

**See:** [service_user_database Playlist Endpoints](API_service_user_database.md#playlist-endpoints) for detailed documentation.

---

### History Routes
```
GET /history
GET /history/top
```

**Description:** Proxies listening history requests to service_user_database.

**Authentication:** Normal JWT required

**Proxied Routes Include:**
- User listening history (`/history`)
- Top played music (`/history/top`)

**See:** [service_user_database History Endpoints](API_service_user_database.md#history-endpoints) for detailed documentation.

---

### Search Routes
```
GET /search/users
GET /search/artists
GET /search/albums
GET /search/music
GET /search/playlists
```

**Description:** Proxies search requests to service_user_database. Search across users, artists, albums, music, and playlists using fuzzy text matching.

**Authentication:** Normal JWT required

**See:** [service_user_database Search Endpoints](API_service_user_database.md#search-endpoints) for detailed documentation.

---

### Recommendation Routes
```
POST /recommend/theme
```

**Description:** Get personalized theme recommendation using multi-armed bandit algorithm.

**Authentication:** Normal JWT required

**Proxied to:** gateway_recommendation → service_bandit_system

**See:** [gateway_recommendation Recommend Theme](API_gateway_recommendation.md#recommend-theme) for detailed documentation.

---

### Popularity Routes
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

**See:** [gateway_recommendation Popularity Endpoints](API_gateway_recommendation.md#all-time-popularity-endpoints) for detailed documentation.

---

### File Routes
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

### Event Routes

**Description:** Track user events for analytics.

**Proxied to:** service_event_ingestion

#### Listen Event
```http
POST /events/listen
```

**Authentication:** Normal JWT required

**Request Body:**
```json
{
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "music_uuid": "123e4567-e89b-12d3-a456-426614174001",
  "artist_uuid": "123e4567-e89b-12d3-a456-426614174002",
  "album_uuid": "123e4567-e89b-12d3-a456-426614174003",
  "listen_duration_seconds": 120,
  "track_duration_seconds": 180,
  "completion_ratio": 0.67
}
```

**Fields:**
- `user_uuid` (required, UUID) - User UUID
- `music_uuid` (required, UUID) - Music track UUID
- `artist_uuid` (required, UUID) - Artist UUID
- `album_uuid` (optional, UUID) - Album UUID
- `listen_duration_seconds` (optional, int, default: 0) - How long the user listened (seconds, must be >= 0)
- `track_duration_seconds` (optional, int, default: 0) - Total track duration (seconds)
- `completion_ratio` (required, float) - Listen completion ratio (0.0-1.0)

**Status Codes:**
- `200 OK` - Event tracked
- `400 Bad Request` - Invalid input
- `422 Unprocessable Entity` - Validation failed
- `500 Internal Server Error` - Server error

---

#### Like Event
```http
POST /events/like
```

**Authentication:** Normal JWT required

**Request Body:**
```json
{
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "music_uuid": "123e4567-e89b-12d3-a456-426614174001",
  "artist_uuid": "123e4567-e89b-12d3-a456-426614174002"
}
```

**Fields:**
- `user_uuid` (required, UUID) - User UUID
- `music_uuid` (required, UUID) - Music track UUID
- `artist_uuid` (required, UUID) - Artist UUID

**Status Codes:**
- `200 OK` - Event tracked
- `400 Bad Request` - Invalid input
- `422 Unprocessable Entity` - Validation failed
- `500 Internal Server Error` - Server error

---

#### Theme Event
```http
POST /events/theme
```

**Authentication:** Normal JWT required

**Request Body:**
```json
{
  "music_uuid": "123e4567-e89b-12d3-a456-426614174001",
  "theme": "rock"
}
```

**Fields:**
- `music_uuid` (required, UUID) - Music track UUID
- `theme` (required, string) - Theme/genre name

**Status Codes:**
- `200 OK` - Event tracked
- `400 Bad Request` - Invalid input (missing music_uuid or theme)
- `422 Unprocessable Entity` - Validation failed
- `500 Internal Server Error` - Server error

---

#### User Dimension Event
```http
POST /events/user
```

**Description:** Track user dimension events for analytics. This endpoint is primarily for internal use and is called automatically when users are created or updated.

**Authentication:** Normal JWT required (via gateway_api). The gateway validates the Normal JWT and adds a Service JWT when proxying to the backend.

**Request Body:**
```json
{
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2024-01-01T00:00:00Z",
  "country": "US"
}
```

**Fields:**
- `user_uuid` (required, UUID) - User UUID
- `created_at` (optional, ISO 8601 timestamp) - User creation timestamp. If not provided, the current time is used.
- `country` (required, string, 2 chars) - ISO 3166-1 alpha-2 country code (uppercase)

**Status Codes:**
- `200 OK` - Event tracked
- `400 Bad Request` - Invalid input (missing user_uuid or country, invalid country format)
- `422 Unprocessable Entity` - Validation failed
- `500 Internal Server Error` - Server error

---

## Service-to-Service Endpoints (Service JWT)

These endpoints are used for internal service communication and require a Service JWT.

### Event Ingestion Routes

These routes are proxied with a Service JWT added by the gateway when called from authenticated clients.

```
POST /events/listen
POST /events/like
POST /events/theme
POST /events/user
```

**Description:** Track user events for analytics. The gateway adds a Service JWT when proxying these requests to service_event_ingestion.

**Authentication:** 
- External clients: Normal JWT required
- Internal services: Service JWT (added by gateway)

**See:** [service_event_ingestion API](API_service_event_ingestion.md) for detailed documentation.
