import time
from contextvars import ContextVar
from typing import Optional

import structlog
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import Response

from .request_id import get_request_id

# Context variable to store user UUID if available from auth
user_uuid_context: ContextVar[Optional[str]] = ContextVar("user_uuid", default=None)


def get_user_uuid() -> Optional[str]:
    """Get the current user UUID from context."""
    return user_uuid_context.get()


def set_user_uuid(user_uuid: str) -> None:
    """Set the user UUID in context."""
    user_uuid_context.set(user_uuid)


def get_ip(request: Request) -> str:
    """Extract the client IP address from the request."""
    xff = request.headers.get("X-Forwarded-For")
    if xff:
        return xff.split(",")[0].strip()
    return request.client.host if request.client else "unknown"


class LoggingMiddleware(BaseHTTPMiddleware):
    """
    Middleware that logs HTTP requests with details including:
    - Method, path, and route pattern
    - Status code and response size
    - Request duration
    - Client IP address
    - Request ID and user UUID (if available)

    Also handles panic recovery with logging.
    """

    def __init__(self, app, logger: structlog.BoundLogger):
        super().__init__(app)
        self.logger = logger

    async def dispatch(self, request: Request, call_next) -> Response:
        start_time = time.time()
        status_code = 500  # Default to 500 in case of uncaught exception
        response_size = 0

        try:
            # Process the request
            response = await call_next(request)
            status_code = response.status_code

            # Calculate response size if available
            if hasattr(response, "body"):
                response_size = len(response.body) if response.body else 0

            return response

        except Exception as exc:
            # Log panic/exception
            duration = time.time() - start_time
            route_path = request.url.path

            self.logger.error(
                "panic recovered",
                panic=str(exc),
                exception_type=type(exc).__name__,
                method=request.method,
                path=route_path,
                route=route_path,  # In FastAPI, we'd need route matching to get template
                remote_addr=get_ip(request),
                duration_ms=int(duration * 1000),
                request_id=get_request_id(),
                user_uuid=get_user_uuid(),
            )
            raise

        finally:
            # Log the request
            duration = time.time() - start_time
            route_path = request.url.path

            # Log at appropriate level based on status code
            if status_code >= 500:
                self.logger.error(
                    "http request",
                    method=request.method,
                    path=route_path,
                    route=route_path,
                    status=status_code,
                    bytes=response_size,
                    duration_ms=int(duration * 1000),
                    remote_addr=get_ip(request),
                    request_id=get_request_id(),
                    user_uuid=get_user_uuid(),
                )
            elif status_code >= 400:
                self.logger.warning(
                    "http request",
                    method=request.method,
                    path=route_path,
                    route=route_path,
                    status=status_code,
                    bytes=response_size,
                    duration_ms=int(duration * 1000),
                    remote_addr=get_ip(request),
                    request_id=get_request_id(),
                    user_uuid=get_user_uuid(),
                )
            else:
                self.logger.info(
                    "http request",
                    method=request.method,
                    path=route_path,
                    route=route_path,
                    status=status_code,
                    bytes=response_size,
                    duration_ms=int(duration * 1000),
                    remote_addr=get_ip(request),
                    request_id=get_request_id(),
                    user_uuid=get_user_uuid(),
                )