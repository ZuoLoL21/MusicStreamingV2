import numpy as np
import structlog
from pydantic import BaseModel, ConfigDict
from typing import Tuple, List, Optional

from src.di.config import Config


class ArmResultLinUCB(BaseModel):
    model_config = ConfigDict(arbitrary_types_allowed=True)

    Theme: str
    Version: int

    Weights: np.ndarray
    Biases: np.ndarray


class LinUCB:
    def __init__(self, config:Config, logger: structlog.BoundLogger):
        self._config = config
        self._logger = logger

    def get_basic(self, dim: int) -> Tuple[np.ndarray, np.ndarray]:
        return self._config.ridge_lambda * np.identity(dim), np.zeros(dim)

    def get_new_arm_result(self, theme: str, dim: int) -> ArmResultLinUCB:
        weight, bias = self.get_basic(dim)
        return ArmResultLinUCB(Theme=theme, Version=0, Weights=weight, Biases=bias)

    def _compute_weight(self, arm: ArmResultLinUCB, features: np.ndarray) -> float:
        A = arm.Weights
        B = arm.Biases
        A_1 = np.linalg.inv(A)

        theta = A_1 @ B
        mean = theta @ features
        std = self._config.alpha * np.sqrt(features.T @ A_1 @ features)

        return mean + std.item()

    def predict(self, arms: List[ArmResultLinUCB], features: List[np.ndarray]) -> Optional[int]:
        if len(arms) != len(features):
            return None

        weights = [
            self._compute_weight(arm, feature)
            for arm, feature in zip(arms, features)
        ]
        return int(np.argmax(weights))

    def update(
        self, arm: ArmResultLinUCB, features: np.ndarray, reward: float
    ) -> ArmResultLinUCB:
        reward = max(0.0, min(reward, 1.0))

        weightAdjustment = features @ features.T
        biasAdjustment = reward * features

        arm.Weights += weightAdjustment
        arm.Biases += biasAdjustment

        return arm

