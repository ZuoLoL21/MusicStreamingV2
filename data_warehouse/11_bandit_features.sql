CREATE MATERIALIZED VIEW user_theme_stats
ENGINE = SummingMergeTree
ORDER BY (user_uuid, theme)
AS
SELECT
    e.user_uuid,
    t.theme,

    exponentialTimeDecayedCount(10)(e.event_time) AS decay_impressions,
    count() AS impressions,
    sum(e.listen_duration_seconds) AS total_listen_seconds,
    exponentialTimeDecayedAvg(10)(e.completion_ratio, e.event_time) AS decay_completion,
    avg(e.completion_ratio) AS avg_completion,

    sumIf(1, e.completion_ratio < 0.9 = 0) AS full_plays
FROM music_listen_events e
JOIN music_theme t
    ON e.music_uuid = t.music_uuid
GROUP BY
    e.user_uuid,
    t.theme;


CREATE MATERIALIZED VIEW user_theme_positive_events
ENGINE = SummingMergeTree
ORDER BY (user_uuid, theme)
AS
SELECT
    l.user_uuid,
    t.theme,
    exponentialTimeDecayedCount(10)(l.event_time) AS decay_likes,
    count() as likes

FROM music_like_events l
JOIN music_theme t
    ON l.music_uuid = t.music_uuid
GROUP BY
    l.user_uuid,
    t.theme;