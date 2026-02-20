import numpy as np
import structlog
from pydantic.v1 import UUID4
from typing import Dict, List, Tuple

from src.di.config import Config
from src.di.db import DBManagers, NUMB_FEATURES
from src.models.linucb import ArmResultLinUCB, LinUCB


class BanditHandler:
    def __init__(self, config:Config, db:DBManagers, logger: structlog.BoundLogger):
        self._config = config
        self._db = db
        self.logger = logger

    def predict(self, uuid: UUID4) -> Tuple[str, np.ndarray]:
        input_data : Dict[str, np.ndarray] = self._db.get_input_data(uuid)
        weight_bias : Dict[str, ArmResultLinUCB] = self._db.get_weight_bias(uuid)

        if len(input_data) == 0:
            self.logger.error("No themes exist in DB")
            raise RuntimeError("No themes exist in DB")

        to_user_features : List[np.ndarray] = []
        to_use_arm_result : List[ArmResultLinUCB] = []

        for key in sorted(input_data.keys()):
            to_user_features.append(input_data[key])
            found_result = weight_bias.pop(key, None)

            if found_result is None:
                found_result = LinUCB.get_new_arm_result(key, NUMB_FEATURES)
            elif found_result.Weights.shape != (NUMB_FEATURES,NUMB_FEATURES) or found_result.Biases.shape != (NUMB_FEATURES,):
                self.logger.error(
                        "deleted/added features WITHOUT modifying weight bias",
                )
                raise RuntimeError("deleted/added features WITHOUT modifying weight bias")

            to_use_arm_result.append(found_result)


        if len(weight_bias) != 0:
            self.logger.warning(
                    "useless data is still in db",
            )

        chosen_index = LinUCB.predict(to_use_arm_result, to_user_features)
        if chosen_index is None:
            self.logger.error(
                    "unknown error"
            )
            raise RuntimeError("unknown error")
        return to_use_arm_result[chosen_index].Theme, to_user_features[chosen_index]


    def update(self, uuid: UUID4, reward: float, theme: str, features: np.ndarray):
        updated = False

        while not updated:
            result = self._db.get_weight_bias_for_one(uuid, theme)
            arm = LinUCB.update(result, features, reward)
            updated = self._db.update_weight_bias(uuid, theme, arm.Weights, arm.Biases, arm.Version)
