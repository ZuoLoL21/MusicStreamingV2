CREATE TABLE music_listen_events
(
    event_time DateTime,
    user_uuid UUID,
    music_uuid UUID,
    artist_uuid UUID,
    album_uuid Nullable(UUID),

    listen_duration_seconds UInt32,
    track_duration_seconds UInt32,
    completion_ratio Float32,

    source Enum8(
        'search' = 1,
        'playlist' = 2,
        'album' = 3,
        'artist_page' = 4,
        'recommendation' = 5
    )
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


CREATE TABLE playlist_add_events
(
    event_time DateTime,
    user_uuid UUID,
    playlist_uuid UUID,
    music_uuid UUID
)
    ENGINE = MergeTree
PARTITION BY toYYYYMM(event_time)
ORDER BY (user_uuid, event_time);


CREATE TABLE follow_artist_add_events
(
    event_time DateTime,
    user_uuid UUID,
    artist_uuid UUID
)
    ENGINE = MergeTree
PARTITION BY toYYYYMM(event_time)
ORDER BY (user_uuid, event_time);


CREATE TABLE follow_user_add_events
(
    event_time DateTime,
    user_uuid UUID,
    poster_uuid UUID
)
    ENGINE = MergeTree
PARTITION BY toYYYYMM(event_time)
ORDER BY (user_uuid, event_time);