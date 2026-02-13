--------------- Artists -----------------
------ GET
-- name: GetArtist :one
SELECT * FROM artist
WHERE uuid = $1;

-- name: GetArtistsAlphabetically :many
SELECT * FROM artist
ORDER BY artist_name;

------ POST
-- name: UpdateArtistProfile :exec
UPDATE artist
SET artist_name = $2,
    bio = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE uuid = $1;

-- name: UpdateArtistPicture :exec
UPDATE artist
SET profile_image_path = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE uuid = $1;

------ PUT
-- name: CreateArtist :exec
    WITH new_artist as (
    INSERT INTO artist (artist_name, bio, profile_image_path)
    VALUES ($2, $3, $4)
    RETURNING uuid
)
INSERT INTO artist_member (artist_uuid, user_uuid, "role")
SELECT uuid, $1, 'owner'
FROM new_artist;

------ DELETE
-- deletes are overrated

--------------- ArtistMember -----------------
------ GET
-- name: GetUsersRepresentingArtist :many
SELECT pu.*, am.role, am.joined_at
FROM artist_member am
         JOIN public_user pu
              ON am.user_uuid = pu.uuid
WHERE am.artist_uuid = $1;

-- name: GetArtistForUser :many
SELECT art.*, am.role, am.joined_at
FROM artist_member am
         JOIN artist art
              ON am.artist_uuid = art.uuid
WHERE am.user_uuid = $1;

------ POST
------ PUT
------ DELETE

-- name: AddUserToArtist :exec

-- name: RemoveUserFromArtist :exec

-- name: ChangeUserRole :exec

-- name: Create :exec

