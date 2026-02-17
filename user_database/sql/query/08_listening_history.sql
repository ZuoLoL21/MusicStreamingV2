--------------- ListeningHistory -----------------
------ GET
-- name: GetListeningHistoryForUser :many
SELECT * FROM listening_history
WHERE user_uuid = $1
ORDER BY played_at DESC;

-- name: GetRecentlyPlayedForUser :many
SELECT * FROM listening_history
WHERE user_uuid = $1
ORDER BY played_at DESC
LIMIT $2;

-- name: GetTopMusicForUser :many
SELECT music_uuid, COUNT(*) as play_count
FROM listening_history
WHERE user_uuid = $1
GROUP BY music_uuid
ORDER BY COUNT(*) DESC
LIMIT $2;

------ PUT
-- name: AddListeningHistoryEntry :exec
INSERT INTO listening_history (user_uuid, music_uuid, listen_duration_seconds, completion_percentage)
VALUES ($1, $2, $3, $4);
