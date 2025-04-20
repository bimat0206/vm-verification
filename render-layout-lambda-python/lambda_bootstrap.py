#!/usr/bin/env python3
import os
import sys
import json
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s %(levelname)s %(message)s'
)
logger = logging.getLogger(__name__)

# Import our lambda handler
from lambda_handler import lambda_handler

# Import AWS Lambda Runtime Interface Client if available
try:
    import awslambdaric
    HAS_LAMBDA_RIC = True
except ImportError:
    HAS_LAMBDA_RIC = False
    logger.warning("AWS Lambda Runtime Interface Client not found - assuming local execution")

def main():
    logger.info("Starting Lambda bootstrap")
    
    # Check environment to determine execution mode
    if os.environ.get('AWS_LAMBDA_RUNTIME_API'):
        logger.info("Running in AWS Lambda environment")
        
        # If awslambdaric is available, use it
        if HAS_LAMBDA_RIC:
            logger.info("Using AWS Lambda Runtime Interface Client")
            from awslambdaric import bootstrap
            bootstrap.run(lambda_handler)
        else:
            # Manual implementation of Lambda Runtime API
            logger.info("Using manual Lambda Runtime API implementation")
            runtime_api = os.environ['AWS_LAMBDA_RUNTIME_API']
            while True:
                # Get next invocation
                import requests
                resp = requests.get(f"http://{runtime_api}/2018-06-01/runtime/invocation/next")
                request_id = resp.headers.get('Lambda-Runtime-Aws-Request-Id')
                
                # Process the event
                try:
                    event = json.loads(resp.content)
                    result = lambda_handler(event, None)
                    
                    # Send the response
                    requests.post(
                        f"http://{runtime_api}/2018-06-01/runtime/invocation/{request_id}/response",
                        json=result
                    )
                except Exception as e:
                    logger.error(f"Error processing request: {e}")
                    error_payload = {
                        "errorMessage": str(e),
                        "errorType": type(e).__name__
                    }
                    requests.post(
                        f"http://{runtime_api}/2018-06-01/runtime/invocation/{request_id}/error",
                        json=error_payload
                    )
    else:
        # Not in Lambda environment, process test event if provided
        logger.info("Not running in AWS Lambda environment")
        
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