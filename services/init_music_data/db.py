"""
Database operations for music data ingestion
"""

import os
import psycopg2
from psycopg2.extras import RealDictCursor
from uuid import uuid4
from datetime import datetime


def connect_postgres():
    """Connect to PostgreSQL database"""
    conn = psycopg2.connect(
        host=os.environ.get('POSTGRES_HOST', 'localhost'),
        port=int(os.environ.get('POSTGRES_PORT', 5432)),
        user=os.environ.get('POSTGRES_USER', 'postgres'),
        password=os.environ.get('POSTGRES_PASSWORD', 'postgres'),
        database=os.environ.get('POSTGRES_DB', 'music_streaming'),
    )
    conn.autocommit = False  # Use transactions
    return conn


def upsert_artist(conn, artist_id, name, bio=None, image_path=None):
    """
    Insert or update artist
    Returns artist UUID
    """
    cursor = conn.cursor(cursor_factory=RealDictCursor)

    # UPSERT: Insert or update on conflict
    query = """
        INSERT INTO artist (uuid, artist_name, bio, profile_image_path, created_at, updated_at)
        VALUES (%s, %s, %s, %s, %s, %s)
        ON CONFLICT (artist_name)
        DO UPDATE SET
            bio = EXCLUDED.bio,
            profile_image_path = EXCLUDED.profile_image_path,
            updated_at = EXCLUDED.updated_at
        RETURNING uuid
    """

    now = datetime.utcnow()
    cursor.execute(query, (
        artist_id,
        name,
        bio,
        image_path,
        now,
        now
    ))

    result = cursor.fetchone()
    cursor.close()

    return result['uuid']


def upsert_album(conn, album_id, artist_uuid, name, description=None, image_path=None):
    """
    Insert or update album
    Returns album UUID
    """
    cursor = conn.cursor(cursor_factory=RealDictCursor)

    # UPSERT: Insert or update on conflict
    query = """
        INSERT INTO album (uuid, from_artist, original_name, description, image_path, created_at, updated_at)
        VALUES (%s, %s, %s, %s, %s, %s, %s)
        ON CONFLICT (from_artist, original_name)
        DO UPDATE SET
            description = EXCLUDED.description,
            image_path = EXCLUDED.image_path,
            updated_at = EXCLUDED.updated_at
        RETURNING uuid
    """

    now = datetime.utcnow()
    cursor.execute(query, (
        album_id,
        artist_uuid,
        name,
        description,
        image_path,
        now,
        now
    ))

    result = cursor.fetchone()
    cursor.close()

    return result['uuid']


def find_music_by_artist_and_name(conn, artist_uuid, song_name):
    """
    Find existing music track by artist and song name
    Returns track data or None
    """
    cursor = conn.cursor(cursor_factory=RealDictCursor)

    query = """
        SELECT uuid, song_name, from_artist, in_album, path_in_file_storage, duration_seconds, uploaded_by
        FROM music
        WHERE from_artist = %s AND song_name = %s
    """

    cursor.execute(query, (artist_uuid, song_name))
    result = cursor.fetchone()
    cursor.close()

    return result


def upsert_music(conn, music_uuid, song_name, artist_uuid, album_uuid, audio_url, duration_ms, uploaded_by):
    """
    Insert or update music track
    Returns music UUID
    """
    cursor = conn.cursor(cursor_factory=RealDictCursor)

    # Check if track exists
    existing = find_music_by_artist_and_name(conn, artist_uuid, song_name)

    if existing:
        # Update existing track
        query = """
            UPDATE music
            SET in_album = %s,
                path_in_file_storage = %s,
                duration_seconds = %s
            WHERE uuid = %s
            RETURNING uuid
        """
        cursor.execute(query, (
            album_uuid,
            audio_url,
            duration_ms,
            existing['uuid']
        ))
        result_uuid = existing['uuid']
    else:
        # Insert new track
        query = """
            INSERT INTO music (uuid, song_name, from_artist, in_album, path_in_file_storage,
                             duration_seconds, uploaded_by)
            VALUES (%s, %s, %s, %s, %s, %s, %s)
            RETURNING uuid
        """
        cursor.execute(query, (
            music_uuid,
            song_name,
            artist_uuid,
            album_uuid,
            audio_url,
            duration_ms,
            uploaded_by
        ))
        result = cursor.fetchone()
        result_uuid = result['uuid']

    cursor.close()
    return result_uuid


def upsert_tag(conn, tag_name):
    """
    Insert tag if not exists (tags are immutable)
    Returns tag name
    """
    cursor = conn.cursor()

    # Insert only if not exists
    query = """
        INSERT INTO music_tags (tag_name, created_at)
        VALUES (%s, %s)
        ON CONFLICT (tag_name) DO NOTHING
    """

    cursor.execute(query, (tag_name, datetime.utcnow()))
    cursor.close()

    return tag_name


def assign_tag_to_music(conn, music_uuid, tag_name):
    """
    Assign tag to music track (idempotent)
    """
    cursor = conn.cursor()

    query = """
        INSERT INTO tag_assignment (music_uuid, tag_name, created_at)
        VALUES (%s, %s, %s)
        ON CONFLICT (music_uuid, tag_name) DO NOTHING
    """

    cursor.execute(query, (music_uuid, tag_name, datetime.utcnow()))
    cursor.close()


def check_data_loaded(conn, uploaded_by):
    """
    Check if data has already been loaded by this user
    Returns True if any music exists from this user
    """
    cursor = conn.cursor()

    query = """
        SELECT COUNT(*) FROM music WHERE uploaded_by = %s
    """

    cursor.execute(query, (uploaded_by,))
    count = cursor.fetchone()[0]
    cursor.close()

    return count > 0


def get_stats(conn):
    """
    Get database statistics
    Returns dict with counts
    """
    cursor = conn.cursor(cursor_factory=RealDictCursor)

    stats = {}

    # Count artists
    cursor.execute("SELECT COUNT(*) as count FROM artist")
    stats['artists'] = cursor.fetchone()['count']

    # Count albums
    cursor.execute("SELECT COUNT(*) as count FROM album")
    stats['albums'] = cursor.fetchone()['count']

    # Count music
    cursor.execute("SELECT COUNT(*) as count FROM music")
    stats['music'] = cursor.fetchone()['count']

    # Count tags
    cursor.execute("SELECT COUNT(*) as count FROM music_tags")
    stats['tags'] = cursor.fetchone()['count']

    cursor.close()
    return stats
