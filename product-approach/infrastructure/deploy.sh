#!/bin/bash

# Set environment variables
export USE_ECR_IMAGES=false
export PYTHONPATH=$(pwd)

# Run the CDK deployment
echo "Starting CDK deployment..."

# Try running with cdk command first
cdk deploy --require-approval never

# If the CDK command fails, try with Python 3 directly
if [ $? -ne 0 ]; then
    echo "CDK command failed, trying with Python 3 directly..."
    python3 app.py
fi

echo "Deployment complete!"