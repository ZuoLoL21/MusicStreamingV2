CREATE FUNCTION is_user_allowed_playlist_edit(user_uuid UUID, playlist_uuid UUID)
    RETURNS boolean
AS $$
SELECT EXISTS (
    SELECT * FROM playlist
    WHERE uuid = playlist_uuid
      AND from_user = user_uuid
)
           $$ language sql;

CREATE FUNCTION is_user_allowed_playlist_view(user_uuid UUID, playlist_uuid UUID)
    RETURNS boolean
AS $$
SELECT EXISTS (
    SELECT * FROM playlist
    WHERE uuid = playlist_uuid
      AND (is_public = TRUE OR from_user = user_uuid)
)
           $$ language sql;

CREATE FUNCTION get_max_playlist_size(playlist_uuid UUID)
    RETURNS integer
AS $$
    SELECT MAX(position) as max_position
    FROM playlist_track
    WHERE playlist_uuid = $1;
$$ language sql;