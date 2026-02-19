from dotenv import load_dotenv

class Config:
    def __init__(self):
        load_dotenv()

        self.db_connection_string = os.getenv("DB_CONNECTION_STRING")
        self.bandit_data_table = os.getenv("BANDIT_DATA_TABLE")
        self.music_theme_table = os.getenv("MUSIC_THEME_TABLE")