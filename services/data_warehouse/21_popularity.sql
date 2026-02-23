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


CREATE VIEW track_popularity AS
SELECT
    music_uuid,
    sumMerge(decay_plays) AS decay_plays,
    sumMerge(decay_listen_seconds) AS decay_listen_seconds
FROM track_popularity_inter
GROUP BY music_uuid;


CREATE VIEW artist_popularity AS
SELECT
    artist_uuid,
    sumMerge(decay_plays) AS decay_plays,
    sumMerge(decay_listen_seconds) AS decay_listen_seconds
FROM artist_popularity_inter
GROUP BY artist_uuid;