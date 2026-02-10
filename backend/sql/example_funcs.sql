-- name: GetUser :one
SELECT * FROM "User"
WHERE UUID = $1 LIMIT 1;

-- name: GetArtists :many
SELECT * FROM "Artist"
ORDER BY artist_name

-- name: GetAlbumsByNew :many
SELECT * FROM "Album"
ORDER BY created_at DESC;

-- name: UpdatePassword :exec
UPDATE "User"
set hashed_password = $2
WHERE UUID = $1;


