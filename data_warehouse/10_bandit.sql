CREATE TABLE bandit_data
(
    user_uuid UUID,
    theme LowCardinality(String),
    views UInt64 DEFAULT 0,
    successes UInt64 DEFAULT 0,
    last_update DateTime DEFAULT now()
)
    ENGINE = ReplacingMergeTree(last_update)
ORDER BY (user_uuid, theme);
