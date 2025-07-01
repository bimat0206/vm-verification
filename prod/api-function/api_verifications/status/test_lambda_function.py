"""
Unit tests for the verification status Lambda function
"""

import json
import os
import unittest
from decimal import Decimal
from unittest.mock import Mock, patch, MagicMock
import lambda_function


class TestLambdaFunction(unittest.TestCase):
    
    def setUp(self):
        """Set up test environment variables"""
        os.environ['DYNAMODB_VERIFICATION_TABLE'] = 'test-verification-table'
        os.environ['DYNAMODB_CONVERSATION_TABLE'] = 'test-conversation-table'
        os.environ['STATE_BUCKET'] = 'test-state-bucket'
        os.environ['AWS_REGION'] = 'us-east-1'
    
    def test_options_request(self):
        """Test CORS OPTIONS request"""
        event = {
            'httpMethod': 'OPTIONS',
            'path': '/api/verifications/status/test-id'
        }
        
        response = lambda_function.lambda_handler(event, None)
        
        self.assertEqual(response['statusCode'], 200)
        self.assertIn('Access-Control-Allow-Origin', response['headers'])
        self.assertEqual(response['body'], '')
    
    def test_invalid_method(self):
        """Test non-GET method returns 405"""
        event = {
            'httpMethod': 'POST',
            'path': '/api/verifications/status/test-id'
        }
        
        response = lambda_function.lambda_handler(event, None)
        
        self.assertEqual(response['statusCode'], 405)
        body = json.loads(response['body'])
        self.assertEqual(body['error'], 'Method not allowed')
    
    def test_missing_verification_id(self):
        """Test missing verification ID returns 400"""
        event = {
            'httpMethod': 'GET',
            'path': '/api/verifications/status/',
            'pathParameters': {},
            'queryStringParameters': {}
        }
        
        response = lambda_function.lambda_handler(event, None)
        
        self.assertEqual(response['statusCode'], 400)
        body = json.loads(response['body'])
        self.assertEqual(body['error'], 'Missing parameter')
    
    @patch('lambda_function.verification_table')
    def test_verification_not_found(self, mock_table):
        """Test verification not found returns 404"""
        mock_table.query.return_value = {'Items': []}
        
        event = {
            'httpMethod': 'GET',
            'pathParameters': {'verificationId': 'non-existent-id'}
        }
        
        response = lambda_function.lambda_handler(event, None)
        
        self.assertEqual(response['statusCode'], 404)
        body = json.loads(response['body'])
        self.assertEqual(body['error'], 'Verification not found')
    
    @patch('lambda_function.verification_table')
    @patch('lambda_function.s3_client')
    def test_successful_verification_retrieval(self, mock_s3, mock_table):
        """Test successful verification retrieval"""
        # Mock DynamoDB response
        mock_table.query.return_value = {
            'Items': [{
                'verificationId': 'test-id',
                'verificationAt': '2025-01-30T10:00:00Z',
                'currentStatus': 'COMPLETED',
                'verificationStatus': 'CORRECT',
                'overallAccuracy': Decimal('100.0'),
                'correctPositions': Decimal('50'),
                'discrepantPositions': Decimal('0'),
                'turn1ProcessedPath': 's3://bucket/turn1.md',
                'turn2ProcessedPath': 's3://bucket/turn2.md'
            }]
        }
        
        # Mock S3 response
        mock_s3.get_object.return_value = {
            'Body': MagicMock(read=lambda: b'LLM Response Content')
        }
        
        event = {
            'httpMethod': 'GET',
            'pathParameters': {'verificationId': 'test-id'}
        }
        
        response = lambda_function.lambda_handler(event, None)
        
        self.assertEqual(response['statusCode'], 200)
        body = json.loads(response['body'])
        self.assertEqual(body['verificationId'], 'test-id')
        self.assertEqual(body['verificationAt'], '2025-01-30T10:00:00Z')
        self.assertIn('message', body)
        self.assertEqual(body['status'], 'COMPLETED')
        self.assertEqual(body['currentStatus'], 'COMPLETED')
        self.assertEqual(body['verificationStatus'], 'CORRECT')
        self.assertIn('s3References', body)
        self.assertEqual(body['s3References']['turn1Processed'], 's3://bucket/turn1.md')
        self.assertEqual(body['s3References']['turn2Processed'], 's3://bucket/turn2.md')
        self.assertIn('llmAnalysis', body)
    
    def test_determine_overall_status(self):
        """Test status determination logic"""
        self.assertEqual(
            lambda_function.determine_overall_status('COMPLETED', 'CORRECT'),
            'COMPLETED'
        )
        self.assertEqual(
            lambda_function.determine_overall_status('RUNNING', 'PENDING'),
            'RUNNING'
        )
        self.assertEqual(
            lambda_function.determine_overall_status('FAILED', 'ERROR'),
            'FAILED'
        )
        self.assertEqual(
            lambda_function.determine_overall_status('', 'PENDING'),
            'RUNNING'
        )
    
    def test_get_status_message(self):
        """Test status message generation"""
        self.assertEqual(
            lambda_function.get_status_message('COMPLETED', 'CORRECT'),
            'Verification completed successfully - No discrepancies found'
        )
        self.assertEqual(
            lambda_function.get_status_message('COMPLETED', 'INCORRECT'),
            'Verification completed - Discrepancies detected'
        )
        self.assertEqual(
            lambda_function.get_status_message('RUNNING', ''),
            'Verification is currently in progress'
        )
        self.assertEqual(
            lambda_function.get_status_message('FAILED', ''),
            'Verification failed during processing'
        )


    def test_decimal_conversion(self):
        """Test Decimal conversion"""
        # Test convert_decimals function
        test_data = {
            'int_decimal': Decimal('100'),
            'float_decimal': Decimal('99.5'),
            'nested': {
                'value': Decimal('50.25')
            },
            'list': [Decimal('1'), Decimal('2.5')],
            'string': 'test'
        }
        
        result = lambda_function.convert_decimals(test_data)
        
        self.assertEqual(result['int_decimal'], 100)
        self.assertEqual(result['float_decimal'], 99.5)
        self.assertEqual(result['nested']['value'], 50.25)
        self.assertEqual(result['list'], [1, 2.5])
        self.assertEqual(result['string'], 'test')


if __name__ == '__main__':
    unittest.main()