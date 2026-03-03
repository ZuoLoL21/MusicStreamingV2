#!/usr/bin/env python3
"""
Music Data Ingestion Script

Reads music-data.yaml config and populates PostgreSQL + MinIO with music content.
Designed to run as a Docker init container.
"""

import os
import sys
from pathlib import Path
from uuid import uuid4

from config import load_config, validate_config, get_mp3_duration, ConfigValidationError
from db import (
    connect_postgres,
    upsert_artist,
    upsert_album,
    upsert_music,
    upsert_tag,
    assign_tag_to_music,
    check_data_loaded,
    get_stats,
)
from storage import (
    connect_minio,
    ensure_bucket_exists,
    upload_audio,
    upload_artist_image,
    upload_album_image,
    upload_music_image,
)


# Configuration paths
SEED_MUSIC_DIR = Path('/seed_music')
CONFIG_FILE = SEED_MUSIC_DIR / 'music-data.yaml'


def log(message, level='INFO'):
    """Simple logging function"""
    print(f"[{level}] {message}")


def main():
    """Main ingestion logic"""

    log("=" * 60)
    log("Music Data Ingestion Starting")
    log("=" * 60)

    # Step 1: Load and validate config
    log("Step 1: Loading configuration...")

    if not CONFIG_FILE.exists():
        log(f"Config file not found: {CONFIG_FILE}", level='ERROR')
        log("Please ensure music-data.yaml is placed in the seed_music directory", level='ERROR')
        sys.exit(1)

    try:
        config = load_config(str(CONFIG_FILE))
        log(f"✓ Config loaded: {CONFIG_FILE}")
    except Exception as e:
        log(f"Failed to load config: {e}", level='ERROR')
        sys.exit(1)

    # Step 2: Validate config
    log("Step 2: Validating configuration...")

    try:
        validate_config(config, SEED_MUSIC_DIR)
        log("✓ Config validation passed")
    except ConfigValidationError as e:
        log(f"Config validation failed:\n{e}", level='ERROR')
        sys.exit(1)

    # Summary
    num_artists = len(config.get('artists', []))
    num_albums = len(config.get('albums', []))
    num_tracks = len(config.get('music', []))
    num_tags = len(config.get('tags', []))

    log(f"  Artists: {num_artists}")
    log(f"  Albums:  {num_albums}")
    log(f"  Tracks:  {num_tracks}")
    log(f"  Tags:    {num_tags}")

    # Step 3: Connect to services
    log("Step 3: Connecting to services...")

    try:
        db_conn = connect_postgres()
        log("✓ Connected to PostgreSQL")
    except Exception as e:
        log(f"Failed to connect to PostgreSQL: {e}", level='ERROR')
        sys.exit(1)

    try:
        s3_client = connect_minio()
        bucket_name = os.environ.get('MINIO_BUCKET_NAME', 'music-streaming')
        ensure_bucket_exists(s3_client, bucket_name)
        log(f"✓ Connected to MinIO (bucket: {bucket_name})")
    except Exception as e:
        log(f"Failed to connect to MinIO: {e}", level='ERROR')
        db_conn.close()
        sys.exit(1)

    # Step 4: Check idempotency (optional - we use UPSERT, so re-runs are safe)
    uploaded_by = config['uploaded_by']
    log(f"Step 4: Checking existing data (uploaded_by: {uploaded_by})...")

    try:
        stats_before = get_stats(db_conn)
        log(f"  Current DB state: {stats_before['artists']} artists, "
            f"{stats_before['albums']} albums, {stats_before['music']} tracks")
    except Exception as e:
        log(f"Warning: Could not get stats: {e}", level='WARN')

    # Step 5: Import data (wrapped in transaction)
    log("Step 5: Starting data import...")

    try:
        # Create artists
        log(f"  Creating {num_artists} artists...")
        artist_uuid_map = {}  # local_id -> UUID

        for idx, artist in enumerate(config.get('artists', []), 1):
            artist_id = str(uuid4())
            artist_name = artist['name']
            bio = artist.get('bio', '').strip() or None
            image_path = None

            # Upload artist image if specified
            if 'image' in artist and artist['image']:
                img_file = SEED_MUSIC_DIR / artist['image']
                try:
                    s3_path = upload_artist_image(s3_client, bucket_name, artist_id, str(img_file))
                    image_path = s3_path
                    log(f"    [{idx}/{num_artists}] Uploaded image for {artist_name}")
                except Exception as e:
                    log(f"    Warning: Failed to upload image for {artist_name}: {e}", level='WARN')

            # Upsert artist
            result_uuid = upsert_artist(db_conn, artist_id, artist_name, bio, image_path)
            artist_uuid_map[artist['id']] = result_uuid
            log(f"    [{idx}/{num_artists}] ✓ {artist_name} (UUID: {result_uuid})")

        log(f"  ✓ Artists created/updated")

        # Create albums
        log(f"  Creating {num_albums} albums...")
        album_uuid_map = {}  # local_id -> UUID

        for idx, album in enumerate(config.get('albums', []), 1):
            album_id = str(uuid4())
            album_name = album['name']
            artist_uuid = artist_uuid_map[album['artist']]
            description = album.get('description', '').strip() or None
            image_path = None

            # Upload album image if specified
            if 'image' in album and album['image']:
                img_file = SEED_MUSIC_DIR / album['image']
                try:
                    s3_path = upload_album_image(s3_client, bucket_name, album_id, str(img_file))
                    image_path = s3_path
                    log(f"    [{idx}/{num_albums}] Uploaded image for {album_name}")
                except Exception as e:
                    log(f"    Warning: Failed to upload image for {album_name}: {e}", level='WARN')

            # Upsert album
            result_uuid = upsert_album(db_conn, album_id, artist_uuid, album_name, description, image_path)
            album_uuid_map[album['id']] = result_uuid
            log(f"    [{idx}/{num_albums}] ✓ {album_name} (UUID: {result_uuid})")

        log(f"  ✓ Albums created/updated")

        # Create tags
        log(f"  Creating {num_tags} tags...")

        for tag in config.get('tags', []):
            tag_name = tag['name']
            upsert_tag(db_conn, tag_name)

        log(f"  ✓ Tags created")

        # Import music tracks
        log(f"  Importing {num_tracks} music tracks...")

        for idx, track in enumerate(config.get('music', []), 1):
            song_name = track['song_name']
            artist_uuid = artist_uuid_map[track['artist']]
            album_uuid = album_uuid_map.get(track.get('album')) if track.get('album') else None
            mp3_file = SEED_MUSIC_DIR / track['file']

            # Calculate duration
            try:
                duration_ms = get_mp3_duration(str(mp3_file))
            except Exception as e:
                log(f"    Warning: Could not get duration for {song_name}: {e}", level='WARN')
                duration_ms = 0

            # Generate music UUID
            music_uuid = str(uuid4())

            # Upload MP3 to MinIO
            try:
                audio_s3_path = upload_audio(s3_client, bucket_name, music_uuid, str(mp3_file))
            except Exception as e:
                log(f"    Error: Failed to upload audio for {song_name}: {e}", level='ERROR')
                raise

            # Upsert music track
            result_uuid = upsert_music(
                db_conn,
                music_uuid,
                song_name,
                artist_uuid,
                album_uuid,
                audio_s3_path,
                duration_ms,
                uploaded_by
            )

            # Assign tags
            for tag_name in track.get('tags', []):
                assign_tag_to_music(db_conn, result_uuid, tag_name)

            # Upload track image if specified
            if 'image' in track and track['image']:
                img_file = SEED_MUSIC_DIR / track['image']
                try:
                    upload_music_image(s3_client, bucket_name, result_uuid, str(img_file))
                except Exception as e:
                    log(f"    Warning: Failed to upload image for {song_name}: {e}", level='WARN')

            log(f"    [{idx}/{num_tracks}] ✓ {song_name} (UUID: {result_uuid})")

        log(f"  ✓ Music tracks imported")

        # Commit transaction
        log("Step 6: Committing transaction...")
        db_conn.commit()
        log("✓ Transaction committed successfully")

        # Final stats
        stats_after = get_stats(db_conn)
        log("\n" + "=" * 60)
        log("Import Complete!")
        log("=" * 60)
        log(f"Database Stats:")
        log(f"  Artists: {stats_after['artists']} (added: {stats_after['artists'] - stats_before['artists']})")
        log(f"  Albums:  {stats_after['albums']} (added: {stats_after['albums'] - stats_before['albums']})")
        log(f"  Tracks:  {stats_after['music']} (added: {stats_after['music'] - stats_before['music']})")
        log(f"  Tags:    {stats_after['tags']} (added: {stats_after['tags'] - stats_before['tags']})")
        log("=" * 60)

    except Exception as e:
        log(f"Import failed: {e}", level='ERROR')
        log("Rolling back transaction...", level='ERROR')
        db_conn.rollback()
        db_conn.close()
        sys.exit(1)

    # Clean up
    db_conn.close()
    log("✓ Database connection closed")
    log("\nMusic data ingestion completed successfully!")
    sys.exit(0)


if __name__ == '__main__':
    main()
