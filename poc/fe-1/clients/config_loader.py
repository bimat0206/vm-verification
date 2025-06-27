import os
import json
import logging
from .aws_client import AWSClient

logger = logging.getLogger(__name__)

class ConfigLoader:
    """
    Configuration loader that can load configuration from either:
    1. CONFIG_SECRET environment variable (pointing to AWS Secrets Manager secret)
    2. Individual environment variables (fallback for backward compatibility)
    """

    def __init__(self):
        self.config = {}
        self._load_config()

    def _load_config(self):
        """Load configuration from CONFIG_SECRET, individual environment variables, or Streamlit secrets"""
        config_secret_name = os.environ.get('CONFIG_SECRET')

        if config_secret_name:
            logger.info(f"Loading configuration from secret: {config_secret_name}")
            try:
                self._load_from_secret(config_secret_name)
                logger.info("Successfully loaded configuration from secret")
                return
            except Exception as e:
                logger.error(f"Failed to load configuration from secret {config_secret_name}: {str(e)}")
                logger.info("Falling back to individual environment variables")

        # Try individual environment variables first
        self._load_from_env_vars()

        # If no API_ENDPOINT found or it's empty, try Streamlit secrets for local development
        if not self.config.get('API_ENDPOINT') or self.config.get('API_ENDPOINT').strip() == '':
            logger.info("No API_ENDPOINT found in environment variables, trying Streamlit secrets for local development")
            self._load_from_streamlit_secrets()

    def _load_from_secret(self, secret_name):
        """Load configuration from AWS Secrets Manager"""
        aws_client = AWSClient()
        response = aws_client.get_secret(secret_name)

        # Parse the secret string as JSON
        secret_data = json.loads(response['SecretString'])

        # Map the secret data to our config
        self.config = {
            'API_ENDPOINT': secret_data.get('API_ENDPOINT', ''),
            'REGION': secret_data.get('REGION', ''),
            'CHECKING_BUCKET': secret_data.get('CHECKING_BUCKET', ''),
            'DYNAMODB_CONVERSATION_TABLE': secret_data.get('DYNAMODB_CONVERSATION_TABLE', ''),
            'DYNAMODB_VERIFICATION_TABLE': secret_data.get('DYNAMODB_VERIFICATION_TABLE', ''),
            'REFERENCE_BUCKET': secret_data.get('REFERENCE_BUCKET', ''),
            'AWS_DEFAULT_REGION': secret_data.get('AWS_DEFAULT_REGION', ''),
            'API_KEY_SECRET_NAME': secret_data.get('API_KEY_SECRET_NAME', ''),
            # Legacy support for existing keys
            'DYNAMODB_TABLE': secret_data.get('DYNAMODB_TABLE', ''),
            'S3_BUCKET': secret_data.get('S3_BUCKET', '')
        }

    def _load_from_env_vars(self):
        """Load configuration from individual environment variables (fallback)"""
        logger.info("Loading configuration from individual environment variables")
        self.config = {
            'API_ENDPOINT': os.environ.get('API_ENDPOINT', ''),
            'REGION': os.environ.get('REGION', ''),
            'CHECKING_BUCKET': os.environ.get('CHECKING_BUCKET', ''),
            'DYNAMODB_CONVERSATION_TABLE': os.environ.get('DYNAMODB_CONVERSATION_TABLE', ''),
            'DYNAMODB_VERIFICATION_TABLE': os.environ.get('DYNAMODB_VERIFICATION_TABLE', ''),
            'REFERENCE_BUCKET': os.environ.get('REFERENCE_BUCKET', ''),
            'AWS_DEFAULT_REGION': os.environ.get('AWS_DEFAULT_REGION', ''),
            'API_KEY_SECRET_NAME': os.environ.get('API_KEY_SECRET_NAME', ''),
            # Legacy support for existing keys
            'DYNAMODB_TABLE': os.environ.get('DYNAMODB_TABLE', ''),
            'S3_BUCKET': os.environ.get('S3_BUCKET', '')
        }

    def get(self, key, default=None):
        """Get a configuration value"""
        return self.config.get(key, default)

    def get_all(self):
        """Get all configuration values"""
        return self.config.copy()

    def _load_from_streamlit_secrets(self):
        """Load configuration from Streamlit secrets for local development"""
        try:
            # Only import streamlit when actually needed
            import streamlit as st
            if hasattr(st, 'secrets') and st.secrets:
                logger.info("Loading configuration from Streamlit secrets for local development")
                # Update config with values from Streamlit secrets, only if they're not empty
                for key in ['API_ENDPOINT', 'REGION', 'CHECKING_BUCKET', 'DYNAMODB_CONVERSATION_TABLE',
                           'DYNAMODB_VERIFICATION_TABLE', 'REFERENCE_BUCKET', 'AWS_DEFAULT_REGION',
                           'API_KEY_SECRET_NAME', 'DYNAMODB_TABLE', 'S3_BUCKET']:
                    secret_value = st.secrets.get(key, '')
                    if secret_value and secret_value.strip():
                        self.config[key] = secret_value
                # Mark that we loaded from Streamlit secrets
                self._loaded_from_streamlit = True
                logger.info("Successfully loaded configuration from Streamlit secrets")
            else:
                logger.debug("Streamlit secrets not available or empty")
                self._loaded_from_streamlit = False
        except ImportError:
            logger.debug("Streamlit not available - likely running in non-Streamlit environment")
            self._loaded_from_streamlit = False
        except Exception as e:
            logger.warning(f"Failed to load from Streamlit secrets: {str(e)}")
            self._loaded_from_streamlit = False

    def is_loaded_from_secret(self):
        """Check if configuration was loaded from AWS Secrets Manager"""
        return os.environ.get('CONFIG_SECRET') is not None

    def is_loaded_from_streamlit(self):
        """Check if configuration was loaded from Streamlit secrets"""
        return getattr(self, '_loaded_from_streamlit', False)
