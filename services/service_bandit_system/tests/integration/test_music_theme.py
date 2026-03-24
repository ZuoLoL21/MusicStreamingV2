"""Integration tests for music_theme table."""

import pytest
from uuid import uuid4

from tests.integration.builders import MusicThemeBuilder


def test_insert_music_theme(db_managers):
    """Test inserting a music-theme association."""
    music_uuid = uuid4()
    theme = "rock"

    # Insert music theme
    MusicThemeBuilder(music_uuid, theme).build(db_managers)

    # Verify it was inserted by checking theme cache
    themes = db_managers._themes_cache.get_all_themes()
    assert theme in themes


def test_insert_multiple_themes_for_music(db_managers):
    """Test that one music track can have multiple themes."""
    music_uuid = uuid4()
    themes = ["rock", "alternative", "indie"]

    # Insert multiple themes for same music
    for theme in themes:
        MusicThemeBuilder(music_uuid, theme).build(db_managers)

    # Verify all themes are in catalog
    all_themes = db_managers._themes_cache.get_all_themes()
    for theme in themes:
        assert theme in all_themes


def test_insert_same_theme_different_music(db_managers):
    """Test that same theme can be associated with multiple music tracks."""
    music_uuids = [uuid4() for _ in range(3)]
    theme = "rock"

    # Insert same theme for different music
    for music_uuid in music_uuids:
        MusicThemeBuilder(music_uuid, theme).build(db_managers)

    # Verify theme appears in catalog
    all_themes = db_managers._themes_cache.get_all_themes()
    assert theme in all_themes


def test_music_theme_with_stats(db_managers):
    """Test inserting music-theme with view and success statistics."""
    music_uuid = uuid4()
    theme = "jazz"

    # Insert with stats
    MusicThemeBuilder(music_uuid, theme).with_stats(views=100, successes=80).build(db_managers)

    # Verify theme exists in catalog
    all_themes = db_managers._themes_cache.get_all_themes()
    assert theme in all_themes


def test_theme_catalog_empty_initially(db_managers):
    """Test that theme catalog is empty when no music themes are inserted."""
    # Don't insert anything
    themes = db_managers._themes_cache.get_all_themes()
    assert len(themes) == 0


def test_theme_catalog_with_multiple_themes(db_managers):
    """Test theme catalog with various music and themes."""
    # Create a diverse catalog
    music_themes = [
        (uuid4(), "rock"),
        (uuid4(), "jazz"),
        (uuid4(), "classical"),
        (uuid4(), "rock"),  # Duplicate theme, different music
        (uuid4(), "pop"),
    ]

    for music_uuid, theme in music_themes:
        MusicThemeBuilder(music_uuid, theme).build(db_managers)

    # Verify distinct themes in catalog
    all_themes = db_managers._themes_cache.get_all_themes()
    expected_themes = {"rock", "jazz", "classical", "pop"}
    assert all_themes == expected_themes


def test_cold_start_with_theme_catalog(db_managers, handler):
    """Test that new user can get predictions from theme catalog."""
    user_uuid = uuid4()

    # Populate theme catalog
    themes = ["rock", "jazz", "pop"]
    for theme in themes:
        music_uuid = uuid4()
        MusicThemeBuilder(music_uuid, theme).build(db_managers)

    # New user should be able to get prediction
    predicted_theme, features = handler.predict(user_uuid)

    # Should predict one of the available themes
    assert predicted_theme in themes
    # Features should be zeros (no user data)
    assert all(f == 0.0 for f in features)
