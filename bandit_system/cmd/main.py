import logging
from contextlib import asynccontextmanager

import structlog
from fastapi import FastAPI

from src.app import AppState, router
from src.middleware import RequestIDMiddleware, LoggingMiddleware

# Configure logging
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
structlog.configure(
    processors=[
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.JSONRenderer(),
    ],
    wrapper_class=structlog.BoundLogger,
    context_class=dict,
    logger_factory=structlog.PrintLoggerFactory(),
    cache_logger_on_first_use=True,
)
logger = structlog.get_logger("bandit_system")


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Bandit system starting up")

    app_state = AppState.create(logger)
    app.state.app_state = app_state

    logger.info(
        "Bandit system ready",
        config={
            "alpha": app_state.config.alpha,
            "ridge_lambda": app_state.config.ridge_lambda,
        },
    )

    yield

    logger.info("Bandit system shutting down")


_app = FastAPI(
    title="Bandit System",
    description="LinUCB contextual bandit for personalized music theme recommendations",
    version="0.1.0",
    lifespan=lifespan,
)

# Middleware
_app.add_middleware(LoggingMiddleware, logger=logger)
_app.add_middleware(RequestIDMiddleware)

# Include routers
_app.include_router(router)
