import logging
from fastapi import FastAPI
from src.middleware import RequestIDMiddleware, LoggingMiddleware

# Configure logging
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

app = FastAPI()

# Add middlewares (order matters - request_id should be first)
app.add_middleware(RequestIDMiddleware)
app.add_middleware(LoggingMiddleware, logger=logger)

@app.get("/")
async def root():
    return {"message": "Hello World"}