FROM python:3.11-slim

# Install system dependencies for Pillow and other requirements
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    libfreetype6-dev \
    libjpeg-dev \
    libpng-dev \
    fonts-dejavu \
    fonts-liberation \
    fonts-freefont-ttf \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Create directories
RUN mkdir -p /app/fonts && \
    mkdir -p /app/output && \
    chmod 777 /app/output

# Copy any available system fonts to our fonts directory
RUN find /usr/share/fonts -type f -name "*.ttf" -exec cp {} /app/fonts/ \; 2>/dev/null || echo "No system fonts found"

# Copy requirements file first (for better caching)
COPY requirements.txt .

# Install Python dependencies including AWS Lambda Runtime Interface Client
RUN pip install --no-cache-dir -r requirements.txt && \
    pip install --no-cache-dir awslambdaric>=2.0.0 requests

# Copy application code
COPY *.py /app/

# Copy bootstrap script
COPY lambda_bootstrap.py /app/
RUN chmod +x /app/lambda_bootstrap.py

# Create an entrypoint script to determine which mode to run in
RUN echo '#!/bin/sh\n\
if [ -n "$AWS_LAMBDA_RUNTIME_API" ]; then\n\
  # Running in Lambda\n\
  echo "Starting in Lambda mode"\n\
  exec /app/lambda_bootstrap.py\n\
else\n\
  # Running as Flask app\n\
  echo "Starting in Flask mode"\n\
  exec python app.py\n\
fi' > /app/entrypoint.sh && \
chmod +x /app/entrypoint.sh

# Set environment variables
ENV PYTHONUNBUFFERED=1

# Expose port 5000 for Flask app (ignored in Lambda)
EXPOSE 5000

# Use the entrypoint script to decide which mode to run in
ENTRYPOINT ["/app/entrypoint.sh"]