FROM golang:1.22-alpine AS build

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o api-images-upload

# Use a minimal alpine image for the final container
FROM alpine:3.18

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/api-images-upload /app/api-images-upload

# Set the entrypoint
ENTRYPOINT ["/app/api-images-upload"]
