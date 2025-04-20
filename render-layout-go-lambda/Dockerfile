# Builder stage
FROM golang:1.24-alpine AS builder

# Install required dependencies for building
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application with CGO disabled for static linking
# Use 'main' as the binary name to match previous build commands
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main -ldflags="-s -w" ./main.go

# Runtime stage
FROM alpine:3.19

# Add necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create app directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main /app/main

# Copy the fonts directory from the builder stage
COPY --from=builder /app/fonts/ /app/fonts/

# Define the entrypoint
ENTRYPOINT ["/app/main"]