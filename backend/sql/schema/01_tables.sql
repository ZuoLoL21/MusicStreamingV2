-- User Table
CREATE TABLE "User" (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    hashed_password VARCHAR(255) NOT NULL,
    bio TEXT,
    profile_image_path VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE VIEW "PublicUser"
AS SELECT uuid, username, email, bio, profile_image_path, created_at, updated_at
FROM User

-- Artist Table
CREATE TABLE "Artist" (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    artist_name VARCHAR(255) NOT NULL UNIQUE,
    bio TEXT,
    profile_image_path VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "ArtistMember" (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    artist_uuid UUID NOT NULL REFERENCES "Artist"(uuid) ON DELETE CASCADE,
    user_uuid UUID NOT NULL REFERENCES "User"(uuid) ON DELETE CASCADE,
    role artist_member_role NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(artist_uuid, user_uuid)
);

CREATE INDEX idx_artistmember_artist_uuid ON "ArtistMember"(artist_uuid);
CREATE INDEX idx_artistmember_user_uuid ON "ArtistMember"(user_uuid);

-- Album Table
CREATE TABLE "Album" (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_artist UUID NOT NULL REFERENCES "Artist"(uuid) ON DELETE CASCADE,
    original_name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_artist, original_name)
);

CREATE INDEX idx_album_from_artist ON "Album"(from_artist);

-- Music Table
CREATE TABLE "Music" (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_artist UUID NOT NULL REFERENCES "Artist"(uuid) ON DELETE CASCADE,
    uploaded_by UUID NOT NULL REFERENCES "User"(uuid) ON DELETE CASCADE,
    in_album UUID REFERENCES "Album"(uuid) ON DELETE SET NULL,
    song_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    path_in_file_storage VARCHAR(500) NOT NULL,
    play_count INTEGER DEFAULT 0,
    duration_seconds INTEGER NOT NULL
);

CREATE INDEX idx_music_from_artist ON "Music"(from_artist);
CREATE INDEX idx_music_in_album ON "Music"(in_album);

-- Likes Table
CREATE TABLE "Likes" (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_user UUID NOT NULL REFERENCES "User"(uuid) ON DELETE CASCADE,
    to_music UUID NOT NULL REFERENCES "Music"(uuid) ON DELETE CASCADE,
    UNIQUE(from_user, to_music)
);

CREATE INDEX idx_likes_to_music ON "Likes"(to_music);
CREATE INDEX idx_likes_from_user ON "Likes"(from_user);

-- Follows Table
CREATE TABLE "Follows" (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_user UUID NOT NULL REFERENCES "User"(uuid) ON DELETE CASCADE,
    to_user UUID REFERENCES "User"(uuid) ON DELETE CASCADE,
    to_artist UUID REFERENCES "Artist"(uuid) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CHECK ((to_user IS NULL) != (to_artist IS NULL)),
    UNIQUE(from_user, to_user),
    UNIQUE(from_user, to_artist)
);

CREATE INDEX idx_follows_from_user ON "Follows"(from_user);
CREATE INDEX idx_follows_to_artist ON "Follows"(to_artist);
CREATE INDEX idx_follows_to_user ON "Follows"(to_user);

-- MusicTags Table
CREATE TABLE "MusicTags" (
    tag_name VARCHAR(50) PRIMARY KEY,
    tag_description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- TagAssignment Table
CREATE TABLE "TagAssignment" (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    music_uuid UUID NOT NULL REFERENCES "Music"(uuid) ON DELETE CASCADE,
    tag_name VARCHAR(50) NOT NULL REFERENCES "MusicTags"(tag_name) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(music_uuid, tag_name)
);

CREATE INDEX idx_tagassignment_music_uuid ON "TagAssignment"(music_uuid);

-- Playlist Table
CREATE TABLE "Playlist" (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_user UUID NOT NULL REFERENCES "User"(uuid) ON DELETE CASCADE,
    original_name VARCHAR(255) NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_user, original_name),
    UNIQUE (playlist_uuid, position)
);

CREATE INDEX idx_playlist_from_user ON "Playlist"(from_user);

-- PlaylistTrack Table
CREATE TABLE "PlaylistTrack" (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    music_uuid UUID NOT NULL REFERENCES "Music"(uuid) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    playlist_uuid UUID NOT NULL REFERENCES "Playlist"(uuid) ON DELETE CASCADE,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_playlisttrack_playlist_position ON "PlaylistTrack"(playlist_uuid, position);

-- ListeningHistory Table
CREATE TABLE "ListeningHistory" (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_uuid UUID NOT NULL REFERENCES "User"(uuid) ON DELETE CASCADE,
    music_uuid UUID NOT NULL REFERENCES "Music"(uuid) ON DELETE CASCADE,
    played_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    listen_duration_seconds INTEGER,
    completion_percentage DECIMAL(5,2)
);

CREATE INDEX idx_listeninghistory_user_uuid ON "ListeningHistory"(user_uuid);
CREATE INDEX idx_listeninghistory_music_uuid ON "ListeningHistory"(music_uuid);
CREATE INDEX idx_listeninghistory_played_at ON "ListeningHistory"(played_at);
