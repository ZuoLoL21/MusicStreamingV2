-- name: GetPublicUser :one
SELECT * FROM "PublicUser"
WHERE uuid = $1 LIMIT 1;

-- name: GetHashPassword :one
SELECT hashed_password FROM "User"
WHERE uuid = $1

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
SET email = $2
WHERE uuid = $1

-- name: UpdateImage :exec
UPDATE "User"
SET profile_image_path = $2
WHERE uuid = %1


-- name: GetArtist :one
SELECT * FROM "Artist"
WHERE uuid = $1

-- name: GetArtistsAlphabetically :many
SELECT * FROM "Artist"
ORDER BY artist_name;

-- name: GetAlbumsByNew :many
SELECT * FROM "Album"
ORDER BY created_at DESC;




