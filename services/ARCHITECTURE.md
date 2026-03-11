# MusicStreamingV2 - Architecture Documentation

**Last Updated:** 2026-03-11

This document describes the architecture, design decisions, and system flows for the MusicStreamingV2 platform. For detailed API endpoint documentation, see [API.md](API.md).

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Service Catalog](#service-catalog)
3. [JWT Architecture](#jwt-architecture)
4. [Authentication Flows](#authentication-flows)
5. [Request Flow Patterns](#request-flow-patterns)
6. [Data Flow](#data-flow)
7. [Design Decisions](#design-decisions)
8. [Port Configuration](#port-configuration)
9. [Related Documentation](#related-documentation)

---

## System Overview

MusicStreamingV2 follows a **microservices architecture** with layered gateways for security, scalability, and separation of concerns.

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Layer                            │
│   ┌──────────────────────────────────────────────────────────┐  │
│   │  Frontend (Next.js) - Port 3000                          │  │
│   │  - Uses Normal JWT for authenticated requests            │  │
│   │  - Uses Refresh JWT for token renewal                    │  │
│   └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ↓ Normal/Refresh JWT
┌─────────────────────────────────────────────────────────────────┐
│                        Gateway Layer                            │
│   ┌──────────────────────────────────────────────────────────┐  │
│   │  gateway_api (Port 8080) - PUBLIC FACING                 │  │
│   │  - Validates user JWTs (Normal/Refresh)                  │  │
│   │  - Generates Service JWTs for backend                    │  │
│   │  - Routes to backend services                            │  │
│   └──────────────────────────────────────────────────────────┘  │
│                              ↓ Service JWT                      │
│   ┌──────────────────────────────────────────────────────────┐  │
│   │  gateway_recommendation (Port 8002)                      │  │
│   │  - Routes recommendation requests                        │  │
│   │  - Validates Service JWT                                 │  │
│   │  - Orchestrates bandit + popularity systems              │  │
│   └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ↓ Service JWT
┌─────────────────────────────────────────────────────────────────┐
│                        Service Layer                            │
│  ┌────────────────────────┐  ┌───────────────────────────────┐  │
│  │  service_user_database │  │  service_popularity_system    │  │
│  │  Port 8001             │  │  Port 8003                    │  │
│  │  - User data           │  │  - Trending metrics           │  │
│  │  - Music catalog       │  │  - Popularity rankings        │  │
│  │  - Playlists/Tags      │  │  - ClickHouse queries         │  │
│  │  - PostgreSQL          │  │                               │  │
│  └────────────────────────┘  └───────────────────────────────┘  │
│  ┌────────────────────────┐  ┌───────────────────────────────┐  │
│  │  service_bandit_system │  │  service_event_ingestion      │  │
│  │  Port 8004             │  │  Port 8080 (internal)         │  │
│  │  - Theme ML recs       │  │  - Event tracking             │  │
│  │  - LinUCB algorithm    │  │  - ClickHouse ingestion       │  │
│  │  - Python/FastAPI      │  │                               │  │
│  └────────────────────────┘  └───────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

**Client Layer:**
- User interface (web frontend)
- Authenticates with user JWTs
- Stores tokens securely (httpOnly cookies)

**Gateway Layer:**
- Public-facing API (gateway_api)
- Authentication validation and transformation
- Request routing and service orchestration
- Service JWT generation

**Service Layer:**
- Business logic implementation
- Data persistence
- ML/analytics processing
- Internal-only access (Service JWT validation)

---

## Service Catalog

### 1. gateway_api

**Port:** 8080
**Language:** Go
**Public Access:** Yes
**Repository:** `services/gateway_api/`

**Purpose:**
Public-facing unified API gateway that serves as the single entry point for all client requests.

**Responsibilities:**
- Accept client requests with Normal or Refresh JWTs
- Validate user tokens (Normal for access, Refresh for renewal)
- Generate Service JWTs containing user context
- Route requests to appropriate backend services
- Handle authentication flows (login, register, token renewal)
- Proxy file requests (public/private)

**Key Routes:**
- `/login`, `/renew` - Authentication
- `/users/*`, `/artists/*`, `/albums/*`, `/music/*` - Data operations (→ service_user_database)
- `/recommend/*`, `/popular/*` - Recommendations (→ gateway_recommendation)
- `/events/*` - Event tracking (→ service_event_ingestion)
- `/files/public/*`, `/files/private/*` - File serving (→ service_user_database)

**Dependencies:**
- service_user_database (user data and authentication)
- gateway_recommendation (recommendations)
- service_event_ingestion (analytics)

**Design Notes:**
- Stateless (no session storage)
- Service JWT generation includes user UUID from validated Normal JWT
- Short Service JWT lifetime (~2 min) for security
- CORS middleware for browser clients

---

### 2. service_user_database

**Port:** 8001 (internal)
**Language:** Go
**Public Access:** No
**Repository:** `services/service_user_database/`

**Purpose:**
Core data backend managing all user data, music content, artists, albums, playlists, tags, and social features.

**Responsibilities:**
- User account management (register, login, profiles)
- Music catalog (tracks, albums, artists)
- Social features (follows, likes)
- Playlist management
- Tag-based organization
- Listening history tracking
- File storage (images, audio via MinIO)
- Search functionality (PostgreSQL full-text)

**Database:** PostgreSQL
**ORM:** sqlc (type-safe SQL code generation)
**File Storage:** MinIO (S3-compatible)

**Key Features:**
- Service JWT validation on all protected routes
- Cursor-based pagination for all list endpoints
- Fuzzy search using PostgreSQL `pg_trgm` extension
- Relative file paths in DB (converted to URLs on read)
- Default images applied dynamically (not stored)

**Dependencies:**
- PostgreSQL database
- MinIO object storage
- HashiCorp Vault (secrets management)

**Design Notes:**
- All routes except `/health`, `/login`, `/files/public/*` require Service JWT
- Uses sqlc for compile-time SQL validation
- File paths stored as relative (e.g., `audio/abc123.mp3`)
- Pagination uses timestamp + UUID cursors for stable ordering

---

### 3. gateway_recommendation

**Port:** 8002 (internal)
**Language:** Go
**Public Access:** No
**Repository:** `services/gateway_recommendation/`

**Purpose:**
Recommendation orchestration gateway that routes recommendation requests to appropriate backend services.

**Responsibilities:**
- Route theme recommendation requests to service_bandit_system
- Route popularity requests to service_popularity_system
- Validate Service JWTs from gateway_api
- Aggregate recommendation data
- Handle failover and error responses

**Key Routes:**
- `/recommend/theme` - Personalized theme recommendation (→ service_bandit_system)
- `/popular/songs/*`, `/popular/artists/*`, `/popular/themes/*` - Popularity rankings (→ service_popularity_system)

**Dependencies:**
- service_bandit_system (ML-based theme recommendations)
- service_popularity_system (trending content)

**Design Notes:**
- Pure orchestration layer (no business logic)
- Validates Service JWT before forwarding
- Handles backend service unavailability gracefully
- Could be extended for A/B testing, multi-model ensembles

---

### 4. service_popularity_system

**Port:** 8003 (internal)
**Language:** Go
**Public Access:** No
**Repository:** `services/service_popularity_system/`

**Purpose:**
Provides popularity rankings and trending metrics for songs, artists, and themes.

**Responsibilities:**
- Calculate popularity scores (normalized 0.0-1.0)
- Provide all-time rankings
- Provide timeframe-based trending (day, week, month, year)
- Compute trend velocity and direction
- Cache frequently accessed rankings

**Data Source:** ClickHouse (OLAP database)
**Metrics:** Decay-weighted play counts and listen time

**Popularity Calculation:**

*Songs (All-Time):*
- 70% total play count
- 30% like count

*Songs (Timeframe):*
- 60% plays in period
- 20% likes
- 20% velocity

*Artists (All-Time):*
- 50% follower count
- 40% total plays
- 10% music count

*Artists (Timeframe):*
- 40% new followers
- 40% plays in period
- 20% velocity

*Themes:*
- 50% total plays
- 30% song count
- 20% unique listeners

**Trend Direction:**
- `"up"` - velocity > 0.1
- `"down"` - velocity < -0.1
- `"stable"` - -0.1 ≤ velocity ≤ 0.1

**Dependencies:**
- ClickHouse data warehouse
- service_event_ingestion (data source)

**Design Notes:**
- All-time rankings pre-computed hourly
- Timeframe rankings computed on-demand with caching
- Uses ClickHouse materialized views for performance
- 5-minute cache for all-time, 1-minute for timeframe

---

### 5. service_bandit_system

**Port:** 8004 (internal)
**Language:** Python (FastAPI)
**Public Access:** No
**Repository:** `services/service_bandit_system/`

**Purpose:**
Personalized theme recommendations using multi-armed bandit machine learning.

**Responsibilities:**
- Provide personalized theme recommendations using LinUCB algorithm
- Learn from user feedback (reward signals)
- Balance exploration (trying new themes) vs exploitation (known preferences)
- Persist user-specific models in PostgreSQL
- Handle cold start for new users

**Algorithm:** LinUCB (Linear Upper Confidence Bound)
**Framework:** FastAPI with NumPy/SciPy
**Model Storage:** PostgreSQL (per-user weight matrices)

**How LinUCB Works:**

1. **Feature Extraction** - User context → feature vector
2. **Prediction** - Calculate UCB for each theme:
   ```
   UCB = θ^T * x + α * sqrt(x^T * A^-1 * x)
   ```
   Where:
   - `θ` = learned weight vector
   - `x` = user feature vector
   - `A` = covariance matrix
   - `α` = exploration parameter (1.0)

3. **Update** - Receive reward (0.0-1.0), update model:
   ```
   A = A + x * x^T
   b = b + reward * x
   θ = A^-1 * b
   ```

**Reward Signal:**
- `1.0` - Loved (listened fully + liked + added to playlist)
- `0.8` - Really liked (listened fully + liked)
- `0.6` - Liked (listened 75%+ or liked)
- `0.4` - Neutral (listened 50-75%)
- `0.2` - Disliked (listened <50%)
- `0.0` - Strongly disliked (skipped immediately)

**Dependencies:**
- PostgreSQL (model persistence)
- service_event_ingestion (feature data source)

**Design Notes:**
- Per-user model personalization
- Cold start falls back to popularity-based recommendations
- Model updates are synchronous (could be async with task queue)
- Feature vectors are 5-20 dimensions
- Models persisted on every update (or batched for performance)

---

### 6. service_event_ingestion

**Port:** 8080 (internal)
**Language:** Go
**Public Access:** No
**Repository:** `services/service_event_ingestion/`

**Purpose:**
Event tracking and data warehouse ingestion for analytics and ML features.

**Responsibilities:**
- Accept event data from services and clients
- Store events in ClickHouse data warehouse
- Track user dimensions (demographics, preferences)
- Track music facts (play events, listen duration)
- Support analytics queries
- Feed recommendation systems

**Data Warehouse:** ClickHouse
**Event Types:**
- User events (login, profile updates)
- Music events (plays, likes, skips)
- Social events (follows, playlist adds)

**Dependencies:**
- ClickHouse database

**Design Notes:**
- Async event processing (fire-and-forget from client perspective)
- ClickHouse optimized for analytical queries
- Events used by service_popularity_system and service_bandit_system
- Could be extended with Apache Kafka for scale

---

### 7. frontend

**Port:** 3000
**Language:** TypeScript (Next.js 14)
**Public Access:** Yes
**Repository:** `services/frontend/`

**Purpose:**
Web-based user interface for the music streaming platform.

**Responsibilities:**
- User authentication (login/register)
- Music browsing and playback
- Artist and album discovery
- Playlist management
- Search functionality
- User profile management
- Personalized recommendations display

**Tech Stack:**
- Next.js 14 (App Router)
- React 18
- TypeScript
- Tailwind CSS
- Axios for API communication

**Features:**
- Responsive design
- Dark theme UI
- JWT-based authentication with httpOnly cookies
- Music player with play/pause/skip controls
- Real-time search
- Infinite scroll pagination

**Dependencies:**
- gateway_api (all backend communication)

**Design Notes:**
- Server-side rendering for SEO
- Client-side state management for player
- Secure token storage (httpOnly cookies)
- CORS configured for gateway_api

---

## JWT Architecture

### Token Types

| Type            | Subject     | Lifetime    | Issued By             | Used By               | Purpose                                        |
|-----------------|-------------|-------------|-----------------------|-----------------------|------------------------------------------------|
| **Normal JWT**  | `"normal"`  | ~10 minutes | service_user_database | Clients → gateway_api | User access token for API requests             |
| **Refresh JWT** | `"refresh"` | ~10 days    | service_user_database | Clients → gateway_api | Token renewal without re-authentication        |
| **Service JWT** | `"service"` | ~2 minutes  | gateway_api           | Gateways → services   | Inter-service authentication with user context |

**Configuration:**
Token lifetimes are configurable via `libs/consts/jwt_and_ctx.go` with environment variable overrides:
- `JWT_EXPIRATION_NORMAL` (default: 10 minutes)
- `JWT_EXPIRATION_REFRESH` (default: 10 days)
- `JWT_EXPIRATION_SERVICE` (default: 2 minutes)

### JWT Claims

**Normal JWT:**
```json
{
  "sub": "normal",
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "iat": 1640000000,
  "exp": 1640000600
}
```

**Refresh JWT:**
```json
{
  "sub": "refresh",
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "iat": 1640000000,
  "exp": 1640864000
}
```

**Service JWT:**
```json
{
  "sub": "service",
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "iat": 1640000000,
  "exp": 1640000120
}
```

**Key Points:**
- `sub` (subject) differentiates token types
- Service JWT contains `user_uuid` for user context in backend
- Service JWT has shortest lifetime for security
- All JWTs are signed with HS256 (symmetric secret from Vault)

---

## Authentication Flows

### 1. User Login Flow

```
┌────────┐                 ┌─────────────┐                 ┌──────────────────┐
│ Client │                 │ gateway_api │                 │ service_user_db  │
└────┬───┘                 └──────┬──────┘                 └────────┬─────────┘
     │                             │                                │
     │ POST /login                 │                                │
     │ {email, password}           │                                │
     ├────────────────────────────>│                                │
     │                             │                                │
     │                             │ POST /login                    │
     │                             │ {email, password}              │
     │                             ├───────────────────────────────>│
     │                             │                                │
     │                             │    Validate credentials        │
     │                             │    Hash password & compare     │
     │                             │    Generate Normal JWT         │
     │                             │    Generate Refresh JWT        │
     │                             │                                │
     │                             │ {normal_token, refresh_token,  │
     │                             │  user_uuid}                    │
     │                             │<───────────────────────────────┤
     │                             │                                │
     │ {normal_token,              │                                │
     │  refresh_token,             │                                │
     │  user_uuid}                 │                                │
     │<────────────────────────────┤                                │
     │                             │                                │
     │ Store tokens securely       │                                │
     │ (httpOnly cookies)          │                                │
     │                             │                                │
```

**Steps:**
1. Client sends credentials to gateway_api
2. Gateway forwards to service_user_database
3. Service validates password hash
4. Service generates Normal + Refresh JWTs
5. Tokens returned through gateway to client
6. Client stores tokens securely

---

### 2. Protected Request Flow

```
┌────────┐         ┌─────────────┐         ┌──────────────────┐
│ Client │         │ gateway_api │         │ Backend Service  │
└────┬───┘         └──────┬──────┘         └────────┬─────────┘
     │                     │                         │
     │ GET /users/me       │                         │
     │ Auth: Normal JWT    │                         │
     ├────────────────────>│                         │
     │                     │                         │
     │                     │ Validate Normal JWT     │
     │                     │ Extract user_uuid       │
     │                     │                         │
     │                     │ Generate Service JWT    │
     │                     │ {sub: "service",        │
     │                     │  user_uuid: "..."}      │
     │                     │                         │
     │                     │ GET /users/me           │
     │                     │ Auth: Service JWT       │
     │                     ├────────────────────────>│
     │                     │                         │
     │                     │     Validate Service JWT│
     │                     │     Extract user_uuid   │
     │                     │     Query database      │
     │                     │                         │
     │                     │ {user_data}             │
     │                     │<────────────────────────┤
     │                     │                         │
     │ {user_data}         │                         │
     │<────────────────────┤                         │
     │                     │                         │
```

**Steps:**
1. Client sends request with Normal JWT
2. Gateway validates Normal JWT
3. Gateway extracts `user_uuid` from Normal JWT
4. Gateway generates Service JWT with `user_uuid`
5. Gateway forwards request with Service JWT
6. Backend validates Service JWT
7. Backend uses `user_uuid` for authorization
8. Response flows back through gateway

**Key Design Points:**
- Service JWT is short-lived (~2 min) for security
- User context (UUID) is passed through Service JWT claims
- Backend services trust Service JWT (internal network)
- Gateway acts as authentication boundary

---

### 3. Token Renewal Flow

```
┌────────┐                 ┌─────────────┐                 ┌──────────────────┐
│ Client │                 │ gateway_api │                 │ service_user_db  │
└────┬───┘                 └──────┬──────┘                 └────────┬─────────┘
     │                             │                                │
     │ POST /renew                 │                                │
     │ Auth: Refresh JWT           │                                │
     ├────────────────────────────>│                                │
     │                             │                                │
     │                             │ Validate Refresh JWT           │
     │                             │ (subject == "refresh")         │
     │                             │ Extract user_uuid              │
     │                             │                                │
     │                             │ Generate Service JWT           │
     │                             │                                │
     │                             │ POST /renew                    │
     │                             │ Auth: Service JWT              │
     │                             ├───────────────────────────────>│
     │                             │                                │
     │                             │    Validate Service JWT        │
     │                             │    Generate new Normal JWT     │
     │                             │                                │
     │                             │ {normal_token, user_uuid}      │
     │                             │<───────────────────────────────┤
     │                             │                                │
     │ {normal_token,              │                                │
     │  user_uuid}                 │                                │
     │<────────────────────────────┤                                │
     │                             │                                │
     │ Update stored Normal JWT    │                                │
     │ Keep existing Refresh JWT   │                                │
     │                             │                                │
```

**Steps:**
1. Client sends Refresh JWT when Normal JWT expires
2. Gateway validates Refresh JWT (checks `sub == "refresh"`)
3. Gateway generates Service JWT
4. Gateway forwards to service_user_database
5. Service generates new Normal JWT
6. Client receives new Normal JWT
7. Client updates stored Normal JWT (keeps Refresh JWT)

**Security Notes:**
- Refresh JWT is only valid for `/renew` endpoint
- Normal JWT cannot be used for `/renew`
- Refresh JWT has much longer lifetime
- If Refresh JWT expires, user must login again

---

## Request Flow Patterns

### Example 1: Get User Profile

**Client Request:**
```http
GET /users/me
Authorization: Bearer eyJhbGc...
```

**Flow:**
```
Client → gateway_api → service_user_database → Response
```

**Detailed Steps:**
1. Client sends Normal JWT to `gateway_api:8080/users/me`
2. gateway_api validates Normal JWT (checks signature, expiration, subject)
3. gateway_api extracts `user_uuid` from Normal JWT claims
4. gateway_api generates Service JWT with `user_uuid`
5. gateway_api forwards to `service_user_database:8001/users/me` with Service JWT
6. service_user_database validates Service JWT
7. service_user_database extracts `user_uuid` from Service JWT
8. service_user_database queries PostgreSQL for user data
9. service_user_database returns user profile
10. gateway_api forwards response to client

---

### Example 2: Get Theme Recommendation

**Client Request:**
```http
POST /recommend/theme
Authorization: Bearer eyJhbGc...
```

**Flow:**
```
Client → gateway_api → gateway_recommendation → service_bandit_system → Response
```

**Detailed Steps:**
1. Client sends Normal JWT to `gateway_api:8080/recommend/theme`
2. gateway_api validates Normal JWT
3. gateway_api generates Service JWT with `user_uuid`
4. gateway_api forwards to `gateway_recommendation:8002/recommend/theme` with Service JWT
5. gateway_recommendation validates Service JWT
6. gateway_recommendation extracts `user_uuid` from Service JWT
7. gateway_recommendation forwards to `service_bandit_system:8004/api/v1/predict`
8. service_bandit_system runs LinUCB prediction for user
9. service_bandit_system returns `{theme, features, confidence}`
10. Response flows back: bandit → gateway_recommendation → gateway_api → client

**Design Rationale:**
- gateway_recommendation provides abstraction layer
- Could switch ML backend without changing gateway_api
- Could implement A/B testing in gateway_recommendation
- Could aggregate multiple recommendation sources

---

### Example 3: Get Trending Songs

**Client Request:**
```http
GET /popular/songs/timeframe?timeframe=week&limit=50
Authorization: Bearer eyJhbGc...
```

**Flow:**
```
Client → gateway_api → gateway_recommendation → service_popularity_system → Response
```

**Detailed Steps:**
1. Client sends Normal JWT to `gateway_api:8080/popular/songs/timeframe?timeframe=week&limit=50`
2. gateway_api validates Normal JWT
3. gateway_api generates Service JWT
4. gateway_api forwards to `gateway_recommendation:8002/popular/songs/timeframe`
5. gateway_recommendation validates Service JWT
6. gateway_recommendation forwards to `service_popularity_system:8003/popular/songs/timeframe`
7. service_popularity_system queries ClickHouse for trending songs
8. service_popularity_system calculates popularity scores and velocity
9. service_popularity_system returns ranked songs with trend direction
10. Response flows back through gateways to client

**Performance Notes:**
- Timeframe queries cached for 1-5 minutes
- ClickHouse optimized for analytical queries
- Could add Redis cache layer in gateway_recommendation

---

## Data Flow

### File Storage Architecture

**Design Decision:** Store relative paths in database, convert to URLs on read.

**Rationale:**
- Storage backend agnostic (can switch from MinIO to S3/GCS without DB migration)
- Easy to change CDN endpoints
- Smaller database storage footprint
- Single source of truth for storage configuration

**Implementation:**

*Database Storage:*
```sql
-- Stored in DB
image_path = "pictures-profile/abc123.jpg"
audio_path = "audio/xyz789.mp3"
```

*Conversion on Read (in handlers):*
```go
func convertPathToPublicURL(path string) string {
    if path == "" {
        return ""
    }
    return fmt.Sprintf("http://%s/files/public/%s",
        config.PublicEndpoint, path)
}
```

*Response to Client:*
```json
{
  "image_url": "http://localhost:8001/files/public/pictures-profile/abc123.jpg",
  "audio_url": "http://localhost:8001/files/public/audio/xyz789.mp3"
}
```

**Default Images:**
- Applied dynamically on read (not stored in DB)
- Function: `applyDefaultImageIfEmpty(imageURL, resourceType)`
- Default paths: `pictures-profile/default.jpg`, `pictures-artist/default.jpg`, etc.

---

### Pagination Architecture

**Design Decision:** Cursor-based pagination using timestamp + UUID.

**Rationale:**
- Stable results (no skipping/duplicates on concurrent writes)
- Efficient database queries (indexed columns)
- No "page drift" problem of offset-based pagination
- Works well with real-time data

**Implementation:**

*Query Parameters:*
```
?limit=20&cursor=2024-01-01T00:00:00Z&cursor_id=123e4567-...
```

*SQL Query:*
```sql
SELECT * FROM music
WHERE (created_at, uuid) < ($1, $2)
ORDER BY created_at DESC, uuid DESC
LIMIT $3
```

*Response:*
```json
{
  "music": [...],
  "next_cursor": "2023-12-31T23:59:59Z",
  "next_cursor_id": "987e6543-..."
}
```

**Cursor Types:**
- `cursor` + `cursor_id` - Most list endpoints (timestamp + UUID)
- `cursor_name` - Tag/artist alphabetical lists (name-based)
- `cursor_pos` - Playlist tracks (position-based)
- `cursor_count` - Top music (play count-based)
- `cursor_score` + `cursor_ts` - Search results (similarity + timestamp)

**Helper Functions (in handlers/helpers.go):**
- `parsePagination(r)` → `(limit, cursorTS, cursorID)`
- `parsePaginationName(r)` → `(limit, cursorName)`
- `parsePaginationAlpha(r)` → `(limit, cursorName, cursorTS)`
- `parsePaginationPos(r)` → `(limit, cursorPos)`

---

## Design Decisions

### 1. Why Two Gateway Layers?

**Decision:** gateway_api (public) + gateway_recommendation (internal orchestration)

**Rationale:**
- **Separation of Concerns:** gateway_api handles authentication, gateway_recommendation handles ML orchestration
- **Security Boundary:** Only gateway_api exposed to internet
- **Flexibility:** Can swap recommendation backends without changing public API
- **A/B Testing:** gateway_recommendation can implement experimentation logic
- **Scalability:** Can scale recommendation gateway independently

**Trade-offs:**
- Additional network hop (latency +5-10ms)
- More services to deploy and monitor
- Complexity in request routing

**Alternative Considered:**
Single gateway with recommendation routing logic embedded - rejected because it couples authentication logic with ML orchestration.

---

### 2. Why Service JWTs Instead of API Keys?

**Decision:** Generate short-lived Service JWTs from validated user JWTs

**Rationale:**
- **User Context:** Service JWT contains `user_uuid` for authorization
- **Security:** Short lifetime (~2 min) limits blast radius if compromised
- **Stateless:** No need for session storage or token revocation
- **Standard:** Uses JWT standard, compatible with existing libraries
- **Audit Trail:** Can log which user triggered each backend request

**Trade-offs:**
- JWT generation overhead (negligible with modern crypto)
- Slightly larger request headers vs API keys
- Clock synchronization required (NTP)

**Alternative Considered:**
Static API keys per service - rejected because no user context, manual rotation burden.

---

### 3. Why Relative File Paths in Database?

**Decision:** Store `audio/abc123.mp3` instead of `http://minio:9000/audio/abc123.mp3`

**Rationale:**
- **Backend Agnostic:** Can switch from MinIO to S3/GCS/Cloudflare R2 without DB migration
- **Environment Flexibility:** Dev/staging/prod can use different endpoints
- **CDN Integration:** Easy to add CDN layer by changing URL construction
- **Smaller DB:** Saves ~50 bytes per file path record
- **No Stale URLs:** Always generates current endpoint

**Trade-offs:**
- Must convert on every read (minimal overhead)
- URL construction logic in application code

**Alternative Considered:**
Store full URLs - rejected due to inflexibility and migration burden.

---

### 4. Why Cursor-Based Instead of Offset Pagination?

**Decision:** Use `cursor` (timestamp) + `cursor_id` (UUID) for pagination

**Rationale:**
- **Stability:** No skipped/duplicate results on concurrent writes
- **Performance:** Index scans vs table scans (offset requires scanning all previous rows)
- **Real-Time:** Works with constantly changing data
- **Consistent:** User sees stable results while paginating

**Trade-offs:**
- Cannot jump to arbitrary page number
- More complex query logic
- Cursors are opaque to clients

**Alternative Considered:**
Offset/limit - rejected due to page drift and performance issues at high offsets.

---

### 5. Why LinUCB Instead of Deep Learning for Recommendations?

**Decision:** Use LinUCB (contextual bandit) for theme recommendations

**Rationale:**
- **Exploration/Exploitation:** Naturally balances trying new themes vs known preferences
- **Cold Start:** Works well with limited user data
- **Interpretable:** Can explain why theme was recommended
- **Efficient:** Low latency (~10-50ms), small model size
- **Online Learning:** Updates in real-time from user feedback

**Trade-offs:**
- Linear assumption (limited expressiveness vs deep models)
- Requires feature engineering
- May plateau with very large datasets

**Alternative Considered:**
Collaborative filtering or neural networks - rejected due to cold start issues and latency requirements.

---

### 6. Why PostgreSQL for User Data and ClickHouse for Analytics?

**Decision:** PostgreSQL (OLTP) for transactional data, ClickHouse (OLAP) for analytics

**Rationale:**

*PostgreSQL:*
- ACID transactions for user accounts, music catalog
- Strong consistency for critical data
- Mature ecosystem (sqlc, pgx)
- Full-text search (pg_trgm)

*ClickHouse:*
- Columnar storage for analytical queries
- 100x faster for aggregations (plays, likes over time)
- Handles high write throughput (event ingestion)
- Materialized views for pre-aggregation

**Trade-offs:**
- Two databases to manage
- Data synchronization (via event_ingestion service)
- Increased operational complexity

**Alternative Considered:**
PostgreSQL only - rejected due to poor analytical query performance at scale.

---

## Port Configuration

| Service                   | Port | Protocol | Public Access            | Docker Network         |
|---------------------------|------|----------|--------------------------|------------------------|
| **gateway_api**           | 8080 | HTTP     | **Yes** (mapped to host) | musicstreaming_network |
| service_user_database     | 8001 | HTTP     | Internal only            | musicstreaming_network |
| gateway_recommendation    | 8002 | HTTP     | Internal only            | musicstreaming_network |
| service_popularity_system | 8003 | HTTP     | Internal only            | musicstreaming_network |
| service_bandit_system     | 8004 | HTTP     | Internal only            | musicstreaming_network |
| service_event_ingestion   | 8080 | HTTP     | Internal only            | musicstreaming_network |
| **frontend**              | 3000 | HTTP     | **Yes** (mapped to host) | musicstreaming_network |
| PostgreSQL                | 5432 | TCP      | Internal only            | musicstreaming_network |
| ClickHouse                | 9000 | TCP      | Internal only            | musicstreaming_network |
| MinIO                     | 9000 | HTTP     | Internal only            | musicstreaming_network |
| MinIO Console             | 9001 | HTTP     | Dev only                 | musicstreaming_network |
| HashiCorp Vault           | 8200 | HTTP     | Internal only            | musicstreaming_network |

**Security Notes:**
- Only `gateway_api` (8080) and `frontend` (3000) are exposed to public internet
- All other services are internal-only within Docker network
- Production should add HTTPS/TLS termination at gateway_api
- Production should use cloud-managed databases (RDS, Cloud SQL, etc.)

**Port Conflicts:**
- service_event_ingestion uses port 8080 internally (not exposed)
- MinIO uses port 9000 internally (same as ClickHouse, different containers)

---

## Related Documentation

**API Documentation:**
- [API.md](API.md) - Complete API reference for all services

**Development Documentation:**
- [.thoughts/DOCUMENTATION_STRATEGY.md](../.thoughts/DOCUMENTATION_STRATEGY.md) - How to maintain documentation
- [.thoughts/LOGGING_STRATEGY.md](../.thoughts/LOGGING_STRATEGY.md) - Logging guidelines and best practices
- [CLAUDE.md](../CLAUDE.md) - Project-specific instructions

**Service-Specific Documentation:**
- `services/gateway_api/README.md` - Deployment and development setup
- `services/service_user_database/README.md` - Database schema and migrations
- `services/service_bandit_system/README.md` - ML model details
- `services/frontend/README.md` - Frontend setup and development

---

**Document Version:** 1.0
**Last Updated:** 2026-03-06
**Maintained By:** Development Team
