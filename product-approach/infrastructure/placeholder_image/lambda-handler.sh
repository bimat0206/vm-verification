#!/bin/sh

# Start nginx in the background
nginx -g "daemon off;" &

# Echo a simple response for Lambda invocations
while true; do
  echo '{"statusCode": 200, "body": "This is a placeholder Lambda function. Replace with actual implementation."}'
  sleep 1
done