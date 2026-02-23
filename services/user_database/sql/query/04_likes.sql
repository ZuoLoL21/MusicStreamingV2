--------------- Likes -----------------
------ GET
-- name: GetLikeCountMusic :one
SELECT COUNT(*) FROM likes
WHERE to_music = $1;

-- name: GetLikeCountUser :one
SELECT COUNT(*) FROM likes
WHERE from_user = $1;

-- name: GetLikesForUser :many
SELECT m.*
FROM likes l
JOIN music m
    ON l.to_music = m.uuid
WHERE l.from_user = $1
AND (
    $3::timestamptz IS NULL
    OR (
         l.created_at < $3
         OR (l.created_at = $3 AND l.uuid < $4)
       )
    )
ORDER BY l.created_at DESC, l.uuid DESC
LIMIT $2;

-- name: IsLiked :one
SELECT EXISTS (
    SELECT 1
    FROM likes l
    WHERE l.from_user = $1
    AND l.to_music = $2
);


------ PUT
-- name: LikeMusic :exec
INSERT INTO likes (from_user, to_music)
VALUES ($1, $2);

------ DELETE
-- name: UnlikeMusic :exec
DELETE FROM likes
WHERE from_user = $1 AND to_music = $2;
