CREATE TABLE music_listen_events
(
    event_time DateTime,
    user_uuid UUID,
    music_uuid UUID,
    artist_uuid UUID,
    album_uuid Nullable(UUID),

    listen_duration_seconds UInt32,
    track_duration_seconds UInt32,
    completion_ratio Float32
)
    ENGINE = MergeTree
PARTITION BY toYYYYMM(event_time)
ORDER BY (user_uuid, event_time);


CREATE TABLE music_like_events
(
    event_time DateTime,
    user_uuid UUID,
    music_uuid UUID,
    artist_uuid UUID
)
    ENGINE = MergeTree
PARTITION BY toYYYYMM(event_time)
ORDER BY (user_uuid, event_time);

