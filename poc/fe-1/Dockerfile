FROM python:3.10.12-alpine

WORKDIR /app

COPY requirements.txt .
RUN --mount=type=cache,target=/root/.cache/pip pip install --no-cache-dir -r requirements.txt

# Create .streamlit directory for Streamlit configuration
RUN mkdir -p /app/.streamlit

COPY clients/ /app/clients/
COPY pages/ /app/pages/
COPY app.py /app/

# Copy .streamlit directory for local development
# In production, configuration comes from CONFIG_SECRET environment variable
COPY .streamlit/ /app/.streamlit/

# Set environment variables for Streamlit configuration
# API configuration will be provided via ECS Task Definition environment variables
# pointing to AWS Secrets Manager secrets
ENV PYTHONUNBUFFERED=1 \
    STREAMLIT_SERVER_PORT=8501 \
    STREAMLIT_SERVER_ADDRESS=0.0.0.0 \
    STREAMLIT_SERVER_HEADLESS=true

HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8501/_stcore/health || exit 1

EXPOSE 8501

CMD ["streamlit", "run", "app.py"]
