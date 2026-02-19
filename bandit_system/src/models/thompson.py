import numpy as np
from pydantic import BaseModel
from typing import List

class ArmResultThompson(BaseModel):
    ArmName: str
    Success: int
    Failures: int

class Thompson:
    @staticmethod
    def predict(data: List[ArmResultThompson]) -> int:
        samples = [np.random.beta(arm.Success + 1, arm.Failures + 1) for arm in data]
        best_arm = np.argmax(samples)

        return int(best_arm)

    @staticmethod
    def update(
        arms: List[ArmResultThompson], action: int, reward: float
    ) -> List[ArmResultThompson]:
        reward = max(0.0, min(reward, 1.0))

        arm = arms[action]
        arm.Success += reward
        arm.Failures += 1 - reward
        arms[action] = arm

        return arms