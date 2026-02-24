import os
from dataclasses import dataclass

from dotenv import load_dotenv

@dataclass
class Config:
    db_warehouse_string: str
    bandit_data_table: str
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

        db_warehouse_string = os.getenv("DB_CONNECTION_STRING_WAREHOUSE")
        bandit_data_table = os.getenv("WAREHOUSE_BANDIT_DATA_TABLE")
        db_params_string = os.getenv("DB_CONNECTION_STRING_PARAMS")
        bandit_params_table = os.getenv("BANDIT_DATA_TABLE")
        alpha = float(os.getenv("BANDIT_ALPHA", "0.5"))
        ridge_lambda = float(os.getenv("BANDIT_RIDGE_LAMBDA", "1.0"))
        sherman_morrison_recompute_interval = int(
            os.getenv("SHERMAN_MORRISON_RECOMPUTE_INTERVAL", "100")
        )
        sherman_morrison_divergence_threshold = float(
            os.getenv("SHERMAN_MORRISON_DIVERGENCE_THRESHOLD", "1e-6")
        )
        max_retries = 3
        initial_backoff_ms = 100

        return cls(
            db_warehouse_string=db_warehouse_string,
            bandit_data_table=bandit_data_table,
            db_params_string=db_params_string,
            bandit_params_table=bandit_params_table,
            alpha=alpha,
            ridge_lambda=ridge_lambda,
            sherman_morrison_recompute_interval=sherman_morrison_recompute_interval,
            sherman_morrison_divergence_threshold=sherman_morrison_divergence_threshold,
            max_retries=max_retries,
            initial_backoff_ms=initial_backoff_ms,
        )
