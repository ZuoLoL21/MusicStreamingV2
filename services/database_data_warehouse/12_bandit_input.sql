CREATE VIEW bandit_input_per_theme AS
WITH
    uts AS (
        SELECT
            user_uuid,
            theme,
            finalizeAggregation(decay_impressions)  AS decay_impressions,
            finalizeAggregation(impressions)        AS impressions,
            finalizeAggregation(decay_completion)   AS decay_completion_weighted_sum,
            finalizeAggregation(full_plays)         AS full_plays
        FROM user_theme_stats_inter
    ),
    utpe AS (
        SELECT
            user_uuid,
            theme,
            finalizeAggregation(decay_likes)        AS decay_likes
        FROM user_theme_positive_events_inter
    ),
    gus AS (
        SELECT
            user_uuid,
            finalizeAggregation(decay_impressions)  AS g_decay_impressions,
            finalizeAggregation(decay_completion)   AS g_decay_completion_weighted_sum
        FROM global_user_stats
    ),
    gupe AS (
        SELECT
            user_uuid,
            finalizeAggregation(decay_likes)        AS g_decay_likes
        FROM global_user_positive_event_stats
    ),
    gts AS (
        SELECT
            theme,
            finalizeAggregation(decay_impressions)  AS t_decay_impressions,
            finalizeAggregation(decay_completion)   AS t_decay_completion_weighted_sum
        FROM global_theme_stats
    ),
    gtpe AS (
        SELECT
            theme,
            finalizeAggregation(decay_likes)        AS t_decay_likes
        FROM global_theme_positive_event_stats
    )
SELECT
    uts.user_uuid AS user_uuid,
    uts.theme AS theme,

    -- User x Theme features
    uts.decay_impressions AS f_user_theme_decay_impressions,
    if(uts.decay_impressions > 0, uts.decay_completion_weighted_sum / uts.decay_impressions, 0.0) AS f_user_theme_decay_completion,
    if(uts.impressions > 0, uts.full_plays / uts.impressions, 0.0) AS f_user_theme_full_play_rate,
    if(uts.decay_impressions > 0, utpe.decay_likes / uts.decay_impressions, 0.0) AS f_user_theme_decay_like_rate,

    -- User global features
    gus.g_decay_impressions AS f_user_decay_impressions,
    if(gus.g_decay_impressions > 0, gus.g_decay_completion_weighted_sum / gus.g_decay_impressions, 0.0) AS f_user_decay_completion,
    if(gus.g_decay_impressions > 0, gupe.g_decay_likes / gus.g_decay_impressions, 0.0) AS f_user_decay_like_rate,

    -- Theme global features
    gts.t_decay_impressions AS f_theme_decay_impressions,
    if(gts.t_decay_impressions > 0, gts.t_decay_completion_weighted_sum / gts.t_decay_impressions, 0.0) AS f_theme_decay_completion,
    if(gts.t_decay_impressions > 0, gtpe.t_decay_likes / gts.t_decay_impressions, 0.0) AS f_theme_decay_like_rate,

    -- Relative preference: user x theme vs user global
    if(gus.g_decay_completion_weighted_sum > 0, uts.decay_completion_weighted_sum / gus.g_decay_completion_weighted_sum, 1.0) AS f_relative_completion,
    if(gus.g_decay_impressions > 0, uts.decay_impressions / gus.g_decay_impressions, 0.0) AS f_relative_exposure

FROM uts
LEFT JOIN utpe  USING (user_uuid, theme)
LEFT JOIN gus   USING (user_uuid)
LEFT JOIN gupe  USING (user_uuid)
LEFT JOIN gts   USING (theme)
LEFT JOIN gtpe  USING (theme);