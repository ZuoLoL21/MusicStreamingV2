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

- [ARCHITECTURE.md](../ARCHITECTURE.md) - System architecture, design decisions, and request flows
- [.thoughts/DOCUMENTATION_STRATEGY.md](../../.thoughts/DOCUMENTATION_STRATEGY.md) - Documentation maintenance rules
- [.thoughts/LOGGING_STRATEGY.md](../../.thoughts/LOGGING_STRATEGY.md) - Logging guidelines

---

**Document Version:** 1.0
**Last Updated:** 2026-03-06
**Maintained By:** Development Team

