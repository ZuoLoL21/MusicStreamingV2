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
