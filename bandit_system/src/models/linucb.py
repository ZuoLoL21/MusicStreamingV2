import numpy as np
from pydantic import BaseModel, ConfigDict
from typing import Tuple, List

from src.dependencies.config import Config


class ArmResultLinUCB(BaseModel):
    model_config = ConfigDict(arbitrary_types_allowed=True)

    ArmName: str
    Weights: np.ndarray
    Biases: np.ndarray


config = Config()


class LinUCB:
    @staticmethod
    def get_basic(dim: int) -> Tuple[np.ndarray, np.ndarray]:
        return np.identity(dim), np.zeros(dim)

    @staticmethod
    def _compute_weight(arm: ArmResultLinUCB, features: np.ndarray) -> float:
        A = arm.Weights
        B = arm.Biases
        A_1 = np.linalg.inv(A)

        theta = A_1 @ B
        ls = theta @ features
        weight = config.alpha * np.sqrt(features @ A_1 @ features)

        return ls + weight.item()

    @staticmethod
    def predict(arms: List[ArmResultLinUCB], features: np.ndarray) -> int:
        weights = [LinUCB._compute_weight(arm, features) for arm in arms]
        return int(np.argmax(weights))

    @staticmethod
    def update(
        arms: List[ArmResultLinUCB], features: np.ndarray, action: int, reward: float
    ) -> List[ArmResultLinUCB]:
        reward = max(0.0, min(reward, 1.0))

        weightAdjustment = features @ features.T
        biasAdjustment = reward * features

        arm = arms[action]
        arm.Weights += weightAdjustment
        arm.Biases += biasAdjustment
        arms[action] = arm

        return arms
