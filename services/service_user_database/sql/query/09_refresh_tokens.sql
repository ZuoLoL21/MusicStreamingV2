-- name: UpsertRefreshToken :one
INSERT INTO refresh_tokens (user_uuid, device_id, token_hash, device_name, expires_at, created_at, last_used_at)
VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (user_uuid, device_id)
DO UPDATE SET
    token_hash = EXCLUDED.token_hash,
    device_name = EXCLUDED.device_name,
    expires_at = EXCLUDED.expires_at,
    last_used_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens
WHERE token_hash = $1 AND expires_at > CURRENT_TIMESTAMP;

-- name: ValidateAndUpdateRefreshToken :one
UPDATE refresh_tokens
SET last_used_at = CURRENT_TIMESTAMP
WHERE token_hash = $1 AND expires_at > CURRENT_TIMESTAMP
RETURNING *;

-- name: GetDevicesForUser :many
SELECT uuid, device_id, device_name, created_at, last_used_at, expires_at
FROM refresh_tokens
WHERE user_uuid = $1 AND expires_at > CURRENT_TIMESTAMP
ORDER BY last_used_at DESC;

-- name: RevokeDevice :exec
DELETE FROM refresh_tokens
WHERE user_uuid = $1 AND device_id = $2;

-- name: RevokeAllDevicesForUser :exec
DELETE FROM refresh_tokens
WHERE user_uuid = $1;

-- name: DeleteExpiredTokens :exec
DELETE FROM refresh_tokens
WHERE expires_at <= CURRENT_TIMESTAMP;
