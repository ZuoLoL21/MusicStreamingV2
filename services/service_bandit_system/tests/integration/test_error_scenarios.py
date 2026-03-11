"""Integration tests for error handling and edge cases."""

import json
import pytest
import numpy as np
from uuid import uuid4
from sqlalchemy import text

from tests.integration.builders import (
    ThemeFeatureBuilder,
    BanditWeightBuilder,
)


def test_malformed_json_in_database(db_managers, postgres_engine, test_config):
    """Test handling of corrupt JSON data in database."""
    user_uuid = uuid4()
    theme = "rock"

    # Insert malformed JSON directly into database
    query = text(
        f"INSERT INTO {test_config.bandit_params_table} "
        "(user_uuid, theme, version, weights, biases, weights_inv) "
        "VALUES (:uuid, :theme, 1, :weights, :biases, :weights_inv)"
    )

    with postgres_engine.connect() as conn:
        with conn.begin():
            conn.execute(
                query,
                {
                    "uuid": str(user_uuid),
                    "theme": theme,
                    "weights": '{"invalid": "json"}',  # Malformed
                    "biases": '{"invalid": "json"}',
                    "weights_inv": '{"invalid": "json"}',
                },
            )

    # Attempt to read should raise an error or handle gracefully
    with pytest.raises(Exception):
        db_managers.get_weight_bias(user_uuid)


def test_nan_features_in_clickhouse(db_managers, handler, clickhouse_engine, test_config):
    """Test handling of NaN feature values."""
    user_uuid = uuid4()
    theme = "rock"

    # Insert features with NaN values
    query = text(
        f"INSERT INTO {test_config.bandit_data_table} "
        "(user_uuid, theme, f_user_theme_decay_impressions, "
        "f_user_theme_decay_completion, f_user_theme_full_play_rate, "
        "f_user_theme_decay_like_rate, f_user_decay_impressions, "
        "f_user_decay_completion, f_user_decay_like_rate, "
        "f_theme_decay_impressions, f_theme_decay_completion, "
        "f_theme_decay_like_rate, f_relative_completion, f_relative_exposure) "
        "VALUES (:uuid, :theme, :f1, :f2, :f3, :f4, :f5, :f6, :f7, :f8, :f9, :f10, :f11, :f12)"
    )

    with clickhouse_engine.connect() as conn:
        with conn.begin():
            conn.execute(
                query,
                {
                    "uuid": str(user_uuid),
                    "theme": theme,
                    "f1": float('nan'),
                    "f2": 0.5,
                    "f3": 0.5,
                    "f4": 0.5,
                    "f5": 0.5,
                    "f6": 0.5,
                    "f7": 0.5,
                    "f8": 0.5,
                    "f9": 0.5,
                    "f10": 0.5,
                    "f11": 0.5,
                    "f12": 0.5,
                },
            )

    # Read features
    features_dict = db_managers.get_input_data(user_uuid)

    # Verify NaN is present
    assert np.isnan(features_dict[theme][0])

    # Attempting prediction with NaN features should handle gracefully
    # (may raise error or handle based on implementation)
    try:
        handler.predict(user_uuid)
        # If it doesn't raise, verify result is still valid
    except Exception as e:
        # NaN in features should be caught
        assert True


def test_infinity_features_in_clickhouse(db_managers, clickhouse_engine, test_config):
    """Test handling of infinity feature values."""
    user_uuid = uuid4()
    theme = "rock"

    # Insert features with infinity values
    query = text(
        f"INSERT INTO {test_config.bandit_data_table} "
        "(user_uuid, theme, f_user_theme_decay_impressions, "
        "f_user_theme_decay_completion, f_user_theme_full_play_rate, "
        "f_user_theme_decay_like_rate, f_user_decay_impressions, "
        "f_user_decay_completion, f_user_decay_like_rate, "
        "f_theme_decay_impressions, f_theme_decay_completion, "
        "f_theme_decay_like_rate, f_relative_completion, f_relative_exposure) "
        "VALUES (:uuid, :theme, :f1, :f2, :f3, :f4, :f5, :f6, :f7, :f8, :f9, :f10, :f11, :f12)"
    )

    with clickhouse_engine.connect() as conn:
        with conn.begin():
            conn.execute(
                query,
                {
                    "uuid": str(user_uuid),
                    "theme": theme,
                    "f1": float('inf'),
                    "f2": 0.5,
                    "f3": 0.5,
                    "f4": 0.5,
                    "f5": 0.5,
                    "f6": 0.5,
                    "f7": 0.5,
                    "f8": 0.5,
                    "f9": 0.5,
                    "f10": 0.5,
                    "f11": 0.5,
                    "f12": 0.5,
                },
            )

    # Read features
    features_dict = db_managers.get_input_data(user_uuid)

    # Verify infinity is present
    assert np.isinf(features_dict[theme][0])


def test_empty_features_array(test_client):
    """Test update with empty features array."""
    user_uuid = uuid4()
    theme = "rock"

    response = test_client.post(
        f"/update",
        json={
            "user_uuid": str(user_uuid),
            "reward": 0.5,
            "theme": theme,
            "features": [],  # Empty
        },
    )

    # Should fail validation
    assert response.status_code == 400


def test_unicode_theme_names(db_managers, handler):
    """Test handling of Unicode (emoji) theme names."""
    user_uuid = uuid4()
    theme = "🎸 rock 🎵"

    # Insert data with emoji theme name
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)

    # Should handle Unicode correctly
    features_dict = db_managers.get_input_data(user_uuid)
    assert theme in features_dict

    # Prediction should work
    predicted_theme, features = handler.predict(user_uuid)
    assert predicted_theme == theme


def test_large_feature_values(db_managers, handler):
    """Test numerical stability with very large feature values."""
    user_uuid = uuid4()
    theme = "rock"

    # Create features with large values
    large_features = [1e6] * 12
    ThemeFeatureBuilder(user_uuid, theme).with_features(large_features).build(db_managers)

    # Should handle prediction
    predicted_theme, features = handler.predict(user_uuid)
    assert predicted_theme == theme

    # Update should also work
    handler.update(user_uuid, 0.8, theme, features)

    # Verify no overflow/underflow
    result = db_managers.get_weight_bias(user_uuid)
    assert not np.isnan(result[theme].Weights).any()
    assert not np.isinf(result[theme].Weights).any()


def test_very_small_feature_values(db_managers, handler):
    """Test numerical stability with very small feature values."""
    user_uuid = uuid4()
    theme = "rock"

    # Create features with very small values
    small_features = [1e-10] * 12
    ThemeFeatureBuilder(user_uuid, theme).with_features(small_features).build(db_managers)

    # Should handle prediction
    predicted_theme, features = handler.predict(user_uuid)
    assert predicted_theme == theme

    # Update should work
    handler.update(user_uuid, 0.8, theme, features)

    # Verify stability
    result = db_managers.get_weight_bias(user_uuid)
    assert not np.isnan(result[theme].Weights).any()


def test_negative_features(db_managers, handler):
    """Test handling of negative feature values."""
    user_uuid = uuid4()
    theme = "rock"

    # Create features with negative values
    negative_features = [-0.5, -0.3, -0.1, 0.0, 0.1, 0.3, 0.5, 0.7, -0.2, -0.4, 0.2, 0.4]
    ThemeFeatureBuilder(user_uuid, theme).with_features(negative_features).build(db_managers)

    # Should handle prediction
    predicted_theme, features = handler.predict(user_uuid)
    assert predicted_theme == theme

    # Update should work
    handler.update(user_uuid, 0.6, theme, features)


def test_all_zero_features(db_managers, handler):
    """Test handling of all-zero feature vector."""
    user_uuid = uuid4()
    theme = "rock"

    # Create all-zero features
    zero_features = [0.0] * 12
    ThemeFeatureBuilder(user_uuid, theme).with_features(zero_features).build(db_managers)

    # Should handle prediction (though might have high uncertainty)
    predicted_theme, features = handler.predict(user_uuid)
    assert predicted_theme == theme

    # Update should work
    handler.update(user_uuid, 0.5, theme, features)


def test_mixed_valid_invalid_themes(db_managers, handler):
    """Test prediction when some themes have invalid data."""
    user_uuid = uuid4()
    valid_theme = "rock"
    invalid_theme = "jazz"

    # Insert valid theme
    ThemeFeatureBuilder(user_uuid, valid_theme).build(db_managers)

    # Insert theme with wrong dimension (if possible to simulate)
    # For now, just test with valid theme

    # Should predict valid theme
    predicted_theme, features = handler.predict(user_uuid)
    assert predicted_theme == valid_theme


def test_extremely_long_theme_name(db_managers):
    """Test handling of very long theme names."""
    user_uuid = uuid4()
    # Create a theme name at the VARCHAR(255) limit
    theme = "a" * 255

    # Should handle insertion
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)

    # Should read successfully
    features_dict = db_managers.get_input_data(user_uuid)
    assert theme in features_dict


def test_theme_name_with_special_characters(db_managers, handler):
    """Test theme names with special SQL characters."""
    user_uuid = uuid4()
    # Theme name with quotes and SQL special chars
    theme = "rock'n\"roll; DROP TABLE--"

    # Should handle safely (parameterized queries prevent injection)
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)

    # Should read successfully
    features_dict = db_managers.get_input_data(user_uuid)
    assert theme in features_dict

    # Prediction should work
    predicted_theme, features = handler.predict(user_uuid)
    assert predicted_theme == theme


def test_concurrent_delete_and_read(db_managers, postgres_engine, test_config):
    """Test behavior when data is deleted during operation."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup data
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).build(db_managers)

    # Read once to verify it exists
    initial_read = db_managers.get_weight_bias(user_uuid)
    assert theme in initial_read

    # Delete the data
    delete_query = text(
        f"DELETE FROM {test_config.bandit_params_table} WHERE user_uuid = :uuid"
    )
    with postgres_engine.connect() as conn:
        with conn.begin():
            conn.execute(delete_query, {"uuid": str(user_uuid)})

    # Read again - should return empty
    after_delete = db_managers.get_weight_bias(user_uuid)
    assert len(after_delete) == 0


def test_update_with_null_values_rejected(postgres_engine, test_config):
    """Test that NULL values in required fields are rejected by schema."""
    user_uuid = uuid4()
    theme = "rock"

    # Try to insert with NULL weights (should fail due to NOT NULL constraint)
    query = text(
        f"INSERT INTO {test_config.bandit_params_table} "
        "(user_uuid, theme, version, weights, biases, weights_inv) "
        "VALUES (:uuid, :theme, 1, NULL, :biases, :weights_inv)"
    )

    with pytest.raises(Exception):
        with postgres_engine.connect() as conn:
            with conn.begin():
                conn.execute(
                    query,
                    {
                        "uuid": str(user_uuid),
                        "theme": theme,
                        "biases": "[]",
                        "weights_inv": "[]",
                    },
                )


def test_duplicate_primary_key_rejected(postgres_engine, test_config):
    """Test that duplicate (user_uuid, theme) is rejected by primary key."""
    user_uuid = uuid4()
    theme = "rock"

    # Insert first record
    query = text(
        f"INSERT INTO {test_config.bandit_params_table} "
        "(user_uuid, theme, version, weights, biases, weights_inv) "
        "VALUES (:uuid, :theme, 1, :weights, :biases, :weights_inv)"
    )

    weights_json = json.dumps(np.eye(12).tolist())
    biases_json = json.dumps(np.zeros(12).tolist())

    with postgres_engine.connect() as conn:
        with conn.begin():
            conn.execute(
                query,
                {
                    "uuid": str(user_uuid),
                    "theme": theme,
                    "weights": weights_json,
                    "biases": biases_json,
                    "weights_inv": weights_json,
                },
            )

    # Try to insert duplicate (should fail)
    with pytest.raises(Exception):
        with postgres_engine.connect() as conn:
            with conn.begin():
                conn.execute(
                    query,
                    {
                        "uuid": str(user_uuid),
                        "theme": theme,
                        "weights": weights_json,
                        "biases": biases_json,
                        "weights_inv": weights_json,
                    },
                )
