# MusicStreamingV2 - Services Overview

This document provides an overview of all microservices in the MusicStreamingV2 architecture and how they interact.

## Architecture Overview

```
┌──────────────────────────────────────────────────────┐
│                    Client (Browser)                   │
└──────────────────────────┬───────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────┐
│         frontend (Next.js - Port 3000)               │
│  - Web UI for music streaming                        │
│  - User authentication                               │
│  - Music player interface                            │
└──────────────────────────┬───────────────────────────┘
                           │ User JWT (Normal/Refresh)
                           ▼
┌──────────────────────────────────────────────────────┐
│         gateway_api (Port 8080)                      │
│  - Public-facing API Gateway                         │
│  - Validates user JWTs                               │
│  - Generates service JWTs                            │
│  - Routes to backend services                        │
└──────────────┬─────────────────────────┬─────────────┘
               │ Service JWT             │ Service JWT
               ▼                         ▼
┌──────────────────────────┐  ┌──────────────────────────┐
│ service_user_database    │  │ gateway_recommendation   │
│  (Port 8001)             │  │    (Port 8002)           │
│  - User management       │  │ - Routes recommendations │
│  - Music catalog         │  └─────┬───────────┬────────┘
│  - Playlists             │        │           │ Service JWT
└──────────────────────────┘        │           ▼
                                    │  ┌─────────────────────┐
                                    │  │ service_bandit      │
                                    │  │  (Port 8004)        │
                                    │  │ - LinUCB algorithm  │
                                    │  └─────────────────────┘
                                    │ Service JWT
                                    ▼
                           ┌─────────────────────┐
                           │ service_popularity  │
                           │  (Port 8003)        │
                           │ - Trending content  │
                           └─────────────────────┘
```

---

## Services

### 1. gateway_api
**Port:** 8080\
**Language:** Go\
**Role:** Public-facing unified API gateway

**Responsibilities:**
- Accept client requests with user JWTs
- Validate normal and refresh tokens
- Generate service JWTs for backend communication
- Route requests to appropriate backend services
- Handle authentication flows (login, register, renew)

**[Full API Documentation →](gateway_api/API.md)**

**Key Endpoints:**
- `POST /login` - User authentication
- `PUT /login` - User registration
- `POST /renew` - Token refresh
- `/users/*`, `/artists/*`, `/albums/*`, `/music/*`, `/tags/*`, `/playlists/*`, `/history/*` - Proxy to user database
- `/recommendation/*` - Proxy to recommendation gateway

---

### 2. service_user_database
**Port:** 8001\
**Language:** Go\
**Role:** Core data backend for users, music, and content

**Responsibilities:**
- Manage user accounts and profiles
- Store and serve music metadata
- Handle artist and album information
- Manage playlists and tags
- Track likes, follows, and listening history
- Validate service JWTs on all protected routes

**[Full API Documentation →](service_user_database/API.md)**

**Database:** PostgreSQL\
**ORM:** sqlc (type-safe SQL)

**Key Features:**
- User authentication (login, register, renew)
- Social features (follows, likes)
- Music catalog management
- Playlist creation and management
- Tag-based organization
- Listening history tracking

---

### 3. gateway_recommendation
**Port:** 8002\
**Language:** Go\
**Role:** Recommendation orchestration gateway

**Responsibilities:**
- Route recommendation requests to appropriate services
- Coordinate between bandit and popularity systems
- Validate service JWTs
- Aggregate recommendation data

**[Full API Documentation →](gateway_recommendation/API.md)**

**Key Endpoints:**
- `POST /recommend/theme` - Get personalized theme recommendation (→ bandit)
- `GET /popular/songs/all-time` - Get popular songs (→ popularity)
- `GET /popular/artists/all-time` - Get popular artists (→ popularity)
- `GET /popular/themes/all-time` - Get popular themes (→ popularity)
- Timeframe endpoints for trending content

---

### 4. service_popularity_system
**Port:** 8003\
**Language:** Go\
**Role:** Popularity metrics and trending content

**Responsibilities:**
- Calculate popularity scores for songs, artists, and themes
- Provide all-time rankings
- Provide timeframe-based trending data (day, week, month, year)
- Compute velocity and trend direction
- Cache frequently accessed rankings

**[Full API Documentation →](service_popularity_system/API.md)**

**Key Features:**
- All-time popularity rankings
- Timeframe-based trending (day/week/month/year)
- Popularity score calculation (0.0-1.0)
- Trend velocity and direction
- Theme-specific popularity

---

### 5. service_bandit_system
**Port:** 8004\
**Language:** Python (FastAPI)\
**Role:** Personalized theme recommendations using ML

**Responsibilities:**
- Provide personalized theme recommendations using LinUCB algorithm
- Learn from user feedback (rewards)
- Balance exploration vs exploitation
- Persist user-specific models

**[Full API Documentation →](service_bandit_system/API.md)**

**Algorithm:** LinUCB (Linear Upper Confidence Bound)
**Framework:** FastAPI with NumPy/SciPy

**Key Features:**
- Contextual multi-armed bandit
- Per-user personalization
- Reinforcement learning from rewards
- Model persistence
- Cold start handling

---

### 6. frontend
**Port:** 3000\
**Language:** TypeScript (Next.js)\
**Role:** Web-based user interface for the music streaming platform

**Responsibilities:**
- User authentication (login/register)
- Music browsing and playback
- Artist and album discovery
- Playlist management
- Search functionality
- User profile management

**Tech Stack:**
- Next.js 14 (App Router)
- React 18
- TypeScript
- Tailwind CSS
- Axios for API communication

**Features:**
- Responsive design
- Dark theme UI
- JWT-based authentication with secure cookie storage
- Music player with play/pause/skip controls
- Artist, album, and playlist browsing
- Personalized recommendations
- Search for songs, artists, and albums

---

## JWT Architecture

### JWT Types

| Type | Subject | Lifetime | Purpose | Issued By |
|------|---------|----------|---------|-----------|
| Normal JWT | "normal" | ~10 min | User access token | service_user_database |
| Refresh JWT | "refresh" | ~10 days | Token renewal | service_user_database |
| Service JWT | "service" | ~2 min | Inter-service auth | gateway_api |

### Authentication Flow

```
1. Client → gateway_api: POST /login with credentials
2. gateway_api → service_user_database: Forward login request
3. service_user_database: Validate credentials
4. service_user_database → gateway_api: Return normal + refresh JWTs
5. gateway_api → Client: Return JWTs

Protected Request Flow:
1. Client → gateway_api: Request with normal JWT
2. gateway_api: Validate normal JWT
3. gateway_api: Generate service JWT (contains user UUID)
4. gateway_api → backend service: Forward with service JWT
5. backend service: Validate service JWT
6. backend service: Process request
7. backend service → gateway_api → Client: Return response
```

### Service JWT Claims

```json
{
  "sub": "service",
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "iat": 1640000000,
  "exp": 1640000120
}
```

---

## Request Flow Examples

### Example 1: Get User Profile

```
1. Client: GET /users/me
   Authorization: Bearer <normal-jwt>

2. gateway_api:
   - Validates normal JWT
   - Extracts user UUID
   - Generates service JWT
   - Forwards to service_user_database

3. service_user_database:
   - Validates service JWT
   - Retrieves user data
   - Returns profile

4. gateway_api → Client: Profile data
```

### Example 2: Get Theme Recommendation

```
1. Client: POST /recommendation/recommend/theme
   Authorization: Bearer <normal-jwt>
   Body: {"user_uuid": "..."}

2. gateway_api:
   - Validates normal JWT
   - Generates service JWT
   - Forwards to gateway_recommendation

3. gateway_recommendation:
   - Validates service JWT
   - Forwards to service_bandit_system

4. service_bandit_system:
   - Extracts user UUID from service JWT
   - Runs LinUCB prediction
   - Returns theme + features

5. gateway_recommendation → gateway_api → Client:
   {"theme": "rock", "features": [...]}
```

### Example 3: Get Trending Songs

```
1. Client: GET /recommendation/popular/songs/timeframe?timeframe=week
   Authorization: Bearer <normal-jwt>

2. gateway_api:
   - Validates normal JWT
   - Generates service JWT
   - Forwards to gateway_recommendation

3. gateway_recommendation:
   - Validates service JWT
   - Forwards to service_popularity_system

4. service_popularity_system:
   - Validates service JWT
   - Queries popularity metrics
   - Returns trending songs with velocity

5. gateway_recommendation → gateway_api → Client:
   Trending songs list
```

---

## Port Configuration

| Service | Port | Protocol | Public Access |
|---------|------|----------|---------------|
| gateway_api | 8080 | HTTP | Yes |
| service_user_database | 8001 | HTTP | Internal only |
| gateway_recommendation | 8002 | HTTP | Internal only |
| service_popularity_system | 8003 | HTTP | Internal only |
| service_bandit_system | 8004 | HTTP | Internal only |

**Note:** Only `gateway_api` should be exposed to public internet. All other services should be internal-only.

## API Documentation

- **[Gateway API](gateway_api/API.md)** - Public-facing gateway (port 8080)
- **[Service User Database](service_user_database/API.md)** - Core data backend (port 8001)
- **[Gateway Recommendation](gateway_recommendation/API.md)** - Recommendation gateway (port 8002)
- **[Service Popularity System](service_popularity_system/API.md)** - Popularity metrics (port 8003)
- **[Service Bandit System](service_bandit_system/API.md)** - ML recommendations (port 8004)
