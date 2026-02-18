--------------- Tags -----------------
------ GET
-- name: GetAllTags :many
SELECT * FROM music_tags
WHERE (
    $2::text IS NULL
    OR tag_name > $2
)
ORDER BY tag_name
LIMIT $1;

-- name: GetTag :one
SELECT * FROM music_tags
WHERE tag_name = $1;

-- name: GetTagsForMusic :many
SELECT mt.*
FROM tag_assignment ta
JOIN music_tags mt
    ON ta.tag_name = mt.tag_name
WHERE ta.music_uuid = $1
AND (
    $3::text IS NULL
    OR mt.tag_name > $3
)
ORDER BY mt.tag_name
LIMIT $2;

-- name: GetMusicForTag :many
SELECT m.*
FROM tag_assignment ta
JOIN music m
    ON ta.music_uuid = m.uuid
WHERE ta.tag_name = $1
AND (
    $3::timestamptz IS NULL
    OR (
         m.created_at < $3
         OR (m.created_at = $3 AND m.uuid < $4)
       )
    )
ORDER BY m.created_at DESC, m.uuid DESC
LIMIT $2;

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
