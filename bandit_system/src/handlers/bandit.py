import numpy as np
import structlog
from pydantic.v1 import UUID4
from typing import Dict, List

from src.di.config import Config
from src.di.db import DBManagers, NUMB_FEATURES
from src.models.linucb import ArmResultLinUCB, LinUCB


class BanditHandler:
    def __init__(self, config:Config, db:DBManagers, logger: structlog.BoundLogger):
        self._config = config
        self._db = db
        self._bandit = LinUCB
        self.logger = logger

    def predict(self, uuid: UUID4) -> str:
        input_data : Dict[str, np.ndarray] = self._db.get_input_data(uuid)
        weight_bias : Dict[str, ArmResultLinUCB] = self._db.get_weight_bias(uuid)

        to_use_input : List[np.ndarray] = []
        to_use_arm_result : List[ArmResultLinUCB] = []

        for key in sorted(input_data.keys()):
            to_use_input.append(input_data[key])
            found_result = weight_bias.pop(key, None)

            if found_result is None:
                found_result = self._bandit.get_new_arm_result(key, NUMB_FEATURES)
            elif found_result.Weights.shape != (NUMB_FEATURES,NUMB_FEATURES) and found_result.Biases.shape != (NUMB_FEATURES,NUMB_FEATURES,):
                self.logger.error(
                        "deleted/added features WITHOUT modifying weight bias",
                )
                raise RuntimeError("deleted/added features WITHOUT modifying weight bias")

            to_use_arm_result.append(found_result)


        if len(weight_bias) != 0:
            self.logger.warning(
                    "useless data is still in db",
            )

        chosen_index = self._bandit.predict(to_use_arm_result, to_use_input)
        return to_use_arm_result[chosen_index].Theme


    def update(self, uuid: UUID4, reward: float, action: str):
        pass
