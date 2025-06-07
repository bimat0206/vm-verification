# Use the official AWS Lambda Go runtime as the base image
FROM public.ecr.aws/lambda/go:1.2023.08.16.12

# Copy the Go binary to the Lambda task root
COPY api-verifications-status ${LAMBDA_TASK_ROOT}

# Set the CMD to your handler (could also be done as a parameter override outside of the Dockerfile)
CMD [ "api-verifications-status" ]
