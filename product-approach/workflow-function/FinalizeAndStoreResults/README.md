# FinalizeAndStoreResults

This Lambda function finalizes a vending machine verification workflow by loading
results from S3, storing them in DynamoDB and updating conversation history.
It expects S3 references to initialization data and processed Turn 2 results
from the Step Functions state machine.

## Input Structure

The function supports the standard nested s3References structure:

```json
{
  "schemaVersion": "2.1.0",
  "s3References": {
    "processing_initialization": {
      "bucket": "bucket-name",
      "key": "path/to/initialization.json"
    },
    "responses": {
      "turn2Processed": {
        "bucket": "bucket-name",
        "key": "path/to/turn2-processed-response.md",
        "size": 151
      }
    }
  },
  "verificationId": "verif-20250605025241-3145",
  "status": "TURN2_COMPLETED"
}
```

## Processing

The function retrieves initialization metadata, parses the Turn 2 output to
extract the verification summary, then writes a consolidated record to the
`VerificationResults` table and marks the conversation as completed in the
`ConversationHistory` table.

The output of the Lambda provides a concise summary including the final
verification status and accuracy metrics.

