## Simple Backend

### Authentification: 
- Hash passwords using Argon2
- Store hashed passwords
- Add OAuth 2.0


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