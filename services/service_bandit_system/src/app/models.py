from pydantic import BaseModel, Field, UUID4
from typing import List


class PredictRequest(BaseModel):
    user_uuid: UUID4 = Field(..., description="UUID of the user")


class PredictResponse(BaseModel):
    theme: str = Field(..., description="Selected theme")
    features: List[float] = Field(..., description="Feature vector used for prediction")


class UpdateRequest(BaseModel):
    user_uuid: UUID4 = Field(..., description="UUID of the user")
    theme: str = Field(..., description="Theme that was shown")
    reward: float = Field(..., ge=0.0, le=1.0, description="Reward value (0-1)")
    features: List[float] = Field(..., description="Feature vector that was used")


class UpdateResponse(BaseModel):
    success: bool = Field(..., description="Whether the update succeeded")


class HealthResponse(BaseModel):
    status: str = Field(..., description="Service status")
    service: str = Field(..., description="Service name")
