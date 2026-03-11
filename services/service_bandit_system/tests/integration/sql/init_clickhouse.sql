-- ClickHouse initialization for integration tests
-- Simple table (not view) for testing bandit feature inputs
--
-- IMPORTANT: This schema must match the columns in:
--   ../../../database_data_warehouse/12_bandit_input.sql
--
-- If you modify the production view, update these columns accordingly.
-- The production view defines the contract; this table implements it for testing.

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
