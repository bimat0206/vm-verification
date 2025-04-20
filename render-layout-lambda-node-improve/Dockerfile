FROM --platform=linux/amd64 node:18-slim

# Set working directory
WORKDIR /app

# Install dependencies for canvas with ARM64 compatibility
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    pkg-config \
    libcairo2-dev \
    libpango1.0-dev \
    libjpeg-dev \
    libgif-dev \
    libpng-dev \
    fonts-dejavu-core \
    libfontconfig1-dev \
    python3 \
    make \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Copy package files
COPY package*.json ./

# Install dependencies with canvas binary settings
ENV CANVAS_PREBUILT_PATH=/tmp/canvas-prebuilt
RUN npm install --build-from-source --production

# Copy application code
COPY index.js services.js renderer.js utils.js config.js common.js ./

# Create directories
RUN mkdir -p /tmp/fonts /tmp/fontconfig /tmp/s3-mock

# Copy fonts to accessible location
RUN cp /usr/share/fonts/truetype/dejavu/DejaVuSans.ttf /tmp/fonts/dejavu-sans.ttf

# Set environment variables
ENV NODE_ENV=production
ENV AWS_REGION=us-east-1
ENV FONTCONFIG_PATH=/tmp/fontconfig
ENV FONTCONFIG_FILE=/tmp/fonts.conf
ENV USE_MOCK_S3=false
ENV TEMP_DIR=/tmp/s3-mock

# Set command
CMD ["node", "index.js"]