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

Endpoints that accept file uploads use `multipart/form-data`. There are two types of file uploads: image uploads and audio uploads.

#### Image Uploads

**Endpoints:**
| Endpoint | Description | Required Fields |
|----------|-------------|-----------------|
| `POST /users/me/image` | Upload user profile image | `image` |
| `POST /artists/{uuid}/image` | Upload artist image | `image` |
| `POST /albums/{uuid}/image` | Upload album cover art | `image` |
| `POST /music/{uuid}/image` | Upload music track cover art | `image` |
| `POST /playlists/{uuid}/image` | Upload playlist cover image | `image` |

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
- Profile images: 10 MB max
- Album/Music/Playlist/Artist images: 10 MB max

**Response:**
```json
{
  "image_url": "http://localhost:8001/files/public/pictures-profile/abc123.jpg"
}
```

#### Audio/Music Uploads

**Endpoints:**
| Endpoint | Description | Required Fields | Optional Fields |
|----------|-------------|-----------------|-----------------|
| `POST /artists/{uuid}/music` | Upload music track with audio file | `title`, `duration_seconds`, `audio` | `in_album`, `image` |

**Request Format:**
```
Content-Type: multipart/form-data

title: "Song Title"
duration_seconds: 180
audio: [audio file]
in_album: "album-uuid (optional)"
image: [cover image (optional)]
```

**Supported Audio Formats:**
- MP3 (`.mp3`)
- WAV (`.wav`)
- FLAC (`.flac`)
- OGG (`.ogg`)
- AAC (`.aac`)
- M4A (`.m4a`)

**Audio Size Limits:**
- Max file size: 100 MB per track

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Song Title",
  "artist_uuid": "123e4567-e89b-12d3-a456-426614174001",
  "artist_name": "Artist Name",
  "duration_seconds": 180,
  "image_url": "http://localhost:8001/files/public/pictures-music/music.jpg",
  "audio_url": "http://localhost:8001/files/public/audio/abc123.mp3",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Status Codes for File Uploads:**
- `200 OK` - Image uploaded successfully
- `201 Created` - Music track created successfully
- `400 Bad Request` - Invalid file format, size exceeded, or missing required fields
- `403 Forbidden` - Not authorized (must be artist member for music uploads)
- `404 Not Found` - Artist or album not found
- `415 Unsupported Media Type` - Wrong Content-Type (must be multipart/form-data)
- `500 Internal Server Error` - Server error

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

### Role-Based Access Control

The API implements role-based access control (RBAC) for artist resources and playlist ownership.

#### Artist Member Roles

| Role | Description | Permissions |
|------|-------------|-------------|
| `owner` | Full control | Can manage all aspects of the artist including members, content, settings, and transfer ownership |
| `admin` | Manage members and content | Can add/remove members, upload music, create albums, edit artist info |
| `member` | Upload content only | Can upload music tracks and images for the artist |

**Role Hierarchy:**
```
owner > admin > member
```

**Role Requirements:**
- Only the `owner` can transfer ownership to another member
- Only `owner` and `admin` can manage artist members (add, remove, change roles)
- All members (`owner`, `admin`, `member`) can upload music and images
- All members can view artist information

**Changing Roles:**
Use `POST /artists/{uuid}/members/{userUuid}/role` to update a member's role. The request body should be:
```json
{
  "role": "admin"  // or "member"
}
```

**Response:**
```json
{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "username": "user123",
  "role": "admin",
  "joined_at": "2024-01-01T00:00:00Z"
}
```

#### Playlist Ownership

Playlists have a simple owner model:
- The creator of a playlist is automatically the `owner`
- Only the owner can:
  - Update playlist details (name, description, visibility)
  - Upload/change playlist cover image
  - Delete the playlist
  - Add or remove tracks from the playlist
  - Reorder tracks within the playlist

**Playlist Visibility:**
- Public playlists can be viewed by any authenticated user
- Private playlists can only be viewed/modified by the owner

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

| Code | Name | Description |
|------|------|-------------|
| `200 OK` | Success | Successful GET, PUT, PATCH, DELETE operations |
| `201 Created` | Created | Successful resource creation (POST resulting in new entity) |
| `202 Accepted` | Accepted | Request accepted for async processing |
| `400 Bad Request` | Bad Request | Invalid input, malformed JSON, validation failure, invalid parameters |
| `401 Unauthorized` | Unauthorized | Invalid/missing/expired JWT token |
| `403 Forbidden` | Forbidden | Valid JWT but insufficient permissions (role-based or ownership) |
| `404 Not Found` | Not Found | Resource does not exist |
| `409 Conflict` | Conflict | Resource already exists, duplicate entry, constraint violation |
| `415 Unsupported Media Type` | Unsupported Media Type | Wrong Content-Type header (e.g., not multipart/form-data for file uploads) |
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

**Permission Denied (Role-Based):**
```json
{
  "error": "Forbidden",
  "detail": "You are not authorized to modify this resource"
}
```
Status: `403 Forbidden`

**Permission Denied (Not Owner):**
```json
{
  "error": "Forbidden",
  "detail": "Not owner of this playlist"
}
```
Status: `403 Forbidden`

**Resource Conflict:**
```json
{
  "error": "Conflict",
  "detail": "Artist name already exists"
}
```
Status: `409 Conflict`

**Wrong Content-Type:**
```json
{
  "error": "Unsupported Media Type",
  "detail": "Content-Type must be multipart/form-data"
}
```
Status: `415 Unsupported Media Type`

**Service Unavailable:**
```json
{
  "error": "Backend service unavailable"
}
```
Status: `502 Bad Gateway`

---

## Related Documentation

- [ARCHITECTURE.md](../ARCHITECTURE.md) - System architecture, design decisions, and request flows
- [API_service_user_database.md](API_service_user_database.md) - User database service API reference
- [API_gateway_api.md](API_gateway_api.md) - Gateway API reference

---

**Document Version:** 1.2
**Last Updated:** 2026-03-16
**Maintained By:** Development Team
