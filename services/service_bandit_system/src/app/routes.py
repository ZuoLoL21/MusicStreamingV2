import numpy as np
import structlog
import time
from fastapi import APIRouter, HTTPException, Depends

from src.di.db import NUMB_FEATURES
from src.services.bandit import BanditHandler
from src.middleware import get_request_id
from src.app.models import (
    PredictRequest,
    PredictResponse,
    UpdateRequest,
    UpdateResponse,
    HealthResponse,
)
from src.app.helpers import get_handler, is_features_valid
from src import metrics

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
    start_time = time.time()
    try:
        theme, features = handler.predict(request.user_uuid)

        # Track metrics
        duration = time.time() - start_time
        metrics.bandit_prediction_duration_seconds.observe(duration)
        metrics.track_prediction(theme)

        logger.info(
            "prediction generated successfully",
            user_uuid=str(request.user_uuid),
            theme=theme,
            request_id=get_request_id(),
        )
        return PredictResponse(
            theme=theme,
            features=features.tolist(),
        )
    except HTTPException as e:
        metrics.track_error("predict", "http_exception")
        logger.error(
                "input are invalid",
                user_uuid=str(request.user_uuid),
                theme=request.theme,
                error=str(e),
                request_id=get_request_id(),
        )
        raise e
    except RuntimeError as e:
        metrics.track_error("predict", "runtime_error")
        logger.error(
            "prediction failed",
            user_uuid=str(request.user_uuid),
            error=str(e),
            request_id=get_request_id(),
        )
        raise HTTPException(status_code=500, detail=str(e))
    except Exception as e:
        metrics.track_error("predict", "unexpected_error")
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
    start_time = time.time()
    try:
        features_array = np.array(request.features, dtype=np.float64)
        if not is_features_valid(features_array):
            raise HTTPException(status_code=400, detail=f"Invalid features, must be of shape 1x{NUMB_FEATURES}")

        handler.update(
            request.user_uuid, request.reward, request.theme, features_array
        )

        # Track metrics
        duration = time.time() - start_time
        metrics.bandit_update_duration_seconds.observe(duration)
        metrics.track_update(request.theme)

        logger.info(
            "model updated successfully",
            user_uuid=str(request.user_uuid),
            theme=request.theme,
            reward=request.reward,
            request_id=get_request_id(),
        )
        return UpdateResponse(success=True)
    except HTTPException as e:
        metrics.track_error("update", "http_exception")
        logger.error(
            "input are invalid",
            user_uuid=str(request.user_uuid),
            theme=request.theme,
            error=str(e),
            request_id=get_request_id(),
        )
        raise e
    except RuntimeError as e:
        metrics.track_error("update", "runtime_error")
        logger.error(
            "update failed",
            user_uuid=str(request.user_uuid),
            theme=request.theme,
            error=str(e),
            request_id=get_request_id(),
        )
        raise HTTPException(status_code=500, detail=str(e))
    except Exception as e:
        metrics.track_error("update", "unexpected_error")
        logger.error(
            "unexpected error during update",
            user_uuid=str(request.user_uuid),
            theme=request.theme,
            error=str(e),
            exception_type=type(e).__name__,
            request_id=get_request_id(),
        )
        raise HTTPException(status_code=500, detail="Internal server error")
