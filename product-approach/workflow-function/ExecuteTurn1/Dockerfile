FROM --platform=linux/arm64 golang:1.22-alpine AS builder

# Install git for dependencies
RUN apk add --no-cache git

# Set up the working directory
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker caching
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with enhanced error checking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bootstrap -ldflags="-s -w" -trimpath ./cmd/main.go

# Create the lambda runtime image
FROM --platform=linux/arm64 public.ecr.aws/lambda/provided:al2-arm64

# Add function version metadata
LABEL function.version="1.3.0" 
LABEL function.description="ExecuteTurn1 - Bedrock InvokeModel API implementation"

# Copy the compiled binary
COPY --from=builder /app/bootstrap /var/runtime/bootstrap

# Make the bootstrap file executable
RUN chmod +x /var/runtime/bootstrap

# Set the entry point for AWS Lambda
CMD ["bootstrap"]