FROM golang:1.24-alpine AS builder

# Install git for dependencies
RUN apk add --no-cache git

# Set up the working directory
WORKDIR /app

# Copy source code with modules first
COPY . .

# Environment setup
ENV GO111MODULE=on
ENV GOFLAGS=-mod=mod

# Special handling for Docker build - ensuring modules are properly setup
RUN go env -w GOMODCACHE=/go/pkg/mod

# Download dependencies
RUN go mod tidy && go mod download

# Build the application with enhanced error checking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bootstrap -ldflags="-s -w" -trimpath ./cmd/main.go

# Use AWS Lambda provided base image for ARM64
FROM public.ecr.aws/lambda/provided:al2-arm64

# Add function version metadata
LABEL function.version="1.3.0" 
LABEL function.description="ExecuteTurn1 - Bedrock InvokeModel API implementation"

# Copy the compiled binary
COPY --from=builder /app/bootstrap /var/runtime/bootstrap

# Make the bootstrap file executable
RUN chmod +x /var/runtime/bootstrap

# Set the entry point for AWS Lambda
CMD ["bootstrap"]