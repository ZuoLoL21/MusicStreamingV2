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
