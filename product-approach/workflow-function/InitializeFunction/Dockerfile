FROM golang:1.24-alpine AS builder

# Install necessary packages
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata && \
    update-ca-certificates

# Set up working directory structure
WORKDIR /go/src/workflow-function/

# First copy the shared packages to ensure they're available for the build
COPY ../shared/schema/ ./shared/schema/
COPY ../shared/logger/ ./shared/logger/
COPY ../shared/dbutils/ ./shared/dbutils/
COPY ../shared/s3utils/ ./shared/s3utils/

# Now copy the function code
COPY ./ ./InitializeFunction/

# Build from the function directory
WORKDIR /go/src/workflow-function/InitializeFunction

# Build the application for AWS Lambda ARM64 (Graviton)
RUN GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags="-s -w" -o /main

# Use AWS Lambda provided base image for ARM64
FROM public.ecr.aws/lambda/provided:al2-arm64

# Copy compiled binary from builder stage
COPY --from=builder /main /var/task/main
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set the binary as the Lambda handler
ENTRYPOINT ["/var/task/main"]