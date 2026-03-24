"""Unit tests for ThemesCache."""
import time
from unittest.mock import Mock, MagicMock, patch
import pytest

from src.cache.themes import ThemesCache


@pytest.fixture
def mock_engine():
    """Create a mock SQLAlchemy engine."""
    engine = Mock()
    return engine


@pytest.fixture
def mock_connection(mock_engine):
    """Create a mock database connection."""
    conn = MagicMock()
    mock_engine.connect.return_value.__enter__ = Mock(return_value=conn)
    mock_engine.connect.return_value.__exit__ = Mock(return_value=None)
    return conn


def test_cache_initialization(mock_engine):
    """Test that cache initializes with correct parameters."""
    cache = ThemesCache(mock_engine, "test_table", ttl_minutes=15)

    assert cache._warehouse_engine == mock_engine
    assert cache._bandit_data_table == "test_table"
    assert cache._ttl.total_seconds() == 15 * 60
    assert cache._cache is None
    assert cache._expires_at is None


def test_get_all_themes_first_call(mock_engine, mock_connection):
    """Test that first call queries the database and caches result."""
    # Setup mock to return themes
    mock_result = MagicMock()
    mock_result.fetchall.return_value = [("rock",), ("jazz",), ("classical",)]
    mock_connection.execute.return_value = mock_result

    cache = ThemesCache(mock_engine, "test_table")

    # First call should query database
    themes = cache.get_all_themes()

    assert themes == {"rock", "jazz", "classical"}
    assert mock_connection.execute.called
    assert cache._cache is not None
    assert cache._expires_at is not None


def test_get_all_themes_cached(mock_engine, mock_connection):
    """Test that subsequent calls use cache without querying database."""
    # Setup mock
    mock_result = MagicMock()
    mock_result.fetchall.return_value = [("rock",), ("jazz",)]
    mock_connection.execute.return_value = mock_result

    cache = ThemesCache(mock_engine, "test_table")

    # First call
    themes1 = cache.get_all_themes()
    call_count_after_first = mock_connection.execute.call_count

    # Second call (should use cache)
    themes2 = cache.get_all_themes()
    call_count_after_second = mock_connection.execute.call_count

    assert themes1 == themes2
    assert call_count_after_first == call_count_after_second  # No additional DB call


def test_get_all_themes_cache_expiration(mock_engine, mock_connection):
    """Test that cache refreshes after TTL expires."""
    # Setup mock
    mock_result = MagicMock()
    mock_result.fetchall.return_value = [("rock",), ("jazz",)]
    mock_connection.execute.return_value = mock_result

    # Use very short TTL for testing (1 second)
    cache = ThemesCache(mock_engine, "test_table", ttl_minutes=0.02)  # ~1 second

    # First call
    themes1 = cache.get_all_themes()
    call_count_after_first = mock_connection.execute.call_count

    # Wait for cache to expire
    time.sleep(1.5)

    # Update mock to return different themes
    mock_result.fetchall.return_value = [("rock",), ("jazz",), ("classical",)]

    # Second call (should refresh cache)
    themes2 = cache.get_all_themes()
    call_count_after_second = mock_connection.execute.call_count

    assert themes1 == {"rock", "jazz"}
    assert themes2 == {"rock", "jazz", "classical"}
    assert call_count_after_second > call_count_after_first  # Additional DB call


def test_invalidate_cache(mock_engine, mock_connection):
    """Test manual cache invalidation."""
    # Setup mock
    mock_result = MagicMock()
    mock_result.fetchall.return_value = [("rock",)]
    mock_connection.execute.return_value = mock_result

    cache = ThemesCache(mock_engine, "test_table")

    # Populate cache
    cache.get_all_themes()
    assert cache._cache is not None

    # Invalidate
    cache.invalidate()
    assert cache._cache is None
    assert cache._expires_at is None

    # Next call should query database again
    themes = cache.get_all_themes()
    assert themes == {"rock"}
    assert mock_connection.execute.call_count == 2  # Once before, once after invalidation


def test_get_all_themes_empty_database(mock_engine, mock_connection):
    """Test behavior when database has no themes."""
    # Setup mock to return empty result
    mock_result = MagicMock()
    mock_result.fetchall.return_value = []
    mock_connection.execute.return_value = mock_result

    cache = ThemesCache(mock_engine, "test_table")

    themes = cache.get_all_themes()

    assert themes == set()
    assert isinstance(themes, set)


def test_thread_safety(mock_engine, mock_connection):
    """Test that cache uses locking correctly."""
    import threading

    # Setup mock
    mock_result = MagicMock()
    mock_result.fetchall.return_value = [("rock",), ("jazz",)]
    mock_connection.execute.return_value = mock_result

    cache = ThemesCache(mock_engine, "test_table")
    results = []

    def get_themes():
        themes = cache.get_all_themes()
        results.append(themes)

    # Call from multiple threads simultaneously
    threads = [threading.Thread(target=get_themes) for _ in range(10)]
    for thread in threads:
        thread.start()
    for thread in threads:
        thread.join()

    # All threads should get the same result
    assert len(results) == 10
    for result in results:
        assert result == {"rock", "jazz"}

    # Database should only be queried once (cache protects against concurrent queries)
    # Note: might be called more than once due to race conditions, but should be minimal
    assert mock_connection.execute.call_count <= 3


def test_sql_query_construction(mock_engine, mock_connection):
    """Test that SQL query is constructed correctly."""
    mock_result = MagicMock()
    mock_result.fetchall.return_value = [("rock",)]
    mock_connection.execute.return_value = mock_result

    cache = ThemesCache(mock_engine, "my_bandit_table")
    cache.get_all_themes()

    # Verify the SQL query contains the correct table name
    call_args = mock_connection.execute.call_args
    query = str(call_args[0][0])

    assert "my_bandit_table" in query
    assert "SELECT DISTINCT theme" in query
    assert "ORDER BY theme" in query
