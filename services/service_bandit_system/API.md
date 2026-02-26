# Service Bandit System - API Documentation

**Base URL:** `http://localhost:8004/api/v1`
**Service:** Theme recommendation using multi-armed bandit (LinUCB)
**Framework:** FastAPI (Python)
**Authentication:** Service JWT (from gateway)

## Overview

The Service Bandit System provides personalized theme recommendations using the LinUCB (Linear Upper Confidence Bound) multi-armed bandit algorithm. It learns user preferences over time by balancing exploration of new themes and exploitation of known preferences.

## Authentication

Protected endpoints require a valid Service JWT in the Authorization header.

### Headers
```
Authorization: Bearer <service-jwt-token>
```

---

## Endpoints

### Health Check
```http
GET /api/v1/health
```

**Description:** Health check endpoint to verify service availability.

**Authentication:** None

**Response:**
```json
{
  "status": "healthy",
  "service": "bandit-system"
}
```

**Status Codes:**
- `200 OK` - Service is healthy and operational

---

### Predict Theme
```http
POST /api/v1/predict
```

**Description:** Get personalized theme recommendation for a user using the LinUCB algorithm. Returns the most promising theme based on historical data and exploration-exploitation balance.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Request Fields:**
- `user_uuid` (UUID4, required) - UUID of the user requesting recommendation

**Response:**
```json
{
  "theme": "rock",
  "features": [0.75, 0.42, 0.88, 0.31, 0.65]
}
```

**Response Fields:**
- `theme` (string) - Recommended music theme/genre
- `features` (array of floats) - Feature vector used for prediction (for feedback loop)

**Status Codes:**
- `200 OK` - Prediction successful
- `401 Unauthorized` - Invalid or missing service JWT
- `422 Unprocessable Entity` - Invalid request body (e.g., malformed UUID)
- `500 Internal Server Error` - Prediction failed (model error, missing features, etc.)

**Example cURL:**
```bash
curl -X POST http://localhost:8004/api/v1/predict \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <service-jwt>" \
  -d '{"user_uuid": "123e4567-e89b-12d3-a456-426614174000"}'
```

---

### Update Model
```http
POST /api/v1/update
```

**Description:** Update the bandit model with user feedback (reward). This endpoint accepts the reward signal for a previously shown theme to improve future recommendations through reinforcement learning.

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "theme": "rock",
  "reward": 0.85,
  "features": [0.75, 0.42, 0.88, 0.31, 0.65]
}
```

**Request Fields:**
- `user_uuid` (UUID4, required) - UUID of the user
- `theme` (string, required) - Theme that was shown to the user
- `reward` (float, required) - Reward value between 0.0 and 1.0
  - `0.0` = User disliked/skipped immediately
  - `0.5` = User listened partially
  - `1.0` = User loved it (liked, added to playlist, listened fully)
- `features` (array of floats, required) - Feature vector from the prediction (must match what was returned)

**Response:**
```json
{
  "success": true
}
```

**Response Fields:**
- `success` (boolean) - Whether the update succeeded

**Status Codes:**
- `202 Accepted` - Update accepted and processed
- `401 Unauthorized` - Invalid or missing service JWT
- `422 Unprocessable Entity` - Invalid request body (e.g., reward out of range)
- `500 Internal Server Error` - Update failed (model error)

**Example cURL:**
```bash
curl -X POST http://localhost:8004/api/v1/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <service-jwt>" \
  -d '{
    "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
    "theme": "rock",
    "reward": 0.85,
    "features": [0.75, 0.42, 0.88, 0.31, 0.65]
  }'
```

---

## Algorithm: LinUCB (Linear Upper Confidence Bound)

### Overview
LinUCB is a contextual multi-armed bandit algorithm that:
- **Learns** user preferences from feedback (rewards)
- **Balances** exploration (trying new themes) vs exploitation (recommending known preferences)
- **Personalizes** recommendations based on user context (features)

### How It Works

1. **Feature Extraction:**
   - System extracts user features (listening history, preferences, demographics)
   - Features are normalized to a vector (typically 5-20 dimensions)

2. **Prediction:**
   - For each theme (arm), calculates upper confidence bound:
     ```
     UCB = θ^T * x + α * sqrt(x^T * A^-1 * x)
     ```
     - `θ` = learned weight vector for the theme
     - `x` = user feature vector
     - `A` = covariance matrix
     - `α` = exploration parameter (default: 1.0)
   - Returns theme with highest UCB

3. **Update:**
   - Receives reward (0.0-1.0) for recommended theme
   - Updates weight vector and covariance matrix:
     ```
     A = A + x * x^T
     θ = A^-1 * b
     b = b + reward * x
     ```

### Feature Vector

The feature vector typically includes:
- User listening history (recent themes, play counts)
- User preferences (likes, playlist adds)
- Temporal features (time of day, day of week)
- User demographics (optional)
- Session context (device, location)

**Dimension:** Usually 5-20 elements, normalized to [-1, 1]

---

## Reward Signal Guidelines

### Calculating Rewards

The reward should reflect user engagement and satisfaction:

| Reward | User Action | Example |
|--------|-------------|---------|
| `1.0` | Loved it | Listened fully + liked + added to playlist |
| `0.8-0.9` | Really liked | Listened fully + liked |
| `0.6-0.7` | Liked | Listened 75%+ OR liked |
| `0.4-0.5` | Neutral | Listened 50-75% |
| `0.2-0.3` | Disliked | Listened <50%, skipped |
| `0.0-0.1` | Strongly disliked | Skipped immediately (<10 seconds) |

### Reward Formula Example

```python
def calculate_reward(listen_duration, total_duration, liked, added_to_playlist):
    # Base reward from listening duration
    listen_ratio = min(listen_duration / total_duration, 1.0)
    base_reward = listen_ratio * 0.6

    # Bonus for explicit actions
    if liked:
        base_reward += 0.2
    if added_to_playlist:
        base_reward += 0.2

    return min(base_reward, 1.0)
```

---

## Integration Flow

### Typical User Session

1. **Get Recommendation:**
   ```
   Client → gateway_api → gateway_recommendation → bandit_system (predict)
   ```
   - User UUID passed through service JWTs
   - Returns theme + features

2. **User Listens:**
   - Play songs from recommended theme
   - Track user engagement (duration, likes, etc.)

3. **Send Feedback:**
   ```
   Client → gateway_api → gateway_recommendation → bandit_system (update)
   ```
   - Calculate reward based on engagement
   - Pass same features from prediction
   - Model learns and improves

### Example Implementation

```python
# 1. Get recommendation
response = requests.post(
    "http://localhost:8080/recommendation/recommend/theme",
    headers={"Authorization": f"Bearer {user_jwt}"},
    json={"user_uuid": user_uuid}
)
theme_data = response.json()

# 2. Play songs from theme
# ... user listens to songs ...

# 3. Calculate reward
reward = calculate_reward(
    listen_duration=120,
    total_duration=180,
    liked=True,
    added_to_playlist=False
)  # Returns 0.87

# 4. Send feedback
requests.post(
    "http://localhost:8080/recommendation/recommend/theme/feedback",
    headers={"Authorization": f"Bearer {user_jwt}"},
    json={
        "user_uuid": user_uuid,
        "theme": theme_data["theme"],
        "reward": reward,
        "features": theme_data["features"]
    }
)
```

---

## Model Persistence

### Storage
- Model parameters (A, b, θ) are persisted to disk
- Saved on every update (or batched for performance)
- Loaded on service startup

### File Structure
```
/data/
  /models/
    /user_{uuid}/
      theta.npy          # Weight vectors per theme
      A.npy             # Covariance matrices per theme
      b.npy             # Reward vectors per theme
      metadata.json     # Model metadata
```

---

## Performance Characteristics

### Prediction
- **Latency:** ~10-50ms
- **Throughput:** 100+ predictions/second
- **Complexity:** O(k * d^2) where k=themes, d=feature dimensions

### Update
- **Latency:** ~20-100ms
- **Throughput:** 50+ updates/second
- **Complexity:** O(d^3) due to matrix operations

### Memory
- **Per User:** ~10-50 KB (depends on feature dimensions and theme count)
- **Cold Start:** Falls back to popularity-based recommendations

---

## Error Handling

### Common Errors

**Prediction Failures:**
- Missing user features → Falls back to default feature vector
- Unknown user → Cold start with exploration-heavy strategy
- Model not trained → Uses popularity-based fallback

**Update Failures:**
- Invalid reward range → Returns 422 with validation error
- Feature dimension mismatch → Returns 422
- Model corruption → Reinitializes from backup

### Error Response Format

```json
{
  "detail": "Error message describing what went wrong"
}
```

---

## Monitoring & Metrics

### Key Metrics to Track

- **Prediction latency:** p50, p95, p99
- **Update latency:** p50, p95, p99
- **Exploration rate:** % of predictions with high uncertainty
- **Theme distribution:** Balance across themes
- **Average reward:** Overall model performance
- **Reward by theme:** Per-theme effectiveness

### Health Indicators

- Service responds to `/health` endpoint
- Prediction success rate > 99.5%
- Update success rate > 99.5%
- Average latency < 100ms

---

## Configuration

### Environment Variables

```bash
# Server Configuration
HOST=0.0.0.0
PORT=8004
LOG_LEVEL=INFO

# Model Parameters
ALPHA=1.0                    # Exploration parameter (higher = more exploration)
FEATURE_DIM=10              # Feature vector dimension
MODEL_SAVE_INTERVAL=100     # Updates between saves

# Storage
MODEL_DATA_PATH=/data/models
BACKUP_PATH=/data/backups
```

---

## API Versioning

Current version: `v1`
- Base path: `/api/v1`
- Versioned to allow future breaking changes
- Old versions supported for 6 months after deprecation

---

## Access Notes

**Via Gateway (Recommended):**
```
Client → http://localhost:8080/recommendation/recommend/theme
       → gateway_api (validates user JWT)
       → gateway_recommendation (generates service JWT)
       → bandit_system
```

**Direct Access (Internal Services Only):**
```
http://localhost:8004/api/v1/predict
```

**Authentication:**
- All requests must include valid service JWT
- JWT must contain user UUID in claims
- Generated by gateway_api from validated user tokens
