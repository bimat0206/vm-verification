# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Cache Go modules
RUN --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the entire application source
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o /verification-service ./cmd/server

FROM alpine:3.17

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /verification-service .

EXPOSE 3000

CMD ["./verification-service"]
