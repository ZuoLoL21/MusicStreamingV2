--------------- Likes -----------------
------ GET
-- name: GetLikesForMusic :many
SELECT pu.*
FROM likes l
JOIN public_user pu
    ON l.from_user = pu.uuid
WHERE l.to_music = $1;

-- name: GetLikesForUser :many
SELECT m.*
FROM likes l
JOIN music m
    ON l.to_music = m.uuid
WHERE l.from_user = $1;

-- name: IsLiked :one
SELECT EXISTS (
    SELECT 1
    FROM likes l
    WHERE l.from_user = $1
    AND l.to_music = $2
);

-- name: GetLikeCount :one
SELECT COUNT(*) FROM likes
WHERE to_music = $1;

------ PUT
-- name: LikeMusic :exec
INSERT INTO likes (from_user, to_music)
VALUES ($1, $2);

------ DELETE
-- name: UnlikeMusic :exec
DELETE FROM likes
WHERE from_user = $1 AND to_music = $2;
