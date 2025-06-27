FROM golang:1.24-alpine AS builder

# Install necessary packages
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata && \
    update-ca-certificates

# Set up working directory
WORKDIR /build

# Copy the entire context
COPY . .

# Verify the structure (for debugging - can be removed in production)
RUN echo "Build context contents:" && ls -la
RUN echo "Shared modules:" && ls -la ./shared/ || echo "Shared modules not accessible in this context"

# Download dependencies and build
RUN go mod download

# Build the application for AWS Lambda ARM64 (Graviton)
RUN GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags="-s -w" -o main ./cmd/main.go

# Use AWS Lambda provided base image for ARM64
FROM public.ecr.aws/lambda/provided:al2-arm64

# Copy compiled binary from builder stage
COPY --from=builder /build/main /var/task/main

# Copy templates with explicit directory structure preservation
COPY --from=builder /build/templates /opt/templates
# Ensure correct permissions
RUN chmod -R 755 /opt/templates

# Create directory structure listing for debugging
RUN echo "Templates directory structure:" && \
    ls -la /opt/templates && \
    find /opt/templates -type f -name "*.tmpl" | sort
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Version information
LABEL version="4.0.11"
LABEL description="PrepareTurn1Prompt with template loading fixes following PrepareSystemPrompt design"

# Set the binary as the Lambda handler
ENTRYPOINT ["/var/task/main"]