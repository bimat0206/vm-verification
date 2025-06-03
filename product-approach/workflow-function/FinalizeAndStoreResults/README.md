# FinalizeAndStoreResults

This Lambda function finalizes a vending machine verification workflow by loading
results from S3, storing them in DynamoDB and updating conversation history.
It expects S3 references to initialization data and processed Turn 2 results
from the Step Functions state machine.

The function retrieves initialization metadata, parses the Turn 2 output to
extract the verification summary, then writes a consolidated record to the
`VerificationResults` table and marks the conversation as completed in the
`ConversationHistory` table.

The output of the Lambda provides a concise summary including the final
verification status and accuracy metrics.

