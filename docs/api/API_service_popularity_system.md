## Popularity Backend: service_popularity_system

**Internal service** - accessed via `gateway_recommendation` → `gateway_api` at `http://localhost:8080/popular/*`
**Authentication:** Service JWT (auto-generated)

**Note:** This service is accessed through gateway_recommendation. All endpoints listed above in the [Recommendation Backend](API_gateway_recommendation.md#recommendation-backend-gateway_recommendation) section are implemented by this service.

### Implementation Details

**Popularity Score Calculation:**

Songs (All-Time):
- 70% total play count
- 30% like count
- Normalized to 0.0-1.0

Songs (Timeframe):
- 60% plays in period
- 20% likes
- 20% velocity (rate of change)

Artists (All-Time):
- 50% follower count
- 40% total plays
- 10% music count

Artists (Timeframe):
- 40% new followers in period
- 40% plays in period
- 20% velocity

Themes:
- 50% total plays
- 30% song count
- 20% unique listeners

**Trend Direction:**
- `"up"` - velocity > 0.1
- `"down"` - velocity < -0.1
- `"stable"` - -0.1 ≤ velocity ≤ 0.1

**Data Source:** ClickHouse data warehouse with decay-weighted metrics

---

