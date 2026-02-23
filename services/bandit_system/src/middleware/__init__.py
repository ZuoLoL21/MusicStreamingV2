"""
Middleware package for the bandit system.

This package provides middleware for FastAPI applications, similar to
the middleware structure in the user_database Go service.
"""

from .logging import LoggingMiddleware, get_user_uuid, set_user_uuid
from .request_id import RequestIDMiddleware, get_request_id

__all__ = [
    "RequestIDMiddleware",
    "LoggingMiddleware",
    "get_request_id",
    "get_user_uuid",
    "set_user_uuid",
]
