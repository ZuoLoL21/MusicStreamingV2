#!/usr/bin/env python3
"""
MP3 Config Generator - Semi-automatic music entry tool
"""

import sys
from pathlib import Path
import click
import yaml


DEFAULT_USER_UUID = "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"  # john


def scan_directory(path):
    """Recursively find all .mp3 files in directory"""
    mp3_files = []
    path_obj = Path(path)

    if not path_obj.exists():
        raise FileNotFoundError(f"Directory not found: {path}")

    for mp3_file in path_obj.rglob("*.mp3"):
        if mp3_file.is_file():
            mp3_files.append(mp3_file)

    return sorted(mp3_files)


def parse_filename(mp3_path):
    """Parse 'Artist - Song' from filename (expects flat structure)"""
    filename_stem = mp3_path.stem

    # Expected format: "Artist - Song.mp3"
    if " - " in filename_stem:
        parts = filename_stem.split(" - ", 1)
        artist = parts[0].strip()
        song = parts[1].strip()
    else:
        # No delimiter: use filename as song, artist unknown
        artist = "Unknown Artist"
        song = filename_stem

    return artist, song


def load_config(config_path):
    """Load existing config or create new one"""
    if Path(config_path).exists():
        with open(config_path, 'r', encoding='utf-8') as f:
            return yaml.safe_load(f)

    return {
        'version': '1.0',
        'uploaded_by': DEFAULT_USER_UUID,
        'tags': [],
        'artists': [],
        'albums': [],
        'music': []
    }


def save_config(config, config_path):
    """Save config to YAML file"""
    with open(config_path, 'w', encoding='utf-8') as f:
        f.write("# MusicStreamingV2 - MP3 Configuration\n\n")
        yaml.dump(config, f,
                  default_flow_style=False,
                  allow_unicode=True,
                  sort_keys=False,
                  width=100)
    click.echo(f"✓ Saved to {config_path}")


def sanitize_id(name):
    """Convert name to valid ID"""
    safe_name = ''.join(c if c.isalnum() or c in (' ', '_') else '_' for c in name)
    safe_name = safe_name.replace(' ', '_').lower()
    while '__' in safe_name:
        safe_name = safe_name.replace('__', '_')
    return safe_name.strip('_')


def get_or_create_artist(config, artist_name):
    """Get existing artist ID or create new one"""
    artist_id = f"artist_{sanitize_id(artist_name)}"

    # Check if artist exists
    for artist in config['artists']:
        if artist['id'] == artist_id:
            return artist_id

    # Create new artist
    config['artists'].append({
        'id': artist_id,
        'name': artist_name,
        'bio': ''
    })
    click.echo(f"  → Created new artist: {artist_name}")
    return artist_id


def get_or_create_album(config, album_name, artist_id):
    """Get existing album ID or create new one"""
    album_id = f"album_{sanitize_id(album_name)}"

    # Check if album exists
    for album in config['albums']:
        if album['id'] == album_id and album['artist'] == artist_id:
            return album_id

    # Create new album
    config['albums'].append({
        'id': album_id,
        'artist': artist_id,
        'name': album_name,
        'description': ''
    })
    click.echo(f"  → Created new album: {album_name}")
    return album_id


def get_or_create_tag(config, tag_name):
    """Get existing tag or create new one"""
    # Check if tag exists
    for tag in config['tags']:
        if tag['name'].lower() == tag_name.lower():
            return tag['name']

    # Create new tag
    config['tags'].append({
        'name': tag_name,
        'description': f'{tag_name} music'
    })
    click.echo(f"  → Created new tag: {tag_name}")
    return tag_name


def validate_tags(tag_list):
    """Validate tag names"""
    for tag in tag_list:
        if not tag or not tag.strip():
            return False, "Empty tag not allowed"
        if len(tag) > 50:
            return False, f"Tag too long: {tag}"
    return True, None


@click.command()
@click.argument('music_dir', type=click.Path(exists=True))
@click.option('--config', '-c', default='music-data.yaml', help='Output config file')
def generate(music_dir, config):
    """Scan directory and interactively add music entries"""

    # Load existing config or create new
    config_data = load_config(config)
    click.echo(f"Loading config: {config}")
    click.echo(f"Current: {len(config_data['artists'])} artists, {len(config_data['music'])} tracks\n")

    # Scan for MP3 files
    click.echo(f"Scanning {music_dir} for MP3 files...")
    mp3_files = scan_directory(music_dir)

    if not mp3_files:
        click.echo("No MP3 files found!", err=True)
        sys.exit(1)

    click.echo(f"Found {len(mp3_files)} MP3 files\n")

    # Track which files already exist in config
    existing_files = {track['file'] for track in config_data.get('music', [])}

    added_count = 0
    skipped_count = 0

    for mp3_file in mp3_files:
        # Use just the filename (flat structure)
        filename = mp3_file.name

        # Skip if already in config
        if filename in existing_files:
            click.echo(f"⊘ Skipping (already in config): {filename}")
            skipped_count += 1
            continue

        click.echo("─" * 60)
        click.echo(f"File: {filename}")

        # Parse filename for suggestions
        suggested_artist, suggested_song = parse_filename(mp3_file)

        # Prompt with suggestions
        artist_name = click.prompt("Artist name", default=suggested_artist, type=str)
        if not artist_name:
            click.echo("Artist name required! Skipping...")
            continue

        song_name = click.prompt("Song name", default=suggested_song, type=str)
        if not song_name:
            click.echo("Song name required! Skipping...")
            continue

        # Get album (optional)
        album_name = click.prompt("Album name (leave empty for none)", default="", type=str)

        # Get tags (with validation)
        while True:
            tags_input = click.prompt("Tags (comma-separated)", default="", type=str)
            tag_list = [t.strip() for t in tags_input.split(',') if t.strip()]

            valid, error = validate_tags(tag_list)
            if valid:
                break
            click.echo(f"✗ Invalid tags: {error}")

        # Create artist and tags
        artist_id = get_or_create_artist(config_data, artist_name)
        tag_names = [get_or_create_tag(config_data, tag) for tag in tag_list]

        # Create album if specified
        album_id = None
        if album_name:
            album_id = get_or_create_album(config_data, album_name, artist_id)

        # Add music entry (flat structure - just filename)
        music_entry = {
            'file': filename,
            'song_name': song_name,
            'artist': artist_id,
            'tags': tag_names
        }

        if album_id:
            music_entry['album'] = album_id

        config_data['music'].append(music_entry)

        click.echo(f"✓ Added: {artist_name} - {song_name}" + (f" ({album_name})" if album_name else ""))
        added_count += 1

    # Save config
    click.echo("\n" + "─" * 60)
    if added_count > 0:
        save_config(config_data, config)

    click.echo(f"\nSummary:")
    click.echo(f"  Added:   {added_count} tracks")
    click.echo(f"  Skipped: {skipped_count} tracks")
    click.echo(f"  Total:   {len(config_data['artists'])} artists, {len(config_data['music'])} tracks")


if __name__ == '__main__':
    generate()
