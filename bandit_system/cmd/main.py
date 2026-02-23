import logging

import structlog
from fastapi import FastAPI
from src.middleware import RequestIDMiddleware, LoggingMiddleware

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
logger.info("Bandit system starting up")
app = FastAPI()

# Middleware
app.add_middleware(RequestIDMiddleware)
app.add_middleware(LoggingMiddleware, logger=logger)


@app.get("/")
async def root():
    return {"message": "Hello World"}
