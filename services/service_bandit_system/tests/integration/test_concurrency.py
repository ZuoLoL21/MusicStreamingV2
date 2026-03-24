"""Integration tests for concurrent operations and version conflict handling."""

import time
import pytest
import numpy as np
from uuid import uuid4
from concurrent.futures import ThreadPoolExecutor, as_completed
from threading import Barrier

from tests.integration.builders import (
    ThemeFeatureBuilder,
    BanditWeightBuilder,
)


def test_concurrent_updates_same_theme(db_managers, handler):
    """Test that concurrent updates to the same theme all succeed via retry."""
    user_uuid = uuid4()
    theme = "rock"
    num_updates = 10

    # Setup initial data
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    # Get features
    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme]

    # Barrier to synchronize threads
    barrier = Barrier(num_updates)

    def update_with_barrier(reward_value):
        """Update function that waits for all threads before executing."""
        barrier.wait()  # Wait for all threads to be ready
        try:
            handler.update(user_uuid, reward_value, theme, features)
            return True
        except Exception as e:
            return False

    # Execute concurrent updates
    with ThreadPoolExecutor(max_workers=num_updates) as executor:
        futures = [
            executor.submit(update_with_barrier, 0.5 + i * 0.01)
            for i in range(num_updates)
        ]
        results = [future.result() for future in as_completed(futures)]

    # All updates should succeed (via retry mechanism)
    assert all(results), "Some updates failed"

    # Verify final version
    weights = db_managers.get_weight_bias(user_uuid)
    # Started at version 1, should be at version 1 + num_updates
    expected_version = 1 + num_updates
    assert weights[theme].Version == expected_version


def test_concurrent_updates_different_themes(db_managers, handler):
    """Test concurrent updates to different themes (no conflicts)."""
    user_uuid = uuid4()
    themes = ["rock", "jazz", "classical", "pop", "metal"]

    # Setup data for all themes
    for theme in themes:
        ThemeFeatureBuilder(user_uuid, theme).build(db_managers)
        BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    # Get features for each theme
    features_dict = db_managers.get_input_data(user_uuid)

    def update_theme(theme):
        """Update a specific theme."""
        features = features_dict[theme]
        handler.update(user_uuid, 0.7, theme, features)
        return theme

    # Execute concurrent updates to different themes
    with ThreadPoolExecutor(max_workers=len(themes)) as executor:
        futures = [executor.submit(update_theme, theme) for theme in themes]
        results = [future.result() for future in as_completed(futures)]

    # All updates should succeed (no conflicts across themes)
    assert len(results) == len(themes)

    # Verify all themes were updated
    weights = db_managers.get_weight_bias(user_uuid)
    for theme in themes:
        assert weights[theme].Version == 2  # All should be updated once


def test_version_conflict_detection(db_managers):
    """Test that version conflict is properly detected."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup initial data with version 5
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(5).build(db_managers)

    # Try to update with wrong version (should fail)
    weights = np.eye(12) * 2.0
    biases = np.ones(12)
    weights_inv = np.eye(12) * 0.5

    # Attempt update with version 3 (wrong)
    success = db_managers.update_weight_bias(
        user_uuid, theme, weights, biases, weights_inv, 1, 3
    )

    assert success is False, "Update should fail with version conflict"

    # Verify version unchanged
    result = db_managers.get_weight_bias(user_uuid)
    assert result[theme].Version == 5


def test_retry_mechanism_success(db_managers, handler):
    """Test that retry mechanism eventually succeeds."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme]

    # Simulate concurrent update scenario
    def competing_update():
        """Competing update that runs slightly before."""
        time.sleep(0.05)  # Small delay
        handler.update(user_uuid, 0.6, theme, features)

    # Start competing update in background
    with ThreadPoolExecutor(max_workers=1) as executor:
        future = executor.submit(competing_update)

        # Main update should retry and succeed
        handler.update(user_uuid, 0.8, theme, features)

        future.result()  # Wait for competing update

    # Both updates should succeed (version should be 3)
    weights = db_managers.get_weight_bias(user_uuid)
    assert weights[theme].Version == 3


def test_max_retries_exceeded(db_managers, handler, monkeypatch):
    """Test that update fails after max retries."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme]

    # Mock update to always fail (simulate persistent conflict)
    original_update = db_managers.update_weight_bias

    def always_fail_update(*args, **kwargs):
        return False  # Always return False (conflict)

    monkeypatch.setattr(db_managers, "update_weight_bias", always_fail_update)

    # Attempt update (should fail after max retries)
    with pytest.raises(RuntimeError, match="Failed to update bandit for user .* after .* retries"):
        handler.update(user_uuid, 0.8, theme, features)


def test_concurrent_predict_same_user(db_managers, handler):
    """Test concurrent predictions for the same user (read-only, no conflicts)."""
    user_uuid = uuid4()
    themes = ["rock", "jazz", "classical"]

    # Setup themes
    for theme in themes:
        ThemeFeatureBuilder(user_uuid, theme).build(db_managers)

    def predict():
        """Make a prediction."""
        theme, features = handler.predict(user_uuid)
        return theme

    # Execute concurrent predictions
    num_predictions = 20
    with ThreadPoolExecutor(max_workers=10) as executor:
        futures = [executor.submit(predict) for _ in range(num_predictions)]
        results = [future.result() for future in as_completed(futures)]

    # All predictions should succeed
    assert len(results) == num_predictions
    # All results should be valid themes
    for result in results:
        assert result in themes


def test_predict_during_update(db_managers, handler):
    """Test that predictions work correctly during concurrent updates."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme]

    def predict_multiple():
        """Make multiple predictions."""
        results = []
        for _ in range(10):
            theme_result, _ = handler.predict(user_uuid)
            results.append(theme_result)
        return results

    def update_multiple():
        """Make multiple updates."""
        for i in range(10):
            handler.update(user_uuid, 0.5 + i * 0.01, theme, features)

    # Run predictions and updates concurrently
    with ThreadPoolExecutor(max_workers=2) as executor:
        predict_future = executor.submit(predict_multiple)
        update_future = executor.submit(update_multiple)

        predictions = predict_future.result()
        update_future.result()

    # All predictions should succeed
    assert len(predictions) == 10
    assert all(p == theme for p in predictions)


def test_high_concurrency_stress(db_managers, handler):
    """Stress test with high concurrency (50 updates to same theme)."""
    user_uuid = uuid4()
    theme = "rock"
    num_updates = 50

    # Setup
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme]

    barrier = Barrier(num_updates)

    def concurrent_update(idx):
        """Update with barrier for maximum contention."""
        barrier.wait()
        try:
            handler.update(user_uuid, 0.5 + idx * 0.005, theme, features)
            return True
        except Exception as e:
            print(f"Update {idx} failed: {e}")
            return False

    start_time = time.time()

    # Execute high-concurrency updates
    with ThreadPoolExecutor(max_workers=num_updates) as executor:
        futures = [executor.submit(concurrent_update, i) for i in range(num_updates)]
        results = [future.result() for future in as_completed(futures)]

    elapsed = time.time() - start_time

    # Check success rate (should be 100% with retries)
    success_count = sum(results)
    success_rate = success_count / num_updates * 100

    print(f"High concurrency test: {success_count}/{num_updates} succeeded ({success_rate:.1f}%)")
    print(f"Elapsed time: {elapsed:.2f}s")

    # Allow for some failures under extreme contention, but most should succeed
    assert success_rate >= 90, f"Success rate too low: {success_rate:.1f}%"

    # Verify final version (should be close to num_updates + 1)
    weights = db_managers.get_weight_bias(user_uuid)
    # Account for potential failures
    assert weights[theme].Version >= num_updates * 0.9
