FROM golang:1.24-alpine AS builder

# Install necessary packages
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata && \
    update-ca-certificates

# Set up working directory
WORKDIR /build

# Copy the ExecuteTurn2Combined directory and shared modules
COPY . .

# Download dependencies and build
RUN go mod tidy && go mod download

# Build the application for AWS Lambda ARM64 (Graviton)
RUN --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags="-s -w" -o main ./cmd/main.go

# Use AWS Lambda provided base image for ARM64
FROM public.ecr.aws/lambda/provided:al2-arm64

# Copy compiled binary from builder stage
COPY --from=builder /build/main /var/task/main
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy templates with proper naming to match what the loader expects
# This flattens the versioned directory structure
COPY --from=builder /build/templates/turn2-layout-vs-checking/v1.0.tmpl /opt/templates/turn2-layout-vs-checking.tmpl
COPY --from=builder /build/templates/turn2-previous-vs-current/v1.0.tmpl /opt/templates/turn2-previous-vs-current.tmpl

# Set the binary as the Lambda handler
ENTRYPOINT ["/var/task/main"]