# ExecuteTurn2Combined

This Lambda function represents the second phase of the vending machine verification workflow. It loads the checking image, applies a prompt that references the analysis from Turn 1, invokes the LLM via Bedrock and stores the results in S3 and DynamoDB.
Prompt templates are cached in memory for faster load times. Set `TEMPLATE_CACHE_ENABLED=false` to disable caching.

This directory only provides a skeleton implementation to outline the intended structure. The real implementation should mirror the architecture of `ExecuteTurn1Combined` and make use of the shared packages in `workflow-function/shared`.
