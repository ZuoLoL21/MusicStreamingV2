-- ClickHouse initialization for integration tests
-- Simple tables (not views) for testing bandit feature inputs
--
-- IMPORTANT: These schemas must match the production schemas in:
--   ../../../database_data_warehouse/02_features.sql (music_theme)
--   ../../../database_data_warehouse/12_bandit_input.sql (bandit_input_per_theme)
--
-- If you modify the production schemas, update these accordingly.

-- Music theme catalog (from 02_features.sql)
CREATE TABLE IF NOT EXISTS music_theme (
    music_uuid UUID,
    theme LowCardinality(String),
    views UInt64 DEFAULT 0,
    successes UInt64 DEFAULT 0,
    last_update DateTime64(3, 'UTC') DEFAULT now64(3)
) ENGINE = ReplacingMergeTree(last_update)
ORDER BY (music_uuid, theme);

-- Bandit feature inputs (from 12_bandit_input.sql)
CREATE TABLE IF NOT EXISTS bandit_input_per_theme (
    user_uuid UUID,
    theme String,

    -- User x Theme features
    f_user_theme_decay_impressions Float64,
    f_user_theme_decay_completion Float64,
    f_user_theme_full_play_rate Float64,
    f_user_theme_decay_like_rate Float64,

    -- User global features
    f_user_decay_impressions Float64,
    f_user_decay_completion Float64,
    f_user_decay_like_rate Float64,

    -- Theme global features
    f_theme_decay_impressions Float64,
    f_theme_decay_completion Float64,
    f_theme_decay_like_rate Float64,

    -- Relative features
    f_relative_completion Float64,
    f_relative_exposure Float64
) ENGINE = MergeTree()
ORDER BY (user_uuid, theme);
