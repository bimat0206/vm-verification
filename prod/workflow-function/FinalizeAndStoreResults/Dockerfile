FROM public.ecr.aws/docker/library/golang:1.22-alpine AS builder

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
    GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags="-s -w" -o /main ./cmd/finalizeandstoreresults

# Use AWS Lambda provided base image for ARM64
FROM public.ecr.aws/lambda/provided:al2-arm64

# Add function version metadata
LABEL function.version="1.0.1" 
LABEL function.description="FinalizeAndStoreResults - Store final verification results and complete workflow"

# Copy compiled binary from builder stage
COPY --from=builder /main /var/task/main
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set the binary as the Lambda handler
ENTRYPOINT ["/var/task/main"]