# FinalizeWithError Lambda Function

This Lambda finalizes the verification workflow when an error occurs. It updates DynamoDB tables with failure status and logs details for RCA.

## Overview
- Updates `VerificationResults` with failure status and error details.
- Optionally updates `ConversationHistory` if present.
- Loads minimal context from `initialization.json` in S3 when available.
- Returns a structured response used by Step Functions to terminate the workflow.

See `CHANGELOG.md` for release history.
