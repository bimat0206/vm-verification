# syntax=docker/dockerfile:1.4
# Builder stage
FROM golang:1.24-alpine AS builder

# Install minimal dependencies for building
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy fonts directory first so it persists
COPY ./fonts/ /app/fonts/

# Build with cache mounts for Go modules and source code
RUN --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,target=. \
    GOOS=linux GOARCH=arm64 go build -o /bootstrap -ldflags="-s -w" -trimpath .

# Runtime stage
FROM public.ecr.aws/lambda/provided:al2

# Copy binary and make executable
COPY --from=builder /bootstrap ${LAMBDA_RUNTIME_DIR}/bootstrap
RUN chmod +x ${LAMBDA_RUNTIME_DIR}/bootstrap

# Copy fonts directory
COPY --from=builder /app/fonts/ ${LAMBDA_TASK_ROOT}/fonts/

# Create directories for temporary files
RUN mkdir -p /tmp/image_cache && chmod 777 /tmp/image_cache

# Set environment variables
ENV DYNAMODB_LAYOUT_TABLE=VendingMachineLayoutMetadata
ENV AWS_REGION=us-east-1

# Set the handler
CMD [ "handler" ]