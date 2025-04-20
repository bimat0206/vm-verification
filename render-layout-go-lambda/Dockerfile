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
RUN CGO_ENABLED=0 GOOS=linux go build -o lambda -ldflags="-s -w" .

# Runtime stage
FROM alpine:3.19

# Add necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create fonts directory
RUN mkdir -p /app/fonts

# Copy the binary from the builder stage
COPY --from=builder /app/lambda /app/

# Copy the fonts directory
COPY fonts/ /app/fonts/

# Set the working directory
WORKDIR /app

# Define the entrypoint
ENTRYPOINT ["/app/lambda"]