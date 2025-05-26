module workflow-function/ExecuteTurn2Combined

go 1.24.0

require (
    github.com/aws/aws-lambda-go v1.48.0
    github.com/aws/aws-sdk-go-v2 v1.36.3
    github.com/aws/aws-sdk-go-v2/config v1.29.14
    github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.19.0
    github.com/aws/aws-sdk-go-v2/service/dynamodb v1.43.1
    workflow-function/shared/bedrock v0.0.0-00010101000000-000000000000
    workflow-function/shared/errors v0.0.0
    workflow-function/shared/logger v0.0.0
    workflow-function/shared/s3state v0.0.0-00010101000000-000000000000
    workflow-function/shared/schema v0.0.0-00010101000000-000000000000
    workflow-function/shared/templateloader v0.0.0-00010101000000-000000000000
)

require github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression v1.7.82

replace workflow-function/shared/bedrock => ../shared/bedrock
replace workflow-function/shared/errors => ../shared/errors
replace workflow-function/shared/logger => ../shared/logger
replace workflow-function/shared/s3state => ../shared/s3state
replace workflow-function/shared/schema => ../shared/schema
replace workflow-function/shared/templateloader => ../shared/templateloader
