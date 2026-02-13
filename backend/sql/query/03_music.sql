--------------- Music -----------------
------ GET
-- name: GetMusic :one
SELECT * FROM music
WHERE uuid = $1;

-- name: GetMusicForArtist :many
SELECT * FROM music
WHERE from_artist = $1
ORDER BY created_at DESC;

-- name: GetMusicForAlbum :many
SELECT * FROM music
WHERE in_album = $1
ORDER BY created_at DESC;

-- name: GetMusicForUser :many
SELECT * FROM music
WHERE uploaded_by = $1
ORDER BY create_at DESC;

-- TODO: name: SearchForMusic :many

------ POST
-- name: UpdateMusicDetails :exec
UPDATE music
SET song_name = $2,
    in_album = $3
WHERE uuid = $1;

-- name: IncrementPlayCount :exec
UPDATE music
SET play_count = play_count + 1
WHERE uuid = $1;

-- name: UpdateMusicStorage :exec
UPDATE music
SET path_in_file_storage = $2,
    duration_seconds = $3
WHERE uuid = $1;

------ PUT
-- name: CreateMusic :exec
INSERT INTO music (from_artist, uploaded_by, in_album, song_name, path_in_file_storage, duration_seconds)
VALUES ($1, $2, $3, $4, $5, $6);

------ DELETE
-- name: DeleteMusic :exec
DELETE FROM music
WHERE uuid = $1;

-- name: ResetPlayCount :exec
UPDATE music
SET play_count = 0
WHERE uuid = $1;
