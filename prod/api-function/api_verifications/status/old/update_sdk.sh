#!/bin/bash

# Script to update all AWS SDK Go v2 modules to compatible versions
# This fixes the "ResolveEndpointV2" error caused by version incompatibilities

echo "Updating AWS SDK Go v2 modules to compatible versions..."

# Remove the go.sum to ensure clean dependency resolution
rm -f go.sum

# Update all AWS SDK modules to the latest compatible versions
go get -u github.com/aws/aws-sdk-go-v2@latest
go get -u github.com/aws/aws-sdk-go-v2/config@latest
go get -u github.com/aws/aws-sdk-go-v2/credentials@latest
go get -u github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue@latest
go get -u github.com/aws/aws-sdk-go-v2/service/dynamodb@latest
go get -u github.com/aws/aws-sdk-go-v2/service/s3@latest
go get -u github.com/aws/smithy-go@latest

# Tidy up the modules
go mod tidy

echo "AWS SDK modules updated. Please rebuild your application."