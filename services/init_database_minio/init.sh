#!/bin/sh
set -e

mc alias set myminio http://database-minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"

mc mb myminio/music-streaming --ignore-existing;

mc anonymous set download myminio/music-streaming;

mc cp /defaults/default_profile.jpeg myminio/music-streaming/defaults/profile.jpg --quiet || true;
mc cp /defaults/default_profile.jpeg myminio/music-streaming/defaults/artist.jpg --quiet || true;
mc cp /defaults/default_music.jpeg myminio/music-streaming/defaults/music.jpg --quiet || true;
mc cp /defaults/default_music.jpeg myminio/music-streaming/defaults/album.jpg --quiet || true;
mc cp /defaults/default_music.jpeg myminio/music-streaming/defaults/playlist.jpg --quiet || true;

echo 'MinIO initialized with default images';
exit 0;