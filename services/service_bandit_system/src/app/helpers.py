from fastapi import Request
from src.services.bandit import BanditHandler
from src.app.state import AppState


def get_app_state(request: Request) -> AppState:
    if not hasattr(request.app.state, "app_state"):
        raise RuntimeError("Application state not initialized")
    return request.app.state.app_state


def get_handler(request: Request) -> BanditHandler:
    app_state = get_app_state(request)
    return app_state.handler
