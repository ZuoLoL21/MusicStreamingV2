import structlog
from dataclasses import dataclass

from src.di.config import Config
from src.di.db import DBManagers
from src.services.bandit import BanditHandler
from src.models.linucb import LinUCB


@dataclass
class AppState:
    config: Config
    logger: structlog.BoundLogger
    bandit: LinUCB
    db: DBManagers
    handler: BanditHandler

    @classmethod
    def create(cls, logger: structlog.BoundLogger) -> "AppState":
        config = Config.create()
        bandit = LinUCB(config, logger)
        db = DBManagers(config, bandit)
        handler = BanditHandler(config, db, logger, bandit)

        return cls(
            config=config,
            logger=logger,
            bandit=bandit,
            db=db,
            handler=handler,
        )
