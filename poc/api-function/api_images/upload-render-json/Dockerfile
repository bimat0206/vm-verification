# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files first to leverage Docker layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bootstrap main.go

# Runtime stage
FROM public.ecr.aws/lambda/provided:al2

# Copy binary and make executable
COPY --from=builder /app/bootstrap ${LAMBDA_RUNTIME_DIR}/bootstrap
RUN chmod +x ${LAMBDA_RUNTIME_DIR}/bootstrap

# Copy fonts directory
COPY --from=builder /app/fonts/ ${LAMBDA_TASK_ROOT}/fonts/

# Create directories for temporary files
RUN mkdir -p /tmp/image_cache && chmod 777 /tmp/image_cache

# Set the handler
CMD [ "handler" ]