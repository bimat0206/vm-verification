# Build stage
# Use Go 1.24 or newer
FROM golang:1.24-alpine AS build

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go app for ARM64 (Graviton) with static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o /main

# Final stage using AWS Lambda ARM64 base image
FROM public.ecr.aws/lambda/provided:al2-arm64

# Copy the binary from the build stage
COPY --from=build /main /var/task/main

# Set the Lambda handler
CMD [ "/var/task/main" ]