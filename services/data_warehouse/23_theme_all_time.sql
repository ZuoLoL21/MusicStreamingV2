
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


CREATE VIEW theme_popularity AS
SELECT
    theme,
    sumMerge(decay_plays) AS decay_plays,
    sumMerge(decay_listen_seconds) AS decay_listen_seconds
FROM theme_popularity_inter
GROUP BY theme;


CREATE VIEW track_by_theme_popularity AS
SELECT
    music_uuid,
    theme,
    sumMerge(decay_plays) AS decay_plays,
    sumMerge(decay_listen_seconds) AS decay_listen_seconds
FROM track_by_theme_popularity_inter
GROUP BY music_uuid, theme;
