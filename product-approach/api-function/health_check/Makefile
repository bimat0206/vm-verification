.PHONY: build clean test docker-build docker-push

# Variables
FUNCTION_NAME=kootoro-dev-lambda-health-check-f6d3xl
DOCKER_REPO=879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-health-check-f6d3xl
DOCKER_TAG=latest
AWS_REGION=us-east-1

# Build the binary
build:
	go build -o $(BINARY_NAME) *.go

# Clean up
clean:
	rm -f $(BINARY_NAME)

# Run tests
test:
	go test -v ./...

# Build Docker image locally
docker-build:
	docker build -t $(DOCKER_REPO):$(DOCKER_TAG) .

# Login to ECR and push Docker image
docker-push:
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(DOCKER_REPO)
	docker push $(DOCKER_REPO):$(DOCKER_TAG)

# Deploy to AWS Lambda (requires AWS CLI and proper IAM permissions)
deploy: docker-build docker-push
	aws lambda update-function-code \
		--function-name $(FUNCTION_NAME) \
		--image-uri $(DOCKER_REPO):$(DOCKER_TAG) \
		--region $(AWS_REGION)

# Run locally
run:
	go run *.go