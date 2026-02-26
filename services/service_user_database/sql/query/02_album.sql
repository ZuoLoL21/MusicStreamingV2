--------------- Album -----------------
------ GET
-- name: GetAlbum :one
SELECT * FROM album
WHERE uuid = $1;

-- name: GetAlbumsForArtist :many
SELECT * FROM album
WHERE from_artist = $1
AND (
    $3::timestamptz IS NULL
    OR (
        created_at < $3
        OR (created_at = $3 AND uuid < $4)
    )
)
ORDER BY created_at DESC, uuid DESC
LIMIT $2;

-- name: SearchForAlbum :many
SELECT
    al.*,
    similarity(al.original_name, $1)::float AS similarity_score
FROM album al
WHERE al.original_name % $1
AND (
    $3 < 0
    OR (
        similarity(al.original_name, $1) < $3
        OR (similarity(al.original_name, $1) = $3 AND al.created_at < $4)
    )
)
ORDER BY similarity(al.original_name, $1) DESC, al.created_at DESC
LIMIT $2;

------ POST
-- name: UpdateAlbum :exec
UPDATE album
SET original_name = $2,
    description = $3
WHERE uuid = $1;

-- name: UpdateAlbumImage :exec
UPDATE album
SET image_path = $2
WHERE uuid = $1;

------ PUT
-- name: CreateAlbum :exec
INSERT INTO album (from_artist, original_name, description, image_path)
VALUES ($1, $2, $3, $4);

------ DELETE
-- name: DeleteAlbum :exec
DELETE FROM album
WHERE uuid = $1;
