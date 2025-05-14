FROM golang:1.24-alpine AS builder

# Install necessary packages
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata && \
    update-ca-certificates

# Set up working directory
WORKDIR /build

# Copy the entire workflow-function directory (build context is workflow-function/)
# This copies both FetchImages and shared directories
COPY . .

# Verify the structure (for debugging - can be removed in production)
RUN echo "Build context contents:" && ls -la
RUN echo "FetchImages contents:" && ls -la FetchImages/
RUN echo "Shared contents:" && ls -la shared/

# Set working directory to FetchImages for the build
WORKDIR /build/FetchImages

# Download dependencies
RUN go mod tidy

# Build the application for AWS Lambda ARM64 (Graviton)
RUN --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags="-s -w" -o main

# Use AWS Lambda provided base image for ARM64
FROM public.ecr.aws/lambda/provided:al2-arm64

# Copy compiled binary from builder stage
COPY --from=builder /build/FetchImages/main /var/task/main
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set the binary as the Lambda handler
ENTRYPOINT ["/var/task/main"]