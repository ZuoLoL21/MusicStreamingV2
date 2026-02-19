
CREATE TABLE bandit_choices
(
    event_time DateTime64(3, 'UTC') DEFAULT now64(3),
    user_uuid UUID,
    action UInt32,
    reward Float32
)
    ENGINE = MergeTree
PARTITION BY toYYYYMM(event_time)
ORDER BY (event_time, user_uuid);

