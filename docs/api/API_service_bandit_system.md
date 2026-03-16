## Bandit ML Backend: service_bandit_system

**Internal service** - accessed via `gateway_recommendation` → `gateway_api` at `http://localhost:8080/recommend/theme`
**Authentication:** Service JWT (auto-generated)
**Framework:** FastAPI (Python)

### Health Check

```http
GET /api/v1/health
```

**Authentication:** None

**Response:**
```json
{
  "status": "healthy",
  "service": "bandit-system"
}
```

**Status Codes:**
- `200 OK` - Service healthy

---

### Predict Theme

```http
POST /api/v1/predict
```

**Description:** Get personalized theme recommendation using LinUCB algorithm.

**Note:** This endpoint is accessed via gateway_recommendation at `POST /recommend/theme`.

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
  "features": [0.75, 0.42, 0.88, 0.31, 0.65, 0.53, 0.91, 0.22, 0.68, 0.44, 0.77, 0.59]
}
```

**Response Fields:**
- `theme` (string) - Recommended music theme/genre
- `features` (array of 12 floats) - Feature vector for feedback loop

**Status Codes:**
- `200 OK` - Prediction successful
- `401 Unauthorized` - Invalid service JWT
- `422 Unprocessable Entity` - Invalid request body
- `500 Internal Server Error` - Model error

---

### Update Model

```http
POST /api/v1/update
```

**Description:** Update the bandit model with user feedback (reward).

**Authentication:** Service JWT required

**Request Body:**
```json
{
  "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
  "theme": "rock",
  "reward": 0.85,
  "features": [0.75, 0.42, 0.88, 0.31, 0.65, 0.53, 0.91, 0.22, 0.68, 0.44, 0.77, 0.59]
}
```

**Request Fields:**
- `user_uuid` (UUID4, required) - User UUID
- `theme` (string, required) - Theme that was shown
- `reward` (float, required) - Reward value 0.0-1.0
- `features` (array, required) - Feature vector from prediction (must be exactly 12 floats)

**Reward Guidelines:**

| Reward | User Action |
|--------|-------------|
| `1.0` | Listened fully + liked + added to playlist |
| `0.8-0.9` | Listened fully + liked |
| `0.6-0.7` | Listened 75%+ OR liked |
| `0.4-0.5` | Listened 50-75% (neutral) |
| `0.2-0.3` | Listened <50% (disliked) |
| `0.0-0.1` | Skipped immediately |

**Response:**
```json
{
  "success": true
}
```

**Status Codes:**
- `202 Accepted` - Update accepted
- `400 Bad Request` - Invalid features length (must be 12 elements)
- `401 Unauthorized` - Invalid service JWT
- `422 Unprocessable Entity` - Invalid reward value
- `500 Internal Server Error` - Update failed

---

### Algorithm: LinUCB

**Overview:**
LinUCB (Linear Upper Confidence Bound) is a contextual multi-armed bandit algorithm that balances exploration (trying new themes) and exploitation (recommending known preferences).

**How It Works:**

1. **Feature Extraction** - User context is converted to a feature vector (5-20 dimensions)
2. **Prediction** - Calculate upper confidence bound for each theme:
   ```
   UCB = θ^T * x + α * sqrt(x^T * A^-1 * x)
   ```
   - `θ` = learned weight vector
   - `x` = user feature vector
   - `A` = covariance matrix
   - `α` = exploration parameter (default: 1.0)
3. **Update** - After receiving reward, update model parameters:
   ```
   A = A + x * x^T
   b = b + reward * x
   θ = A^-1 * b
   ```

**Feature Vector:**
Typically includes:
- User listening history
- Recent themes and play counts
- Time of day, day of week
- User preferences (likes, playlist adds)
- Session context

**Model Persistence:**
Models are stored per-user in PostgreSQL and loaded on service startup.

**Cold Start:**
Falls back to popularity-based recommendations for new users.

---

