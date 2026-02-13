--------------- Album -----------------
------ GET
-- name: GetAlbum :one
SELECT * FROM album
WHERE uuid = $1;

-- name: GetAlbumsForArtist :many
SELECT * FROM album
WHERE from_artist = $1
ORDER BY created_at DESC;


------ POST
-- name: UpdateAlbum :exec
UPDATE album
SET original_name = $2,
    description = $3
WHERE uuid = $1;

------ PUT
-- name: CreateAlbum :exec
INSERT INTO album (from_artist, original_name, description)
VALUES ($1, $2, $3);

------ DELETE
-- name: DeleteAlbum :exec
DELETE FROM album
WHERE uuid = $1;
