-- Seed data for data_warehouse (ClickHouse)
-- This creates realistic data for user 'john_doe' (a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11)

-- ============================================================================
-- User Dimension Table
-- ============================================================================
INSERT INTO user_dim (user_uuid, created_at, country) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '2025-01-15 10:30:00', 'US');

-- ============================================================================
-- Music Theme Mappings
-- ============================================================================
-- Map the music tracks to themes (similar to tags but for the bandit system)
INSERT INTO music_theme (music_uuid, theme) VALUES
    ('76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'indie'),
    ('76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'chill'),
    ('86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'indie'),
    ('86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'chill'),
    ('96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 'folk'),
    ('96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 'acoustic'),
    ('a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', 'electronic'),
    ('a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', 'upbeat'),
    ('a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', 'energetic'),
    ('b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a28', 'rock'),
    ('b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a28', 'classic');

-- ============================================================================
-- Music Listen Events
-- ============================================================================
-- Simulating 30 days of listening history for john_doe
-- Pattern: User likes indie/chill and electronic, moderate on folk, less on rock

-- Week 1 (2026-01-20 to 2026-01-26) - Discovery phase
INSERT INTO music_listen_events (event_time, user_uuid, music_uuid, artist_uuid, album_uuid, listen_duration_seconds, track_duration_seconds, completion_ratio) VALUES
    ('2026-01-20 08:15:30', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-01-20 08:20:15', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 198, 198, 1.0),
    ('2026-01-20 14:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-01-21 09:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-01-21 09:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-01-21 16:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 280, 320, 0.875),
    ('2026-01-22 10:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 'f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', '56eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 180, 215, 0.837),
    ('2026-01-22 15:20:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 'f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', '56eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 215, 215, 1.0),
    ('2026-01-23 08:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-01-23 12:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-01-23 18:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 198, 198, 1.0),
    ('2026-01-24 11:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'b6eebc99-9c0b-4ef8-bb6d-6bb9bd380a28', '16eebc99-9c0b-4ef8-bb6d-6bb9bd380a18', '66eebc99-9c0b-4ef8-bb6d-6bb9bd380a23', 120, 259, 0.463),
    ('2026-01-24 15:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-01-25 10:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-01-25 10:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-01-25 14:20:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-01-25 16:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 198, 198, 1.0),
    ('2026-01-26 11:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-01-26 15:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0);

-- Week 2 (2026-01-27 to 2026-02-02) - Established preferences
INSERT INTO music_listen_events (event_time, user_uuid, music_uuid, artist_uuid, album_uuid, listen_duration_seconds, track_duration_seconds, completion_ratio) VALUES
    ('2026-01-27 09:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-01-27 17:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-01-28 08:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 198, 198, 1.0),
    ('2026-01-28 14:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-01-29 10:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-01-29 16:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 'f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', '56eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 200, 215, 0.930),
    ('2026-01-30 11:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-01-30 18:20:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-01-31 09:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 198, 198, 1.0),
    ('2026-01-31 15:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-02-01 12:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-02-01 17:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-02-02 10:20:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0);

-- Week 3 (2026-02-03 to 2026-02-09) - Continued engagement
INSERT INTO music_listen_events (event_time, user_uuid, music_uuid, artist_uuid, album_uuid, listen_duration_seconds, track_duration_seconds, completion_ratio) VALUES
    ('2026-02-03 08:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-02-03 14:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-02-04 09:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 198, 198, 1.0),
    ('2026-02-04 16:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 305, 320, 0.953),
    ('2026-02-05 11:20:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-02-05 18:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 'f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', '56eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 215, 215, 1.0),
    ('2026-02-06 10:10:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-02-06 15:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-02-07 12:40:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 198, 198, 1.0),
    ('2026-02-07 17:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-02-08 09:25:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-02-08 14:50:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-02-09 11:05:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 198, 198, 1.0),
    ('2026-02-09 16:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0);

-- Week 4 (2026-02-10 to 2026-02-16) - Recent activity
INSERT INTO music_listen_events (event_time, user_uuid, music_uuid, artist_uuid, album_uuid, listen_duration_seconds, track_duration_seconds, completion_ratio) VALUES
    ('2026-02-10 08:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-02-10 13:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-02-11 10:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 198, 198, 1.0),
    ('2026-02-11 17:20:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-02-12 09:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-02-12 15:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 'f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', '56eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 190, 215, 0.884),
    ('2026-02-13 11:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-02-13 18:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-02-14 10:10:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 198, 198, 1.0),
    ('2026-02-14 16:40:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-02-15 12:20:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-02-15 17:50:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-02-16 09:35:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 198, 198, 1.0),
    ('2026-02-16 14:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0);

-- Most recent days (2026-02-17 to 2026-02-19) - Very recent
INSERT INTO music_listen_events (event_time, user_uuid, music_uuid, artist_uuid, album_uuid, listen_duration_seconds, track_duration_seconds, completion_ratio) VALUES
    ('2026-02-17 08:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-02-17 14:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-02-18 10:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 198, 198, 1.0),
    ('2026-02-18 16:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0),
    ('2026-02-19 09:20:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '46eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', 245, 245, 1.0),
    ('2026-02-19 15:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', NULL, 320, 320, 1.0);

-- ============================================================================
-- Music Like Events
-- ============================================================================
-- User likes certain tracks (corresponds to strong preferences in listen data)
INSERT INTO music_like_events (event_time, user_uuid, music_uuid, artist_uuid) VALUES
    ('2026-01-20 08:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15'),
    ('2026-01-21 09:35:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17'),
    ('2026-01-22 15:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '96eebc99-9c0b-4ef8-bb6d-6bb9bd380a26', 'f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a16'),
    ('2026-01-23 12:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '86eebc99-9c0b-4ef8-bb6d-6bb9bd380a25', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15'),
    ('2026-02-10 13:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '76eebc99-9c0b-4ef8-bb6d-6bb9bd380a24', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15'),
    ('2026-02-14 17:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a6eebc99-9c0b-4ef8-bb6d-6bb9bd380a27', '06eebc99-9c0b-4ef8-bb6d-6bb9bd380a17');

-- ============================================================================
-- Bandit Choices
-- ============================================================================
-- Simulated bandit algorithm decisions and rewards
-- Action IDs would map to theme indices (e.g., 0=indie, 1=electronic, 2=folk, 3=rock, etc.)
INSERT INTO bandit_choices (event_time, user_uuid, action, reward) VALUES
    ('2026-01-20 08:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-01-21 08:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 1, 1.0),
    ('2026-01-22 10:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 2, 0.837),
    ('2026-01-23 07:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-01-24 10:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 3, 0.463),
    ('2026-01-25 09:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 1, 1.0),
    ('2026-01-26 11:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-01-27 09:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-01-28 08:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-01-29 10:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 1, 1.0),
    ('2026-01-30 10:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-01-31 09:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-02-01 11:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 1, 1.0),
    ('2026-02-02 10:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-02-03 08:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 1, 1.0),
    ('2026-02-04 08:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-02-05 11:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-02-06 10:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 1, 1.0),
    ('2026-02-07 12:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-02-08 09:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-02-09 11:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-02-10 08:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 1, 1.0),
    ('2026-02-11 10:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-02-12 08:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-02-13 11:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 1, 1.0),
    ('2026-02-14 10:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-02-15 12:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-02-16 09:30:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-02-17 07:45:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 1, 1.0),
    ('2026-02-18 10:00:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0),
    ('2026-02-19 09:15:00', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 0, 1.0);
