--------------- Playlist -----------------
------ GET
-- name: GetPlaylist :one
SELECT * FROM playlist
WHERE uuid = $1;

-- name: IsPlaylistPublicOrOwnedByUser :one
SELECT is_user_allowed_playlist_view($2, $1) AS is_accessible;

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

-- name: GetPlaylists :many
SELECT * FROM playlist
WHERE is_user_allowed_playlist_view($1, uuid)
AND (
    $3::timestamptz IS NULL
    OR (
        updated_at < $3
        OR (updated_at = $3 AND uuid < $4)
    )
)
ORDER BY updated_at DESC, uuid DESC
    LIMIT $2;

-- name: SearchForPlaylist :many
SELECT
    p.*,
    similarity(p.original_name, $1)::float AS similarity_score
FROM playlist p
WHERE p.original_name % $1
AND is_user_allowed_playlist_view($3, p.uuid)
AND (
    $4 < 0
    OR (
        similarity(p.original_name, $1) < $4
        OR (similarity(p.original_name, $1) = $4 AND p.created_at < $5)
    )
)
ORDER BY similarity(p.original_name, $1) DESC, p.created_at DESC
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
SET original_name = $3,
    description = $4,
    is_public = $5
WHERE uuid = $2
AND is_user_allowed_playlist_edit($1, $2);

-- TODO: Current implementation just doesn't work
    -- Requires validating position (ensure positive < max position)
    -- Requires shifting other 
-- name: UpdateTrackPosition :exec
UPDATE playlist_track
SET position = $4
WHERE uuid = $3
AND is_user_allowed_playlist_edit($1, $2);

-- name: UpdatePlaylistImage :exec
UPDATE playlist
SET image_path = $3
WHERE uuid = $2
AND is_user_allowed_playlist_edit($1, $2);

------ PUT
-- name: CreatePlaylist :exec
INSERT INTO playlist (from_user, original_name, description, is_public, image_path)
VALUES ($1, $2, $3, $4, $5);

-- name: AddTrackToPlaylist :exec
INSERT INTO playlist_track (music_uuid, position, playlist_uuid)
VALUES (
    $1,
    COALESCE((SELECT get_max_playlist_size($1)+1),0),
    $2
);

------ DELETE
-- name: RemoveTrackFromPlaylist :exec
DELETE FROM playlist_track
WHERE music_uuid = $3 AND playlist_uuid = $2
AND is_user_allowed_playlist_edit($1, $2);

-- name: DeletePlaylist :exec
DELETE FROM playlist
WHERE uuid = $2
AND is_user_allowed_playlist_edit($1, $2);
