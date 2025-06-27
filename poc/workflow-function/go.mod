module workflow-function

replace (
	workflow-function/shared/dbutils => ./shared/dbutils
	workflow-function/shared/logger => ./shared/logger
	workflow-function/shared/s3utils => ./shared/s3utils
	workflow-function/shared/schema => ./shared/schema
)

go 1.22

toolchain go1.24.0
