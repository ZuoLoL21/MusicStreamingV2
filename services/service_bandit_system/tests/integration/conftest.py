"""Pytest configuration and fixtures for integration tests."""

import os
import pytest
import structlog
from sqlalchemy import create_engine, text
from fastapi.testclient import TestClient

from src.app.state import AppState
from src.di.config import Config
from src.di.db import DBManagers
from src.services.bandit import BanditHandler
from src.models.linucb import LinUCB
from main import app

import logging

# Test database connection strings
# Support both containerized tests (using service names) and local tests (using localhost)
POSTGRES_HOST = os.getenv("POSTGRES_HOST", "localhost")
POSTGRES_PORT = os.getenv("POSTGRES_PORT", "5434")
POSTGRES_USER = os.getenv("POSTGRES_USER", "test")
POSTGRES_PASSWORD = os.getenv("POSTGRES_PASSWORD", "test")
POSTGRES_DB = os.getenv("POSTGRES_DB", "bandit_test")
POSTGRES_TEST_URL = f"postgresql://{POSTGRES_USER}:{POSTGRES_PASSWORD}@{POSTGRES_HOST}:{POSTGRES_PORT}/{POSTGRES_DB}"

CLICKHOUSE_HOST = os.getenv("CLICKHOUSE_HOST", "localhost")
CLICKHOUSE_HTTP_PORT = os.getenv("CLICKHOUSE_HTTP_PORT", "8124")
CLICKHOUSE_USER = os.getenv("CLICKHOUSE_USER", "default")
CLICKHOUSE_DB = os.getenv("CLICKHOUSE_DB", "default")
CLICKHOUSE_TEST_URL = f"clickhouse://{CLICKHOUSE_USER}:@{CLICKHOUSE_HOST}:{CLICKHOUSE_HTTP_PORT}/{CLICKHOUSE_DB}"


@pytest.fixture(scope="session")
def test_config():
    """Create a Config object for testing with test database connections."""
    return Config(
        db_warehouse_string=CLICKHOUSE_TEST_URL,
        bandit_data_table="bandit_input_per_theme",
        db_params_string=POSTGRES_TEST_URL,
        bandit_params_table="bandit_data",
        alpha=0.5,
        ridge_lambda=1.0,
        max_retries=5,
        initial_backoff_ms=100,
        sherman_morrison_recompute_interval=100,
        sherman_morrison_divergence_threshold=1e-6,
    )


@pytest.fixture(scope="session")
def logger():
    """Create a test logger."""
    return structlog.get_logger("integration_test")


@pytest.fixture(scope="session")
def bandit(test_config, logger):
    """Create a LinUCB instance for testing."""
    return LinUCB(test_config, logger)


@pytest.fixture(scope="function")
def db_managers(test_config, bandit):
    """Create DatabaseManagers connected to test databases.

    Function-scoped to ensure fresh connections for each test.
    """
    return DBManagers(test_config, bandit)


@pytest.fixture(scope="function")
def handler(test_config, db_managers, logger, bandit):
    """Create a BanditHandler with real database connections."""
    return BanditHandler(test_config, db_managers, logger, bandit)


@pytest.fixture(scope="function")
def app_state(test_config, logger, bandit, db_managers, handler):
    """Create an AppState instance for testing."""
    return AppState(
        config=test_config,
        logger=logger,
        bandit=bandit,
        db=db_managers,
        handler=handler,
    )


@pytest.fixture(scope="function")
def test_client(app_state):
    """Create a FastAPI test client with real app state."""
    app.state.app_state = app_state
    return TestClient(app)


@pytest.fixture(autouse=True, scope="function")
def clean_databases(test_config):
    """Clean databases before each test to ensure isolation.

    This fixture runs automatically before each test function.
    """
    # Clean PostgreSQL
    postgres_engine = create_engine(test_config.db_params_string)
    with postgres_engine.connect() as conn:
        with conn.begin():
            conn.execute(text(f"DELETE FROM {test_config.bandit_params_table}"))
    postgres_engine.dispose()

    # Clean ClickHouse
    clickhouse_engine = create_engine(test_config.db_warehouse_string)
    with clickhouse_engine.connect() as conn:
        with conn.begin():
            conn.execute(text(f"ALTER TABLE {test_config.bandit_data_table} DELETE WHERE 1=1"))
    clickhouse_engine.dispose()

    yield

    # Cleanup after test (optional, but good practice)
    # The databases will be cleaned again before the next test


@pytest.fixture
def postgres_engine(test_config):
    """Provide a PostgreSQL engine for direct database access in tests."""
    engine = create_engine(test_config.db_params_string)
    yield engine
    engine.dispose()


@pytest.fixture
def clickhouse_engine(test_config):
    """Provide a ClickHouse engine for direct database access in tests."""
    engine = create_engine(test_config.db_warehouse_string)
    yield engine
    engine.dispose()
