import os
from dataclasses import dataclass

from dotenv import load_dotenv


def get_env_required(key: str) -> str:
    """Get required environment variable or raise an error if not set."""
    value = os.getenv(key)
    if not value:
        raise ValueError(f"{key} environment variable is not set")
    return value

@dataclass
class Config:
    db_warehouse_string: str
    bandit_data_table: str
    theme_catalog_table: str
    db_params_string: str
    bandit_params_table: str
    alpha: float
    ridge_lambda: float
    sherman_morrison_recompute_interval: int
    sherman_morrison_divergence_threshold: float
    max_retries: int
    initial_backoff_ms: float


    @classmethod
    def create(cls) -> "Config":
        load_dotenv()

        # Load required environment variables
        db_warehouse_string = get_env_required("DB_CONNECTION_STRING_WAREHOUSE")
        db_params_string = get_env_required("DB_CONNECTION_STRING_PARAMS")

        # Load optional environment variables with defaults
        bandit_data_table = os.getenv("WAREHOUSE_BANDIT_DATA_TABLE", "bandit_input_per_theme")
        theme_catalog_table = os.getenv("WAREHOUSE_THEME_CATALOG_TABLE", "music_theme")
        bandit_params_table = os.getenv("BANDIT_DATA_TABLE", "bandit_data")

        alpha = float(os.getenv("BANDIT_ALPHA", "0.5"))
        ridge_lambda = float(os.getenv("BANDIT_RIDGE_LAMBDA", "1.0"))
        sherman_morrison_recompute_interval = int(os.getenv("SHERMAN_MORRISON_RECOMPUTE_INTERVAL", "100"))
        sherman_morrison_divergence_threshold = float(os.getenv("SHERMAN_MORRISON_DIVERGENCE_THRESHOLD", "1e-6"))

        max_retries = int(os.getenv("MAX_RETRIES", "3"))
        initial_backoff_ms = float(os.getenv("BACKOFF_MS", "100"))

        return cls(
            db_warehouse_string=db_warehouse_string,
            bandit_data_table=bandit_data_table,
            theme_catalog_table=theme_catalog_table,
            db_params_string=db_params_string,
            bandit_params_table=bandit_params_table,
            alpha=alpha,
            ridge_lambda=ridge_lambda,
            sherman_morrison_recompute_interval=sherman_morrison_recompute_interval,
            sherman_morrison_divergence_threshold=sherman_morrison_divergence_threshold,
            max_retries=max_retries,
            initial_backoff_ms=initial_backoff_ms,
        )
