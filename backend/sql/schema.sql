-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create ENUM types
CREATE TYPE artist_member_role AS ENUM ('owner', 'manager', 'member');

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

-- ============================================
-- TRIGGERS FOR AUTO-UPDATING updated_at
-- ============================================

-- Create function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add triggers to tables with updated_at columns
CREATE TRIGGER update_user_updated_at
    BEFORE UPDATE ON "User"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_artist_updated_at
    BEFORE UPDATE ON "Artist"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_album_updated_at
    BEFORE UPDATE ON "Album"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_music_updated_at
    BEFORE UPDATE ON "Music"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_playlist_updated_at
    BEFORE UPDATE ON "Playlist"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- SAMPLE DATA
-- ============================================

-- Insert Sample Users
INSERT INTO "User" (uuid, username, email, hashed_password, bio) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'john_doe', 'john@example.com', '$2a$10$hashedpassword1', 'Music lover and indie artist'),
    ('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'jane_smith', 'jane@example.com', '$2a$10$hashedpassword2', 'Singer-songwriter from NYC'),
    ('c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'mike_jones', 'mike@example.com', '$2a$10$hashedpassword3', 'Electronic music producer'),
    ('d3eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'sarah_wilson', 'sarah@example.com', '$2a$10$hashedpassword4', 'Jazz enthusiast');

-- Insert Sample Artists
INSERT INTO "Artist" (uuid, user_uuid, artist_name, bio) VALUES
    ('e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'The Indie Collective', 'Alternative indie band from Portland'),
    ('f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'Jane Smith', 'Singer-songwriter specializing in acoustic folk'),
    ('06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', 'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'Digital Dreams', 'Electronic/EDM artist'),
    ('16eebc99-9c0b-4ef8-bb6d-6bb9bd380a18', NULL, 'The Beatles', 'Legendary rock band from Liverpool');

-- Insert ArtistMembers
INSERT INTO "ArtistMember" (uuid, artist_uuid, user_uuid, role) VALUES
    ('26eebc99-9c0b-4ef8-bb6d-6bb9bd380a19', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'owner'),
    ('36eebc99-9c0b-4ef8-bb6d-6bb9bd380a20', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'member');

-- Insert Albums
INSERT INTO "Album" (uuid, from_artist, original_name, description) VALUES
    ('46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'Summer Vibes', 'Our debut album recorded in summer 2025'),
    ('56eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', 'Whispers in the Wind', 'Acoustic folk album'),
    ('66eebc99-9c0b-4ef8-bb6d-6bb9bd380a23', '16eebc99-9c0b-4ef8-bb6d-6bb9bd380a18', 'Abbey Road', 'Classic album from 1969');

-- Insert Music
INSERT INTO "Music" (uuid, from_artist, uploaded_by, in_album, song_name, path_in_file_storage, play_count, duration_seconds) VALUES
    ('76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 'Sunset Boulevard', '/music/indie/sunset_boulevard.mp3', 1250, 245),
    ('86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 'Morning Coffee', '/music/indie/morning_coffee.mp3', 890, 198),
    ('96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 'f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '56eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Broken Wings', '/music/folk/broken_wings.mp3', 2100, 215),
    ('a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', 'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', NULL, 'Digital Horizon', '/music/edm/digital_horizon.mp3', 3400, 320),
    ('b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a28', '16eebc99-9c0b-4ef8-bb6d-6bb9bd380a18', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '66eebc99-9c0b-4ef8-bb6d-6bb9bd380a23', 'Come Together', '/music/beatles/come_together.mp3', 15000, 259);

-- Insert MusicTags
INSERT INTO "MusicTags" (tag_name, tag_description) VALUES
    ('indie', 'Independent music'),
    ('folk', 'Folk and acoustic music'),
    ('electronic', 'Electronic and EDM music'),
    ('rock', 'Rock music'),
    ('chill', 'Relaxing and chill vibes'),
    ('upbeat', 'Energetic and upbeat music');

-- Insert TagAssignments
INSERT INTO "TagAssignment" (uuid, music_uuid, tag_name) VALUES
    ('c6eebc99-9c0b-4ef8-bb6d-6bb9bd380a29', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'indie'),
    ('d6eebc99-9c0b-4ef8-bb6d-6bb9bd380a30', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'chill'),
    ('e6eebc99-9c0b-4ef8-bb6d-6bb9bd380a31', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'indie'),
    ('f6eebc99-9c0b-4ef8-bb6d-6bb9bd380a32', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 'folk'),
    ('07eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', 'electronic'),
    ('17eebc99-9c0b-4ef8-bb6d-6bb9bd380a34', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', 'upbeat'),
    ('27eebc99-9c0b-4ef8-bb6d-6bb9bd380a35', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a28', 'rock');

-- Insert Likes
INSERT INTO "Likes" (uuid, from_user, to_music) VALUES
    ('37eebc99-9c0b-4ef8-bb6d-6bb9bd380a36', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26'),
    ('47eebc99-9c0b-4ef8-bb6d-6bb9bd380a37', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27'),
    ('57eebc99-9c0b-4ef8-bb6d-6bb9bd380a38', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24'),
    ('67eebc99-9c0b-4ef8-bb6d-6bb9bd380a39', 'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a28');

-- Insert Follows (users following artists)
INSERT INTO "Follows" (uuid, from_user, to_user, to_artist) VALUES
    ('77eebc99-9c0b-4ef8-bb6d-6bb9bd380a40', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NULL, 'f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16'),
    ('87eebc99-9c0b-4ef8-bb6d-6bb9bd380a41', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', NULL, 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15'),
    ('97eebc99-9c0b-4ef8-bb6d-6bb9bd380a42', 'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', NULL);

-- Insert Playlists
INSERT INTO "Playlist" (uuid, from_user, original_name, description, is_public) VALUES
    ('a7eebc99-9c0b-4ef8-bb6d-6bb9bd380a43', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'My Favorites', 'Collection of my favorite tracks', false),
    ('b7eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'Workout Mix', 'High energy tracks for workouts', true);

-- Insert PlaylistTracks
INSERT INTO "PlaylistTrack" (uuid, music_uuid, position, playlist_uuid) VALUES
    ('c7eebc99-9c0b-4ef8-bb6d-6bb9bd380a45', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 1, 'a7eebc99-9c0b-4ef8-bb6d-6bb9bd380a43'),
    ('d7eebc99-9c0b-4ef8-bb6d-6bb9bd380a46', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', 2, 'a7eebc99-9c0b-4ef8-bb6d-6bb9bd380a43'),
    ('e7eebc99-9c0b-4ef8-bb6d-6bb9bd380a47', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 3, 'a7eebc99-9c0b-4ef8-bb6d-6bb9bd380a43'),
    ('f7eebc99-9c0b-4ef8-bb6d-6bb9bd380a48', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', 1, 'b7eebc99-9c0b-4ef8-bb6d-6bb9bd380a44'),
    ('08eebc99-9c0b-4ef8-bb6d-6bb9bd380a49', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a28', 2, 'b7eebc99-9c0b-4ef8-bb6d-6bb9bd380a44');

-- Insert ListeningHistory
INSERT INTO "ListeningHistory" (uuid, user_uuid, music_uuid, listen_duration_seconds, completion_percentage) VALUES
    ('18eebc99-9c0b-4ef8-bb6d-6bb9bd380a50', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 245, 100.00),
    ('28eebc99-9c0b-4ef8-bb6d-6bb9bd380a51', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 180, 83.72),
    ('38eebc99-9c0b-4ef8-bb6d-6bb9bd380a52', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', 320, 100.00),
    ('48eebc99-9c0b-4ef8-bb6d-6bb9bd380a53', 'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a28', 259, 100.00);
