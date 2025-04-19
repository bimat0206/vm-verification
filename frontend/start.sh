#!/bin/bash
set -e
echo "Starting Streamlit application..."
echo "Environment information:"
echo "SECRET_ARN: ${SECRET_ARN:-Not set}"
echo "REGION: ${REGION:-Not set}"
echo "API_ENDPOINT: ${API_ENDPOINT:-Not set}"
echo "Working directory: $(pwd)"
echo "Files in current directory:"
ls -la

# Wait for network services to be fully available
echo "Waiting for network services..."
sleep 5

# Create a static health check file to respond quickly to initial health checks
mkdir -p /tmp/streamlit_static
echo "{\"status\":\"ok\"}" > /tmp/streamlit_static/health.json

echo "Starting Streamlit with detailed logging..."
exec streamlit run app.py \
  --server.port=8501 \
  --server.address=0.0.0.0 \
  --server.headless=true \
  --server.enableCORS=false \
  --server.enableXsrfProtection=false \
  --server.maxUploadSize=200 \
  --server.fileWatcherType=none \
  --logger.level=info