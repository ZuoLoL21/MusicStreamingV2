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
    WeightsInv: np.ndarray

    UpdatesSinceRecompute: int = 0


class LinUCB:
    def __init__(self, config: Config, logger: structlog.BoundLogger):
        self._config = config
        self._logger = logger

    def get_basic(self, dim: int) -> Tuple[np.ndarray, np.ndarray, np.ndarray]:
        return (
            self._config.ridge_lambda * np.identity(dim),
            np.zeros(dim),
            (1.0 / self._config.ridge_lambda) * np.identity(dim),
        )

    def get_new_arm_result(self, theme: str, dim: int) -> ArmResultLinUCB:
        weight, bias, inverse_weight = self.get_basic(dim)
        return ArmResultLinUCB(
            Theme=theme,
            Version=0,
            Weights=weight,
            Biases=bias,
            WeightsInv=inverse_weight,
            UpdatesSinceRecompute=0,
        )

    def _compute_weight(self, arm: ArmResultLinUCB, features: np.ndarray) -> float:
        A_inv = arm.WeightsInv
        B = arm.Biases

        theta = A_inv @ B
        mean = theta @ features
        std = self._config.alpha * np.sqrt(features.T @ A_inv @ features)

        return mean + std.item()

    def predict(
        self, arms: List[ArmResultLinUCB], features: List[np.ndarray]
    ) -> Optional[int]:
        if len(arms) != len(features):
            return None

        weights = [
            self._compute_weight(arm, feature) for arm, feature in zip(arms, features)
        ]
        return int(np.argmax(weights))

    def update(
        self, arm: ArmResultLinUCB, features: np.ndarray, reward: float
    ) -> ArmResultLinUCB:
        reward = max(0.0, min(reward, 1.0))

        # Normal
        weightAdjustment = np.outer(features, features)
        biasAdjustment = reward * features
        arm.Weights += weightAdjustment
        arm.Biases += biasAdjustment

        # Sherman-Morrison update for A_inv
        A_inv_x = arm.WeightsInv @ features
        denominator = 1.0 + features.T @ A_inv_x
        arm.WeightsInv -= np.outer(A_inv_x, A_inv_x) / denominator
        arm.UpdatesSinceRecompute += 1

        # Divergence check
        if (
            arm.UpdatesSinceRecompute
            >= self._config.sherman_morrison_recompute_interval
        ):
            divergence = _check_divergence(arm.Weights, arm.WeightsInv)

            if divergence > self._config.sherman_morrison_divergence_threshold:
                self._logger.warning(
                    "Sherman-Morrison divergence exceeded threshold, recomputing inverse",
                    theme=arm.Theme,
                    divergence=divergence,
                    threshold=self._config.sherman_morrison_divergence_threshold,
                    updates_since_recompute=arm.UpdatesSinceRecompute,
                )
                arm.WeightsInv = np.linalg.inv(arm.Weights)
                arm.UpdatesSinceRecompute = 0
            else:
                self._logger.debug(
                    "Sherman-Morrison divergence check passed",
                    theme=arm.Theme,
                    divergence=divergence,
                    updates_since_recompute=arm.UpdatesSinceRecompute,
                )
                arm.UpdatesSinceRecompute = 0
        return arm


def _check_divergence(A: np.ndarray, A_inv: np.ndarray) -> float:
    identity = np.eye(A.shape[0])
    product = A @ A_inv
    divergence = np.linalg.norm(identity - product, ord="fro")
    return float(divergence)
