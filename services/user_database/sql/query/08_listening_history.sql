--------------- ListeningHistory -----------------
------ GET
-- name: GetListeningHistoryForUser :many
SELECT *
FROM listening_history
WHERE user_uuid = $1
AND (
    $3::timestamptz IS NULL
    OR (
         played_at < $3
         OR (played_at = $3 AND uuid < $4)
       )
)
ORDER BY played_at DESC, uuid DESC
LIMIT $2;

-- name: GetTopMusicForUser :many
SELECT music_uuid, COUNT(*) as play_count
FROM listening_history
WHERE user_uuid = $1
GROUP BY music_uuid
HAVING (
    $3::uuid IS NULL
    OR (
        COUNT(*) < $4
        OR (COUNT(*) = $4 AND music_uuid < $3)
    )
)
ORDER BY play_count DESC, music_uuid DESC
LIMIT $2;

------ PUT
-- name: AddListeningHistoryEntry :exec
INSERT INTO listening_history (user_uuid, music_uuid, listen_duration_seconds, completion_percentage)
VALUES ($1, $2, $3, $4);
