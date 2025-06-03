module workflow-function/FinalizeAndStoreResults

go 1.21

replace workflow-function/shared/logger => ../shared/logger
replace workflow-function/shared/schema => ../shared/schema
replace workflow-function/shared/s3state => ../shared/s3state
replace workflow-function/shared/errors => ../shared/errors

require (
    github.com/aws/aws-lambda-go v1.48.0
    github.com/aws/aws-sdk-go-v2 v1.36.3
    github.com/aws/aws-sdk-go-v2/config v1.27.9
    github.com/aws/aws-sdk-go-v2/service/dynamodb v1.43.1
    github.com/aws/aws-sdk-go-v2/service/s3 v1.79.3
    workflow-function/shared/logger v0.0.0-00010101000000-000000000000
    workflow-function/shared/s3state v0.0.0-00010101000000-000000000000
    workflow-function/shared/schema v0.0.0-00010101000000-000000000000
    workflow-function/shared/errors v0.0.0
)

require (
    github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
    github.com/aws/aws-sdk-go-v2/credentials v1.17.9 // indirect
    github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.0 // indirect
    github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
    github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
    github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
    github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
    github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
    github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.1 // indirect
    github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
    github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
    github.com/aws/aws-sdk-go-v2/service/sso v1.20.3 // indirect
    github.com/aws/aws-sdk-go-v2/service/ssooidc v1.23.3 // indirect
    github.com/aws/aws-sdk-go-v2/service/sts v1.28.5 // indirect
    github.com/aws/smithy-go v1.22.3 // indirect
)
