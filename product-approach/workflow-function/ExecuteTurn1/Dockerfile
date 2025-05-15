# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS build

WORKDIR /app

# Create shared directory for module replacements
RUN mkdir -p /app/shared/bedrock /app/shared/errors /app/shared/schema

# Copy the main module files
COPY go.mod go.sum ./

# Copy source code
COPY . .

# Copy shared modules (these must be included in the build context)
COPY shared/bedrock/ /app/shared/bedrock/
COPY shared/errors/ /app/shared/errors/
COPY shared/schema/ /app/shared/schema/

# Build the application with shared packages
# We're using the workdir path for module replacements
RUN go mod edit -replace workflow-function/shared/bedrock=./shared/bedrock \
    -replace workflow-function/shared/errors=./shared/errors \
    -replace workflow-function/shared/schema=./shared/schema \
    && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /main ./cmd/main.go

FROM public.ecr.aws/lambda/provided:al2-arm64

WORKDIR /

COPY --from=build /main /main

# Set the Lambda handler
ENV LAMBDA_TASK_ROOT=/var/task

# Create necessary directories
RUN mkdir -p /var/task

# Set the Lambda entrypoint
ENTRYPOINT ["/main"]
