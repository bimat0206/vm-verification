#!/usr/bin/env python3
import os
import sys
import json
import logging
import requests
import traceback

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s %(levelname)s %(message)s'
)
logger = logging.getLogger(__name__)

# Import our lambda handler
from lambda_handler import lambda_handler

def main():
    logger.info("Starting Lambda bootstrap in direct mode")
    
    # Check if we're running in Lambda
    if os.environ.get('AWS_LAMBDA_RUNTIME_API'):
        runtime_api = os.environ['AWS_LAMBDA_RUNTIME_API']
        logger.info(f"Using Lambda Runtime API at: {runtime_api}")
        
        # Main processing loop
        while True:
            # Get the next event
            logger.info("Waiting for next event...")
            event_url = f"http://{runtime_api}/2018-06-01/runtime/invocation/next"
            try:
                response = requests.get(event_url)
                request_id = response.headers.get('Lambda-Runtime-Aws-Request-Id')
                logger.info(f"Received event with request ID: {request_id}")
                
                # Mock Lambda context object
                class LambdaContext:
                    def __init__(self, request_id):
                        self.aws_request_id = request_id
                        self.function_name = os.environ.get('AWS_LAMBDA_FUNCTION_NAME', 'unknown')
                        self.memory_limit_in_mb = os.environ.get('AWS_LAMBDA_FUNCTION_MEMORY_SIZE', '128')
                        self.log_group_name = os.environ.get('AWS_LAMBDA_LOG_GROUP_NAME', 'unknown')
                        self.log_stream_name = os.environ.get('AWS_LAMBDA_LOG_STREAM_NAME', 'unknown')
                
                # Create context and parse event
                context = LambdaContext(request_id)
                event = json.loads(response.content)
                
                # Process the event
                logger.info(f"Processing event: {json.dumps(event)}")
                result = lambda_handler(event, context)
                
                # Send the response
                response_url = f"http://{runtime_api}/2018-06-01/runtime/invocation/{request_id}/response"
                logger.info(f"Sending response for request ID: {request_id}")
                requests.post(response_url, json=result)
                
            except Exception as e:
                logger.error(f"Error processing request: {str(e)}")
                error_traceback = traceback.format_exc()
                logger.error(f"Traceback: {error_traceback}")
                
                # Try to send error response if we have a request_id
                if 'request_id' in locals():
                    error_url = f"http://{runtime_api}/2018-06-01/runtime/invocation/{request_id}/error"
                    error_payload = {
                        "errorMessage": str(e),
                        "errorType": type(e).__name__,
                        "stackTrace": error_traceback.split("\n")
                    }
                    try:
                        requests.post(error_url, json=error_payload)
                    except Exception as e2:
                        logger.error(f"Failed to send error response: {str(e2)}")
    else:
        # Not in Lambda environment
        logger.info("Not running in AWS Lambda environment")
        
        # Process a test event if provided
        test_event_path = os.environ.get('TEST_EVENT_PATH')
        if test_event_path and os.path.exists(test_event_path):
            logger.info(f"Processing test event from {test_event_path}")
            with open(test_event_path, 'r') as f:
                test_event = json.load(f)
            
            result = lambda_handler(test_event, None)
            logger.info(f"Test result: {json.dumps(result, indent=2)}")
        else:
            logger.info("No test event provided, exiting.")

if __name__ == "__main__":
    main()