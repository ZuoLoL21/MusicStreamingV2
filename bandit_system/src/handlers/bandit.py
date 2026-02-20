from pydantic.v1 import UUID4

from src.di.config import Config
from src.di.db import DBManagers


class BanditHandler:
    def __init__(self, config:Config, db:DBManagers):
        self._config = config
        self._db = db

    def predict(self, uuid: UUID4) -> str:
        self._db.get_input_data(uuid)

    def update(self, uuid: UUID4, reward: float, action: str):
        pass
