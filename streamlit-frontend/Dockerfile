FROM python:3.9-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY app.py .
COPY .streamlit/ .streamlit/

# Set environment variables
ENV API_URL=http://verification-service:3000

# Expose port
EXPOSE 8501

# Command to run the application
CMD ["streamlit", "run", "app.py"]