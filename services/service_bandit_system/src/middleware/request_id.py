import uuid
from contextvars import ContextVar
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import Response

# Context variable to store request ID across the request lifecycle
request_id_context: ContextVar[str] = ContextVar("request_id", default="")


def generate_request_id() -> str:
    """Generate a new UUID for the request."""
    return str(uuid.uuid4())


def get_request_id() -> str:
    """Get the current request ID from context."""
    return request_id_context.get()


class RequestIDMiddleware(BaseHTTPMiddleware):
    """
    Middleware that extracts or generates a request ID for each request.
    The request ID is stored in a context variable and added to response headers.
    """

    async def dispatch(self, request: Request, call_next) -> Response:
        # Get request ID from header or generate a new one
        request_id = request.headers.get("X-Request-ID", generate_request_id())

        # Store in context variable for access throughout the request
        request_id_context.set(request_id)

        # Process the request
        response = await call_next(request)

        # Add request ID to response headers
        response.headers["X-Request-ID"] = request_id

        return response
