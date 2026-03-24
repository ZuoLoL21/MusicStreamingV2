"""Factory classes for creating test data in integration tests."""

import json
import numpy as np
from uuid import UUID
from sqlalchemy import text
from typing import List

from src.di.db import DBManagers, NUMB_FEATURES


class ThemeFeatureBuilder:
    """Builds feature vectors for ClickHouse warehouse database.

    Example usage:
        builder = ThemeFeatureBuilder(user_uuid, "rock")
        builder.with_features([0.8, 0.7, 0.6, ...]).build(db_managers)
    """

    def __init__(self, user_uuid: UUID, theme: str):
        self.user_uuid = user_uuid
        self.theme = theme
        # Default neutral features
        self.features = [0.5] * NUMB_FEATURES

    def with_features(self, features: List[float]) -> "ThemeFeatureBuilder":
        """Set custom feature values.

        Args:
            features: List of 12 float values for the feature vector

        Returns:
            self for method chaining
        """
        if len(features) != NUMB_FEATURES:
            raise ValueError(f"Expected {NUMB_FEATURES} features, got {len(features)}")
        self.features = features
        return self

    def with_high_engagement(self) -> "ThemeFeatureBuilder":
        """Set features indicating high user engagement with this theme."""
        self.features = [
            0.9,  # f_user_theme_decay_impressions
            0.85,  # f_user_theme_decay_completion
            0.8,  # f_user_theme_full_play_rate
            0.75,  # f_user_theme_decay_like_rate
            0.7,  # f_user_decay_impressions
            0.65,  # f_user_decay_completion
            0.6,  # f_user_decay_like_rate
            0.55,  # f_theme_decay_impressions
            0.5,  # f_theme_decay_completion
            0.45,  # f_theme_decay_like_rate
            0.8,  # f_relative_completion
            0.75,  # f_relative_exposure
        ]
        return self

    def with_low_engagement(self) -> "ThemeFeatureBuilder":
        """Set features indicating low user engagement with this theme."""
        self.features = [
            0.1,  # f_user_theme_decay_impressions
            0.15,  # f_user_theme_decay_completion
            0.2,  # f_user_theme_full_play_rate
            0.1,  # f_user_theme_decay_like_rate
            0.3,  # f_user_decay_impressions
            0.35,  # f_user_decay_completion
            0.25,  # f_user_decay_like_rate
            0.4,  # f_theme_decay_impressions
            0.45,  # f_theme_decay_completion
            0.3,  # f_theme_decay_like_rate
            0.2,  # f_relative_completion
            0.15,  # f_relative_exposure
        ]
        return self

    def build(self, db_managers: DBManagers):
        """Insert the theme feature into ClickHouse.

        Args:
            db_managers: DatabaseManagers instance with warehouse connection
        """
        query = text(
            f"INSERT INTO {db_managers._config.bandit_data_table} "
            "(user_uuid, theme, f_user_theme_decay_impressions, "
            "f_user_theme_decay_completion, f_user_theme_full_play_rate, "
            "f_user_theme_decay_like_rate, f_user_decay_impressions, "
            "f_user_decay_completion, f_user_decay_like_rate, "
            "f_theme_decay_impressions, f_theme_decay_completion, "
            "f_theme_decay_like_rate, f_relative_completion, f_relative_exposure) "
            "VALUES (:uuid, :theme, :f1, :f2, :f3, :f4, :f5, :f6, :f7, :f8, :f9, :f10, :f11, :f12)"
        )

        with db_managers._warehouse_engine.connect() as conn:
            with conn.begin():
                conn.execute(
                    query,
                    {
                        "uuid": str(self.user_uuid),
                        "theme": self.theme,
                        "f1": self.features[0],
                        "f2": self.features[1],
                        "f3": self.features[2],
                        "f4": self.features[3],
                        "f5": self.features[4],
                        "f6": self.features[5],
                        "f7": self.features[6],
                        "f8": self.features[7],
                        "f9": self.features[8],
                        "f10": self.features[9],
                        "f11": self.features[10],
                        "f12": self.features[11],
                    },
                )


class BanditWeightBuilder:
    """Builds model weights for PostgreSQL params database.

    Example usage:
        builder = BanditWeightBuilder(user_uuid, "rock")
        builder.with_version(5).build(db_managers)
    """

    def __init__(self, user_uuid: UUID, theme: str):
        self.user_uuid = user_uuid
        self.theme = theme
        self.version = 0
        # Initialize with identity matrix + small ridge regularization
        self.weights = np.eye(NUMB_FEATURES) * 1.0
        self.biases = np.zeros(NUMB_FEATURES)
        self.weights_inv = np.eye(NUMB_FEATURES)
        self.updates_since_recompute = 0

    def with_version(self, version: int) -> "BanditWeightBuilder":
        """Set the version number.

        Args:
            version: Version number for optimistic locking

        Returns:
            self for method chaining
        """
        self.version = version
        return self

    def with_weights(
        self, weights: np.ndarray, biases: np.ndarray, weights_inv: np.ndarray
    ) -> "BanditWeightBuilder":
        """Set custom weights, biases, and inverse matrix.

        Args:
            weights: 12x12 weight matrix
            biases: 12-element bias vector
            weights_inv: 12x12 inverse weight matrix

        Returns:
            self for method chaining
        """
        self.weights = weights
        self.biases = biases
        self.weights_inv = weights_inv
        return self

    def with_updates_since_recompute(self, count: int) -> "BanditWeightBuilder":
        """Set the number of updates since last recompute.

        Args:
            count: Number of Sherman-Morrison updates since last full recompute

        Returns:
            self for method chaining
        """
        self.updates_since_recompute = count
        return self

    def build(self, db_managers: DBManagers):
        """Insert the model weights into PostgreSQL.

        Args:
            db_managers: DatabaseManagers instance with params connection
        """
        query = text(
            f"INSERT INTO {db_managers._config.bandit_params_table} "
            "(user_uuid, theme, version, weights, biases, weights_inv, updates_since_recompute) "
            "VALUES (:uuid, :theme, :version, :weights, :biases, :weights_inv, :updates)"
        )

        with db_managers._storage_engine.connect() as conn:
            with conn.begin():
                conn.execute(
                    query,
                    {
                        "uuid": str(self.user_uuid),
                        "theme": self.theme,
                        "version": self.version,
                        "weights": json.dumps(self.weights.tolist()),
                        "biases": json.dumps(self.biases.tolist()),
                        "weights_inv": json.dumps(self.weights_inv.tolist()),
                        "updates": self.updates_since_recompute,
                    },
                )


class MusicThemeBuilder:
    """Builds music-theme associations for ClickHouse warehouse database.

    Example usage:
        builder = MusicThemeBuilder(music_uuid, "rock")
        builder.with_stats(views=100, successes=80).build(db_managers)
    """

    def __init__(self, music_uuid: UUID, theme: str):
        self.music_uuid = music_uuid
        self.theme = theme
        self.views = 0
        self.successes = 0

    def with_stats(self, views: int = 0, successes: int = 0) -> "MusicThemeBuilder":
        """Set view and success statistics.

        Args:
            views: Number of times this theme was viewed
            successes: Number of successful interactions with this theme

        Returns:
            self for method chaining
        """
        self.views = views
        self.successes = successes
        return self

    def build(self, db_managers: DBManagers):
        """Insert the music-theme association into ClickHouse.

        Args:
            db_managers: DatabaseManagers instance with warehouse connection
        """
        query = text(
            f"INSERT INTO {db_managers._config.theme_catalog_table} "
            "(music_uuid, theme, views, successes) "
            "VALUES (:music_uuid, :theme, :views, :successes)"
        )

        with db_managers._warehouse_engine.connect() as conn:
            with conn.begin():
                conn.execute(
                    query,
                    {
                        "music_uuid": str(self.music_uuid),
                        "theme": self.theme,
                        "views": self.views,
                        "successes": self.successes,
                    },
                )


def create_test_user_with_themes(
    db_managers: DBManagers, user_uuid: UUID, themes: List[str]
) -> None:
    """Helper function to create a test user with multiple themes.

    Args:
        db_managers: DatabaseManagers instance
        user_uuid: UUID of the test user
        themes: List of theme names to create
    """
    for theme in themes:
        # Create features in ClickHouse
        ThemeFeatureBuilder(user_uuid, theme).build(db_managers)

        # Create weights in PostgreSQL (optional, will be initialized if missing)
        # Uncomment if you want pre-existing weights:
        # BanditWeightBuilder(user_uuid, theme).build(db_managers)
