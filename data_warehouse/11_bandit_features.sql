CREATE MATERIALIZED VIEW user_theme_stats_inter
ENGINE = AggregatingMergeTree
ORDER BY (user_uuid, theme)
AS
SELECT
    e.user_uuid,
    t.theme,

    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(e.event_time)) / 2592000)) AS decay_impressions,
    countState() AS impressions,
    sumState(e.listen_duration_seconds) AS total_listen_seconds,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(e.event_time)) / 2592000) * e.completion_ratio) AS decay_completion,
    avgState(e.completion_ratio) AS avg_completion,

    sumIfState(1, e.completion_ratio >= 0.9) AS full_plays
FROM music_listen_events e
JOIN music_theme t
    ON e.music_uuid = t.music_uuid
GROUP BY
    e.user_uuid,
    t.theme;


CREATE MATERIALIZED VIEW user_theme_positive_events_inter
ENGINE = AggregatingMergeTree
ORDER BY (user_uuid, theme)
AS
SELECT
    l.user_uuid,
    t.theme,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(l.event_time)) / 2592000)) AS decay_likes,
    countState() AS likes
FROM music_like_events l
JOIN music_theme t
    ON l.music_uuid = t.music_uuid
GROUP BY
    l.user_uuid,
    t.theme;


CREATE MATERIALIZED VIEW global_user_stats
ENGINE = AggregatingMergeTree
ORDER BY user_uuid
AS
SELECT
    user_uuid,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(event_time)) / 2592000)) AS decay_impressions,
    countState() AS impressions,
    sumState(listen_duration_seconds) AS total_listen_seconds,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(event_time)) / 2592000) * completion_ratio) AS decay_completion,
    avgState(completion_ratio) AS avg_completion
FROM music_listen_events
GROUP BY
    user_uuid;


CREATE MATERIALIZED VIEW global_user_positive_event_stats
ENGINE = AggregatingMergeTree
ORDER BY user_uuid
AS
SELECT
    user_uuid,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(event_time)) / 2592000)) AS decay_likes,
    countState() AS likes
FROM music_like_events
GROUP BY
    user_uuid;

CREATE MATERIALIZED VIEW global_theme_stats
ENGINE = AggregatingMergeTree
ORDER BY theme
AS
SELECT
    t.theme,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(e.event_time)) / 2592000)) AS decay_impressions,
    countState() AS impressions,
    sumState(e.listen_duration_seconds) AS total_listen_seconds,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(e.event_time)) / 2592000) * e.completion_ratio) AS decay_completion,
    avgState(e.completion_ratio) AS avg_completion
FROM music_listen_events e
JOIN music_theme t
    ON e.music_uuid = t.music_uuid
GROUP BY
    t.theme;


CREATE MATERIALIZED VIEW global_theme_positive_event_stats
ENGINE = AggregatingMergeTree
ORDER BY theme
AS
SELECT
    t.theme,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(l.event_time)) / 2592000)) AS decay_likes,
    countState() AS likes
FROM music_like_events l
JOIN music_theme t
    ON l.music_uuid = t.music_uuid
GROUP BY
    t.theme;
