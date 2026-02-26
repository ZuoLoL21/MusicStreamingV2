--------------- Users -----------------
------ GET
-- name: GetPublicUser :one
SELECT * FROM public_user
WHERE uuid = $1 LIMIT 1;

-- name: GetHashPassword :one
SELECT hashed_password FROM users
WHERE uuid = $1;

-- name: GetUserByEmail :one
SELECT uuid, username, email, hashed_password, bio, profile_image_path, created_at, updated_at
FROM users WHERE email = $1 LIMIT 1;

-- name: SearchForUser :many
SELECT
    pu.*,
    similarity(pu.username, $1)::float AS similarity_score
FROM public_user pu
WHERE pu.username % $1
AND (
    $3 < 0
    OR (
        similarity(pu.username, $1) < $3
        OR (similarity(pu.username, $1) = $3 AND pu.created_at < $4)
    )
)
ORDER BY similarity(pu.username, $1) DESC, pu.created_at DESC
LIMIT $2;

------ POST
-- name: UpdatePassword :exec
UPDATE users
SET hashed_password = $2
WHERE uuid = $1;

-- name: UpdateProfile :exec
UPDATE users
SET username = $2,
    bio = $3
WHERE uuid = $1;

-- name: UpdateEmail :exec
UPDATE users
SET email = $2
WHERE uuid = $1;

-- name: UpdateImage :exec
UPDATE users
SET profile_image_path = $2
WHERE uuid = $1;

------ PUT
-- name: CreateUser :one
INSERT INTO users (username, email, hashed_password, bio, profile_image_path)
VALUES ($1, $2, $3, $4, $5)
RETURNING uuid;

------ DELETE
-- name: DeleteUser :exec
-- lol you thought, I get yo data fo-eva
