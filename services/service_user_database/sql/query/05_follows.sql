--------------- Follows -----------------
------ GET
-- name: GetFollowersForUser :many
SELECT pu.*
FROM follows f
JOIN public_user pu
    ON f.from_user = pu.uuid
WHERE f.to_user = $1
AND (
    $3::timestamptz IS NULL
    OR (
        f.created_at < $3
        OR (f.created_at = $3 AND f.uuid < $4)
    )
)
ORDER BY f.created_at DESC, f.uuid DESC
LIMIT $2;

-- name: GetFollowersForArtist :many
SELECT pu.*
FROM follows f
JOIN public_user pu
    ON f.from_user = pu.uuid
WHERE f.to_artist = $1
AND (
    $3::timestamptz IS NULL
    OR (
        f.created_at < $3
        OR (f.created_at = $3 AND f.uuid < $4)
    )
)
ORDER BY f.created_at DESC, f.uuid DESC
LIMIT $2;

-- name: GetFollowedUsersForUser :many
SELECT pu.*
FROM follows f
JOIN public_user pu
    ON f.to_user = pu.uuid
WHERE f.from_user = $1
AND (
    $3::timestamptz IS NULL
    OR (
        f.created_at < $3
        OR (f.created_at = $3 AND f.uuid < $4)
    )
)
ORDER BY f.created_at DESC, f.uuid DESC
LIMIT $2;

-- name: GetFollowedArtistsForUser :many
SELECT a.*
FROM follows f
         JOIN artist a
              ON f.to_artist = a.uuid
WHERE f.from_user = $1
  AND (
    $3::timestamptz IS NULL
    OR (
        f.created_at < $3
        OR (f.created_at = $3 AND f.uuid < $4)
    )
    )
ORDER BY f.created_at DESC, f.uuid DESC
    LIMIT $2;

-- name: GetFollowerCountForUser :one
SELECT COUNT(*)
FROM follows
WHERE to_user = $1;

-- name: GetFollowingCountForArtist :one
SELECT COUNT(*)
FROM follows
WHERE to_artist = $1;

-- name: GetFollowCount :one
SELECT COUNT(*)
FROM follows
WHERE from_user = $1;

-- name: IsFollowingUser :one
SELECT EXISTS (
    SELECT 1
    FROM follows f
    WHERE f.from_user = $1
    AND f.to_user = $2
);

-- name: IsFollowingArtist :one
SELECT EXISTS (
    SELECT 1
    FROM follows f
    WHERE f.from_user = $1
    AND f.to_artist = $2
);

------ PUT
-- name: FollowUser :exec
INSERT INTO follows (from_user, to_user)
VALUES ($1, $2);

-- name: FollowArtist :exec
INSERT INTO follows (from_user, to_artist)
VALUES ($1, $2);

------ DELETE
-- name: UnfollowUser :exec
DELETE FROM follows
WHERE from_user = $1 AND to_user = $2;

-- name: UnfollowArtist :exec
DELETE FROM follows
WHERE from_user = $1 AND to_artist = $2;
