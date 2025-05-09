# Build stage
FROM golang:1.24-alpine AS build

WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /main cmd/main.go

# Use AWS Lambda provided base image for ARM64
FROM public.ecr.aws/lambda/provided:al2-arm64

WORKDIR /

# Copy the binary
COPY --from=builder /main /main

# Create directory for templates
RUN mkdir -p /opt/templates

# Copy templates
COPY templates/ /opt/templates/

# Set the entrypoint
ENTRYPOINT ["/main"]