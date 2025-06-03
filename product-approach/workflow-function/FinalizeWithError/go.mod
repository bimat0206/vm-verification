module workflow-function/FinalizeWithErrorFunction

go 1.24

replace workflow-function/shared/logger => ../shared/logger

replace workflow-function/shared/schema => ../shared/schema

replace workflow-function/shared/s3state => ../shared/s3state

replace workflow-function/shared/errors => ../shared/errors

require (
    github.com/aws/aws-lambda-go v1.48.0
    github.com/aws/aws-sdk-go-v2 v1.36.3
    github.com/aws/aws-sdk-go-v2/config v1.29.14
    github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.19.0
    github.com/aws/aws-sdk-go-v2/service/dynamodb v1.43.1
    github.com/aws/aws-sdk-go-v2/service/s3 v1.79.3
    workflow-function/shared/logger v0.0.0
    workflow-function/shared/schema v0.0.0
    workflow-function/shared/s3state v0.0.0
)
