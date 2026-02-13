--------------- Follows -----------------
------ GET
-- name: GetFollowersForUser :many
SELECT pu.*
FROM follows f
JOIN public_user pu
    ON f.from_user = pu.uuid
WHERE f.to_user = $1;

-- name: GetFollowingUsersForUser :many
SELECT pu.*
FROM follows f
JOIN public_user pu
    ON f.to_user = pu.uuid
WHERE f.from_user = $1
    AND f.to_user IS NOT NULL;

-- name: GetFollowersForArtist :many
SELECT pu.*
FROM follows f
JOIN public_user pu
    ON f.from_user = pu.uuid
WHERE f.to_artist = $1;

-- name: GetFollowingArtistsForUser :many
SELECT art.*
FROM follows f
    JOIN artist art
        ON f.to_artist = art.uuid
WHERE f.from_user = $1
    AND f.to_artist IS NOT NULL;

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
