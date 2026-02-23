
CREATE MATERIALIZED VIEW track_popularity_daily
ENGINE = SummingMergeTree
ORDER BY (music_uuid, for_day)
AS
SELECT
    music_uuid,
    toDate(event_time) AS for_day,
    count() AS plays,
    sum(listen_duration_seconds) AS listen_seconds
FROM music_listen_events
GROUP BY
    music_uuid,
    for_day;


CREATE MATERIALIZED VIEW artist_popularity_daily
ENGINE = SummingMergeTree
ORDER BY (artist_uuid, for_day)
AS
SELECT
    artist_uuid,
    toDate(event_time) AS for_day,
    count() AS plays,
    sum(listen_duration_seconds) AS listen_seconds
FROM music_listen_events
GROUP BY
    artist_uuid,
    for_day;


CREATE MATERIALIZED VIEW theme_popularity_daily
ENGINE = SummingMergeTree
ORDER BY (theme, for_day)
AS
SELECT
    mt.theme,
    toDate(mle.event_time) AS for_day,
    count() AS plays,
    sum(mle.listen_duration_seconds) AS listen_seconds
FROM music_listen_events mle
         INNER JOIN music_theme mt ON mle.music_uuid = mt.music_uuid
GROUP BY mt.theme, for_day;


CREATE MATERIALIZED VIEW track_by_theme_popularity_daily
ENGINE = SummingMergeTree
ORDER BY (music_uuid, theme, for_day)
AS
SELECT
    mle.music_uuid,
    mt.theme,
    toDate(mle.event_time) AS for_day,
    count() AS plays,
    sum(mle.listen_duration_seconds) AS listen_seconds
FROM music_listen_events mle
         INNER JOIN music_theme mt ON mle.music_uuid = mt.music_uuid
GROUP BY mle.music_uuid, mt.theme, for_day;

