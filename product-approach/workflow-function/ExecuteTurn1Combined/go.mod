module workflow-function/ExecuteTurn1Combined

go 1.24.0

require (
	github.com/aws/aws-lambda-go v1.48.0
	github.com/aws/aws-sdk-go-v2 v1.36.3
	github.com/aws/aws-sdk-go-v2/config v1.29.14
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.19.0
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.43.1
	workflow-function/shared/bedrock v0.0.0-00010101000000-000000000000
	workflow-function/shared/errors v0.0.0 // NEW
	workflow-function/shared/logger v0.0.0 // NEW
	workflow-function/shared/s3state v0.0.0-00010101000000-000000000000
	workflow-function/shared/schema v0.0.0-00010101000000-000000000000
        workflow-function/shared/templateloader v0.0.0-00010101000000-000000000000
       golang.org/x/image v0.26.0
)

require github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression v1.7.82

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.67 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.30.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.79.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.19 // indirect
	github.com/aws/smithy-go v1.22.3 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace workflow-function/shared/bedrock => ../shared/bedrock

replace workflow-function/shared/errors => ../shared/errors

replace workflow-function/shared/logger => ../shared/logger

replace workflow-function/shared/s3state => ../shared/s3state

replace workflow-function/shared/schema => ../shared/schema

replace workflow-function/shared/templateloader => ../shared/templateloader
