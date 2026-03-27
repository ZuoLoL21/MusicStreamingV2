"""
Prometheus metrics for the Bandit System service.

This module provides metrics tracking for:
- HTTP requests
- Bandit predictions
- Model updates
- Database queries
"""

from prometheus_client import Counter, Histogram, Gauge
import time
from typing import Callable, Any

http_requests_total = Counter(
    "http_requests_total",
    "Total HTTP requests",
    ["method", "endpoint", "status"],
)

http_request_duration_seconds = Histogram(
    "http_request_duration_seconds",
    "HTTP request latency",
    ["method", "endpoint"],
    buckets=(0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0, 7.5, 10.0),
)

bandit_predictions_total = Counter(
    "bandit_predictions_total",
    "Total bandit predictions made",
    ["theme"],
)

bandit_updates_total = Counter(
    "bandit_updates_total",
    "Total bandit model updates",
    ["theme"],
)

bandit_prediction_duration_seconds = Histogram(
    "bandit_prediction_duration_seconds",
    "Time to generate a prediction",
    buckets=(0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0),
)

bandit_update_duration_seconds = Histogram(
    "bandit_update_duration_seconds",
    "Time to update the model",
    buckets=(0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0),
)

bandit_errors_total = Counter(
    "bandit_errors_total",
    "Total errors in bandit operations",
    ["operation", "error_type"],
)

db_queries_total = Counter(
    "db_queries_total",
    "Total database queries",
    ["operation", "status"],
)

db_query_duration_seconds = Histogram(
    "db_query_duration_seconds",
    "Database query latency",
    ["operation"],
    buckets=(0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0),
)

bandit_model_parameters = Gauge(
    "bandit_model_parameters",
    "Current bandit model parameters",
    ["parameter"],
)


def track_prediction(theme: str) -> None:
    """Track a successful prediction."""
    bandit_predictions_total.labels(theme=theme).inc()


def track_update(theme: str) -> None:
    """Track a successful model update."""
    bandit_updates_total.labels(theme=theme).inc()


def track_error(operation: str, error_type: str) -> None:
    """Track an error occurrence."""
    bandit_errors_total.labels(operation=operation, error_type=error_type).inc()


def track_db_query(operation: str, func: Callable[[], Any]) -> Any:
    """
    Track a database query execution.

    Args:
        operation: Name of the operation (e.g., "get_user_params", "update_theme")
        func: Function that executes the query

    Returns:
        The result of func()
    """
    start = time.time()
    try:
        result = func()
        duration = time.time() - start
        db_queries_total.labels(operation=operation, status="success").inc()
        db_query_duration_seconds.labels(operation=operation).observe(duration)
        return result
    except Exception as e:
        duration = time.time() - start
        db_queries_total.labels(operation=operation, status="error").inc()
        db_query_duration_seconds.labels(operation=operation).observe(duration)
        raise e


def update_model_parameters(alpha: float, ridge_lambda: float) -> None:
    """Update the model parameter gauges."""
    bandit_model_parameters.labels(parameter="alpha").set(alpha)
    bandit_model_parameters.labels(parameter="ridge_lambda").set(ridge_lambda)
