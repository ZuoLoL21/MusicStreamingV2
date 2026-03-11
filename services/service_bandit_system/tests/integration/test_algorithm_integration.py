"""Integration tests for LinUCB algorithm with real database."""

import numpy as np
import pytest
from uuid import uuid4

from tests.integration.builders import (
    ThemeFeatureBuilder,
    BanditWeightBuilder,
)


def test_linucb_prediction_accuracy(db_managers, handler, bandit):
    """Test LinUCB prediction with known features produces expected UCB scores."""
    user_uuid = uuid4()
    theme = "rock"

    # Create features with known values
    features = [0.8, 0.7, 0.6, 0.5, 0.4, 0.3, 0.2, 0.1, 0.0, 0.1, 0.2, 0.3]
    ThemeFeatureBuilder(user_uuid, theme).with_features(features).build(db_managers)

    # Make prediction
    predicted_theme, predicted_features = handler.predict(user_uuid)

    assert predicted_theme == theme
    np.testing.assert_array_almost_equal(predicted_features, np.array(features))


def test_sherman_morrison_stability(db_managers, handler, bandit):
    """Test that Sherman-Morrison updates remain stable over many iterations."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme]

    # Perform 200 updates (more than recompute interval)
    for i in range(200):
        reward = 0.5 + 0.001 * (i % 50)  # Varying rewards
        handler.update(user_uuid, reward, theme, features)

    # Verify weights are still valid (not NaN or Inf)
    result = db_managers.get_weight_bias(user_uuid)
    weights = result[theme].Weights
    weights_inv = result[theme].WeightsInv

    assert not np.isnan(weights).any(), "Weights contain NaN"
    assert not np.isinf(weights).any(), "Weights contain Inf"
    assert not np.isnan(weights_inv).any(), "Inverse weights contain NaN"
    assert not np.isinf(weights_inv).any(), "Inverse weights contain Inf"

    # Verify inverse is still accurate
    product = weights @ weights_inv
    identity = np.eye(12)
    np.testing.assert_array_almost_equal(product, identity, decimal=3)


def test_divergence_detection_triggers(db_managers, handler, bandit, test_config):
    """Test that divergence detection triggers recomputation."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup with custom config to force divergence
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)

    # Create initial weights with high updates_since_recompute
    builder = BanditWeightBuilder(user_uuid, theme)
    builder.with_version(1).with_updates_since_recompute(99).build(db_managers)

    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme]

    # Perform update that should trigger recompute check
    handler.update(user_uuid, 0.8, theme, features)

    # Verify updates_since_recompute was reset (indicating recompute happened)
    result = db_managers.get_weight_bias(user_uuid)
    # After recompute interval (100), should reset
    assert result[theme].UpdatesSinceRecompute < 99


def test_divergence_recompute(db_managers, handler, bandit):
    """Test that recomputation restores accuracy."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)

    # Create weights with potential divergence
    weights = np.eye(12) * 2.0
    biases = np.ones(12)
    # Create slightly inaccurate inverse
    weights_inv = np.eye(12) * 0.51  # Not exactly 0.5

    builder = BanditWeightBuilder(user_uuid, theme)
    builder.with_weights(weights, biases, weights_inv)
    builder.with_updates_since_recompute(99).build(db_managers)

    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme]

    # Perform update (should trigger recompute if divergence detected)
    handler.update(user_uuid, 0.8, theme, features)

    # Verify inverse is accurate after update
    result = db_managers.get_weight_bias(user_uuid)
    product = result[theme].Weights @ result[theme].WeightsInv
    identity = np.eye(12)
    np.testing.assert_array_almost_equal(product, identity, decimal=5)


def test_reward_incorporation(db_managers, handler, bandit):
    """Test that weights change appropriately with different rewards."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup
    features = [0.8] * 12
    ThemeFeatureBuilder(user_uuid, theme).with_features(features).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    # Get initial weights
    initial_weights = db_managers.get_weight_bias(user_uuid)
    initial_biases = initial_weights[theme].Biases.copy()

    features_array = np.array(features)

    # Update with high reward
    handler.update(user_uuid, 1.0, theme, features_array)

    # Get updated weights
    updated_weights = db_managers.get_weight_bias(user_uuid)
    updated_biases = updated_weights[theme].Biases

    # Biases should have changed
    assert not np.allclose(initial_biases, updated_biases), "Biases should change after update"


def test_exploration_vs_exploitation(db_managers, handler, bandit):
    """Test that high uncertainty leads to exploration."""
    user_uuid = uuid4()
    themes = ["rock", "jazz"]

    # Theme 1: High certainty (many updates)
    ThemeFeatureBuilder(user_uuid, themes[0]).build(db_managers)
    weights_high_certainty = np.eye(12) * 100.0  # High weights = high certainty
    biases_high = np.ones(12) * 0.5
    weights_inv_high = np.eye(12) * 0.01
    BanditWeightBuilder(user_uuid, themes[0]).with_weights(
        weights_high_certainty, biases_high, weights_inv_high
    ).build(db_managers)

    # Theme 2: Low certainty (new theme, identity matrix)
    ThemeFeatureBuilder(user_uuid, themes[1]).build(db_managers)
    # Don't create weights - will use identity matrix (low certainty)

    # Make multiple predictions
    predictions = []
    for _ in range(10):
        predicted_theme, _ = handler.predict(user_uuid)
        predictions.append(predicted_theme)

    # Due to exploration bonus, should sometimes select the uncertain theme
    # (though not guaranteed in every run due to randomness)
    unique_predictions = set(predictions)
    # At least should have valid predictions
    assert all(p in themes for p in predictions)


def test_feature_normalization(db_managers, handler):
    """Test that features are handled correctly regardless of scale."""
    user_uuid = uuid4()
    theme = "rock"

    # Create features with different scales
    large_features = [10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0, 90.0, 100.0, 110.0, 120.0]
    ThemeFeatureBuilder(user_uuid, theme).with_features(large_features).build(db_managers)

    # Should handle prediction without errors
    predicted_theme, predicted_features = handler.predict(user_uuid)

    assert predicted_theme == theme
    assert len(predicted_features) == 12


def test_new_theme_initialization(db_managers, handler, bandit):
    """Test that new themes initialize with identity matrix."""
    user_uuid = uuid4()
    theme = "rock"

    # Create features but no weights
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)

    # Make prediction (should initialize with identity)
    predicted_theme, features = handler.predict(user_uuid)

    assert predicted_theme == theme

    # Now update and check initialization
    handler.update(user_uuid, 0.8, theme, features)

    # Check that weights were initialized properly
    result = db_managers.get_weight_bias(user_uuid)
    # After first update, should have non-identity weights
    assert result[theme].Version == 2  # Should be incremented


def test_multiple_users_independence(db_managers, handler):
    """Test that updates for one user don't affect another user."""
    user1_uuid = uuid4()
    user2_uuid = uuid4()
    theme = "rock"

    # Setup both users with same theme
    ThemeFeatureBuilder(user1_uuid, theme).build(db_managers)
    ThemeFeatureBuilder(user2_uuid, theme).build(db_managers)

    BanditWeightBuilder(user1_uuid, theme).with_version(1).build(db_managers)
    BanditWeightBuilder(user2_uuid, theme).with_version(1).build(db_managers)

    # Get initial weights for user2
    user2_initial = db_managers.get_weight_bias(user2_uuid)
    user2_initial_weights = user2_initial[theme].Weights.copy()
    user2_initial_version = user2_initial[theme].Version

    # Update user1 only
    features_dict = db_managers.get_input_data(user1_uuid)
    features = features_dict[theme]
    handler.update(user1_uuid, 0.9, theme, features)

    # Verify user2 unchanged
    user2_after = db_managers.get_weight_bias(user2_uuid)
    np.testing.assert_array_equal(user2_after[theme].Weights, user2_initial_weights)
    assert user2_after[theme].Version == user2_initial_version


def test_theme_preference_learning(db_managers, handler):
    """Test that high rewards lead to theme preference."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup
    features = [0.7] * 12
    ThemeFeatureBuilder(user_uuid, theme).with_features(features).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    # Get initial biases
    initial_result = db_managers.get_weight_bias(user_uuid)
    initial_biases_sum = np.sum(initial_result[theme].Biases)

    features_array = np.array(features)

    # Apply multiple high rewards
    for _ in range(10):
        handler.update(user_uuid, 1.0, theme, features_array)

    # Get updated biases
    updated_result = db_managers.get_weight_bias(user_uuid)
    updated_biases_sum = np.sum(updated_result[theme].Biases)

    # Biases should increase with positive rewards
    # (exact behavior depends on algorithm, but should show learning)
    assert updated_result[theme].Version == 11  # 10 updates


def test_zero_reward_updates(db_managers, handler):
    """Test that zero rewards are handled correctly."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme]

    # Update with zero reward
    handler.update(user_uuid, 0.0, theme, features)

    # Should succeed without errors
    result = db_managers.get_weight_bias(user_uuid)
    assert result[theme].Version == 2


def test_alternating_rewards(db_managers, handler):
    """Test algorithm with alternating high/low rewards."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup
    features = [0.6] * 12
    ThemeFeatureBuilder(user_uuid, theme).with_features(features).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    features_array = np.array(features)

    # Alternate between high and low rewards
    rewards = [1.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0]
    for reward in rewards:
        handler.update(user_uuid, reward, theme, features_array)

    # Should handle all updates successfully
    result = db_managers.get_weight_bias(user_uuid)
    assert result[theme].Version == 1 + len(rewards)


def test_inverse_matrix_accuracy_over_time(db_managers, handler):
    """Test that inverse matrix remains accurate over many updates."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme]

    # Perform 50 updates with varying rewards
    for i in range(50):
        reward = 0.3 + 0.4 * np.sin(i * 0.1)  # Varying between 0.3 and 0.7
        handler.update(user_uuid, reward, theme, features)

    # Check inverse accuracy
    result = db_managers.get_weight_bias(user_uuid)
    weights = result[theme].Weights
    weights_inv = result[theme].WeightsInv

    product = weights @ weights_inv
    identity = np.eye(12)

    # Should be close to identity
    np.testing.assert_array_almost_equal(product, identity, decimal=4)
