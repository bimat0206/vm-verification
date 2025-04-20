FROM node:18-slim

# Set working directory
WORKDIR /app

# Install dependencies for canvas
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
    && rm -rf /var/lib/apt/lists/*

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install --production

# Copy application code
COPY index.js services.js renderer.js utils.js config.js ./

# Create fonts directories in tmp and copy fonts
RUN mkdir -p /tmp/fonts /tmp/fontconfig && \
    cp /usr/share/fonts/truetype/dejavu/DejaVuSans.ttf /tmp/fonts/dejavu-sans.ttf

# Set environment variables
ENV NODE_ENV=production
ENV AWS_REGION=us-east-1
ENV FONTCONFIG_PATH=/tmp/fontconfig
ENV FONTCONFIG_FILE=/tmp/fonts.conf

# Set command
CMD ["node", "index.js"]