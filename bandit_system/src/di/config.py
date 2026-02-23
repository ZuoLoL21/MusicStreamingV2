import os

from dotenv import load_dotenv


class Config:
    def __init__(self):
        load_dotenv()

        self.db_warehouse_string = os.getenv("DB_CONNECTION_STRING_WAREHOUSE")
        self.bandit_data_table = os.getenv("WAREHOUSE_BANDIT_DATA_TABLE")

        self.db_params_string = os.getenv("DB_CONNECTION_STRING_PARAMS")
        self.bandit_params_table = os.getenv("BANDIT_DATA_TABLE")

        self.alpha = float(os.getenv("BANDIT_ALPHA", "0.5"))
        self.ridge_lambda = float(os.getenv("BANDIT_RIDGE_LAMBDA", "1.0"))

        self.sherman_morrison_recompute_interval = int(os.getenv("SHERMAN_MORRISON_RECOMPUTE_INTERVAL", "100"))
        self.sherman_morrison_divergence_threshold = float(os.getenv("SHERMAN_MORRISON_DIVERGENCE_THRESHOLD", "1e-6"))

        self.max_retries = 3
        self.initial_backoff_ms = 100
