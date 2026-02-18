CREATE FUNCTION is_user_allowed_playlist_edit(user_uuid UUID, album_uuid UUID)
    RETURNS boolean
AS $$
SELECT EXISTS (
    SELECT * FROM playlist
    WHERE uuid = album_uuid
      AND from_user = user_uuid
)
           $$ language sql;