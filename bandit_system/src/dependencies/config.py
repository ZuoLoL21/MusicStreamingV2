from dotenv import load_dotenv

class Config:
    def __init__(self):
        load_dotenv()

        self.db_connection_string = os.getenv("DB_CONNECTION_STRING")