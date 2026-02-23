import json
from typing import Dict, List

import numpy as np
from pydantic import UUID4
from sqlalchemy import create_engine, text

from src.di.config import Config
from src.models.linucb import ArmResultLinUCB, LinUCB

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
    def __init__(self, config: Config):
        self._config = config
        self._storage_engine = create_engine(config.db_params_string)
        self._warehouse_engine = create_engine(config.db_warehouse_string)

    def get_input_data(self, uuid: UUID4) -> Dict[str, np.ndarray]:
        cols = ", ".join(_FEATURE_COLS)
        query = text(
            f"SELECT theme, {cols}"
            f" FROM {self._config.bandit_data_table}"
            f" WHERE user_uuid = :uuid"
            f" ORDER BY theme"
        )
        with self._warehouse_engine.connect() as conn:
            rows = conn.execute(query, {"uuid": str(uuid)}).fetchall()

        return {row[0]: np.array(row[1:], dtype=np.float64) for row in rows}

    def get_weight_bias(self, uuid: UUID4) -> Dict[str, ArmResultLinUCB]:
        query = text(
            f"SELECT theme, weights, biases, version"
            f" FROM {self._config.bandit_params_table}"
            f" WHERE user_uuid = :uuid"
            f" ORDER BY theme"
        )
        with self._storage_engine.connect() as conn:
            rows = conn.execute(query, {"uuid": str(uuid)}).fetchall()

        arms: Dict[str, ArmResultLinUCB] = {}
        for theme, weights_json, biases_json, version in rows:
            arms[theme] = (
                ArmResultLinUCB(
                    Theme=theme,
                    Version=int(version),
                    Weights=np.array(json.loads(weights_json), dtype=np.float64),
                    Biases=np.array(json.loads(biases_json), dtype=np.float64),
                )
            )
        return arms


    def get_weight_bias_for_one(self, uuid: UUID4, theme: str) -> ArmResultLinUCB:
        query = text(
            f"SELECT weights, biases, version"
            f" FROM {self._config.bandit_params_table}"
            f" WHERE user_uuid = :uuid"
            f" AND theme = :theme"
        )
        with self._storage_engine.connect() as conn:
            rows = conn.execute(query, {"uuid": str(uuid), "theme": theme}).fetchall()

        if len(rows) == 0:
            return LinUCB.get_new_arm_result(theme, NUMB_FEATURES)

        weights_json, biases_json, version = rows[0]
        return (
            ArmResultLinUCB(
                Theme=theme,
                Version=int(version),
                Weights=np.array(json.loads(weights_json), dtype=np.float64),
                Biases=np.array(json.loads(biases_json), dtype=np.float64),
            )
        )

    def update_weight_bias(
        self,
        uuid: UUID4,
        theme: str,
        weight: np.ndarray,
        bias: np.ndarray,
        latest_version: int,
    ) -> bool:
        query = text(
            f"UPDATE {self._config.bandit_params_table}"
            " SET weights = :weights, biases = :biases, version = :new_version"
            " WHERE user_uuid = :uuid AND theme = :theme AND version = :latest_version"
        )

        with self._storage_engine.connect() as conn:
            with conn.begin():
                result = conn.execute(
                    query,
                    {
                        "weights": json.dumps(weight.tolist()),
                        "biases": json.dumps(bias.tolist()),
                        "new_version": latest_version + 1,
                        "uuid": str(uuid),
                        "theme": theme,
                        "latest_version": latest_version,
                    },
                )

        return result.rowcount() > 0


if __name__ == "__main__":
    _config = Config()
    _db = DBManagers(_config)
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