"""Pytest configuration and fixtures."""

import pytest
import structlog
from unittest.mock import Mock
from fastapi.testclient import TestClient

from src.app.state import AppState
from src.di.config import Config
from src.di.db import DBManagers
from src.services.bandit import BanditHandler
from src.models.linucb import LinUCB
from main import app


@pytest.fixture
def mock_config():
    """Create a mock Config object."""
    config = Mock(spec=Config)
    config.alpha = 0.5
    config.ridge_lambda = 1.0
    config.max_retries = 3
    config.initial_backoff_ms = 100
    config.sherman_morrison_recompute_interval = 100
    config.sherman_morrison_divergence_threshold = 1e-6
    return config


@pytest.fixture
def logger():
    """Create a test logger."""
    return structlog.get_logger("test")


@pytest.fixture
def mock_db(mock_config, logger):
    """Create a mock DBManagers object."""
    return Mock(spec=DBManagers)


@pytest.fixture
def bandit(mock_config, logger):
    """Create a real LinUCB instance for testing."""
    return LinUCB(mock_config, logger)


@pytest.fixture
def handler(mock_config, mock_db, logger, bandit):
    """Create a BanditHandler with mocked dependencies."""
    return BanditHandler(mock_config, mock_db, logger, bandit)


@pytest.fixture
def app_state(mock_config, logger, bandit, mock_db, handler):
    """Create an AppState instance for testing."""
    return AppState(
        config=mock_config,
        logger=logger,
        bandit=bandit,
        db=mock_db,
        handler=handler,
    )


@pytest.fixture
def client(app_state):
    """Create a FastAPI test client with mocked app state."""
    # Set the app state before creating the client
    app.state.app_state = app_state
    return TestClient(app)
