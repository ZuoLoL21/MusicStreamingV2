-- User Table
CREATE TABLE users (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    hashed_password VARCHAR(255) NOT NULL,
    bio TEXT,
    profile_image_path VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE VIEW public_user
AS SELECT uuid, username, email, bio, profile_image_path, created_at, updated_at
FROM users;

-- Artist Table
CREATE TABLE artist (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    artist_name VARCHAR(255) NOT NULL UNIQUE,
    bio TEXT,
    profile_image_path VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE artist_member (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    artist_uuid UUID NOT NULL REFERENCES artist(uuid) ON DELETE CASCADE,
    user_uuid UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    role artist_member_role NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(artist_uuid, user_uuid)
);

CREATE INDEX idx_artistmember_artist_uuid ON artist_member(artist_uuid);
CREATE INDEX idx_artistmember_user_uuid ON artist_member(user_uuid);

-- Album Table
CREATE TABLE "Album" (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_artist UUID NOT NULL REFERENCES artist(uuid) ON DELETE CASCADE,
    original_name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_artist, original_name)
);

CREATE INDEX idx_album_from_artist ON "Album"(from_artist);

-- Music Table
CREATE TABLE music (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_artist UUID NOT NULL REFERENCES artist(uuid) ON DELETE CASCADE,
    uploaded_by UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    in_album UUID REFERENCES "Album"(uuid) ON DELETE SET NULL,
    song_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    path_in_file_storage VARCHAR(500) NOT NULL,
    play_count INTEGER DEFAULT 0,
    duration_seconds INTEGER NOT NULL
);

CREATE INDEX idx_music_from_artist ON music(from_artist);
CREATE INDEX idx_music_in_album ON music(in_album);

-- Likes Table
CREATE TABLE likes (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_user UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    to_music UUID NOT NULL REFERENCES music(uuid) ON DELETE CASCADE,
    UNIQUE(from_user, to_music)
);

CREATE INDEX idx_likes_to_music ON likes(to_music);
CREATE INDEX idx_likes_from_user ON likes(from_user);

-- Follows Table
CREATE TABLE follows (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_user UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    to_user UUID REFERENCES users(uuid) ON DELETE CASCADE,
    to_artist UUID REFERENCES artist(uuid) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CHECK ((to_user IS NULL) != (to_artist IS NULL)),
    UNIQUE(from_user, to_user),
    UNIQUE(from_user, to_artist)
);

CREATE INDEX idx_follows_from_user ON follows(from_user);
CREATE INDEX idx_follows_to_artist ON follows(to_artist);
CREATE INDEX idx_follows_to_user ON follows(to_user);

-- MusicTags Table
CREATE TABLE music_tags (
    tag_name VARCHAR(50) PRIMARY KEY,
    tag_description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- TagAssignment Table
CREATE TABLE tag_assignment (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    music_uuid UUID NOT NULL REFERENCES music(uuid) ON DELETE CASCADE,
    tag_name VARCHAR(50) NOT NULL REFERENCES music_tags(tag_name) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(music_uuid, tag_name)
);

CREATE INDEX idx_tagassignment_music_uuid ON tag_assignment(music_uuid);

-- Playlist Table
CREATE TABLE playlist (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_user UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    original_name VARCHAR(255) NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_user, original_name),
    UNIQUE (playlist_uuid, position)
);

CREATE INDEX idx_playlist_from_user ON playlist(from_user);

-- PlaylistTrack Table
CREATE TABLE playlist_track (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    music_uuid UUID NOT NULL REFERENCES music(uuid) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    playlist_uuid UUID NOT NULL REFERENCES playlist(uuid) ON DELETE CASCADE,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_playlisttrack_playlist_position ON playlist_track(playlist_uuid, position);

-- ListeningHistory Table
CREATE TABLE listening_history (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_uuid UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    music_uuid UUID NOT NULL REFERENCES music(uuid) ON DELETE CASCADE,
    played_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    listen_duration_seconds INTEGER,
    completion_percentage DECIMAL(5,2)
);

CREATE INDEX idx_listeninghistory_user_uuid ON listening_history(user_uuid);
CREATE INDEX idx_listeninghistory_music_uuid ON listening_history(music_uuid);
CREATE INDEX idx_listeninghistory_played_at ON listening_history(played_at);
