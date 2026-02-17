FROM python:3.12-slim

COPY --from=ghcr.io/astral-sh/uv:latest /uv /bin/uv

WORKDIR /app

COPY pyproject.toml uv.lock ./
ENV UV_SYSTEM_PYTHON=1
RUN uv sync --frozen --no-dev && uv pip install "uvicorn[standard]"

COPY vault.py ./

VOLUME /app/keys.json

EXPOSE 8000

CMD ["uvicorn", "vault:app", "--host", "0.0.0.0", "--port", "8000"]
