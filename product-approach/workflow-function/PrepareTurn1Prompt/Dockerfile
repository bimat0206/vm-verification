# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS build

WORKDIR /app

# Copy go.mod and go.sum first to leverage caching
COPY go.mod go.sum ./

# Download dependencies using cache mounts
RUN --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /main cmd/main.go

FROM public.ecr.aws/lambda/provided:al2-arm64

WORKDIR /

COPY --from=build /main /main

RUN mkdir -p /opt/templates

COPY templates/ /opt/templates/

ENTRYPOINT ["/main"]