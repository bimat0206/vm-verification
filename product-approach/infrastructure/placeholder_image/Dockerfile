FROM public.ecr.aws/nginx/nginx:latest

# Add a simple HTML page
RUN echo '<html><body><h1>Placeholder Lambda Container</h1><p>This is a placeholder container for Lambda functions.</p></body></html>' > /usr/share/nginx/html/index.html

# Expose port 8080 - AWS Lambda container runtime API uses this port
EXPOSE 8080

# Use a simple script to handle Lambda invocations
COPY lambda-handler.sh /
RUN chmod +x /lambda-handler.sh

# Set the entrypoint to the handler script
ENTRYPOINT ["/lambda-handler.sh"]