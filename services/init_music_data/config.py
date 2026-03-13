"""
Configuration validation for music data ingestion
"""

import os
from pathlib import Path
import yaml
from pydub import AudioSegment


class ConfigValidationError(Exception):
    """Raised when config validation fails"""
    pass


def load_config(config_path):
    """Load YAML config file"""
    if not os.path.exists(config_path):
        raise ConfigValidationError(f"Config file not found: {config_path}")

    try:
        with open(config_path, 'r', encoding='utf-8') as f:
            config = yaml.safe_load(f)
        return config
    except yaml.YAMLError as e:
        raise ConfigValidationError(f"Invalid YAML: {e}")


def validate_config(config, base_dir):
    """Validate configuration structure and references."""
    errors = []
    base_path = Path(base_dir)

    # Validate required top-level fields
    required_fields = ['version', 'uploaded_by']
    for field in required_fields:
        if field not in config:
            errors.append(f"Missing required field: {field}")

    if errors:
        raise ConfigValidationError("\n".join(errors))

    # Validate version
    if config['version'] != '1.0':
        errors.append(f"Unsupported config version: {config['version']} (expected 1.0)")

    # Validate uploaded_by is a valid UUID format (basic check)
    uploaded_by = config.get('uploaded_by', '')
    if not uploaded_by or len(uploaded_by) != 36:
        errors.append(f"Invalid uploaded_by UUID: {uploaded_by}")

    # Build reference maps
    artist_ids = set()
    artist_names = set()
    album_ids = set()
    tag_names = set()

    # Validate tags
    for tag in config.get('tags', []):
        if 'name' not in tag:
            errors.append(f"Tag missing 'name' field: {tag}")
            continue

        tag_name = tag['name']
        if tag_name in tag_names:
            errors.append(f"Duplicate tag name: {tag_name}")
        tag_names.add(tag_name)

    # Validate artists
    for idx, artist in enumerate(config.get('artists', [])):
        artist_ref = f"Artist #{idx + 1}"

        # Check required fields
        if 'id' not in artist:
            errors.append(f"{artist_ref}: Missing 'id' field")
            continue
        if 'name' not in artist:
            errors.append(f"{artist_ref}: Missing 'name' field")
            continue

        artist_id = artist['id']
        artist_name = artist['name']

        # Check for duplicates
        if artist_id in artist_ids:
            errors.append(f"{artist_ref}: Duplicate artist ID: {artist_id}")
        artist_ids.add(artist_id)

        if artist_name in artist_names:
            errors.append(f"{artist_ref}: Duplicate artist name: {artist_name}")
        artist_names.add(artist_name)

        # Check image exists if specified
        if 'image' in artist and artist['image']:
            img_path = base_path / artist['image']
            if not img_path.exists():
                errors.append(f"{artist_ref}: Image file not found: {artist['image']}")

    # Validate albums
    for idx, album in enumerate(config.get('albums', [])):
        album_ref = f"Album #{idx + 1}"

        # Check required fields
        if 'id' not in album:
            errors.append(f"{album_ref}: Missing 'id' field")
            continue
        if 'artist' not in album:
            errors.append(f"{album_ref}: Missing 'artist' field")
            continue
        if 'name' not in album:
            errors.append(f"{album_ref}: Missing 'name' field")
            continue

        album_id = album['id']

        # Check for duplicate ID
        if album_id in album_ids:
            errors.append(f"{album_ref}: Duplicate album ID: {album_id}")
        album_ids.add(album_id)

        # Check artist reference
        if album['artist'] not in artist_ids:
            errors.append(f"{album_ref} ({album['name']}): References undefined artist: {album['artist']}")

        # Check image exists if specified
        if 'image' in album and album['image']:
            img_path = base_path / album['image']
            if not img_path.exists():
                errors.append(f"{album_ref}: Image file not found: {album['image']}")

    # Validate music tracks
    if 'music' not in config or not config['music']:
        errors.append("No music tracks defined (config must have at least one track)")
    else:
        for idx, track in enumerate(config['music']):
            track_ref = f"Track #{idx + 1}"
            song_name = track.get('song_name', 'unknown')

            # Check required fields
            if 'file' not in track:
                errors.append(f"{track_ref}: Missing 'file' field")
                continue
            if 'song_name' not in track:
                errors.append(f"{track_ref}: Missing 'song_name' field")
            if 'artist' not in track:
                errors.append(f"{track_ref} ({song_name}): Missing 'artist' field")
                continue

            # Check MP3 file exists
            mp3_path = base_path / track['file']
            if not mp3_path.exists():
                errors.append(f"{track_ref} ({song_name}): MP3 file not found: {track['file']}")
            elif mp3_path.suffix.lower() != '.mp3':
                errors.append(f"{track_ref} ({song_name}): File is not an MP3: {track['file']}")
            # Note: We don't validate MP3 readability here anymore
            # Some files may have issues with ffmpeg but can still be stored/played

            # Check artist reference
            if track['artist'] not in artist_ids:
                errors.append(f"{track_ref} ({song_name}): References undefined artist: {track['artist']}")

            # Check album reference (optional)
            if 'album' in track and track['album']:
                if track['album'] not in album_ids:
                    errors.append(f"{track_ref} ({song_name}): References undefined album: {track['album']}")

            # Check tags - STRICT MODE: all tags must be defined
            for tag_name in track.get('tags', []):
                if tag_name not in tag_names:
                    errors.append(
                        f"{track_ref} ({song_name}): References undefined tag '{tag_name}'. "
                        f"Add it to the 'tags' section first."
                    )

            # Check track image if specified
            if 'image' in track and track['image']:
                img_path = base_path / track['image']
                if not img_path.exists():
                    errors.append(f"{track_ref} ({song_name}): Image file not found: {track['image']}")

    # Raise if any errors
    if errors:
        error_msg = "Configuration validation failed:\n" + "\n".join(f"  - {err}" for err in errors)
        raise ConfigValidationError(error_msg)


def validate_mp3(mp3_path):
    """Validate MP3 file is readable and get duration.

    Returns 0 if file cannot be decoded (e.g., metadata issues).
    """
    try:
        audio = AudioSegment.from_mp3(mp3_path)
        return len(audio)  # Duration in ms
    except Exception as e:
        # Log warning but don't fail - some MP3s may have metadata issues
        # but can still be stored and played by other software
        print(f"Warning: Cannot extract duration from {mp3_path}: {e}")
        return 0


def get_mp3_duration(mp3_path):
    """Get MP3 duration in milliseconds, or 0 if file cannot be processed."""
    return validate_mp3(mp3_path)
