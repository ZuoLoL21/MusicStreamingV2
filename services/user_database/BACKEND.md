## Simple Backend

### Authentification: 
- Hash passwords using Argon2
- Store hashed passwords
- Add OAuth 2.0

- Use JWT tokens (get auth)
- Two tokens: 
  - access token - short lived (sent with every request)
  - refresh token - long-lived (send to get a new access token)
    - TODO: make them expire after one use - rotating refresh tokens
### Format
- Format Queries like the following
```sql
--------------- Artists -----------------
------ GET
------ POST
------ PUT
------ DELETE
```

### Generate SQLC
Bash
```bash
docker run --rm -v $(pwd):/src -w /src sqlc/sqlc generate
```

PS
```bash
docker run --rm -v ${PWD}:/src -w /src sqlc/sqlc generate
```