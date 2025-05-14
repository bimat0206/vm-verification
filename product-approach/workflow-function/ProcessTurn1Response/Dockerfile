FROM public.ecr.aws/lambda/go:1.21

# Copy shared packages first
COPY workflow-function/shared/ /app/workflow-function/shared/

# Copy function code
COPY workflow-function/ProcessTurn1Response/ /app/workflow-function/ProcessTurn1Response/

WORKDIR /app/workflow-function/ProcessTurn1Response

# Build the function
RUN go mod download
RUN GOOS=linux GOARCH=amd64 go build -o main .

# Set the CMD to the handler
CMD ["main"]