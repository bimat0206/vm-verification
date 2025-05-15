# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set up Go workspace
WORKDIR /go/src

# Create directories for our modules
RUN mkdir -p product-approach/workflow-function/ProcessTurn1Response
RUN mkdir -p product-approach/workflow-function/shared

# Copy go.mod and go.sum first for better layer caching
COPY workflow-function/ProcessTurn1Response/go.mod workflow-function/ProcessTurn1Response/go.sum* ./product-approach/workflow-function/ProcessTurn1Response/

# Copy shared packages (required by the replace directives in go.mod)
COPY workflow-function/shared/schema/ ./product-approach/workflow-function/shared/schema/
COPY workflow-function/shared/logger/ ./product-approach/workflow-function/shared/logger/
COPY workflow-function/shared/s3utils/ ./product-approach/workflow-function/shared/s3utils/
COPY workflow-function/shared/dbutils/ ./product-approach/workflow-function/shared/dbutils/

# Copy the rest of the application code
COPY workflow-function/ProcessTurn1Response/ ./product-approach/workflow-function/ProcessTurn1Response/

# Set working directory to the function directory
WORKDIR /go/src/product-approach/workflow-function/ProcessTurn1Response

# Build the function with optimizations
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bootstrap ./cmd

# Runtime stage
FROM public.ecr.aws/lambda/provided:al2-arm64

# Copy the binary as bootstrap (required by Lambda runtime)
COPY --from=builder /go/src/product-approach/workflow-function/ProcessTurn1Response/bootstrap /var/runtime/bootstrap

# Set environment variables
ENV LOG_LEVEL=info

# The CMD should be the handler name for provided runtime
CMD ["bootstrap"]