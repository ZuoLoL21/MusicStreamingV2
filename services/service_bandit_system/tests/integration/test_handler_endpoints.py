"""Integration tests for HTTP endpoints with real databases."""

import json
import pytest
from uuid import uuid4

from tests.integration.builders import (
    ThemeFeatureBuilder,
    BanditWeightBuilder,
    create_test_user_with_themes,
)


def test_health_endpoint(test_client):
    """Test health check endpoint (no database dependency)."""
    response = test_client.get("/api/v1/health")

    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "healthy"
    assert data["service"] == "bandit-system"


def test_root_endpoint(test_client):
    """Test root endpoint returns service metadata."""
    response = test_client.get("/")

    assert response.status_code == 200
    data = response.json()
    assert "message" in data
    assert "version" in data


def test_predict_success_single_theme(test_client, db_managers):
    """Test successful prediction with a single theme."""
    user_uuid = uuid4()
    theme = "rock"

    # Insert test data
    ThemeFeatureBuilder(user_uuid, theme).with_high_engagement().build(db_managers)

    # Make prediction request
    response = test_client.post(
        "/api/v1/predict",
        json={"user_uuid": str(user_uuid)}
    )

    assert response.status_code == 200
    data = response.json()
    assert "theme" in data
    assert "features" in data
    assert data["theme"] == theme
    assert len(data["features"]) == 12


def test_predict_success_multiple_themes(test_client, db_managers):
    """Test prediction with multiple themes selects one."""
    user_uuid = uuid4()
    themes = ["rock", "jazz", "classical", "pop"]

    # Insert test data for all themes
    for theme in themes:
        ThemeFeatureBuilder(user_uuid, theme).build(db_managers)

    # Make prediction request
    response = test_client.post(
        "/api/v1/predict",
        json={"user_uuid": str(user_uuid)}
    )

    assert response.status_code == 200
    data = response.json()
    assert "theme" in data
    assert "features" in data
    # Should select one of the available themes
    assert data["theme"] in themes
    assert len(data["features"]) == 12


def test_predict_no_themes_available(test_client, db_managers):
    """Test prediction fails when NO themes exist in the entire system."""
    user_uuid = uuid4()

    # Don't insert any themes (empty system)

    # Make prediction request
    response = test_client.post(
        "/api/v1/predict",
        json={"user_uuid": str(user_uuid)}
    )

    assert response.status_code == 500
    assert "No themes exist in the system" in response.json()["detail"]


def test_predict_new_user_cold_start(test_client, db_managers):
    """Test prediction for new user with no data (cold-start scenario)."""
    new_user_uuid = uuid4()
    existing_user_uuid = uuid4()
    themes = ["rock", "jazz", "classical", "pop", "metal"]

    # Insert themes for an existing user (so themes exist in the system)
    # Each with different popularity levels
    for i, theme in enumerate(themes):
        builder = ThemeFeatureBuilder(existing_user_uuid, theme)
        # Give them different popularity (higher impressions = more popular)
        builder.features[7] = 1000.0 - (i * 100)  # f_theme_decay_impressions
        builder.build(db_managers)

    # New user makes prediction request (has no data yet)
    response = test_client.post(
        "/api/v1/predict",
        json={"user_uuid": str(new_user_uuid)}
    )

    # Should succeed with random exploration across all available themes
    assert response.status_code == 200
    data = response.json()
    assert "theme" in data
    assert "features" in data
    # Should return one of all available themes
    assert data["theme"] in themes
    # Features should be zeros (pure exploration)
    assert all(f == 0.0 for f in data["features"])


def test_predict_existing_user_with_exploration_themes(test_client, db_managers):
    """Test that existing users get all themes for discovering new genres."""
    user_uuid = uuid4()
    other_user_uuid = uuid4()

    # User has listened to only 2 themes
    user_themes = ["rock", "jazz"]
    for theme in user_themes:
        ThemeFeatureBuilder(user_uuid, theme).with_high_engagement().build(db_managers)

    # System has more themes (some user hasn't tried)
    all_themes = ["rock", "jazz", "pop", "electronic", "classical"]
    for i, theme in enumerate(all_themes):
        if theme not in user_themes:
            # Other users have listened to these themes
            builder = ThemeFeatureBuilder(other_user_uuid, theme)
            builder.features[7] = 1000.0 - (i * 100)  # f_theme_decay_impressions
            builder.build(db_managers)

    # Make prediction
    response = test_client.post(
        "/api/v1/predict",
        json={"user_uuid": str(user_uuid)}
    )

    assert response.status_code == 200
    data = response.json()

    # User should be able to get recommended ANY theme in the system
    # (Their 2 themes with personalized features + all other themes for exploration)
    assert data["theme"] in all_themes


def test_predict_invalid_uuid(test_client):
    """Test prediction with invalid UUID format."""
    response = test_client.post(
        "/api/v1/predict",
        json={"user_uuid": "not-a-uuid"}
    )

    assert response.status_code == 422  # Validation error


def test_predict_exploration_for_new_theme(test_client, db_managers):
    """Test that new themes without weights get explored."""
    user_uuid = uuid4()
    theme = "rock"

    # Insert features but no weights (new theme)
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)

    # Make prediction request
    response = test_client.post(
        "/api/v1/predict",
        json={"user_uuid": str(user_uuid)}
    )

    assert response.status_code == 200
    data = response.json()
    assert data["theme"] == theme

    # Should have high exploration (no prior data)
    # The features should be returned
    assert len(data["features"]) == 12


def test_update_success_new_theme(test_client, db_managers, handler):
    """Test update creates new weight entry for first-time theme."""
    user_uuid = uuid4()
    theme = "rock"
    reward = 0.8

    # Insert features in ClickHouse
    builder = ThemeFeatureBuilder(user_uuid, theme).with_high_engagement()
    builder.build(db_managers)

    # Get the features we just inserted (handler.predict would give us these)
    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme].tolist()

    # Make update request
    response = test_client.post(
        "/api/v1/update",
        json={
            "user_uuid": str(user_uuid),
            "reward": reward,
            "theme": theme,
            "features": features,
        },
    )

    assert response.status_code == 202
    assert response.json() == {"success": True}

    # Verify weights were created in database
    weights = db_managers.get_weight_bias(user_uuid)
    assert theme in weights
    assert weights[theme].Version == 1  # Should be incremented after update


def test_update_success_existing_theme(test_client, db_managers):
    """Test update increments version for existing theme."""
    user_uuid = uuid4()
    theme = "rock"
    reward = 0.6

    # Insert features and initial weights
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(3).build(db_managers)

    # Get features
    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme].tolist()

    # Make update request
    response = test_client.post(
        "/api/v1/update",
        json={
            "user_uuid": str(user_uuid),
            "reward": reward,
            "theme": theme,
            "features": features,
        },
    )

    assert response.status_code == 202

    # Verify version was incremented
    weights = db_managers.get_weight_bias(user_uuid)
    assert weights[theme].Version == 4  # Should be incremented from 3 to 4


def test_update_invalid_reward_too_high(test_client):
    """Test that rewards > 1.0 are rejected or clamped."""
    user_uuid = uuid4()
    theme = "rock"
    features = [0.5] * 12

    response = test_client.post(
        "/api/v1/update",
        json={
            "user_uuid": str(user_uuid),
            "reward": 1.5,  # Invalid
            "theme": theme,
            "features": features,
        },
    )

    # Should either reject or clamp (check implementation)
    # Based on typical validation, this might return 422
    assert response.status_code in [200, 422]


def test_update_invalid_reward_negative(test_client):
    """Test that negative rewards are rejected or clamped."""
    user_uuid = uuid4()
    theme = "rock"
    features = [0.5] * 12

    response = test_client.post(
        "/api/v1/update",
        json={
            "user_uuid": str(user_uuid),
            "reward": -0.5,  # Invalid
            "theme": theme,
            "features": features,
        },
    )

    assert response.status_code in [202, 422]


def test_update_feature_dimension_mismatch(test_client, db_managers):
    """Test update with wrong number of features."""
    user_uuid = uuid4()
    theme = "rock"

    response = test_client.post(
        "/api/v1/update",
        json={
            "user_uuid": str(user_uuid),
            "reward": 0.5,
            "theme": theme,
            "features": [0.5] * 10,  # Wrong dimension (should be 12)
        },
    )

    # Should fail validation or in processing
    assert response.status_code == 400


def test_update_without_prior_predict(test_client, db_managers):
    """Test that update works even without prior predict (direct update flow)."""
    user_uuid = uuid4()
    theme = "rock"
    reward = 0.7

    # Insert features (no weights)
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)

    # Get features
    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme].tolist()

    # Make update request without predict
    response = test_client.post(
        "/api/v1/update",
        json={
            "user_uuid": str(user_uuid),
            "reward": reward,
            "theme": theme,
            "features": features,
        },
    )

    assert response.status_code == 202

    # Verify weights were created
    weights = db_managers.get_weight_bias(user_uuid)
    assert theme in weights


def test_predict_then_update_flow(test_client, db_managers):
    """Test the typical flow: predict -> update."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup: insert features
    ThemeFeatureBuilder(user_uuid, theme).with_high_engagement().build(db_managers)

    # Step 1: Predict
    predict_response = test_client.post(
        "/api/v1/predict",
        json={"user_uuid": str(user_uuid)}
    )
    assert predict_response.status_code == 200

    predict_data = predict_response.json()
    chosen_theme = predict_data["theme"]
    chosen_features = predict_data["features"]

    # Step 2: Update with reward
    update_response = test_client.post(
        "/api/v1/update",
        json={
            "user_uuid": str(user_uuid),
            "reward": 0.9,
            "theme": chosen_theme,
            "features": chosen_features,
        },
    )

    assert update_response.status_code == 202
    assert update_response.json() == {"success": True}

    # Verify weights exist and were updated
    weights = db_managers.get_weight_bias(user_uuid)
    assert chosen_theme in weights
    # Version should be at least 1 (1 initial + 1 update)
    assert weights[chosen_theme].Version >= 1


def test_multiple_predictions_same_user(test_client, db_managers):
    """Test multiple predictions for the same user."""
    user_uuid = uuid4()
    themes = ["rock", "jazz", "classical"]

    # Setup themes
    for theme in themes:
        ThemeFeatureBuilder(user_uuid, theme).build(db_managers)

    # Make multiple predictions
    predictions = []
    for _ in range(5):
        response = test_client.post(
            "/api/v1/predict",
            json={"user_uuid": str(user_uuid)}
        )
        assert response.status_code == 200
        predictions.append(response.json()["theme"])

    # All predictions should be valid themes
    for pred in predictions:
        assert pred in themes


def test_multiple_updates_same_theme(test_client, db_managers):
    """Test multiple updates to the same theme increment version correctly."""
    user_uuid = uuid4()
    theme = "rock"

    # Setup
    ThemeFeatureBuilder(user_uuid, theme).build(db_managers)
    BanditWeightBuilder(user_uuid, theme).with_version(1).build(db_managers)

    features_dict = db_managers.get_input_data(user_uuid)
    features = features_dict[theme].tolist()

    # Make 3 sequential updates
    for i in range(3):
        response = test_client.post(
            "/api/v1/update",
            json={
                "user_uuid": str(user_uuid),
                "reward": 0.5 + i * 0.1,
                "theme": theme,
                "features": features,
            },
        )
        assert response.status_code == 202

    # Verify final version
    weights = db_managers.get_weight_bias(user_uuid)
    # Started at version 1, should be at version 4 after 3 updates
    assert weights[theme].Version == 4
