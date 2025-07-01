"""
AWS Lambda function for verification status API.
Provides GET endpoint to retrieve verification status from DynamoDB.
"""

import json
import os
import logging
from datetime import datetime
from typing import Dict, Optional, Any
from decimal import Decimal
import boto3
from boto3.dynamodb.conditions import Key
from botocore.exceptions import ClientError

# Initialize structured logger similar to logrus
class StructuredLogger:
    """Structured logger that mimics logrus format"""
    
    class StructuredFormatter(logging.Formatter):
        """Custom formatter for structured JSON logging"""
        
        def __init__(self):
            super().__init__(datefmt='%Y-%m-%dT%H:%M:%SZ')
        
        def format(self, record):
            # Base JSON structure
            log_entry = {
                "level": record.levelname.lower(),
                "time": self.formatTime(record, self.datefmt),
                "msg": record.getMessage()
            }
            
            # Add extra fields if they exist
            if hasattr(record, 'fields') and record.fields:
                try:
                    field_dict = json.loads("{" + record.fields + "}")
                    log_entry.update(field_dict)
                except:
                    pass
            
            return json.dumps(log_entry)
    
    def __init__(self):
        self.logger = logging.getLogger()
        log_level = os.environ.get('LOG_LEVEL', 'INFO')
        self.logger.setLevel(getattr(logging, log_level.upper()))
        
        # Configure JSON formatter similar to logrus
        formatter = self.StructuredFormatter()
        
        # Set up handler if not already configured
        if not self.logger.handlers:
            handler = logging.StreamHandler()
            handler.setFormatter(formatter)
            self.logger.addHandler(handler)
    
    def _format_fields(self, **fields):
        """Format additional fields for structured logging"""
        if not fields:
            return ""
        
        field_strings = []
        for key, value in fields.items():
            if isinstance(value, str):
                field_strings.append(f'"{key}":"{value}"')
            elif isinstance(value, (int, float, bool)):
                field_strings.append(f'"{key}":{json.dumps(value)}')
            elif value is None:
                field_strings.append(f'"{key}":null')
            else:
                field_strings.append(f'"{key}":"{str(value)}"')
        
        return "," + ",".join(field_strings)
    
    def with_fields(self, **fields):
        """Return a logger context with fields (similar to logrus WithFields)"""
        return LoggerContext(self, fields)
    
    def debug(self, msg, **fields):
        """Debug level logging with optional fields"""
        self.logger.debug(msg, extra={'fields': self._format_fields(**fields)})
    
    def info(self, msg, **fields):
        """Info level logging with optional fields"""
        self.logger.info(msg, extra={'fields': self._format_fields(**fields)})
    
    def warning(self, msg, **fields):
        """Warning level logging with optional fields"""
        self.logger.warning(msg, extra={'fields': self._format_fields(**fields)})
    
    def warn(self, msg, **fields):
        """Alias for warning (like logrus)"""
        self.warning(msg, **fields)
    
    def error(self, msg, **fields):
        """Error level logging with optional fields"""
        self.logger.error(msg, extra={'fields': self._format_fields(**fields)})
    
    def fatal(self, msg, **fields):
        """Fatal level logging with optional fields"""
        self.logger.critical(msg, extra={'fields': self._format_fields(**fields)})


class LoggerContext:
    """Logger context that carries fields (similar to logrus WithFields)"""
    
    def __init__(self, logger, fields):
        self.logger = logger
        self.fields = fields
    
    def with_fields(self, **additional_fields):
        """Add more fields to the context"""
        combined_fields = {**self.fields, **additional_fields}
        return LoggerContext(self.logger, combined_fields)
    
    def debug(self, msg, **extra_fields):
        combined_fields = {**self.fields, **extra_fields}
        self.logger.debug(msg, **combined_fields)
    
    def info(self, msg, **extra_fields):
        combined_fields = {**self.fields, **extra_fields}
        self.logger.info(msg, **combined_fields)
    
    def warning(self, msg, **extra_fields):
        combined_fields = {**self.fields, **extra_fields}
        self.logger.warning(msg, **combined_fields)
    
    def warn(self, msg, **extra_fields):
        self.warning(msg, **extra_fields)
    
    def error(self, msg, **extra_fields):
        combined_fields = {**self.fields, **extra_fields}
        self.logger.error(msg, **combined_fields)
    
    def fatal(self, msg, **extra_fields):
        combined_fields = {**self.fields, **extra_fields}
        self.logger.fatal(msg, **combined_fields)


# Initialize the structured logger
log = StructuredLogger()


class DecimalEncoder(json.JSONEncoder):
    """Custom JSON encoder for Decimal objects from DynamoDB"""
    def default(self, obj):
        if isinstance(obj, Decimal):
            # Convert Decimal to int or float
            if obj % 1 == 0:
                return int(obj)
            else:
                return float(obj)
        return super(DecimalEncoder, self).default(obj)


def convert_decimals(obj):
    """Recursively convert Decimal objects to int/float in nested structures"""
    if isinstance(obj, list):
        return [convert_decimals(item) for item in obj]
    elif isinstance(obj, dict):
        return {k: convert_decimals(v) for k, v in obj.items()}
    elif isinstance(obj, Decimal):
        if obj % 1 == 0:
            return int(obj)
        else:
            return float(obj)
    else:
        return obj

# Environment variables
VERIFICATION_TABLE = os.environ.get('DYNAMODB_VERIFICATION_TABLE')
CONVERSATION_TABLE = os.environ.get('DYNAMODB_CONVERSATION_TABLE')
STATE_BUCKET = os.environ.get('STATE_BUCKET')
STEP_FUNCTIONS_ARN = os.environ.get('STEP_FUNCTIONS_STATE_MACHINE_ARN')
AWS_REGION = os.environ.get('AWS_REGION', os.environ.get('REGION', 'us-east-1'))

# Validate required environment variables
if not VERIFICATION_TABLE:
    raise ValueError("DYNAMODB_VERIFICATION_TABLE environment variable is required")
if not CONVERSATION_TABLE:
    raise ValueError("DYNAMODB_CONVERSATION_TABLE environment variable is required")

# Initialize AWS clients
dynamodb = boto3.resource('dynamodb', region_name=AWS_REGION)
s3_client = boto3.client('s3', region_name=AWS_REGION)

# Get DynamoDB table
verification_table = dynamodb.Table(VERIFICATION_TABLE)

# Log configuration
log.info("Initialized with configuration",
         region=AWS_REGION,
         verificationTable=VERIFICATION_TABLE,
         conversationTable=CONVERSATION_TABLE,
         stateBucket=STATE_BUCKET)


def lambda_handler(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Main Lambda handler for GET /api/verifications/status/{verificationId}
    """
    # Log request with structured fields
    log.with_fields(
        method=event.get('httpMethod'),
        path=event.get('path'),
        params=event.get('pathParameters'),
        requestId=context.aws_request_id if context else None
    ).info("Get verification status request received")
    
    # CORS headers
    headers = {
        'Content-Type': 'application/json',
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Credentials': 'true',
        'Access-Control-Allow-Headers': 'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token',
        'Access-Control-Allow-Methods': 'GET,OPTIONS'
    }
    
    # Handle OPTIONS request for CORS
    if event.get('httpMethod') == 'OPTIONS':
        return {
            'statusCode': 200,
            'headers': headers,
            'body': ''
        }
    
    # Only allow GET requests
    if event.get('httpMethod') != 'GET':
        return create_error_response(
            405, 
            'Method not allowed', 
            'Only GET requests are supported',
            headers
        )
    
    # Extract verification ID
    verification_id = None
    path_params = event.get('pathParameters', {})
    if path_params:
        verification_id = path_params.get('verificationId')
    
    # Fallback to query parameters
    if not verification_id:
        query_params = event.get('queryStringParameters', {})
        if query_params:
            verification_id = query_params.get('verificationId')
    
    if not verification_id:
        return create_error_response(
            400,
            'Missing parameter',
            'verificationId path or query parameter is required',
            headers
        )
    
    log.with_fields(verificationId=verification_id).info("Processing get verification status request")
    
    try:
        # Query DynamoDB for verification record
        verification_record = get_verification_record(verification_id)
        
        if not verification_record:
            return create_error_response(
                404,
                'Verification not found',
                f'No verification found for verificationId: {verification_id}',
                headers
            )
        
        # Build response
        response = build_status_response(verification_record)
        
        log.with_fields(
            verificationId=verification_id,
            status=response['status'],
            currentStatus=response.get('currentStatus'),
            verificationStatus=response.get('verificationStatus')
        ).info("Get verification status request completed successfully")
        
        return {
            'statusCode': 200,
            'headers': headers,
            'body': json.dumps(response, cls=DecimalEncoder)
        }
        
    except Exception as e:
        log.with_fields(error=str(e)).error("Error processing request")
        return create_error_response(
            500,
            'Internal server error',
            'Failed to process request',
            headers
        )


def get_verification_record(verification_id: str) -> Optional[Dict[str, Any]]:
    """
    Query DynamoDB for the most recent verification record
    """
    try:
        log.with_fields(
            verificationId=verification_id,
            tableName=VERIFICATION_TABLE,
            region=AWS_REGION
        ).debug("Querying DynamoDB for verification record")
        
        # Query using verificationId as hash key
        response = verification_table.query(
            KeyConditionExpression=Key('verificationId').eq(verification_id),
            ScanIndexForward=False,  # Sort by verificationAt descending
            Limit=1  # Get only the most recent record
        )
        
        items = response.get('Items', [])
        if not items:
            log.with_fields(verificationId=verification_id).info("No verification record found")
            return None
        
        record = items[0]
        log.with_fields(
            verificationId=record.get('verificationId'),
            verificationAt=record.get('verificationAt'),
            currentStatus=record.get('currentStatus'),
            verificationStatus=record.get('verificationStatus'),
            turn1ProcessedPath=record.get('turn1ProcessedPath'),
            turn2ProcessedPath=record.get('turn2ProcessedPath')
        ).debug("Successfully retrieved verification record")
        
        return record
        
    except ClientError as e:
        log.with_fields(error=str(e)).error("DynamoDB query failed")
        raise


def build_status_response(record: Dict[str, Any]) -> Dict[str, Any]:
    """
    Build the API response from the verification record
    """
    # Determine overall status
    current_status = record.get('currentStatus', '')
    verification_status = record.get('verificationStatus', '')
    status = determine_overall_status(current_status, verification_status)
    
    # Extract numeric values safely and convert Decimal to appropriate type
    overall_accuracy = record.get('overallAccuracy')
    correct_positions = record.get('correctPositions')
    discrepant_positions = record.get('discrepantPositions')
    
    # Build response
    response = {
        'verificationId': record.get('verificationId', ''),
        'verificationAt': record.get('verificationAt', ''),
        'message': get_status_message(status, verification_status),
        'status': status,
        'currentStatus': current_status,
        'verificationStatus': verification_status,
        's3References': {
            'turn1Processed': record.get('turn1ProcessedPath', ''),
            'turn2Processed': record.get('turn2ProcessedPath', '')
        }
    }
    
    # Add verification summary if present, converting any Decimal values
    if 'verificationSummary' in record:
        response['verificationSummary'] = convert_decimals(record['verificationSummary'])
    
    # Retrieve LLM response from S3 if completed
    if status == 'COMPLETED' and record.get('turn2ProcessedPath'):
        try:
            llm_response = get_s3_content(record['turn2ProcessedPath'])
            if llm_response:
                response['llmAnalysis'] = llm_response  # For frontend compatibility
        except Exception as e:
            log.with_fields(
                s3Path=record.get('turn2ProcessedPath'),
                error=str(e)
            ).warn("Failed to retrieve LLM response from S3")
    
    return response


def determine_overall_status(current_status: str, verification_status: str) -> str:
    """
    Determine the overall status based on current and verification status
    """
    if current_status == 'COMPLETED':
        return 'COMPLETED'
    elif current_status in ['RUNNING', 'TURN1_COMPLETED', 'TURN2_RUNNING', 'PROCESSING']:
        return 'RUNNING'
    elif current_status in ['FAILED', 'ERROR']:
        return 'FAILED'
    else:
        # If current status is empty or unknown, check verification status
        if verification_status in ['PENDING', '']:
            return 'RUNNING'
        return 'COMPLETED'


def get_status_message(status: str, verification_status: str) -> str:
    """
    Generate appropriate status message
    """
    if status == 'COMPLETED':
        if verification_status == 'CORRECT':
            return 'Verification completed successfully - No discrepancies found'
        elif verification_status == 'INCORRECT':
            return 'Verification completed - Discrepancies detected'
        else:
            return 'Verification completed'
    elif status == 'RUNNING':
        return 'Verification is currently in progress'
    elif status == 'FAILED':
        return 'Verification failed during processing'
    else:
        return 'Verification status unknown'


def get_s3_content(s3_path: str) -> Optional[str]:
    """
    Retrieve content from S3
    """
    if not s3_path or not s3_path.startswith('s3://'):
        log.with_fields(s3Path=s3_path).warn("Invalid S3 path")
        return None
    
    # Parse S3 path
    path_parts = s3_path[5:].split('/', 1)
    if len(path_parts) != 2:
        log.with_fields(s3Path=s3_path).warn("Invalid S3 path format")
        return None
    
    bucket = path_parts[0]
    key = path_parts[1]
    
    log.with_fields(
        s3Path=s3_path,
        bucket=bucket,
        key=key
    ).debug("Retrieving content from S3")
    
    try:
        response = s3_client.get_object(Bucket=bucket, Key=key)
        content = response['Body'].read().decode('utf-8')
        log.with_fields(
            s3Path=s3_path,
            contentLength=len(content)
        ).debug("Successfully retrieved S3 content")
        return content
    except ClientError as e:
        log.with_fields(
            s3Path=s3_path,
            bucket=bucket,
            key=key,
            error=str(e)
        ).error("Failed to get S3 object")
        raise


def create_error_response(status_code: int, error: str, message: str, headers: Dict[str, str]) -> Dict[str, Any]:
    """
    Create standardized error response
    """
    error_body = {
        'error': error,
        'message': message,
        'code': f'HTTP_{status_code}'
    }
    
    return {
        'statusCode': status_code,
        'headers': headers,
        'body': json.dumps(error_body, cls=DecimalEncoder)
    }