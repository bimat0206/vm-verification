# Build stage
FROM golang:1.24-alpine AS build

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go app for ARM64 (Graviton) with static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bootstrap

# Use AWS Lambda provided base image for ARM64
FROM public.ecr.aws/lambda/provided:al2-arm64

# Copy the bootstrap executable
COPY --from=build /app/bootstrap ${LAMBDA_TASK_ROOT}

# Set the binary as the Lambda handler
ENTRYPOINT ["/var/task/bootstrap"]
