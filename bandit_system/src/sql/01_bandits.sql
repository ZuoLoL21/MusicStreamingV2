CREATE TABLE bandit_data
(
    user_uuid UUID PRIMARY KEY,
    theme LowCardinality(String),
    weights UInt64 DEFAULT 0,
    biases UInt64 DEFAULT 0,
    last_update DateTime DEFAULT now()
)
ORDER BY (user_uuid, theme);
