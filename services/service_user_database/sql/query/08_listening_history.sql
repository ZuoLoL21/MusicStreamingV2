--------------- ListeningHistory -----------------
------ GET
-- name: GetListeningHistoryForUser :many
SELECT
    lh.uuid,
    lh.user_uuid,
    lh.music_uuid,
    m.song_name,
    a.artist_name,
    m.from_artist as artist_uuid,
    lh.played_at as listened_at,
    lh.listen_duration_seconds,
    lh.completion_percentage
FROM listening_history lh
INNER JOIN music m ON lh.music_uuid = m.uuid
INNER JOIN artist a ON m.from_artist = a.uuid
WHERE lh.user_uuid = $1
AND (
    $3::timestamptz IS NULL
    OR (
         lh.played_at < $3
         OR (lh.played_at = $3 AND lh.uuid < $4)
       )
)
ORDER BY lh.played_at DESC, lh.uuid DESC
LIMIT $2;

-- name: GetTopMusicForUser :many
SELECT
    lh.music_uuid,
    m.song_name,
    a.artist_name,
    m.from_artist as artist_uuid,
    COUNT(*) as play_count
FROM listening_history lh
INNER JOIN music m ON lh.music_uuid = m.uuid
INNER JOIN artist a ON m.from_artist = a.uuid
WHERE lh.user_uuid = $1
GROUP BY lh.music_uuid, m.song_name, a.artist_name, m.from_artist
HAVING (
    $3::uuid IS NULL
    OR (
        COUNT(*) < $4
        OR (COUNT(*) = $4 AND lh.music_uuid < $3)
    )
)
ORDER BY play_count DESC, lh.music_uuid DESC
LIMIT $2;

------ PUT
-- name: AddListeningHistoryEntry :exec
INSERT INTO listening_history (user_uuid, music_uuid, listen_duration_seconds, completion_percentage)
VALUES ($1, $2, $3, $4);
