# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS build

WORKDIR /app
ENV GO111MODULE=on

# Install required tools
RUN apk add --no-cache git

# Copy go module files first for better caching
COPY go.mod go.sum ./

# Copy source code
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY *.md ./

# Create shared directory structure
RUN mkdir -p shared/schema shared/s3utils shared/templateloader shared/logger

# Copy shared modules with proper structure
COPY ../shared/schema shared/schema/
COPY ../shared/s3utils shared/s3utils/
COPY ../shared/templateloader shared/templateloader/
COPY ../shared/logger shared/logger/

# Update go.mod to use local shared modules
RUN go mod edit -replace=workflow-function/shared/schema=./shared/schema \
    && go mod edit -replace=workflow-function/shared/s3utils=./shared/s3utils \
    && go mod edit -replace=workflow-function/shared/templateloader=./shared/templateloader \
    && go mod edit -replace=workflow-function/shared/logger=./shared/logger

# Download dependencies and build
RUN go mod download && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /main cmd/main.go

# Final stage
FROM public.ecr.aws/lambda/provided:al2-arm64

WORKDIR /

# Copy the binary
COPY --from=build /main /main

# Create templates directory and copy templates
RUN mkdir -p /opt/templates
COPY templates/ /opt/templates/

# Set component name for logging
ENV COMPONENT_NAME="PrepareSystemPrompt"

ENTRYPOINT ["/main"]