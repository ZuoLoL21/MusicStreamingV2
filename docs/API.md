# MusicStreamingV2 - Unified API Documentation

**Last Updated:** 2026-03-16

This document provides complete API reference for all services in the MusicStreamingV2 platform. For architecture and design decisions, see [ARCHITECTURE.md](ARCHITECTURE.md).

---

## Table of Contents

1. [Quick Reference](#quick-reference)
2. [Authentication Overview](#authentication-overview)
3. [Public Layer: gateway_api](api/API_gateway_api.md#public-layer-gateway_api)
4. [Data Backend: service_user_database](api/API_service_user_database.md#data-backend-service_user_database)
5. [Recommendation Backend: gateway_recommendation](api/API_gateway_recommendation.md#recommendation-backend-gateway_recommendation)
6. [Popularity Backend: service_popularity_system](api/API_service_popularity_system.md#popularity-backend-service_popularity_system)
7. [Bandit ML Backend: service_bandit_system](api/API_service_bandit_system.md#bandit-ml-backend-service_bandit_system)
8. [Common Patterns](api/API_common.md#common-patterns)
9. [Error Responses](api/API_common.md#error-responses)

---

## Quick Reference

### Service Access Patterns

| Service | Public Access | Client Access Pattern |
|---------|---------------|----------------------|
| gateway_api | **Yes** | Direct access with Normal/Refresh JWT at `http://localhost:8080` |
| service_user_database | Internal only | Via gateway_api at `/users/*`, `/artists/*`, `/albums/*`, `/music/*`, `/tags/*`, `/playlists/*`, `/history/*`, `/search/*` |
| gateway_recommendation | Internal only | Via gateway_api at `/recommend/*` and `/popular/*` |
| service_popularity_system | Internal only | Via gateway_recommendation (automatic routing) |
| service_bandit_system | Internal only | Via gateway_recommendation (automatic routing) |

### JWT Token Types

| Type | Subject | Lifetime | Use Case | Header Format |
|------|---------|----------|----------|---------------|
| Normal JWT | `"normal"` | ~10 min | User access token | `Authorization: Bearer <token>` |
| Refresh JWT | `"refresh"` | ~10 days | Token renewal | `Authorization: Bearer <token>` |
| Service JWT | `"service"` | ~2 min | Inter-service auth | `Authorization: Bearer <token>` |

**Client developers:** Only use Normal and Refresh JWTs. Service JWTs are generated automatically by gateways.

---

## Authentication Overview

### Login Flow
```
1. POST /login with credentials
   → gateway_api validates and forwards to service_user_database
   → Returns Normal JWT + Refresh JWT + User UUID
2. Store tokens securely (httpOnly cookies recommended)
3. Use Normal JWT for all protected requests
```

### Protected Request Flow
```
1. Client sends request with Normal JWT to gateway_api
2. Gateway validates JWT and generates Service JWT
3. Gateway forwards to backend with Service JWT
4. Backend validates Service JWT and processes request
5. Response returns through gateway to client
```

### Token Renewal Flow
```
1. POST /renew with Refresh JWT
   → gateway_api validates with service_user_database
   → Returns new Normal JWT
2. Update stored Normal JWT
```

---
