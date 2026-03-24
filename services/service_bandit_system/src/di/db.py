import json
from typing import Any, Dict

import numpy as np
import structlog
from pydantic import UUID4
from sqlalchemy import create_engine, text

from src.cache.themes import ThemesCache
from src.di.config import Config
from src.models.linucb import ArmResultLinUCB, LinUCB


def _ensure_deserialized(data: Any) -> Any:
    """Ensure data is deserialized from JSON string if needed.

    JSONB columns return Python objects directly, while Text columns return strings.
    This helper handles both cases.
    """
    if isinstance(data, str):
        return json.loads(data)
    return data

_FEATURE_COLS = [
    "f_user_theme_decay_impressions",
    "f_user_theme_decay_completion",
    "f_user_theme_full_play_rate",
    "f_user_theme_decay_like_rate",
    "f_user_decay_impressions",
    "f_user_decay_completion",
    "f_user_decay_like_rate",
    "f_theme_decay_impressions",
    "f_theme_decay_completion",
    "f_theme_decay_like_rate",
    "f_relative_completion",
    "f_relative_exposure",
]
NUMB_FEATURES = len(_FEATURE_COLS)


class DBManagers:
    def __init__(self, config: Config, bandit: LinUCB):
        self._config = config
        self._bandit = bandit
        self._storage_engine = create_engine(config.db_params_string)
        self._warehouse_engine = create_engine(config.db_warehouse_string)
        self._logger = structlog.get_logger("db_managers")
        self._themes_cache = ThemesCache(self._warehouse_engine, config.bandit_data_table)

    def get_input_data(self, uuid: UUID4) -> Dict[str, np.ndarray]:
        cols = ", ".join(_FEATURE_COLS)
        user_query = text(
            f"SELECT theme, {cols}"
            f" FROM {self._config.bandit_data_table}"
            f" WHERE user_uuid = :uuid"
        )
        with self._warehouse_engine.connect() as conn:
            user_rows = conn.execute(user_query, {"uuid": str(uuid)}).fetchall()

        user_themes = {row[0]: np.array(row[1:], dtype=np.float64) for row in user_rows}

        all_themes = {
            theme: np.zeros(NUMB_FEATURES, dtype=np.float64)
            for theme in self._themes_cache.get_all_themes()
        }

        result = {**all_themes, **user_themes}

        if len(user_themes) == 0:
            self._logger.info("cold_start", user_uuid=str(uuid), theme_count=len(result))

        return result

    def get_weight_bias(self, uuid: UUID4) -> Dict[str, ArmResultLinUCB]:
        query = text(
            f"SELECT theme, weights, biases, weights_inv, updates_since_recompute, version"
            f" FROM {self._config.bandit_params_table}"
            f" WHERE user_uuid = :uuid"
        )
        with self._storage_engine.connect() as conn:
            rows = conn.execute(query, {"uuid": str(uuid)}).fetchall()

        return {
            row[0]: self._build_arm_result(*row)
            for row in rows
        }

    def get_weight_bias_for_one(self, uuid: UUID4, theme: str) -> ArmResultLinUCB:
        query = text(
            f"SELECT weights, biases, weights_inv, updates_since_recompute, version"
            f" FROM {self._config.bandit_params_table}"
            f" WHERE user_uuid = :uuid AND theme = :theme"
        )
        with self._storage_engine.connect() as conn:
            rows = conn.execute(query, {"uuid": str(uuid), "theme": theme}).fetchall()

        if not rows:
            return self._bandit.get_new_arm_result(theme, NUMB_FEATURES)

        return self._build_arm_result(theme, *rows[0])

    def _build_arm_result(
        self, theme: str, weights_json, biases_json, weights_inv_json, updates_since_recompute, version
    ) -> ArmResultLinUCB:
        weights = np.array(_ensure_deserialized(weights_json), dtype=np.float64)
        biases = np.array(_ensure_deserialized(biases_json), dtype=np.float64)

        # Compute inverse if not stored
        if weights_inv_json:
            weights_inv = np.array(_ensure_deserialized(weights_inv_json), dtype=np.float64)
        else:
            weights_inv = np.linalg.inv(weights)
            updates_since_recompute = 0

        return ArmResultLinUCB(
            Theme=theme,
            Version=int(version),
            Weights=weights,
            Biases=biases,
            WeightsInv=weights_inv,
            UpdatesSinceRecompute=int(updates_since_recompute or 0),
        )

    def update_weight_bias(
        self,
        uuid: UUID4,
        theme: str,
        weight: np.ndarray,
        bias: np.ndarray,
        weight_inv: np.ndarray,
        updates_since_recompute: int,
        latest_version: int,
    ) -> bool:
        query = text(
            f"INSERT INTO {self._config.bandit_params_table}"
            " (user_uuid, theme, weights, biases, weights_inv, updates_since_recompute, version)"
            " VALUES (:uuid, :theme, CAST(:weights AS jsonb), CAST(:biases AS jsonb), CAST(:weights_inv AS jsonb), :updates_since_recompute, :new_version)"
            " ON CONFLICT (user_uuid, theme)"
            " DO UPDATE SET"
            "   weights = CAST(:weights AS jsonb),"
            "   biases = CAST(:biases AS jsonb),"
            "   weights_inv = CAST(:weights_inv AS jsonb),"
            "   updates_since_recompute = :updates_since_recompute,"
            "   version = :new_version"
            f" WHERE {self._config.bandit_params_table}.version = :latest_version"
        )

        with self._storage_engine.connect() as conn:
            with conn.begin():
                result = conn.execute(
                    query,
                    {
                        "weights": json.dumps(weight.tolist()),
                        "biases": json.dumps(bias.tolist()),
                        "weights_inv": json.dumps(weight_inv.tolist()),
                        "updates_since_recompute": updates_since_recompute,
                        "new_version": latest_version + 1,
                        "uuid": str(uuid),
                        "theme": theme,
                        "latest_version": latest_version,
                    },
                )

        return result.rowcount > 0


if __name__ == "__main__":
    _config = Config.create()
    _logger = structlog.get_logger("db_test")
    _bandit = LinUCB(_config, _logger)
    _db = DBManagers(_config, _bandit)

    _theme_dict = _db.get_input_data(UUID4("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"))

    for key in _theme_dict:
        print(key)
        print(_theme_dict[key])
        print(len(_theme_dict[key]))
        print("")

    _weights_dict = _db.get_weight_bias(UUID4("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"))

    for key in _weights_dict:
        print(key)
        print(_weights_dict[key])
