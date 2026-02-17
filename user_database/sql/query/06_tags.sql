--------------- Tags -----------------
------ GET
-- name: GetAllTags :many
SELECT * FROM music_tags
ORDER BY tag_name;

-- name: GetTag :one
SELECT * FROM music_tags
WHERE tag_name = $1;

-- name: GetTagsForMusic :many
SELECT mt.*
FROM tag_assignment ta
JOIN music_tags mt
    ON ta.tag_name = mt.tag_name
WHERE ta.music_uuid = $1;

-- name: GetMusicForTag :many
SELECT m.*
FROM tag_assignment ta
JOIN music m
    ON ta.music_uuid = m.uuid
WHERE ta.tag_name = $1;

------ PUT
-- name: CreateTag :exec
INSERT INTO music_tags (tag_name, tag_description)
VALUES ($1, $2);

-- name: AssignTagToMusic :exec
INSERT INTO tag_assignment (music_uuid, tag_name)
VALUES ($1, $2);

------ DELETE
-- name: RemoveTagFromMusic :exec
DELETE FROM tag_assignment
WHERE music_uuid = $1 AND tag_name = $2;
