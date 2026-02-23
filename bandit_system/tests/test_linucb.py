"""Tests for LinUCB algorithm."""

import numpy as np
import pytest
from src.models.linucb import LinUCB, ArmResultLinUCB, _check_divergence


def test_get_basic(bandit):
    """Test initialization of basic matrices."""
    dim = 5
    weight, bias, inverse_weight = bandit.get_basic(dim)

    # Check shapes
    assert weight.shape == (dim, dim)
    assert bias.shape == (dim,)
    assert inverse_weight.shape == (dim, dim)

    # Check values
    expected_weight = bandit._config.ridge_lambda * np.identity(dim)
    assert np.allclose(weight, expected_weight)
    assert np.allclose(bias, np.zeros(dim))

    # Check inverse is correct
    expected_inv = (1.0 / bandit._config.ridge_lambda) * np.identity(dim)
    assert np.allclose(inverse_weight, expected_inv)


def test_get_new_arm_result(bandit):
    """Test creation of new arm result."""
    theme = "jazz"
    dim = 5

    arm = bandit.get_new_arm_result(theme, dim)

    assert arm.Theme == theme
    assert arm.Version == 0
    assert arm.UpdatesSinceRecompute == 0
    assert arm.Weights.shape == (dim, dim)
    assert arm.Biases.shape == (dim,)
    assert arm.WeightsInv.shape == (dim, dim)


def test_predict_single_arm(bandit):
    """Test prediction with a single arm."""
    arm = bandit.get_new_arm_result("jazz", 3)
    features = [np.array([1.0, 0.5, 0.2])]

    chosen_idx = bandit.predict([arm], features)

    assert chosen_idx == 0


def test_predict_multiple_arms(bandit):
    """Test prediction with multiple arms."""
    arm1 = bandit.get_new_arm_result("jazz", 3)
    arm2 = bandit.get_new_arm_result("rock", 3)

    # Update arm1 to have higher expected reward
    arm1.Biases = np.array([10.0, 5.0, 2.0])

    features = [
        np.array([1.0, 0.5, 0.2]),
        np.array([1.0, 0.5, 0.2])
    ]

    chosen_idx = bandit.predict([arm1, arm2], features)

    # arm1 should be chosen because it has higher biases
    assert chosen_idx == 0


def test_predict_length_mismatch(bandit):
    """Test prediction with mismatched arms and features."""
    arm = bandit.get_new_arm_result("jazz", 3)
    features = [np.array([1.0, 0.5, 0.2]), np.array([1.0, 0.5, 0.2])]

    result = bandit.predict([arm], features)

    assert result is None


def test_update_basic(bandit):
    """Test basic update operation."""
    arm = bandit.get_new_arm_result("jazz", 3)
    features = np.array([1.0, 0.5, 0.2])
    reward = 0.8

    initial_weights = arm.Weights.copy()
    initial_biases = arm.Biases.copy()

    updated_arm = bandit.update(arm, features, reward)

    # Weights should have increased
    assert not np.allclose(updated_arm.Weights, initial_weights)

    # Biases should have increased
    assert not np.allclose(updated_arm.Biases, initial_biases)

    # UpdatesSinceRecompute should increment
    assert updated_arm.UpdatesSinceRecompute == 1


def test_update_reward_clamping(bandit):
    """Test that rewards are clamped to [0, 1]."""
    arm = bandit.get_new_arm_result("jazz", 3)
    features = np.array([1.0, 0.5, 0.2])

    # Test reward > 1.0 gets clamped to 1.0
    arm_high = bandit.update(arm.model_copy(deep=True), features, 1.5)

    # Test reward < 0.0 gets clamped to 0.0
    arm_low = bandit.update(arm.model_copy(deep=True), features, -0.5)

    # Both should succeed without error
    assert arm_high.UpdatesSinceRecompute == 1
    assert arm_low.UpdatesSinceRecompute == 1


def test_sherman_morrison_inverse_tracking(bandit):
    """Test Sherman-Morrison inverse tracking."""
    arm = bandit.get_new_arm_result("jazz", 3)
    features = np.array([1.0, 0.5, 0.2])

    # Perform update
    updated_arm = bandit.update(arm, features, 0.8)

    # Check that A_inv * A ≈ I
    product = updated_arm.WeightsInv @ updated_arm.Weights
    identity = np.eye(3)
    divergence = np.linalg.norm(identity - product, ord="fro")

    assert divergence < 1e-10


def test_check_divergence():
    """Test divergence checking function."""
    dim = 3
    A = np.array([[2.0, 0.5, 0.1], [0.5, 3.0, 0.2], [0.1, 0.2, 1.5]])
    A_inv = np.linalg.inv(A)

    divergence = _check_divergence(A, A_inv)

    # Should be very close to 0 for exact inverse
    assert divergence < 1e-10


def test_divergence_recompute(bandit, logger):
    """Test that inverse is recomputed when divergence exceeds threshold."""
    # Set low recompute interval for testing
    bandit._config.sherman_morrison_recompute_interval = 2

    arm = bandit.get_new_arm_result("jazz", 3)
    features = np.array([1.0, 0.5, 0.2])

    # Perform multiple updates
    for _ in range(3):
        arm = bandit.update(arm, features, 0.8)

    # After 3 updates with interval 2:
    # - Update 1: count=1
    # - Update 2: count=2 (>= 2, triggers check, resets to 0)
    # - Update 3: count=1
    assert arm.UpdatesSinceRecompute == 1
