#!/bin/sh
set -e

mc alias set myminio http://database-minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"

mc mb myminio/music-streaming --ignore-existing;

mc anonymous set download myminio/music-streaming;

mc cp /defaults/default_profile.jpeg myminio/music-streaming/pictures-profile/default.jpg --quiet || true;
mc cp /defaults/default_profile.jpeg myminio/music-streaming/pictures-artist/default.jpg --quiet || true;
mc cp /defaults/default_music.jpeg myminio/music-streaming/pictures-music/default.jpg --quiet || true;
mc cp /defaults/default_music.jpeg myminio/music-streaming/pictures-album/default.jpg --quiet || true;
mc cp /defaults/default_music.jpeg myminio/music-streaming/pictures-playlist/default.jpg --quiet || true;

echo 'MinIO initialized with default images';
exit 0;