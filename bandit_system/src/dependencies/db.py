from typing import Any

from pydantic import UUID4
from sqlalchemy import create_engine, CursorResult

from src.dependencies.config import Config


class DB:
    def __init__(self, config: Config):
        self.engine = create_engine(config.db_connection_string)

    def get_weight_bias(self, uuid:UUID4) -> CursorResult[Any]:
        with self.engine.connect() as conn:
            answer = conn.execute(f"SELECT * FROM {self.config.bandit_data_table} WHERE user_uuid = {uuid}")

        return answer