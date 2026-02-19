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
    sum(listen_durection_seconds) AS listen_seconds
FROM music_listen_events
GROUP BY
    artist_uuid,
    for_day;


CREATE MATERIALIZED VIEW track_popularity
ENGINE = SummingMergeTree
ORDER BY (music_uuid)
AS
SELECT
    music_uuid,
    exponentialTimeDecayedSum(30)(for_day, plays),
    exponentialTimeDecayedSum(30)(for_day, listen_seconds)
FROM track_popularity_daily
GROUP BY
    music_uuid;


CREATE MATERIALIZED VIEW artist_popularity
ENGINE = SummingMergeTree
ORDER BY (artist_uuid)
AS
SELECT
    artist_uuid,
    exponentialTimeDecayedSum(30)(for_day, plays),
    exponentialTimeDecayedSum(30)(for_day, listen_seconds)
FROM artist_popularity_daily
GROUP BY
    artist_uuid;