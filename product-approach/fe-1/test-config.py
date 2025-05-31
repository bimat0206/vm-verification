#!/usr/bin/env python3
"""
Test script to verify configuration loading for local development
Run this before starting the Streamlit app to ensure everything is configured correctly
"""

import os
import sys
import logging

# Add the current directory to Python path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from clients.config_loader import ConfigLoader
from clients.api_client import APIClient

def test_configuration():
    """Test configuration loading and API client initialization"""
    print("🧪 Testing configuration loading...")
    
    try:
        # Test ConfigLoader
        print("\n1️⃣ Testing ConfigLoader...")
        config_loader = ConfigLoader()
        
        # Check configuration source
        if config_loader.is_loaded_from_secret():
            print("✅ Configuration loaded from AWS Secrets Manager")
        elif config_loader.is_loaded_from_streamlit():
            print("✅ Configuration loaded from Streamlit secrets (local development)")
        else:
            print("✅ Configuration loaded from environment variables")
        
        # Check required configuration
        api_endpoint = config_loader.get('API_ENDPOINT', '')
        region = config_loader.get('REGION', '')
        
        print(f"   API Endpoint: {api_endpoint[:50]}{'...' if len(api_endpoint) > 50 else ''}")
        print(f"   Region: {region}")
        
        if not api_endpoint:
            print("❌ API_ENDPOINT not found in configuration")
            return False
        
        # Test APIClient
        print("\n2️⃣ Testing APIClient...")
        api_client = APIClient()
        print("✅ APIClient initialized successfully")
        
        # Test API connection (optional)
        print("\n3️⃣ Testing API connection...")
        try:
            health_response = api_client.health_check()
            print("✅ API health check successful")
            print(f"   Response: {health_response}")
        except Exception as e:
            print(f"⚠️  API health check failed: {str(e)}")
            print("   This might be normal if the API is not accessible from your network")
        
        print("\n🎉 Configuration test completed successfully!")
        print("🚀 You can now run: streamlit run app.py")
        return True
        
    except Exception as e:
        print(f"\n❌ Configuration test failed: {str(e)}")
        print("\n💡 Troubleshooting tips:")
        print("   1. Run ./setup-local-dev.sh to auto-configure")
        print("   2. Check .streamlit/secrets.toml exists and has correct values")
        print("   3. Verify AWS credentials: aws sts get-caller-identity")
        print("   4. See LOCAL_DEVELOPMENT.md for detailed setup instructions")
        return False

def show_configuration_info():
    """Show current configuration information"""
    print("📋 Current configuration sources:")
    
    # Check environment variables
    env_vars = ['API_ENDPOINT', 'API_KEY', 'REGION', 'CONFIG_SECRET', 'API_KEY_SECRET_NAME']
    print("\n🌍 Environment Variables:")
    for var in env_vars:
        value = os.environ.get(var, '')
        if value:
            # Mask sensitive values
            if 'KEY' in var:
                display_value = '*' * len(value)
            else:
                display_value = value[:30] + '...' if len(value) > 30 else value
            print(f"   {var}: {display_value}")
        else:
            print(f"   {var}: ❌ Not set")
    
    # Check Streamlit secrets
    secrets_file = '.streamlit/secrets.toml'
    print(f"\n🔧 Streamlit Secrets ({secrets_file}):")
    if os.path.exists(secrets_file):
        print("   ✅ File exists")
        try:
            with open(secrets_file, 'r') as f:
                lines = f.readlines()
                config_lines = [line.strip() for line in lines if '=' in line and not line.strip().startswith('#')]
                print(f"   📝 Contains {len(config_lines)} configuration entries")
        except Exception as e:
            print(f"   ❌ Error reading file: {e}")
    else:
        print("   ❌ File not found")
    
    # Check AWS credentials
    print("\n🔑 AWS Credentials:")
    try:
        import boto3
        sts = boto3.client('sts')
        identity = sts.get_caller_identity()
        print(f"   ✅ Account: {identity.get('Account', 'Unknown')}")
        print(f"   ✅ User/Role: {identity.get('Arn', 'Unknown').split('/')[-1]}")
    except Exception as e:
        print(f"   ❌ Not configured or invalid: {e}")

if __name__ == "__main__":
    print("🔍 Configuration Test for Streamlit App")
    print("=" * 50)
    
    # Show configuration info
    show_configuration_info()
    
    print("\n" + "=" * 50)
    
    # Test configuration
    success = test_configuration()
    
    sys.exit(0 if success else 1)
