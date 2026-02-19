CREATE TABLE music_theme
(
    music_uuid UUID,
    theme LowCardinality(String),
    views UInt64 DEFAULT 0,
    successes UInt64 DEFAULT 0,
    last_update DateTime DEFAULT now()
)
    ENGINE = ReplacingMergeTree(last_update)
ORDER BY (music_uuid, theme);

CREATE TABLE user_dim
(
    user_uuid UUID,
    created_at DateTime,
    country LowCardinality(String)
)
    ENGINE = ReplacingMergeTree
ORDER BY user_uuid;
