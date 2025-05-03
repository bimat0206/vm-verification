import requests
import os
import logging
from aws_client import AWSClient

logger = logging.getLogger(__name__)

class APIClient:
    def __init__(self):
        self.base_url = os.environ.get('API_ENDPOINT', '')
        if not self.base_url:
            raise ValueError("API_ENDPOINT environment variable not set")
        self.api_key = self.get_api_key()

    def get_api_key(self):
        if 'API_KEY' in os.environ:
            return os.environ['API_KEY']
        elif 'API_KEY_SECRET_NAME' in os.environ:
            aws_client = AWSClient()
            secret_name = os.environ['API_KEY_SECRET_NAME']
            try:
                secret = aws_client.get_secret(secret_name)
                return secret['SecretString']
            except Exception as e:
                logger.error(f"Failed to retrieve API key from Secrets Manager: {str(e)}")
                raise
        else:
            raise ValueError("API key not found in environment or secrets")

    def make_request(self, method, endpoint, params=None, data=None):
        headers = {'x-api-key': self.api_key}
        url = f"{self.base_url}/{endpoint}"
        try:
            response = requests.request(method, url, headers=headers, params=params, json=data, timeout=10)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            logger.error(f"API request failed: {method} {endpoint} - {str(e)}")
            raise Exception(f"Failed to {method.lower()} {endpoint}: {str(e)}")

    def initiate_verification(self, data):
        return self.make_request('POST', 'api/v1/verifications', data=data)

    def list_verifications(self, params=None):
        return self.make_request('GET', 'api/v1/verifications', params=params)

    def get_verification_details(self, verification_id):
        return self.make_request('GET', f'api/v1/verifications/{verification_id}')

    def get_verification_conversation(self, verification_id):
        return self.make_request('GET', f'api/v1/verifications/{verification_id}/conversation')

    def browse_images(self, path=''):
        return self.make_request('GET', f'api/v1/images/browser/{path}')

    def get_image_url(self, key):
        return self.make_request('GET', f'api/v1/images/{key}/view')

    def health_check(self):
        return self.make_request('GET', 'health')