# syntax=docker/dockerfile:1.4
FROM node:18-alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app
COPY package.json package-lock.json* ./
RUN --mount=type=cache,target=/root/.npm \
  if [ -f package-lock.json ]; then npm ci; \
  else echo "Lockfile not found." && exit 1; \
  fi

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1
# Use the correct API Gateway endpoint for build
ENV NEXT_PUBLIC_API_BASE_URL=https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1
ENV NEXT_PUBLIC_API_ENDPOINT=https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1
ENV NEXT_PUBLIC_API_KEY=dummy-key-for-build
ENV NEXT_PUBLIC_AWS_REGION=us-east-1
ENV NEXT_PUBLIC_REFERENCE_BUCKET=kootoro-dev-s3-reference-f6d3xl
ENV NEXT_PUBLIC_CHECKING_BUCKET=kootoro-dev-s3-checking-f6d3xl
RUN --mount=type=cache,target=/root/.npm npm run build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Install additional tools
RUN apk add --no-cache curl jq

# Create a non-root user
RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 nextjs

# Copy built app
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# Create a startup script that maps environment variables correctly
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'set -e' >> /app/start.sh && \
    echo '' >> /app/start.sh && \
    echo '# Map CONFIG_SECRET to NEXT_PUBLIC_CONFIG_SECRET for server-side access' >> /app/start.sh && \
    echo 'if [ -n "$CONFIG_SECRET" ]; then' >> /app/start.sh && \
    echo '  export NEXT_PUBLIC_CONFIG_SECRET="$CONFIG_SECRET"' >> /app/start.sh && \
    echo '  echo "CONFIG_SECRET mapped to NEXT_PUBLIC_CONFIG_SECRET"' >> /app/start.sh && \
    echo 'fi' >> /app/start.sh && \
    echo '' >> /app/start.sh && \
    echo '# Map API_KEY_SECRET_NAME for server-side access' >> /app/start.sh && \
    echo 'if [ -n "$API_KEY_SECRET_NAME" ]; then' >> /app/start.sh && \
    echo '  export NEXT_PUBLIC_API_KEY_SECRET_NAME="$API_KEY_SECRET_NAME"' >> /app/start.sh && \
    echo '  echo "API_KEY_SECRET_NAME mapped to NEXT_PUBLIC_API_KEY_SECRET_NAME"' >> /app/start.sh && \
    echo 'fi' >> /app/start.sh && \
    echo '' >> /app/start.sh && \
    echo '# Set AWS region if provided' >> /app/start.sh && \
    echo 'if [ -n "$AWS_REGION" ]; then' >> /app/start.sh && \
    echo '  export NEXT_PUBLIC_AWS_REGION="$AWS_REGION"' >> /app/start.sh && \
    echo '  echo "AWS_REGION mapped to NEXT_PUBLIC_AWS_REGION: $AWS_REGION"' >> /app/start.sh && \
    echo 'fi' >> /app/start.sh && \
    echo '' >> /app/start.sh && \
    echo '# Override API endpoint with correct value' >> /app/start.sh && \
    echo 'export NEXT_PUBLIC_API_BASE_URL="https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1"' >> /app/start.sh && \
    echo 'export NEXT_PUBLIC_API_ENDPOINT="https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1"' >> /app/start.sh && \
    echo 'echo "API endpoints set to: $NEXT_PUBLIC_API_BASE_URL"' >> /app/start.sh && \
    echo '' >> /app/start.sh && \
    echo '# Log environment for debugging' >> /app/start.sh && \
    echo 'echo "Starting React/Next.js application..."' >> /app/start.sh && \
    echo 'echo "NODE_ENV: $NODE_ENV"' >> /app/start.sh && \
    echo 'echo "PORT: $PORT"' >> /app/start.sh && \
    echo 'echo "HOSTNAME: $HOSTNAME"' >> /app/start.sh && \
    echo 'echo "NEXT_PUBLIC_API_BASE_URL: $NEXT_PUBLIC_API_BASE_URL"' >> /app/start.sh && \
    echo '' >> /app/start.sh && \
    echo '# Start the server' >> /app/start.sh && \
    echo 'exec node server.js' >> /app/start.sh && \
    chmod +x /app/start.sh

# Set permissions
RUN chown nextjs:nodejs /app/start.sh

# Switch to non-root user
USER nextjs

# Configure network
EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:3000/api/health || exit 1

# Start the application with our wrapper script
CMD ["/app/start.sh"]