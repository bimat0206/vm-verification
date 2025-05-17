FROM public.ecr.aws/lambda/provided:al2-x86_64 as builder

# Install necessary build tools
RUN yum install -y golang git

# Set up the working directory
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker caching
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimizations
RUN go build -o main ./cmd/main.go

# Create the lambda runtime image
FROM public.ecr.aws/lambda/provided:al2-x86_64

# Copy the compiled binary
COPY --from=builder /app/main ${LAMBDA_RUNTIME_DIR}

# Set the entry point for AWS Lambda
CMD ["main"]