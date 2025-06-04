# FinalizeWithError Lambda Function

This Lambda finalizes the verification workflow when an error occurs. It updates DynamoDB tables with failure status and logs details for RCA.

## Overview
 - Sets the `status` field in the results table to `VERIFICATION_FAILED`.
 - Updates `VerificationResults` with failure status and error details.
 - Sets the `status` field in the results table to `VERIFICATION_FAILED`.
 - Optionally updates `ConversationHistory` if present.
   - Looks up the latest conversation entry and sets `turnStatus` to `FAILED_WORKFLOW`.
- Loads minimal context from `initialization.json` in S3 when available.
- Returns a structured response used by Step Functions to terminate the workflow.

See `CHANGELOG.md` for release history.
