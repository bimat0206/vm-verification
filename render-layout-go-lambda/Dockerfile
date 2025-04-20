# Builder stage
FROM golang:1.24-alpine AS builder

# Install minimal dependencies for building
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build for ARM64
RUN GOOS=linux GOARCH=arm64 go build -o bootstrap -ldflags="-s -w" -trimpath ./main.go

# Runtime stage
FROM public.ecr.aws/lambda/provided:al2

# Copy binary and make executable
COPY --from=builder /app/bootstrap ${LAMBDA_RUNTIME_DIR}/bootstrap
RUN chmod +x ${LAMBDA_RUNTIME_DIR}/bootstrap

# Copy fonts
COPY --from=builder /app/fonts/ ${LAMBDA_TASK_ROOT}/fonts/

# Set the handler
CMD [ "handler" ]