import numpy as np
import structlog
from fastapi import APIRouter, HTTPException, Depends

from src.services.bandit import BanditHandler
from src.middleware import get_request_id
from src.app.models import (
    PredictRequest,
    PredictResponse,
    UpdateRequest,
    UpdateResponse,
    HealthResponse,
)
from src.app.helpers import get_handler

logger = structlog.get_logger("bandit_routes")

router = APIRouter(prefix="/api/v1", tags=["bandit"])


@router.get("/health", response_model=HealthResponse)
async def health():
    return HealthResponse(status="healthy", service="bandit-system")


@router.post("/predict", response_model=PredictResponse)
async def predict(
    request: PredictRequest,
    handler: BanditHandler = Depends(get_handler),
):
    """
    Predict the best theme for a user using LinUCB.

    This endpoint synchronously returns the predicted theme and features.
    """
    try:
        theme, features = handler.predct(request.user_uuid)
        return PredictResponse(
            theme=theme,
            features=features.tolist(),
        )
    except RuntimeError as e:
        logger.error(
            "prediction failed",
            user_uuid=str(request.user_uuid),
            error=str(e),
            request_id=get_request_id(),
        )
        raise HTTPException(status_code=500, detail=str(e))
    except Exception as e:
        logger.error(
            "unexpected error during prediction",
            user_uuid=str(request.user_uuid),
            error=str(e),
            exception_type=type(e).__name__,
            request_id=get_request_id(),
        )
        raise HTTPException(status_code=500, detail="Internal server error")


@router.post("/update", response_model=UpdateResponse, status_code=202)
async def update(
    request: UpdateRequest,
    handler: BanditHandler = Depends(get_handler),
):
    """
    Update the bandit model with feedback.

    This endpoint accepts the reward for a previously shown theme.
    Returns 202 Accepted as the update is processed synchronously but
    could be made async in the future with a task queue.
    """
    try:
        features_array = np.array(request.features, dtype=np.float64)
        handler.update(
            request.user_uuid, request.reward, request.theme, features_array
        )
        return UpdateResponse(success=True)
    except RuntimeError as e:
        logger.error(
            "update failed",
            user_uuid=str(request.user_uuid),
            theme=request.theme,
            error=str(e),
            request_id=get_request_id(),
        )
        raise HTTPException(status_code=500, detail=str(e))
    except Exception as e:
        logger.error(
            "unexpected error during update",
            user_uuid=str(request.user_uuid),
            theme=request.theme,
            error=str(e),
            exception_type=type(e).__name__,
            request_id=get_request_id(),
        )
        raise HTTPException(status_code=500, detail="Internal server error")
