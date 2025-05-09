# syntax=docker/dockerfile:1.4

# 1. Build stage
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first for better cache usage
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Set the target architecture (supports arm64 and amd64)
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

# Build the binary for Linux, statically linked, named bootstrap (Lambda requirement)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -ldflags="-s -w" -o bootstrap .

# Use AWS Lambda provided base image for ARM64
FROM public.ecr.aws/lambda/provided:al2-arm64

# Copy the binary from the builder
COPY --from=builder /app/bootstrap /var/task/bootstrap

# (Optional) Set environment variables, e.g. for AWS SDK
# ENV AWS_REGION=us-east-1

# (Optional) Set the Lambda handler (for custom runtime, bootstrap is default)
# CMD ["bootstrap"]

# Document architecture support
# amd64: docker buildx build --platform linux/amd64 ...
# arm64: docker buildx build --platform linux/arm64 ...
