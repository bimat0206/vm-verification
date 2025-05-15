module workflow-function/shared/dbutils

replace (
	workflow-function/shared/logger => ../logger
	workflow-function/shared/schema => ../schema
)

go 1.22

toolchain go1.24.0

require (
	github.com/aws/aws-sdk-go-v2 v1.36.3
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.19.0
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.43.1
	workflow-function/shared/logger v0.0.0-00010101000000-000000000000
	workflow-function/shared/schema v0.0.0-00010101000000-000000000000
)

require (
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.15 // indirect
	github.com/aws/smithy-go v1.22.2 // indirect
)
