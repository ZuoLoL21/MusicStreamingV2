"""Tests for API routes."""

import numpy as np
import pytest
from unittest.mock import Mock
from uuid import uuid4


def test_health_check(client):
    """Test the health check endpoint."""
    response = client.get("/api/v1/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "healthy"
    assert data["service"] == "bandit-system"


def test_root_endpoint(client):
    """Test the root endpoint."""
    response = client.get("/")
    assert response.status_code == 200
    data = response.json()
    assert "message" in data
    assert "version" in data


def test_predict_success(client, app_state):
    """Test successful prediction."""
    user_uuid = uuid4()
    theme = "jazz"
    features = np.array([0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1, 1.2])

    # Mock the handler's predict method on the actual app_state handler
    app_state.handler.predict = Mock(return_value=(theme, features))

    response = client.post(
        "/api/v1/predict",
        json={"user_uuid": str(user_uuid)}
    )

    assert response.status_code == 200
    data = response.json()
    assert data["theme"] == theme
    assert len(data["features"]) == 12
    # Request ID is in headers, not body
    assert "X-Request-ID" in response.headers


def test_predict_invalid_uuid(client):
    """Test prediction with invalid UUID."""
    response = client.post(
        "/api/v1/predict",
        json={"user_uuid": "not-a-uuid"}
    )
    assert response.status_code == 422  # Validation error


def test_predict_handler_error(client, app_state):
    """Test prediction when handler raises an error."""
    user_uuid = uuid4()

    # Mock the handler to raise an error on the app_state handler
    app_state.handler.predict = Mock(side_effect=RuntimeError("No themes in DB"))

    response = client.post(
        "/api/v1/predict",
        json={"user_uuid": str(user_uuid)}
    )

    assert response.status_code == 500
    assert "No themes in DB" in response.json()["detail"]


def test_update_success(client, app_state):
    """Test successful update."""
    user_uuid = uuid4()
    theme = "jazz"
    reward = 0.75
    features = [0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1, 1.2]

    # Mock the handler's update method on the app_state handler
    app_state.handler.update = Mock(return_value=None)

    response = client.post(
        "/api/v1/update",
        json={
            "user_uuid": str(user_uuid),
            "theme": theme,
            "reward": reward,
            "features": features
        }
    )

    assert response.status_code == 202  # Accepted
    data = response.json()
    assert data["success"] is True
    # Request ID is in headers, not body
    assert "X-Request-ID" in response.headers

    # Verify handler was called with correct arguments
    app_state.handler.update.assert_called_once()
    call_args = app_state.handler.update.call_args[0]
    assert call_args[0] == user_uuid
    assert call_args[1] == reward
    assert call_args[2] == theme
    assert np.array_equal(call_args[3], np.array(features, dtype=np.float64))


def test_update_invalid_reward(client):
    """Test update with invalid reward value."""
    user_uuid = uuid4()

    # Reward > 1.0
    response = client.post(
        "/api/v1/update",
        json={
            "user_uuid": str(user_uuid),
            "theme": "jazz",
            "reward": 1.5,
            "features": [0.1] * 12
        }
    )
    assert response.status_code == 422  # Validation error

    # Reward < 0.0
    response = client.post(
        "/api/v1/update",
        json={
            "user_uuid": str(user_uuid),
            "theme": "jazz",
            "reward": -0.5,
            "features": [0.1] * 12
        }
    )
    assert response.status_code == 422  # Validation error


def test_update_handler_error(client, app_state):
    """Test update when handler raises an error."""
    user_uuid = uuid4()

    # Mock the handler to raise an error on the app_state handler
    app_state.handler.update = Mock(side_effect=RuntimeError("Update failed"))

    response = client.post(
        "/api/v1/update",
        json={
            "user_uuid": str(user_uuid),
            "theme": "jazz",
            "reward": 0.5,
            "features": [0.1] * 12
        }
    )

    assert response.status_code == 500
    assert "Update failed" in response.json()["detail"]
