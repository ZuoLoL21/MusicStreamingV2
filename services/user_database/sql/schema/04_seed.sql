
-- Insert Sample Users
INSERT INTO users (uuid, username, email, hashed_password, bio) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'john_doe', 'john@example.com', '$2a$10$hashedpassword1', 'Music lover and indie artist'),
    ('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'jane_smith', 'jane@example.com', '$2a$10$hashedpassword2', 'Singer-songwriter from NYC'),
    ('c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'mike_jones', 'mike@example.com', '$2a$10$hashedpassword3', 'Electronic music producer'),
    ('d3eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'sarah_wilson', 'sarah@example.com', '$2a$10$hashedpassword4', 'Jazz enthusiast');

-- Insert Sample Artists
INSERT INTO artist (uuid, artist_name, bio) VALUES
    ('e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'The Indie Collective', 'Alternative indie band from Portland'),
    ('f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', 'Jane Smith', 'Singer-songwriter specializing in acoustic folk'),
    ('06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', 'Digital Dreams', 'Electronic/EDM artist'),
    ('16eebc99-9c0b-4ef8-bb6d-6bb9bd380a18', 'The Beatles', 'Legendary rock band from Liverpool');

-- Insert ArtistMembers
INSERT INTO artist_member (uuid, artist_uuid, user_uuid, role) VALUES
    ('26eebc99-9c0b-4ef8-bb6d-6bb9bd380a19', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'owner'),
    ('36eebc99-9c0b-4ef8-bb6d-6bb9bd380a20', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'member');

-- Insert Albums
INSERT INTO album (uuid, from_artist, original_name, description) VALUES
    ('46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'Summer Vibes', 'Our debut album recorded in summer 2025'),
    ('56eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', 'Whispers in the Wind', 'Acoustic folk album'),
    ('66eebc99-9c0b-4ef8-bb6d-6bb9bd380a23', '16eebc99-9c0b-4ef8-bb6d-6bb9bd380a18', 'Abbey Road', 'Classic album from 1969');

-- Insert Music
INSERT INTO music (uuid, from_artist, uploaded_by, in_album, song_name, path_in_file_storage, play_count, duration_seconds) VALUES
    ('76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 'Sunset Boulevard', '/music/indie/sunset_boulevard.mp3', 1250, 245),
    ('86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 'Morning Coffee', '/music/indie/morning_coffee.mp3', 890, 198),
    ('96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 'f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '56eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Broken Wings', '/music/folk/broken_wings.mp3', 2100, 215),
    ('a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', 'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', NULL, 'Digital Horizon', '/music/edm/digital_horizon.mp3', 3400, 320),
    ('b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a28', '16eebc99-9c0b-4ef8-bb6d-6bb9bd380a18', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '66eebc99-9c0b-4ef8-bb6d-6bb9bd380a23', 'Come Together', '/music/beatles/come_together.mp3', 15000, 259);

-- Insert MusicTags
INSERT INTO music_tags (tag_name, tag_description) VALUES
    ('indie', 'Independent music'),
    ('folk', 'Folk and acoustic music'),
    ('electronic', 'Electronic and EDM music'),
    ('rock', 'Rock music'),
    ('chill', 'Relaxing and chill vibes'),
    ('upbeat', 'Energetic and upbeat music');

-- Insert TagAssignments
INSERT INTO tag_assignment (uuid, music_uuid, tag_name) VALUES
    ('c6eebc99-9c0b-4ef8-bb6d-6bb9bd380a29', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'indie'),
    ('d6eebc99-9c0b-4ef8-bb6d-6bb9bd380a30', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'chill'),
    ('e6eebc99-9c0b-4ef8-bb6d-6bb9bd380a31', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'indie'),
    ('f6eebc99-9c0b-4ef8-bb6d-6bb9bd380a32', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 'folk'),
    ('07eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', 'electronic'),
    ('17eebc99-9c0b-4ef8-bb6d-6bb9bd380a34', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', 'upbeat'),
    ('27eebc99-9c0b-4ef8-bb6d-6bb9bd380a35', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a28', 'rock');

-- Insert Likes
INSERT INTO likes (uuid, from_user, to_music) VALUES
    ('37eebc99-9c0b-4ef8-bb6d-6bb9bd380a36', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26'),
    ('47eebc99-9c0b-4ef8-bb6d-6bb9bd380a37', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27'),
    ('57eebc99-9c0b-4ef8-bb6d-6bb9bd380a38', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24'),
    ('67eebc99-9c0b-4ef8-bb6d-6bb9bd380a39', 'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a28');

-- Insert Follows (users following artists)
INSERT INTO follows (uuid, from_user, to_user, to_artist) VALUES
    ('77eebc99-9c0b-4ef8-bb6d-6bb9bd380a40', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NULL, 'f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16'),
    ('87eebc99-9c0b-4ef8-bb6d-6bb9bd380a41', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', NULL, 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15'),
    ('97eebc99-9c0b-4ef8-bb6d-6bb9bd380a42', 'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', NULL);

-- Insert Playlists
INSERT INTO playlist (uuid, from_user, original_name, description, is_public) VALUES
    ('a7eebc99-9c0b-4ef8-bb6d-6bb9bd380a43', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'My Favorites', 'Collection of my favorite tracks', false),
    ('b7eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'Workout Mix', 'High energy tracks for workouts', true);

-- Insert PlaylistTracks
INSERT INTO playlist_track (uuid, music_uuid, position, playlist_uuid) VALUES
    ('c7eebc99-9c0b-4ef8-bb6d-6bb9bd380a45', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 1, 'a7eebc99-9c0b-4ef8-bb6d-6bb9bd380a43'),
    ('d7eebc99-9c0b-4ef8-bb6d-6bb9bd380a46', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', 2, 'a7eebc99-9c0b-4ef8-bb6d-6bb9bd380a43'),
    ('e7eebc99-9c0b-4ef8-bb6d-6bb9bd380a47', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 3, 'a7eebc99-9c0b-4ef8-bb6d-6bb9bd380a43'),
    ('f7eebc99-9c0b-4ef8-bb6d-6bb9bd380a48', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', 1, 'b7eebc99-9c0b-4ef8-bb6d-6bb9bd380a44'),
    ('08eebc99-9c0b-4ef8-bb6d-6bb9bd380a49', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a28', 2, 'b7eebc99-9c0b-4ef8-bb6d-6bb9bd380a44');

-- Insert ListeningHistory
INSERT INTO listening_history (uuid, user_uuid, music_uuid, listen_duration_seconds, completion_percentage) VALUES
    ('18eebc99-9c0b-4ef8-bb6d-6bb9bd380a50', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 245, 100.00),
    ('28eebc99-9c0b-4ef8-bb6d-6bb9bd380a51', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 180, 83.72),
    ('38eebc99-9c0b-4ef8-bb6d-6bb9bd380a52', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', 320, 100.00),
    ('48eebc99-9c0b-4ef8-bb6d-6bb9bd380a53', 'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a28', 259, 100.00);
