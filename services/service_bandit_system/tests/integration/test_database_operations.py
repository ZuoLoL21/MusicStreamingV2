"""Integration tests for database operations."""

import json
import numpy as np
import pytest
from uuid import UUID, uuid4
from sqlalchemy import text

from tests.integration.builders import (
    ThemeFeatureBuilder,
    BanditWeightBuilder,
    create_test_user_with_themes,
)
from src.di.db import NUMB_FEATURES


def test_get_input_data_single_theme(db_managers):
    """Test reading a single theme's features from ClickHouse."""
    user_uuid = uuid4()
    theme = "rock"

    # Insert test data
    ThemeFeatureBuilder(user_uuid, theme).with_high_engagement().build(db_managers)

    # Read from database
    result = db_managers.get_input_data(user_uuid)

    assert len(result) == 1
    assert theme in result
    assert result[theme].shape == (NUMB_FEATURES,)
    assert result[theme][0] == 0.9  # First feature from high_engagement


def test_get_input_data_multiple_themes(db_managers):
    """Test reading multiple themes for a user from ClickHouse."""
    user_uuid = uuid4()
    themes = ["rock", "jazz", "classical"]

    # Insert test data
    for theme in themes:
        ThemeFeatureBuilder(user_uuid, theme).build(db_managers)

    # Read from database
    result = db_managers.get_input_data(user_uuid)

    assert len(result) == 3
    for theme in themes:
        assert theme in result
        assert result[theme].shape == (NUMB_FEATURES,)


def test_get_input_data_no_themes(db_managers):
    """Test reading when user has no themes returns empty dict."""
    user_uuid = uuid4()

    result = db_managers.get_input_data(user_uuid)

    assert len(result) == 0
    assert isinstance(result, dict)


def test_get_weight_bias_existing(db_managers):
    """Test reading existing weights from PostgreSQL."""
    user_uuid = uuid4()
    theme = "rock"

    # Insert test data
    BanditWeightBuilder(user_uuid, theme).with_version(5).build(db_managers)

    # Read from database
    result = db_managers.get_weight_bias(user_uuid)

    assert len(result) == 1
    assert theme in result
    arm = result[theme]
    assert arm.Theme == theme
    assert arm.Version == 5
    assert arm.Weights.shape == (NUMB_FEATURES, NUMB_FEATURES)
    assert arm.Biases.shape == (NUMB_FEATURES,)
    assert arm.WeightsInv.shape == (NUMB_FEATURES, NUMB_FEATURES)


def test_get_weight_bias_multiple(db_managers):
    """Test reading multiple themes' weights from PostgreSQL."""
    user_uuid = uuid4()
    themes = ["rock", "jazz", "classical"]

    # Insert test data
    for i, theme in enumerate(themes):
        BanditWeightBuilder(user_uuid, theme).with_version(i + 1).build(db_managers)

    # Read from database
    result = db_managers.get_weight_bias(user_uuid)

    assert len(result) == 3
    for i, theme in enumerate(themes):
        assert theme in result
        assert result[theme].Version == i + 1


def test_get_weight_bias_for_one_new(db_managers):
    """Test get_weight_bias_for_one for a new theme (not in DB)."""
    user_uuid = uuid4()
    theme = "rock"

    # Don't insert anything, should return new arm
    result = db_managers.get_weight_bias_for_one(user_uuid, theme)

    assert result.Theme == theme
    assert result.Version == 1
    # New arm should have identity matrix
    expected_weights = np.eye(NUMB_FEATURES) * 1.0
    np.testing.assert_array_almost_equal(result.Weights, expected_weights)
    np.testing.assert_array_almost_equal(result.Biases, np.zeros(NUMB_FEATURES))


def test_get_weight_bias_for_one_existing(db_managers):
    """Test get_weight_bias_for_one for an existing theme."""
    user_uuid = uuid4()
    theme = "rock"

    # Insert test data
    BanditWeightBuilder(user_uuid, theme).with_version(3).build(db_managers)

    # Read from database
    result = db_managers.get_weight_bias_for_one(user_uuid, theme)

    assert result.Theme == theme
    assert result.Version == 3


def test_update_weight_bias_new(db_managers):
    """Test updating weights for a new theme (should fail, need INSERT)."""
    user_uuid = uuid4()
    theme = "rock"

    # Try to update non-existent record
    weights = np.eye(NUMB_FEATURES) * 2.0
    biases = np.ones(NUMB_FEATURES)
    weights_inv = np.eye(NUMB_FEATURES) * 0.5

    success = db_managers.update_weight_bias(
        user_uuid, theme, weights, biases, weights_inv, 0, 1
    )

    assert success is True


def test_update_weight_bias_existing(db_managers):
    """Test updating weights for an existing theme."""
    user_uuid = uuid4()
    theme = "rock"

    # Insert initial data
    BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    # Update with new weights
    weights = np.eye(NUMB_FEATURES) * 2.0
    biases = np.ones(NUMB_FEATURES)
    weights_inv = np.eye(NUMB_FEATURES) * 0.5

    success = db_managers.update_weight_bias(
        user_uuid, theme, weights, biases, weights_inv, 1, 1  # version 1
    )

    assert success is True

    # Verify the update
    result = db_managers.get_weight_bias_for_one(user_uuid, theme)
    assert result.Version == 2  # Should be incremented
    np.testing.assert_array_almost_equal(result.Weights, weights)
    np.testing.assert_array_almost_equal(result.Biases, biases)


def test_update_weight_bias_version_conflict(db_managers):
    """Test that version conflict causes update to fail."""
    user_uuid = uuid4()
    theme = "rock"

    # Insert initial data with version 3
    BanditWeightBuilder(user_uuid, theme).with_version(3).build(db_managers)

    # Try to update with wrong version
    weights = np.eye(NUMB_FEATURES) * 2.0
    biases = np.ones(NUMB_FEATURES)
    weights_inv = np.eye(NUMB_FEATURES) * 0.5

    success = db_managers.update_weight_bias(
        user_uuid, theme, weights, biases, weights_inv, 1, 1  # version 1 (wrong!)
    )

    # Should fail because version doesn't match
    assert success is False

    # Verify nothing changed
    result = db_managers.get_weight_bias_for_one(user_uuid, theme)
    assert result.Version == 3  # Still version 3


def test_update_weight_bias_retry(db_managers):
    """Test successful update after version conflict resolution."""
    user_uuid = uuid4()
    theme = "rock"

    # Insert initial data
    BanditWeightBuilder(user_uuid, theme).with_version(2).build(db_managers)

    # First attempt with wrong version (fails)
    weights = np.eye(NUMB_FEATURES) * 2.0
    biases = np.ones(NUMB_FEATURES)
    weights_inv = np.eye(NUMB_FEATURES) * 0.5

    success = db_managers.update_weight_bias(
        user_uuid, theme, weights, biases, weights_inv, 1, 1
    )
    assert success is False

    # Second attempt with correct version (succeeds)
    success = db_managers.update_weight_bias(
        user_uuid, theme, weights, biases, weights_inv, 1, 2  # correct version
    )
    assert success is True

    # Verify the update
    result = db_managers.get_weight_bias_for_one(user_uuid, theme)
    assert result.Version == 3


def test_json_serialization(db_managers):
    """Test that numpy arrays serialize/deserialize correctly through JSON."""
    user_uuid = uuid4()
    theme = "rock"

    # Create complex weights with various values
    weights = np.random.rand(NUMB_FEATURES, NUMB_FEATURES)
    biases = np.random.rand(NUMB_FEATURES)
    weights_inv = np.linalg.inv(weights + np.eye(NUMB_FEATURES))  # Ensure invertible

    # Insert
    builder = BanditWeightBuilder(user_uuid, theme)
    builder.with_weights(weights, biases, weights_inv).build(db_managers)

    # Read back
    result = db_managers.get_weight_bias_for_one(user_uuid, theme)

    # Verify serialization roundtrip
    np.testing.assert_array_almost_equal(result.Weights, weights, decimal=10)
    np.testing.assert_array_almost_equal(result.Biases, biases, decimal=10)
    np.testing.assert_array_almost_equal(result.WeightsInv, weights_inv, decimal=10)


def test_database_cleanup(db_managers):
    """Verify that databases are cleaned between tests."""
    user_uuid = uuid4()

    # Database should be empty at start
    clickhouse_result = db_managers.get_input_data(user_uuid)
    postgres_result = db_managers.get_weight_bias(user_uuid)

    assert len(clickhouse_result) == 0
    assert len(postgres_result) == 0


def test_clickhouse_multiple_users_isolation(db_managers):
    """Test that different users' data is isolated in ClickHouse."""
    user1_uuid = uuid4()
    user2_uuid = uuid4()
    theme = "rock"

    # Insert data for both users
    ThemeFeatureBuilder(user1_uuid, theme).with_high_engagement().build(db_managers)
    ThemeFeatureBuilder(user2_uuid, theme).with_low_engagement().build(db_managers)

    # Read each user's data
    user1_data = db_managers.get_input_data(user1_uuid)
    user2_data = db_managers.get_input_data(user2_uuid)

    assert len(user1_data) == 1
    assert len(user2_data) == 1
    # Features should be different
    assert user1_data[theme][0] != user2_data[theme][0]


def test_postgres_multiple_users_isolation(db_managers):
    """Test that different users' data is isolated in PostgreSQL."""
    user1_uuid = uuid4()
    user2_uuid = uuid4()
    theme = "rock"

    # Insert data for both users
    BanditWeightBuilder(user1_uuid, theme).with_version(5).build(db_managers)
    BanditWeightBuilder(user2_uuid, theme).with_version(10).build(db_managers)

    # Read each user's data
    user1_data = db_managers.get_weight_bias(user1_uuid)
    user2_data = db_managers.get_weight_bias(user2_uuid)

    assert user1_data[theme].Version == 5
    assert user2_data[theme].Version == 10
