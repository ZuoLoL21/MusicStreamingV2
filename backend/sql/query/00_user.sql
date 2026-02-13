--------------- Users -----------------
------ GET
-- name: GetPublicUser :one
SELECT * FROM "PublicUser"
WHERE uuid = $1 LIMIT 1;

-- name: GetHashPassword :one
SELECT hashed_password FROM "User"
WHERE uuid = $1;

------ POST
-- name: UpdatePassword :exec
UPDATE "User"
SET hashed_password = $2
WHERE uuid = $1;

-- name: UpdateProfile :exec
UPDATE "User"
SET username = $2,
    bio = $3,
    updated = CURRENT_TIMESTAMP
WHERE uuid = $1;

-- name: UpdateEmail :exec
UPDATE "User"
SET email = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE uuid = $1;

-- name: UpdateImage :exec
UPDATE "User"
SET profile_image_path = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE uuid = $1;

------ PUT
-- name: CreateUser :exec
INSERT INTO "User" (username, email, hashed_password, bio, profile_image_path)
VALUES ($1, $2, $3, $4, $5)

------ DELETE
-- name: DeleteUser :exec
-- lol you thought, I get yo data fo-eva
