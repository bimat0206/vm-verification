# syntax=docker/dockerfile:1.5
FROM golang:1.24-alpine AS builder

# Install necessary packages
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata && \
    update-ca-certificates

# Set up working directory
WORKDIR /build

# Build the application for AWS Lambda ARM64 (Graviton)
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags="-s -w" -o /main ./cmd/fetchimages

# Use AWS Lambda provided base image for ARM64
FROM public.ecr.aws/lambda/provided:al2-arm64

# Copy compiled binary from builder stage
COPY --from=builder /main /var/task/main
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set the binary as the Lambda handler
ENTRYPOINT ["/var/task/main"]