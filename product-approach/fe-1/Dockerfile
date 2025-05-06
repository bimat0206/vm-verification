FROM python:3.10.12-alpine

WORKDIR /app

COPY requirements.txt .
RUN --mount=type=cache,target=/root/.cache/pip pip install --no-cache-dir -r requirements.txt

# Create .streamlit directory and secrets.toml file
RUN mkdir -p /app/.streamlit

# We'll create the secrets.toml file directly instead of copying it

COPY clients/ /app/clients/
COPY pages/ /app/pages/
COPY app.py /app/

# Default values for API configuration
ARG API_ENDPOINT='https://tgea33n7wb.execute-api.us-east-1.amazonaws.com'
ARG API_KEY='dnXCwHqkGi97P3YzUg62A7qc1RNmmtouBr4smuE0'

# Set environment variables including API configuration
ENV PYTHONUNBUFFERED=1 \
    STREAMLIT_SERVER_PORT=8501 \
    STREAMLIT_SERVER_ADDRESS=0.0.0.0 \
    STREAMLIT_SERVER_HEADLESS=true \
    API_ENDPOINT=${API_ENDPOINT} \
    API_KEY=${API_KEY}

# Create secrets.toml with API configuration
RUN echo "# API Configuration\nAPI_ENDPOINT = \"${API_ENDPOINT}\"\nAPI_KEY = \"${API_KEY}\"" > /app/.streamlit/secrets.toml

HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8501/_stcore/health || exit 1

EXPOSE 8501

CMD ["streamlit", "run", "app.py"]
