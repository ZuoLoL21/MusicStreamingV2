import numpy as np
from fastapi import Request

from src.di.db import NUMB_FEATURES
from src.services.bandit import BanditHandler
from src.app.state import AppState


def get_app_state(request: Request) -> AppState:
    if not hasattr(request.app.state, "app_state"):
        raise RuntimeError("Application state not initialized")
    return request.app.state.app_state


def get_handler(request: Request) -> BanditHandler:
    app_state = get_app_state(request)
    return app_state.handler

def is_features_valid(features: np.ndarray) -> bool:
    if len(features) != NUMB_FEATURES:
        return False
    return True