# vending-machine-verification

This repository contains Terraform modules, Lambda functions, and helper scripts used to verify vending machine images. The workflow runs on AWS and relies on S3, DynamoDB and Bedrock to process and track verification results.

## Project structure

- **product-approach/workflow-function** – Go Lambda functions implementing the verification workflow.
- **product-approach/iac** – Terraform configuration and modules.
- **product-approach/fe-1** – Example Streamlit front‑end utilities.
- **lib** – supporting layout rendering libraries.
- Utility scripts at the repository root (`cloudwatch_loggroup_bulk_delete.py`, `ecr_bulk_delete.py`, `s3_bulk_delete.py`).

## Setup

Install Go, Docker, the AWS CLI and Terraform. After cloning the repo run:

```bash
go mod download
```

Define the following environment variables so the workflow functions can access your AWS resources:

```bash
export AWS_REGION=us-east-1
export STATE_BUCKET=your-state-bucket
export BEDROCK_MODEL=anthropic.claude-3-sonnet-20250219-v1:0
export DYNAMODB_VERIFICATION_TABLE=your-verification-table
export DYNAMODB_CONVERSATION_TABLE=your-conversation-table
```

## Running workflow functions locally

Each function resides in its own directory under `product-approach/workflow-function`. To run one locally:

```bash
cd product-approach/workflow-function/<FunctionName>
go run ./cmd/main.go
```

For example:

```bash
cd product-approach/workflow-function/ExecuteTurn1Combined
go run ./cmd/main.go
```

Make sure the environment variables above are exported in your shell when launching the function.
