# Gateway Recommendation - API Documentation

**Base URL:** `http://localhost:8002`
**Service:** Recommendation gateway
**Authentication:** Service JWT (from gateway_api)

## Overview

The Gateway Recommendation service orchestrates recommendation requests by routing them to appropriate backend services (Bandit System for personalized theme recommendations and Popularity System for trending content).

## Authentication

All endpoints (except `/` health check) require a valid Service JWT in the Authorization header.

### Headers
```
Authorization: Bearer <service-jwt-token>
```

---

## Public Endpoints

### Health Check
```http
GET /
```

**Description:** Health check endpoint to verify service availability.

**Authentication:** None

**Response:**
```json
{
  "status": "healthy",
  "service": "gateway-recommendation"
}
```

---

## Recommendation Endpoints

### Recommend Theme
```http
POST /recommend/theme
```

**Description:** Get personalized theme recommendation using multi-armed bandit algorithm (LinUCB). Returns the best theme for the user based on their listening history and preferences.

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
  "features": [0.75, 0.42, 0.88, 0.31, 0.65],
  "confidence": 0.87
}
```

**Fields:**
- `theme` - Recommended music theme/genre
- `features` - Feature vector used for prediction
- `confidence` - Confidence score for the recommendation

**Status Codes:**
- `200 OK` - Recommendation generated successfully
- `401 Unauthorized` - Invalid or missing service JWT
- `500 Internal Server Error` - Bandit service unavailable
- `502 Bad Gateway` - Backend service error

**Backend Service:** Routes to service_bandit_system

---

## All-Time Popularity Endpoints

### Get Popular Songs (All-Time)
```http
GET /popular/songs/all-time
```

**Description:** Get the most popular songs of all time based on play count and likes.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20, max: 100) - Number of results

**Response:**
```json
{
  "songs": [
    {
      "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Song Title",
      "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "artist_name": "Artist Name",
      "play_count": 15000,
      "like_count": 2000,
      "popularity_score": 0.95
    }
  ]
}
```

**Backend Service:** Routes to service_popularity_system

---

### Get Popular Artists (All-Time)
```http
GET /popular/artists/all-time
```

**Description:** Get the most popular artists of all time based on follower count and music plays.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)

**Response:**
```json
{
  "artists": [
    {
      "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Artist Name",
      "follower_count": 50000,
      "total_plays": 500000,
      "popularity_score": 0.98
    }
  ]
}
```

**Backend Service:** Routes to service_popularity_system

---

### Get Popular Themes (All-Time)
```http
GET /popular/themes/all-time
```

**Description:** Get the most popular music themes/genres of all time.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)

**Response:**
```json
{
  "themes": [
    {
      "theme": "rock",
      "play_count": 1000000,
      "song_count": 5000,
      "popularity_score": 0.92
    }
  ]
}
```

**Backend Service:** Routes to service_popularity_system

---

### Get Popular Songs by Theme (All-Time)
```http
GET /popular/songs/theme/{theme}
```

**Description:** Get the most popular songs for a specific theme/genre of all time.

**Authentication:** Service JWT required

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

**Backend Service:** Routes to service_popularity_system

---

## Timeframe Popularity Endpoints

### Get Popular Songs (Timeframe)
```http
GET /popular/songs/timeframe
```

**Description:** Get trending songs within a specific time period.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `timeframe` (required) - Time period: "day", "week", "month", "year"

**Example:**
```
GET /popular/songs/timeframe?timeframe=week&limit=50
```

**Response:**
```json
{
  "timeframe": "week",
  "songs": [
    {
      "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Trending Song",
      "artist_name": "Artist Name",
      "play_count": 5000,
      "like_count": 800,
      "popularity_score": 0.89,
      "trend_direction": "up"
    }
  ]
}
```

**Backend Service:** Routes to service_popularity_system

---

### Get Popular Artists (Timeframe)
```http
GET /popular/artists/timeframe
```

**Description:** Get trending artists within a specific time period.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `timeframe` (required) - Time period: "day", "week", "month", "year"

**Response:**
```json
{
  "timeframe": "week",
  "artists": [
    {
      "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Trending Artist",
      "new_followers": 500,
      "total_plays": 25000,
      "popularity_score": 0.85,
      "trend_direction": "up"
    }
  ]
}
```

**Backend Service:** Routes to service_popularity_system

---

### Get Popular Themes (Timeframe)
```http
GET /popular/themes/timeframe
```

**Description:** Get trending themes/genres within a specific time period.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `timeframe` (required) - Time period: "day", "week", "month", "year"

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

**Backend Service:** Routes to service_popularity_system

---

### Get Popular Songs by Theme (Timeframe)
```http
GET /popular/songs/theme/{theme}/timeframe
```

**Description:** Get trending songs for a specific theme within a time period.

**Authentication:** Service JWT required

**Path Parameters:**
- `theme` - Theme/genre name

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `timeframe` (required) - Time period: "day", "week", "month", "year"

**Example:**
```
GET /popular/songs/theme/rock/timeframe?timeframe=month&limit=30
```

**Response:**
```json
{
  "theme": "rock",
  "timeframe": "month",
  "songs": [
    {
      "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Trending Rock Song",
      "artist_name": "Rock Artist",
      "play_count": 10000,
      "like_count": 1500,
      "popularity_score": 0.87,
      "trend_direction": "up"
    }
  ]
}
```

**Backend Service:** Routes to service_popularity_system

---

## Timeframe Values

Valid timeframe values for timeframe-based endpoints:
- `day` - Last 24 hours
- `week` - Last 7 days
- `month` - Last 30 days
- `year` - Last 365 days

---

## How It Works

1. **Theme Recommendation Flow:**
   - Client → gateway_api → gateway_recommendation → service_bandit_system
   - Bandit system uses LinUCB algorithm for personalized recommendations
   - Returns theme with feature vector and confidence score

2. **Popularity Flow:**
   - Client → gateway_api → gateway_recommendation → service_popularity_system
   - Popularity system queries aggregated metrics
   - Returns ranked results by popularity score

3. **Service JWT:**
   - All requests must include valid service JWT from gateway_api
   - JWT contains user UUID for personalization
   - Validated by gateway_recommendation before forwarding

---

## Error Responses

- `400 Bad Request` - Invalid parameters (e.g., invalid timeframe)
- `401 Unauthorized` - Invalid or missing service JWT
- `404 Not Found` - Theme not found (for theme-specific endpoints)
- `500 Internal Server Error` - Server error
- `502 Bad Gateway` - Backend service unavailable (bandit/popularity)

---

## Integration Notes

**For Frontend Developers:**
1. Access all these endpoints through `gateway_api` with the `/recommendation` prefix
2. Example: `POST http://localhost:8080/recommendation/recommend/theme`
3. Use your normal JWT token - gateway_api handles service JWT generation

**For Backend Developers:**
1. This gateway validates service JWTs on all protected routes
2. Forwards requests to appropriate backend services
3. Handles response formatting and error propagation
