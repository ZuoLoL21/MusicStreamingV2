from datetime import datetime, timedelta
from threading import Lock
from typing import Optional, Set

import structlog
from sqlalchemy import Engine, text


class ThemesCache:

    def __init__(self, warehouse_engine: Engine, bandit_data_table: str, ttl_minutes: int = 30):
        self._warehouse_engine = warehouse_engine
        self._bandit_data_table = bandit_data_table
        self._logger = structlog.get_logger("themes_cache")

        self._ttl = timedelta(minutes=ttl_minutes)

        self._cache: Optional[Set[str]] = None
        self._expires_at: Optional[datetime] = None
        self._lock = Lock()

    def get_all_themes(self) -> Set[str]:
        now = datetime.now()

        with self._lock:
            if self._cache is not None and self._expires_at is not None and now < self._expires_at:
                return self._cache

            self._logger.debug("themes_cache_refresh_started")

            query = text(
                f"SELECT DISTINCT theme"
                f" FROM {self._bandit_data_table}"
                f" ORDER BY theme"
            )

            with self._warehouse_engine.connect() as conn:
                rows = conn.execute(query).fetchall()

            self._cache = {row[0] for row in rows}
            self._expires_at = now + self._ttl

            self._logger.info(
                "themes_cache_refreshed",
                count=len(self._cache),
                ttl_minutes=self._ttl.total_seconds() / 60,
            )

            return self._cache

    def invalidate(self):
        with self._lock:
            self._cache = None
            self._expires_at = None
            self._logger.info("themes_cache_invalidated")
