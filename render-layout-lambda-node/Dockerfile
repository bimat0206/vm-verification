FROM node:18-slim
WORKDIR /app
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
COPY package*.json ./
RUN npm install --production
COPY lambda-handler.js ./
RUN mkdir -p /app/fonts && \
    cp /usr/share/fonts/truetype/dejavu/DejaVuSans.ttf /app/fonts/dejavu-sans.ttf
ENV NODE_ENV=production
ENV AWS_REGION=us-east-1
ENV FONTCONFIG_PATH=/tmp/fontconfig
ENV FONTCONFIG_FILE=/tmp/fonts.conf
CMD ["node", "lambda-handler.js"]