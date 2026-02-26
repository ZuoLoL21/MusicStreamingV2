# Gateway API - API Documentation

**Base URL:** `http://localhost:8080`
**Service:** Public-facing unified API gateway
**Authentication:** JWT-based (Normal/Refresh tokens)

## Overview

The Gateway API is the main entry point for all client requests. It validates user JWT tokens (normal or refresh), generates service JWTs for backend communication, and routes requests to appropriate backend services.

## Authentication

### JWT Types
- **Normal JWT** (subject: "normal") - User access tokens (~10 min lifetime)
- **Refresh JWT** (subject: "refresh") - User refresh tokens (~10 day lifetime)
- **Service JWT** (subject: "service") - Inter-service tokens (~2 min lifetime)

### Headers
```
Authorization: Bearer <jwt-token>
```

---

## Public Endpoints

### Health Check
```http
GET /health
```

**Description:** Health check endpoint to verify service availability.

**Authentication:** None

**Response:**
```json
{
  "status": "healthy",
  "service": "gateway-api"
}
```

---

### Login
```http
POST /login
```

**Description:** Authenticate user and receive JWT tokens.

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

**Status Codes:**
- `200 OK` - Successful authentication
- `401 Unauthorized` - Invalid credentials
- `502 Bad Gateway` - Backend service unavailable

---

### Register
```http
PUT /login
```

**Description:** Register a new user account.

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

**Status Codes:**
- `201 Created` - User registered successfully
- `400 Bad Request` - Invalid input
- `409 Conflict` - Email already exists
- `502 Bad Gateway` - Backend service unavailable

---

## Refresh Token Endpoints

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
  "normal_token": "eyJhbGciOiJIUzI1NiIs...",
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Status Codes:**
- `200 OK` - Token refreshed successfully
- `401 Unauthorized` - Invalid or expired refresh token
- `502 Bad Gateway` - Backend service unavailable

---

## Protected Endpoints (Require Normal JWT)

All endpoints below require a valid Normal JWT in the Authorization header.

### User Routes
```http
GET    /users/*
POST   /users/*
PUT    /users/*
DELETE /users/*
```

**Description:** Proxies all user-related requests to the User Database Service.

**Authentication:** Normal JWT required

**Routes Include:**
- User profile management
- User relationships (following/followers)
- User music and playlists
- User likes and history

**See:** [Service User Database API](../service_user_database/API.md) for detailed endpoint documentation.

---

### Artist Routes
```http
GET    /artists/*
POST   /artists/*
PUT    /artists/*
DELETE /artists/*
```

**Description:** Proxies all artist-related requests to the User Database Service.

**Authentication:** Normal JWT required

**Routes Include:**
- Artist profiles
- Artist members and roles
- Artist albums and music
- Artist followers

**See:** [Service User Database API](../service_user_database/API.md) for detailed endpoint documentation.

---

### Album Routes
```http
GET    /albums/*
POST   /albums/*
PUT    /albums/*
DELETE /albums/*
```

**Description:** Proxies all album-related requests to the User Database Service.

**Authentication:** Normal JWT required

**Routes Include:**
- Album creation and management
- Album music tracks
- Album metadata

**See:** [Service User Database API](../service_user_database/API.md) for detailed endpoint documentation.

---

### Music Routes
```http
GET    /music/*
POST   /music/*
PUT    /music/*
DELETE /music/*
```

**Description:** Proxies all music track-related requests to the User Database Service.

**Authentication:** Normal JWT required

**Routes Include:**
- Music track management
- Play count and listening history
- Music likes
- Music tags

**See:** [Service User Database API](../service_user_database/API.md) for detailed endpoint documentation.

---

### Tag Routes
```http
GET    /tags/*
POST   /tags/*
PUT    /tags/*
DELETE /tags/*
```

**Description:** Proxies all tag-related requests to the User Database Service.

**Authentication:** Normal JWT required

**Routes Include:**
- Tag management
- Music tagging
- Tag browsing

**See:** [Service User Database API](../service_user_database/API.md) for detailed endpoint documentation.

---

### Playlist Routes
```http
GET    /playlists/*
POST   /playlists/*
PUT    /playlists/*
DELETE /playlists/*
```

**Description:** Proxies all playlist-related requests to the User Database Service.

**Authentication:** Normal JWT required

**Routes Include:**
- Playlist creation and management
- Playlist tracks
- Track ordering

**See:** [Service User Database API](../service_user_database/API.md) for detailed endpoint documentation.

---

### History Routes
```http
GET /history/*
```

**Description:** Proxies listening history requests to the User Database Service.

**Authentication:** Normal JWT required

**Routes Include:**
- User listening history
- Top played music

**See:** [Service User Database API](../service_user_database/API.md) for detailed endpoint documentation.

---

### Search Routes
```http
GET /search/users
GET /search/artists
GET /search/albums
GET /search/music
GET /search/playlists
```

**Description:** Proxies search requests to the User Database Service. Search across users, artists, albums, music, and playlists using fuzzy text matching.

**Authentication:** Normal JWT required

**Routes Include:**
- User search by username or email
- Artist search by name
- Album search by name
- Music track search by name
- Playlist search by name (public and owned playlists)

**See:** [Service User Database API](../service_user_database/API.md) for detailed endpoint documentation.

---

### Recommendation Routes
```http
GET  /recommendation/*
POST /recommendation/*
```

**Description:** Proxies recommendation requests to the Recommendation Gateway.

**Authentication:** Normal JWT required

**Routes Include:**
- Theme-based recommendations
- Popularity metrics
- Personalized suggestions

**See:** [Gateway Recommendation API](../gateway_recommendation/API.md) for detailed endpoint documentation.

---

## How It Works

1. **Client Request** → Gateway API validates user JWT
2. **Gateway** → Generates service JWT from validated user JWT
3. **Gateway** → Forwards request to backend service with service JWT
4. **Backend Service** → Validates service JWT and processes request
5. **Response** → Returns through gateway to client

## Service JWT

The gateway automatically generates service JWTs for authenticated requests containing:
- Subject: "service"
- User UUID: Extracted from validated user JWT
- Expiration: ~2 minutes

Backend services validate this service JWT to ensure requests come from authorized gateways.

## Error Responses

All endpoints may return:
- `401 Unauthorized` - Invalid or missing JWT token
- `500 Internal Server Error` - Server error
- `502 Bad Gateway` - Backend service unavailable
