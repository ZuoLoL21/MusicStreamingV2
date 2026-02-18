--------------- Playlist -----------------
------ GET
-- name: GetPlaylist :one
SELECT * FROM playlist
WHERE uuid = $1;

-- name: GetPlaylistsForUser :many
SELECT * FROM playlist
WHERE from_user = $1
AND (
    $3::timestamptz IS NULL
    OR (
        updated_at < $3
        OR (updated_at = $3 AND uuid < $4)
    )
)
ORDER BY updated_at DESC, uuid DESC
LIMIT $2;

-- name: GetPlaylistTracks :many
SELECT m.*
FROM playlist_track pt
JOIN music m
    ON pt.music_uuid = m.uuid
WHERE pt.playlist_uuid = $1
AND (
    $3::int IS NULL
    OR pt.position > $3
    )
ORDER BY pt.position
LIMIT $2;

------ POST
-- name: UpdatePlaylist :exec
UPDATE playlist
SET original_name = $2,
    description = $3,
    is_public = $4
WHERE uuid = $1;

-- name: UpdateTrackPosition :exec
UPDATE playlist_track
SET position = $2
WHERE uuid = $1;

-- name: UpdatePlaylistImage :exec
UPDATE playlist
SET image_path = $2
WHERE uuid = $1;

------ PUT
-- name: CreatePlaylist :exec
INSERT INTO playlist (from_user, original_name, description, is_public, image_path)
VALUES ($1, $2, $3, $4, $5);

-- name: AddTrackToPlaylist :exec
INSERT INTO playlist_track (music_uuid, position, playlist_uuid)
VALUES ($1, $2, $3);

------ DELETE
-- name: RemoveTrackFromPlaylist :exec
DELETE FROM playlist_track
WHERE music_uuid = $1 AND playlist_uuid = $2;

-- name: DeletePlaylist :exec
DELETE FROM playlist
WHERE uuid = $1;
