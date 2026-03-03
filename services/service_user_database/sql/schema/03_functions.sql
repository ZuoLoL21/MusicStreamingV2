CREATE FUNCTION is_user_allowed_playlist_edit(user_uuid UUID, playlist_uuid UUID)
    RETURNS boolean
AS $$
SELECT EXISTS (
    SELECT * FROM playlist
    WHERE uuid = playlist_uuid
      AND from_user = user_uuid
)
           $$ language sql;