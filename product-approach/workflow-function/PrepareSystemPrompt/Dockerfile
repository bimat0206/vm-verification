# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS build

WORKDIR /app
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Install required tools
RUN apk add --no-cache git

# Copy go module files first for better caching
COPY go.mod go.sum ./

# Copy source code with the new directory structure
COPY ./cmd ./cmd
COPY ./internal ./internal 
COPY *.md ./

# Create shared directory structure and copy shared modules
RUN mkdir -p shared/schema shared/templateloader shared/logger shared/s3state shared/errors

# During container build, this assumes the shared modules are in the parent directory
# This is handled differently in the retry-docker-build.sh script
COPY ../shared/schema ./shared/schema
COPY ../shared/templateloader ./shared/templateloader
COPY ../shared/logger ./shared/logger
COPY ../shared/s3state ./shared/s3state
COPY ../shared/errors ./shared/errors

# Update go.mod to use correct paths
RUN go mod edit -replace=workflow-function/shared/schema=./shared/schema \
    && go mod edit -replace=workflow-function/shared/templateloader=./shared/templateloader \
    && go mod edit -replace=workflow-function/shared/logger=./shared/logger \
    && go mod edit -replace=workflow-function/shared/s3state=./shared/s3state \
    && go mod edit -replace=workflow-function/shared/error=./shared/errors

# Download dependencies and build
RUN go mod download && go mod tidy
RUN GOOS=linux GOARCH=arm64 go build -o /main cmd/main.go

# Final stage
FROM public.ecr.aws/lambda/provided:al2-arm64

WORKDIR /

# Copy the binary
COPY --from=build /main /main

# Create templates directory and copy templates
RUN mkdir -p /opt/templates
COPY templates/ /opt/templates/

# Set environment variables
ENV COMPONENT_NAME="PrepareSystemPrompt"
ENV DATE_PARTITION_TIMEZONE="UTC"
ENV TEMPLATE_BASE_PATH="/opt/templates"
ENV DEBUG="false"

ENTRYPOINT ["/main"]