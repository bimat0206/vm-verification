#!/usr/bin/env python3
"""
Local test script for the Lambda function
"""

import os
import json
from decimal import Decimal

# Set environment variables for testing
os.environ['DYNAMODB_VERIFICATION_TABLE'] = 'test-table'
os.environ['DYNAMODB_CONVERSATION_TABLE'] = 'test-conversation-table'
os.environ['AWS_REGION'] = 'us-east-1'
os.environ['LOG_LEVEL'] = 'DEBUG'

# Import after setting environment variables
import lambda_function


def test_structured_logging():
    """Test the structured logging functionality"""
    print("Testing structured logging...")
    
    # Test basic logging
    lambda_function.log.info("Test info message", testField="testValue", count=42)
    lambda_function.log.debug("Test debug message", nested={"key": "value"})
    lambda_function.log.warn("Test warning message", warningType="test")
    lambda_function.log.error("Test error message", errorCode=500)
    
    # Test with_fields functionality
    context_logger = lambda_function.log.with_fields(
        requestId="test-request-123",
        verificationId="test-verification-456"
    )
    context_logger.info("Test context logging")
    context_logger.debug("Test context debug", additionalInfo="extra")
    
    # Test nested with_fields
    nested_logger = context_logger.with_fields(operation="test", step=1)
    nested_logger.info("Test nested context")
    
    print("✓ Structured logging test completed")


def test_decimal_conversion():
    """Test the Decimal conversion functionality"""
    print("Testing Decimal conversion...")
    
    test_data = {
        'int_decimal': Decimal('100'),
        'float_decimal': Decimal('99.5'),
        'nested': {
            'value': Decimal('50.25'),
            'items': [Decimal('1'), Decimal('2.5'), 'string']
        },
        'string': 'test',
        'normal_int': 42
    }
    
    result = lambda_function.convert_decimals(test_data)
    
    print(f"Original: {test_data}")
    print(f"Converted: {result}")
    
    # Test JSON serialization
    json_str = json.dumps(result, cls=lambda_function.DecimalEncoder)
    print(f"JSON serializable: {json_str}")
    
    print("✓ Decimal conversion test passed")


def test_error_response():
    """Test error response creation"""
    print("Testing error response creation...")
    
    headers = {'Content-Type': 'application/json'}
    response = lambda_function.create_error_response(404, "Not Found", "Test message", headers)
    
    print(f"Error response: {response}")
    
    # Ensure it's JSON serializable
    json.loads(response['body'])
    print("✓ Error response test passed")


def test_status_messages():
    """Test status message generation"""
    print("Testing status message generation...")
    
    test_cases = [
        ('COMPLETED', 'CORRECT', 'Verification completed successfully - No discrepancies found'),
        ('COMPLETED', 'INCORRECT', 'Verification completed - Discrepancies detected'),
        ('RUNNING', '', 'Verification is currently in progress'),
        ('FAILED', '', 'Verification failed during processing'),
    ]
    
    for status, verification_status, expected in test_cases:
        result = lambda_function.get_status_message(status, verification_status)
        assert result == expected, f"Expected '{expected}', got '{result}'"
        print(f"✓ {status}/{verification_status}: {result}")
    
    print("✓ Status message tests passed")


if __name__ == '__main__':
    print("Running local tests for Lambda function...")
    print("=" * 50)
    
    try:
        test_structured_logging()
        print()
        test_decimal_conversion()
        print()
        test_error_response()
        print()
        test_status_messages()
        print()
        print("✅ All local tests passed!")
    except Exception as e:
        print(f"❌ Test failed: {e}")
        import traceback
        traceback.print_exc()
        exit(1)