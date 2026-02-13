--------------- Users -----------------
------ GET
-- name: GetPublicUser :one
SELECT * FROM "PublicUser"
WHERE uuid = $1 LIMIT 1;

-- name: GetHashPassword :one
SELECT hashed_password FROM "User"
WHERE uuid = $1;

------ POST
-- name: UpdatePassword :exec
UPDATE "User"
SET hashed_password = $2
WHERE uuid = $1;

-- name: UpdateProfile :exec
UPDATE "User"
SET username = $2,
    bio = $3,
    updated = CURRENT_TIMESTAMP
WHERE uuid = $1;

-- name: UpdateEmail :exec
UPDATE "User"
SET email = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE uuid = $1;

-- name: UpdateImage :exec
UPDATE "User"
SET profile_image_path = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE uuid = $1;

------ PUT
-- name: CreateUser :exec
INSERT INTO "User" (username, email, hashed_password, bio, profile_image_path)
VALUES ($1, $2, $3, $4, $5)

------ DELETE
-- name: DeleteUser :exec
-- lol you thought

--------------- Artists -----------------
------ GET
-- name: GetArtist :one
SELECT * FROM "Artist"
WHERE uuid = $1;

-- name: GetArtistsAlphabetically :many
SELECT * FROM "Artist"
ORDER BY artist_name;

------ POST
-- name: UpdateArtistProfile: exec
UPDATE "Artist"
SET artist_name = $2,
    bio = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE uuid = $1

-- name: UpdateArtistPicture: exec
UPDATE "Artist"
SET profile_image_path = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE uuid = $1

------ PUT
-- name: CreateArtist
WITH new_artist as (
    INSERT INTO "Artist" (artist_name, bio, profile_image_path)
    VALUES ($2, $3, $4)
    RETURNING uuid
)
INSERT INTO "ArtistMember" (artist_uuid, user_uuid, "role")
SELECT uuid, $1, 'owner'
FROM new_artist;

------ DELETE
-- deletes are overrated

--------------- ArtistMember -----------------
------ GET
-- name: GetUsersRepresentingArtist :many
SELECT pu.*, am.role, am.joined_at
FROM "ArtistMember" am
         JOIN "PublicUser" pu
              ON am.user_uuid = pu.uuid
WHERE am.artist_uuid = $1;

-- name: GetArtistForUser :many
SELECT art.*, am.role, am.joined_at
FROM "ArtistMember" am
         JOIN "Artist" art
              ON am.artist_uuid = art.uuid
WHERE am.user_uuid = $1;

------ POST
------ PUT
------ DELETE

-- name: AddUserToArtist :exec

-- name: RemoveUserFromArtist :exec

-- name: ChangeUserRole

-- name: Create

