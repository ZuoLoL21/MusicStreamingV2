import numpy as np
from pydantic import BaseModel
from typing import List

class ArmResult(BaseModel):
    ArmName: str
    Success: int
    Failures: int


def predict(data: List[ArmResult]) -> int:
    samples = [np.random.beta(arm.Success + 1, arm.Failures + 1) for arm in data]
    best_arm = np.argmax(samples)

    return int(best_arm)