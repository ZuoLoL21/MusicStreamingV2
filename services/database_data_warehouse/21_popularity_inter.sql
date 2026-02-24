CREATE MATERIALIZED VIEW track_popularity_inter
ENGINE = SummingMergeTree
ORDER BY (music_uuid)
AS
SELECT
    music_uuid,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(event_time)) / 2592000)) AS decay_plays,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(event_time)) / 2592000) * listen_duration_seconds) AS decay_listen_seconds
FROM music_listen_events
GROUP BY
    music_uuid;


CREATE MATERIALIZED VIEW artist_popularity_inter
ENGINE = SummingMergeTree
ORDER BY (artist_uuid)
AS
SELECT
    artist_uuid,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(event_time)) / 2592000)) AS decay_plays,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(event_time)) / 2592000) * listen_duration_seconds) AS decay_listen_seconds
FROM music_listen_events
GROUP BY
    artist_uuid;


CREATE MATERIALIZED VIEW theme_popularity_inter
ENGINE = SummingMergeTree
ORDER BY (theme)
AS
SELECT
    mt.theme,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(mle.event_time)) / 2592000)) AS decay_plays,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(mle.event_time)) / 2592000) * mle.listen_duration_seconds) AS decay_listen_seconds
FROM music_listen_events mle
    INNER JOIN music_theme mt ON mle.music_uuid = mt.music_uuid
GROUP BY mt.theme;


CREATE MATERIALIZED VIEW track_by_theme_popularity_inter
ENGINE = SummingMergeTree
ORDER BY (music_uuid, theme)
AS
SELECT
    mle.music_uuid,
    mt.theme,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(mle.event_time)) / 2592000)) AS decay_plays,
    sumState(exp(-0.693147 * (toUnixTimestamp(now()) - toUnixTimestamp(mle.event_time)) / 2592000) * mle.listen_duration_seconds) AS decay_listen_seconds
FROM music_listen_events mle
    INNER JOIN music_theme mt ON mle.music_uuid = mt.music_uuid
GROUP BY mle.music_uuid, mt.theme;
