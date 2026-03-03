# MP3 Data Ingestion Guide

Complete guide for populating the MusicStreamingV2 database with music content.

## Overview

The MP3 ingestion system consists of two components:

1. **Config Generator CLI** (`tools/mp3-config-generator/`) - Interactive tool to scan MP3s and create YAML config
2. **Ingestion Container** (`services/init_music_data/`) - Reads the config and populates PostgreSQL + MinIO

**File Structure:** Uses a **flat directory structure** where all MP3 files are in a single `songs/` directory with filenames formatted as `Artist - Song.mp3`

## Quick Start

### Step 1: Prepare Your MP3 Files

Organize your MP3 files in a **flat directory** with clear naming:

```
songs/
├── Nirvana - Smells Like Teen Spirit.mp3
├── Nirvana - Come As You Are.mp3
├── Pearl Jam - Alive.mp3
└── AJR - BANG!.mp3
```

**Naming convention:** `Artist - Song.mp3` (the tool will parse this automatically)

### Step 2: Generate Config File

Install and run the config generator:

```bash
cd tools/mp3-config-generator
uv sync
python generator.py /path/to/songs --config music-data.yaml
```

The tool will:
1. Scan your directory for MP3 files
2. Parse filenames as "Artist - Song.mp3"
3. Prompt you to confirm/edit each track
4. Ask for album and tags
5. Generate a YAML configuration file

**Example interactive session:**

```
Scanning /path/to/songs for MP3 files...
Found 3 MP3 files

────────────────────────────────────────────────────────────
File: Nirvana - Smells Like Teen Spirit.mp3
Artist name [Nirvana]:
Song name [Smells Like Teen Spirit]:
Album name (leave empty for none) []: Nevermind
Tags (comma-separated) []: Rock, Grunge
  → Created new artist: Nirvana
  → Created new album: Nevermind
  → Created new tag: Rock
  → Created new tag: Grunge
✓ Added: Nirvana - Smells Like Teen Spirit (Nevermind)

────────────────────────────────────────────────────────────
File: Nirvana - Come As You Are.mp3
Artist name [Nirvana]:
Song name [Come As You Are]:
Album name (leave empty for none) []: Nevermind
Tags (comma-separated) []: Rock, Grunge
✓ Added: Nirvana - Come As You Are (Nevermind)
```

**Tips:**
- Press **Enter** to accept suggested names (parsed from filename)
- Type to override suggestions
- Artists, albums, and tags are auto-created as you enter them
- Tags are validated (no empty or overly long tags)
- Already-added tracks are automatically skipped
- Press **Ctrl+C** to cancel anytime

### Step 3: Review the Configuration

Open `music-data.yaml` and verify the generated config:

```yaml
version: '1.0'
uploaded_by: a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11

tags:
- name: Rock
  description: Rock music
- name: Grunge
  description: Grunge music

artists:
- id: artist_nirvana
  name: Nirvana
  bio: ''

albums:
- id: album_nevermind
  artist: artist_nirvana
  name: Nevermind
  description: ''

music:
- file: Nirvana - Smells Like Teen Spirit.mp3
  song_name: Smells Like Teen Spirit
  artist: artist_nirvana
  album: album_nevermind
  tags:
  - Rock
  - Grunge
- file: Nirvana - Come As You Are.mp3
  song_name: Come As You Are
  artist: artist_nirvana
  album: album_nevermind
  tags:
  - Rock
  - Grunge
```

**Verify:**
- Artist names are correct
- Song names are correct
- Tags are appropriate
- File names match your actual MP3 files
- Optionally add artist bios and album descriptions

**Note:** Albums are optional. Leave the album prompt empty for tracks without albums.

### Step 4: Prepare Seed Music Directory

Create a `seed_music` directory at the init_music_data root and copy your files:

Your `seed_music` directory should look like this:

```
seed_music/
├── music-data.yaml
├── Nirvana - Smells Like Teen Spirit.mp3
├── Nirvana - Come As You Are.mp3
├── Pearl Jam - Alive.mp3
└── AJR - BANG!.mp3
```

**Important:** File names in `seed_music/` must match the `file:` entries in `music-data.yaml`

### Step 5: Run the Ingestion Container

Start the Docker Compose stack:

```bash
docker-compose up init-music-data
```

The `init-music-data` container will:
1. Load and validate `music-data.yaml`
2. Connect to PostgreSQL and MinIO
3. Create artists
4. Create albums
5. Create tags
6. Import music tracks (upload MP3s to MinIO, insert metadata to PostgreSQL)
7. Assign tags to tracks
8. Exit successfully

**Important:** The container runs once and exits. It will not restart automatically.

### Step 6: Verify Import

Check the container logs:

```bash
docker logs init-music-data
```

You should see output like:
```
[INFO] ============================================================
[INFO] Music Data Ingestion Starting
[INFO] ============================================================
[INFO] Step 1: Loading configuration...
[INFO] ✓ Config loaded: /seed_music/music-data.yaml
[INFO] Step 2: Validating configuration...
[INFO] ✓ Config validation passed
[INFO]   Artists: 2
[INFO]   Tracks:  3
[INFO]   Tags:    2
[INFO] Step 3: Connecting to services...
[INFO] ✓ Connected to PostgreSQL
[INFO] ✓ Connected to MinIO (bucket: music-streaming)
[INFO] Step 5: Starting data import...
[INFO]   Creating 2 artists...
[INFO]     [1/2] ✓ Nirvana (UUID: ...)
[INFO]   ✓ Artists created/updated
...
[INFO] ============================================================
[INFO] Import Complete!
[INFO] ============================================================
```

### Step 7: Access Via API

Your music is now available via the API:

```bash
# List all artists
curl http://localhost:8080/artists

# Search for music
curl "http://localhost:8080/music/search?query=nirvana"

# Get music track details
curl http://localhost:8080/music/{uuid}
```

## Advanced Usage

### Adding More Tracks

To add more tracks to an existing config:

1. Add new MP3 files to your music directory (flat structure)

2. Run the generator again with the same config file:
   ```bash
   python generator.py /path/to/songs --config music-data.yaml
   ```

3. The tool will skip existing tracks and only prompt for new ones

4. Copy updated config and new MP3s to `seed_music/`:
   ```bash
   cp tools/mp3-config-generator/music-data.yaml seed_music/
   cp /path/to/songs/*.mp3 seed_music/
   ```

5. Re-run the ingestion container:
   ```bash
   docker-compose up init-music-data
   ```

### Idempotency

The ingestion process is fully idempotent:
- Re-running with the same config is safe
- Artists are matched by name (unique constraint)
- Albums are matched by (artist, name) (unique constraint)
- Music tracks are matched by (artist, song_name)
- Duplicate tags are ignored
- MinIO files are overwritten (S3 PUT is idempotent)

### Error Handling

The ingestion runs in **STRICT MODE**:
- All referenced MP3 files MUST exist in `/seed_music/`
- All artist references MUST be valid
- All tags MUST be defined before use
- The entire import runs in a transaction
- Any error triggers a rollback (all-or-nothing)

If the container fails:
1. Check logs: `docker logs init-music-data`
2. Fix the error in `music-data.yaml` or add missing files
3. Re-run: `docker-compose up init-music-data`

### Custom User

By default, all imported music is owned by user `john` (UUID: `a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11`).

To assign to a different user, edit `music-data.yaml`:

```yaml
uploaded_by: "your-user-uuid-here"
```

## Configuration Reference

### YAML Structure

```yaml
version: "1.0"                     # Config version (required)
uploaded_by: "uuid"                # User UUID who uploads (required)

tags:                              # Global tag definitions (optional)
  - name: "Rock"                   # Tag name (required)
    description: "Rock genre"      # Tag description (optional)

artists:                           # Artist definitions (required)
  - id: "artist_nirvana"           # Local reference ID (required)
    name: "Nirvana"                # Artist name (required, unique)
    bio: "Band bio"                # Artist bio (optional)

albums:                            # Album definitions (optional)
  - id: "album_nevermind"          # Local reference ID (required)
    artist: "artist_nirvana"       # Artist reference (required)
    name: "Nevermind"              # Album name (required)
    description: "Album info"      # Description (optional)

music:                             # Music tracks (required, at least 1)
  - file: "Artist - Song.mp3"      # MP3 filename (required, flat structure)
    song_name: "Song Title"        # Track title (required)
    artist: "artist_nirvana"       # Artist reference (required)
    album: "album_nevermind"       # Album reference (optional)
    tags: ["Rock", "Grunge"]       # Tag list (optional)
```

### Validation Rules

**Required Fields:**
- `version` - Must be "1.0"
- `uploaded_by` - Valid UUID (36 characters)
- `music` - At least one track
- `music[].file` - MP3 filename (must exist in `seed_music/`)
- `music[].song_name` - Track title
- `music[].artist` - Artist reference ID
- `artists[].id` - Unique artist reference ID
- `artists[].name` - Unique artist name

**Optional Fields:**
- `tags` - Can be empty array
- `albums` - Can be empty array
- `music[].album` - Album reference (can be omitted)
- `music[].tags` - Can be empty array
- `artists[].bio` - Artist biography
- `albums[].description` - Album description

**Constraints:**
- Artist names must be unique
- Album names must be unique per artist
- All file paths are **just filenames** (flat structure in `seed_music/`)
- All artist/album/tag references must be defined
- Tags used by tracks must be declared in `tags` section

## Troubleshooting

### "Config file not found"
- Ensure `music-data.yaml` is in `seed_music/` directory
- Check file name spelling (case-sensitive on Linux)

### "MP3 file not found"
- Verify filenames in config match files in `seed_music/` exactly
- File names are case-sensitive
- Check that MP3 files were copied to `seed_music/` root (no subdirectories)

### "References undefined tag"
- Add the tag to the `tags:` section first
- Tag names are case-sensitive
- Tags are auto-created by the generator, but check for typos if manually edited

### "References undefined artist"
- Verify artist ID in `artists:` section matches the reference
- Check for typos in artist IDs
- Artists are auto-created by the generator

### "Invalid tags" during entry
- Empty tags are not allowed
- Tags longer than 50 characters are rejected
- Re-enter valid tags when prompted

### "Database connection failed"
- Ensure PostgreSQL is healthy: `docker ps | grep postgres`
- Check environment variables in `.env` file
- Verify `init-music-data` depends_on is correct

### "MinIO upload failed"
- Ensure MinIO is running: `docker ps | grep minio`
- Verify bucket was created by `init-minio` container
- Check MinIO credentials in `.env`

## File Locations

- **Config Generator:** `tools/mp3-config-generator/`
  - Tool script: `generator.py`
  - Dependencies: `pyproject.toml`
- **Ingestion Container:** `services/init_music_data/`
  - Main script: `ingest.py`
  - Database ops: `db.py`
  - Storage ops: `storage.py`
  - Config validation: `config.py`
- **Seed Music Directory:** `seed_music/` (gitignored)
  - Configuration: `music-data.yaml`
  - MP3 files: `*.mp3` (flat structure)
- **Docker Compose:** `docker-compose.yml` (init-music-data service)
- **Documentation:** `MP3_INGESTION_GUIDE.md` (this file)

## Architecture Notes

### Flat Directory Structure

The system uses a flat directory structure for simplicity:
- **Generator** expects: `songs/Artist - Song.mp3`
- **Config** contains: `file: Artist - Song.mp3` (just filename)
- **Ingestion** reads from: `/seed_music/Artist - Song.mp3`

Benefits:
- Simple organization
- Easy file management
- Clear filename parsing
- No nested directory complexity

### Semi-Automatic Approach

The config generator uses a semi-automatic approach:
- **Automatic**: Scans files, parses filenames for suggestions
- **Manual**: You confirm/edit names and add tags/albums
- **Benefits**:
  - Fast for well-named files (just press Enter)
  - Full control over metadata
  - No dependency on ID3 tags
  - Transparent and reviewable

### Database Schema

The ingestion inserts into these tables:
- `artist` - Artist metadata
- `album` - Album metadata (linked to artists)
- `music` - Music track metadata
- `tag` - Tag definitions
- `music_tag` - Many-to-many tag assignments

### MinIO Storage

MP3 files are stored in MinIO S3:
- Bucket: `music-streaming`
- Path: `audio/{uuid}.mp3`

### Default User

The seed data includes a default user 'john':
```sql
-- From 04_seed.sql
INSERT INTO "user" (uuid, email, hashed_password, username, created_at, updated_at)
VALUES (
  'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
  'john.doe@example.com',
  'hashed_password_placeholder',
  'john',
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
);
```

All imported music is assigned to this user by default.

## Best Practices

1. **Name your files correctly:** Use `Artist - Song.mp3` format for automatic parsing
2. **Use consistent tags:** Reuse existing tags rather than creating variants (e.g., "Rock" not "rock music")
3. **Keep configs backed up:** The YAML file is your source of truth
4. **Test with small batches:** Import a few tracks first to verify your setup
5. **Review before importing:** Check the YAML file for errors before running the container
6. **Use albums wisely:** Only group tracks into albums if they truly belong together

## Next Steps

After importing your music:
1. Test playback via the frontend UI
2. Test API endpoints for search/browse
3. Add more music by re-running the generator
4. Customize artist bios and album descriptions in the YAML file
5. Tag your music appropriately for better recommendations

## Support

For issues or questions:
- Check the troubleshooting section above
- Review container logs: `docker logs init-music-data`
- Review the generated YAML file for errors
- Verify file naming matches the expected format
- Open an issue in the project repository
