CREATE MATERIALIZED VIEW user_theme_stats_inter
ENGINE = AggregatingMergeTree
ORDER BY (user_uuid, theme)
AS
SELECT
    e.user_uuid,
    t.theme,

    exponentialTimeDecayedCountState(2592000)(e.event_time) AS decay_impressions,
    countState() AS impressions,
    sumState(e.listen_duration_seconds) AS total_listen_seconds,
    exponentialTimeDecayedAvgState(2592000)(e.completion_ratio, e.event_time) AS decay_completion,
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
    exponentialTimeDecayedCountState(2592000)(l.event_time) AS decay_likes,
    countState() AS likes
FROM music_like_events l
JOIN music_theme t
    ON l.music_uuid = t.music_uuid
GROUP BY
    l.user_uuid,
    t.theme;


CREATE VIEW user_theme_stats AS
SELECT
    user_uuid,
    theme,
    exponentialTimeDecayedCountMerge(2592000)(decay_impressions) AS decay_impressions,
    countMerge(impressions) AS impressions,
    sumMerge(total_listen_seconds) AS total_listen_seconds,
    exponentialTimeDecayedAvgMerge(2592000)(decay_completion) AS decay_completion,
    avgMerge(avg_completion) AS avg_completion,
    sumIfMerge(full_plays) AS full_plays
FROM user_theme_stats_inter
GROUP BY user_uuid, theme;


CREATE VIEW user_theme_positive_events AS
SELECT
    user_uuid,
    theme,
    exponentialTimeDecayedCountMerge(2592000)(decay_likes) AS decay_likes,
    countMerge(likes) AS likes
FROM user_theme_positive_events_inter
GROUP BY user_uuid, theme;
