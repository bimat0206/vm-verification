.PHONY: build clean deploy local test

# Configuration variables
FUNCTION_NAME ?= kootoro-initialize-function
LAMBDA_ROLE ?= arn:aws:iam::$(shell aws sts get-caller-identity --query 'Account' --output text):role/lambda-$(FUNCTION_NAME)-role
AWS_REGION ?= ap-southeast-1
ECR_REPOSITORY ?= $(shell aws sts get-caller-identity --query 'Account' --output text).dkr.ecr.$(AWS_REGION).amazonaws.com/$(FUNCTION_NAME)
IMAGE_TAG ?= latest

# Build the Docker image
build:
	docker build -t $(FUNCTION_NAME):$(IMAGE_TAG) .

# Push the Docker image to ECR
ecr-push:
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(shell aws sts get-caller-identity --query 'Account' --output text).dkr.ecr.$(AWS_REGION).amazonaws.com
	aws ecr describe-repositories --repository-names $(FUNCTION_NAME) --region $(AWS_REGION) || aws ecr create-repository --repository-name $(FUNCTION_NAME) --region $(AWS_REGION)
	docker tag $(FUNCTION_NAME):$(IMAGE_TAG) $(ECR_REPOSITORY):$(IMAGE_TAG)
	docker push $(ECR_REPOSITORY):$(IMAGE_TAG)

# Create or update the Lambda function
deploy: build ecr-push
	aws lambda create-function \
		--function-name $(FUNCTION_NAME) \
		--package-type Image \
		--code ImageUri=$(ECR_REPOSITORY):$(IMAGE_TAG) \
		--role $(LAMBDA_ROLE) \
		--architectures arm64 \
		--timeout 30 \
		--memory-size 512 \
		--environment Variables="{DYNAMODB_LAYOUT_TABLE=$(DYNAMODB_LAYOUT_TABLE),DYNAMODB_VERIFICATION_TABLE=$(DYNAMODB_VERIFICATION_TABLE),VERIFICATION_PREFIX=$(VERIFICATION_PREFIX),REFERENCE_BUCKET=$(REFERENCE_BUCKET),CHECKING_BUCKET=$(CHECKING_BUCKET)}" \
		--region $(AWS_REGION) || \
	aws lambda update-function-code \
		--function-name $(FUNCTION_NAME) \
		--image-uri $(ECR_REPOSITORY):$(IMAGE_TAG) \
		--region $(AWS_REGION)
		
	aws lambda update-function-configuration \
		--function-name $(FUNCTION_NAME) \
		--environment Variables="{DYNAMODB_LAYOUT_TABLE=$(DYNAMODB_LAYOUT_TABLE),DYNAMODB_VERIFICATION_TABLE=$(DYNAMODB_VERIFICATION_TABLE),VERIFICATION_PREFIX=$(VERIFICATION_PREFIX),REFERENCE_BUCKET=$(REFERENCE_BUCKET),CHECKING_BUCKET=$(CHECKING_BUCKET)}" \
		--region $(AWS_REGION)

# Run locally using Docker
local:
	docker run --rm \
		-e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
		-e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
		-e AWS_SESSION_TOKEN=$(AWS_SESSION_TOKEN) \
		-e AWS_REGION=$(AWS_REGION) \
		-e DYNAMODB_LAYOUT_TABLE=$(DYNAMODB_LAYOUT_TABLE) \
		-e DYNAMODB_VERIFICATION_TABLE=$(DYNAMODB_VERIFICATION_TABLE) \
		-e VERIFICATION_PREFIX=$(VERIFICATION_PREFIX) \
		-e REFERENCE_BUCKET=$(REFERENCE_BUCKET) \
		-e CHECKING_BUCKET=$(CHECKING_BUCKET) \
		-p 9000:8080 \
		$(FUNCTION_NAME):$(IMAGE_TAG)

# Run tests
test:
	go test -v ./...

# Clean up build artifacts
clean:
	rm -rf ./build
	docker rmi $(FUNCTION_NAME):$(IMAGE_TAG) || true
	docker rmi $(ECR_REPOSITORY):$(IMAGE_TAG) || true