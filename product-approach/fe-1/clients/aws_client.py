import boto3
import logging

logger = logging.getLogger(__name__)

class AWSClient:
    def __init__(self):
        try:
            self.s3 = boto3.client('s3')
            self.dynamodb = boto3.resource('dynamodb')
            self.secretsmanager = boto3.client('secretsmanager')
        except Exception as e:
            logger.error(f"Failed to initialize AWS clients: {str(e)}")
            raise

    def get_secret(self, secret_name):
        try:
            response = self.secretsmanager.get_secret_value(SecretId=secret_name)
            return response
        except Exception as e:
            logger.error(f"Failed to get secret {secret_name}: {str(e)}")
            raise