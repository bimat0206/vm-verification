# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Use cache mounts for Go modules
RUN --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    apk add --no-cache git

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /verification-service ./cmd/server

FROM alpine:3.17

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /verification-service .
COPY .env .

EXPOSE 3000

CMD ["./verification-service"]
