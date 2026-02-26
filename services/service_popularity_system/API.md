# Service Popularity System - API Documentation

**Base URL:** `http://localhost:8003`
**Service:** Popularity metrics and trending content
**Authentication:** Service JWT (from gateway)

## Overview

The Service Popularity System provides popularity rankings and trending metrics for songs, artists, and themes. It aggregates play counts, likes, followers, and other engagement metrics to generate popularity scores.

## Authentication

All endpoints require a valid Service JWT in the Authorization header.

### Headers
```
Authorization: Bearer <service-jwt-token>
```

---

## All-Time Popularity Endpoints

### Get Popular Songs (All-Time)
```http
GET /popular/songs/all-time
```

**Description:** Retrieve the most popular songs of all time based on total play count and likes.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20, max: 100) - Number of results to return

**Response:**
```json
{
  "songs": [
    {
      "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Song Title",
      "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "artist_name": "Artist Name",
      "album_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "album_title": "Album Title",
      "play_count": 15000,
      "like_count": 2000,
      "popularity_score": 0.95,
      "rank": 1
    }
  ],
  "total_count": 1000000
}
```

**Fields:**
- `popularity_score` - Normalized score (0.0-1.0) based on plays and likes
- `rank` - Position in popularity ranking

**Status Codes:**
- `200 OK` - Successfully retrieved popular songs
- `401 Unauthorized` - Invalid or missing service JWT
- `400 Bad Request` - Invalid limit parameter
- `500 Internal Server Error` - Server error

---

### Get Popular Artists (All-Time)
```http
GET /popular/artists/all-time
```

**Description:** Retrieve the most popular artists of all time based on follower count and total music plays.

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
      "bio": "Artist biography",
      "image_url": "https://storage.example.com/images/artist.jpg",
      "follower_count": 50000,
      "total_plays": 500000,
      "music_count": 120,
      "popularity_score": 0.98,
      "rank": 1
    }
  ],
  "total_count": 50000
}
```

**Popularity Calculation:**
- Weighted combination of follower count and total music plays
- Adjusted for artist activity level and music count

---

### Get Popular Themes (All-Time)
```http
GET /popular/themes/all-time
```

**Description:** Retrieve the most popular music themes/genres of all time.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)

**Response:**
```json
{
  "themes": [
    {
      "theme": "rock",
      "description": "Rock music",
      "play_count": 1000000,
      "song_count": 5000,
      "unique_listeners": 50000,
      "popularity_score": 0.92,
      "rank": 1
    }
  ],
  "total_count": 150
}
```

**Fields:**
- `unique_listeners` - Number of unique users who played songs in this theme
- `song_count` - Total number of songs tagged with this theme

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
  "theme_description": "Rock music",
  "songs": [
    {
      "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Rock Song",
      "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "artist_name": "Rock Artist",
      "play_count": 25000,
      "like_count": 3000,
      "popularity_score": 0.91,
      "rank": 1
    }
  ],
  "total_songs_in_theme": 5000
}
```

**Status Codes:**
- `200 OK` - Successfully retrieved songs
- `404 Not Found` - Theme does not exist
- `401 Unauthorized` - Invalid service JWT

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
- `timeframe` (required) - One of: "day", "week", "month", "year"

**Example:**
```
GET /popular/songs/timeframe?timeframe=week&limit=50
```

**Response:**
```json
{
  "timeframe": "week",
  "start_date": "2024-01-15T00:00:00Z",
  "end_date": "2024-01-22T00:00:00Z",
  "songs": [
    {
      "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Trending Song",
      "artist_name": "Artist Name",
      "play_count": 5000,
      "like_count": 800,
      "previous_rank": 5,
      "current_rank": 1,
      "rank_change": 4,
      "popularity_score": 0.89,
      "trend_direction": "up",
      "velocity": 0.85
    }
  ],
  "total_count": 10000
}
```

**Fields:**
- `previous_rank` - Rank in previous timeframe period
- `current_rank` - Current rank in this timeframe
- `rank_change` - Change in rank (positive = moving up)
- `trend_direction` - "up", "down", or "stable"
- `velocity` - Rate of popularity change (0.0-1.0)

---

### Get Popular Artists (Timeframe)
```http
GET /popular/artists/timeframe
```

**Description:** Get trending artists within a specific time period.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `timeframe` (required) - "day", "week", "month", "year"

**Response:**
```json
{
  "timeframe": "week",
  "start_date": "2024-01-15T00:00:00Z",
  "end_date": "2024-01-22T00:00:00Z",
  "artists": [
    {
      "artist_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Trending Artist",
      "image_url": "https://storage.example.com/images/artist.jpg",
      "new_followers": 500,
      "total_plays": 25000,
      "plays_in_period": 10000,
      "previous_rank": 10,
      "current_rank": 3,
      "rank_change": 7,
      "popularity_score": 0.85,
      "trend_direction": "up",
      "velocity": 0.78
    }
  ],
  "total_count": 5000
}
```

**Fields:**
- `new_followers` - New followers gained in this timeframe
- `plays_in_period` - Plays during this specific timeframe

---

### Get Popular Themes (Timeframe)
```http
GET /popular/themes/timeframe
```

**Description:** Get trending themes/genres within a specific time period.

**Authentication:** Service JWT required

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `timeframe` (required) - "day", "week", "month", "year"

**Response:**
```json
{
  "timeframe": "week",
  "start_date": "2024-01-15T00:00:00Z",
  "end_date": "2024-01-22T00:00:00Z",
  "themes": [
    {
      "theme": "electronic",
      "description": "Electronic music",
      "play_count": 50000,
      "song_count": 200,
      "unique_listeners": 5000,
      "previous_rank": 8,
      "current_rank": 2,
      "rank_change": 6,
      "popularity_score": 0.88,
      "trend_direction": "up",
      "velocity": 0.82
    }
  ],
  "total_count": 150
}
```

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
- `timeframe` (required) - "day", "week", "month", "year"

**Example:**
```
GET /popular/songs/theme/rock/timeframe?timeframe=month&limit=30
```

**Response:**
```json
{
  "theme": "rock",
  "timeframe": "month",
  "start_date": "2023-12-22T00:00:00Z",
  "end_date": "2024-01-22T00:00:00Z",
  "songs": [
    {
      "music_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Trending Rock Song",
      "artist_name": "Rock Artist",
      "play_count": 10000,
      "like_count": 1500,
      "previous_rank": 15,
      "current_rank": 3,
      "rank_change": 12,
      "popularity_score": 0.87,
      "trend_direction": "up",
      "velocity": 0.75
    }
  ],
  "total_songs_in_theme": 5000
}
```

---

## Timeframe Values

Valid timeframe values:

| Timeframe | Duration | Description |
|-----------|----------|-------------|
| `day` | 24 hours | Last 24 hours of activity |
| `week` | 7 days | Last 7 days of activity |
| `month` | 30 days | Last 30 days of activity |
| `year` | 365 days | Last 365 days of activity |

---

## Popularity Score Calculation

Popularity scores are normalized values between 0.0 and 1.0:

### Songs
- **All-Time:** Based on total play count (70%) and like count (30%)
- **Timeframe:** Based on plays in period (60%), likes (20%), and velocity (20%)

### Artists
- **All-Time:** Based on follower count (50%), total plays (40%), and music count (10%)
- **Timeframe:** Based on new followers (40%), plays in period (40%), and velocity (20%)

### Themes
- **All-Time:** Based on total plays (50%), song count (30%), and unique listeners (20%)
- **Timeframe:** Based on plays in period (50%), unique listeners (30%), and velocity (20%)

---

## Trend Direction

The `trend_direction` field indicates momentum:

- `up` - Increasing popularity (positive velocity)
- `down` - Decreasing popularity (negative velocity)
- `stable` - Consistent popularity (near-zero velocity)

**Thresholds:**
- `up`: velocity > 0.1
- `down`: velocity < -0.1
- `stable`: -0.1 ≤ velocity ≤ 0.1

---

## Velocity Calculation

Velocity measures rate of change in popularity:
- Compares current period to previous period
- Values range from -1.0 (rapidly declining) to 1.0 (rapidly growing)
- Calculated as: `(current_score - previous_score) / max(current_score, previous_score)`

---

## Performance Notes

- All endpoints use database indexes for efficient querying
- Results are cached for 5 minutes by default
- All-time rankings are pre-computed hourly
- Timeframe rankings are computed on-demand with caching

---

## Error Responses

| Code | Description |
|------|-------------|
| `400 Bad Request` | Invalid parameters (e.g., invalid timeframe or limit) |
| `401 Unauthorized` | Invalid or missing service JWT |
| `404 Not Found` | Theme not found (for theme-specific endpoints) |
| `500 Internal Server Error` | Database error or server issue |

---

## Integration Notes

**Access via Gateway:**
These endpoints are accessed through `gateway_recommendation` at:
- Base: `http://localhost:8002`
- Then through `gateway_api` at: `http://localhost:8080/recommendation`

**Direct Access:**
Only for internal services with valid service JWT:
- Base: `http://localhost:8003`

**Rate Limiting:**
- Recommended: 100 requests per minute per user
- Burst: Up to 20 simultaneous requests

**Caching:**
- Client-side caching recommended (5-15 minutes for all-time, 1-5 minutes for timeframe)
- Use `ETag` headers for efficient cache validation
