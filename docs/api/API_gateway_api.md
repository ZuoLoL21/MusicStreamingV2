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

**See:** [service_user_database User Endpoints](API_service_user_database.md#user-endpoints) for detailed documentation.

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

**See:** [service_user_database Artist Endpoints](API_service_user_database.md#artist-endpoints) for detailed documentation.

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

**See:** [service_user_database Album Endpoints](API_service_user_database.md#album-endpoints) for detailed documentation.

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

**See:** [service_user_database Music Endpoints](API_service_user_database.md#music-endpoints) for detailed documentation.

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

**See:** [service_user_database Tag Endpoints](API_service_user_database.md#tag-endpoints) for detailed documentation.

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

**See:** [service_user_database Playlist Endpoints](API_service_user_database.md#playlist-endpoints) for detailed documentation.

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

**See:** [service_user_database History Endpoints](API_service_user_database.md#history-endpoints) for detailed documentation.

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

**See:** [service_user_database Search Endpoints](API_service_user_database.md#search-endpoints) for detailed documentation.

---

#### Recommendation Routes
```
POST /recommend/theme
```

**Description:** Get personalized theme recommendation using multi-armed bandit algorithm.

**Authentication:** Normal JWT required

**Proxied to:** gateway_recommendation → service_bandit_system

**See:** [gateway_recommendation Recommend Theme](API_gateway_recommendation.md#recommend-theme) for detailed documentation.

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

**See:** [gateway_recommendation Popularity Endpoints](API_gateway_recommendation.md#all-time-popularity-endpoints) for detailed documentation.

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
