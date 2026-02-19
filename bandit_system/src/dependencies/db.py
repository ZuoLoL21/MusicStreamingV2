from typing import Any

import numpy as np
from pydantic import UUID4
from sqlalchemy import create_engine, CursorResult

from src.dependencies.config import Config


class DBManagers:
    def __init__(self, config: Config):
        self.storage_engine = create_engine(config.db_params_string)
        self.warehouse_engine = create_engine(config.db_warehouse_string)

    def get_input_data(self, uuid: UUID4) -> CursorResult[Any]:
        with self.warehouse_engine.connect() as conn:
            answer = conn.execute(
                f"SELECT * FROM {self.config.bandit_data_table} WHERE user_uuid = {uuid}"
            )

        return answer

    def get_weight_bias(self, uuid: UUID4) -> CursorResult[Any]:
        with self.storage_engine.connect() as conn:
            answer = conn.execute(
                f"SELECT * FROM {self.config.bandit_params_table} WHERE user_uuid = {uuid}"
            )

        return answer

    def update_weight_bias(
        self,
        uuid: UUID4,
        weight: np.ndarray,
        bias: np.ndarray,
        latest_version: np.ndarray,
    ) -> bool:
        weight_json = weight.tolist()
        bias_json = bias.tolist()

        with self.storage_engine.connect() as conn:
            answer = conn.execute(f"""
                UPDATE {self.config.bandit_params_table}
                SET weight = {weight_json}, bias = {bias_json}, latest_version = {latest_version + 1}
                WHERE user_uuid = {uuid} AND version = {latest_version}
            """)

        return answer.rowcount > 0
