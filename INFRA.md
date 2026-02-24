```
Frontend
->
API Gateway (Reverse Proxy)
- User Database
  - File storage
  - Database
- Recommendation System
  - Bandit
    - Database
  - Popularity Service
  - Clickhouse
```

API Gateway
- A simple reverse proxy that does basic verifications 
- Allows access to login/ register/
- Validate JWT with claim (refresh) if go to refresh/
- Validates JWT with claim (general)
- Assigns request IDs

User Database
- Holds login/ register/ refresh JWT logic
  - Returns JWT with claims (general or refresh) + key id
- Basic authentification
  - Only allows access to login/ register/ refresh without JWT or without valid JWT
  - Allows rest if JWT is valid
  - Checks role for specific routes
- Talks with file storage
  - Generates intra-service JWT with claim (service) + key id

Recommendation System
- Simple API gateway
- Quick validation of JWT (general)
- Emits new JWT (service) and pass the request along

All the other services (except for databases)
- Quick validation to ensure that JWT has correct claim (service)

