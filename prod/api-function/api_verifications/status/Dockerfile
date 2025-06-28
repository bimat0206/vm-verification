FROM golang:1.24-alpine AS build

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o api-verifications-status

# Use a minimal alpine image for the final container
FROM public.ecr.aws/lambda/provided:al2-arm64

# Install ca-certificates for HTTPS connections
RUN yum update -y && yum install -y ca-certificates && yum clean all

WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/api-verifications-status /app/api-verifications-status

# Set the entrypoint
ENTRYPOINT ["/app/api-verifications-status"]
