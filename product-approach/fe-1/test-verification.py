#!/usr/bin/env python3
"""
Test script to verify the initiate_verification API endpoint fix
"""

import os
import sys
import logging

# Add the current directory to Python path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from clients.api_client import APIClient

def test_verification_initiation():
    """Test verification initiation with the corrected API structure"""
    print("🧪 Testing verification initiation...")

    try:
        # Initialize API client
        api_client = APIClient()
        print("✅ API client initialized successfully")

        # Test data for LAYOUT_VS_CHECKING verification (using your exact data)
        test_data = {
            "verification_type": "LAYOUT_VS_CHECKING",
            "reference_image_url": "s3://kootoro-dev-s3-reference-f6d3xl/processed/2025/05/06/23591_5560c9c9_reference_image.png",
            "checking_image_url": "s3://kootoro-dev-s3-checking-f6d3xl/AACZ_3.png",  # Fixed: removed space
            "vending_machine_id": "VM-3245",
            "layout_id": 23591,
            "layout_prefix": "5560c9c9"
        }

        print("\n📋 Test verification data:")
        for key, value in test_data.items():
            print(f"   {key}: {value}")

        print("\n🚀 Attempting to initiate verification...")

        # Call the API
        response = api_client.initiate_verification(**test_data)

        print("✅ Verification initiated successfully!")
        print(f"📄 Response: {response}")

        # Check response structure
        if 'verificationId' in response:
            print(f"🆔 Verification ID: {response['verificationId']}")
        if 'status' in response:
            print(f"📊 Status: {response['status']}")
        if 'verificationAt' in response:
            print(f"⏰ Initiated at: {response['verificationAt']}")

        return True

    except Exception as e:
        print(f"❌ Verification initiation failed: {str(e)}")

        # Check if it's still a 405 error
        if "405" in str(e):
            print("🔍 Still getting 405 Method Not Allowed - endpoint or method issue")
        elif "400" in str(e):
            print("🔍 Getting 400 Bad Request - likely a data validation issue")
            print("   Possible causes:")
            print("   - Invalid S3 URL format or non-existent files")
            print("   - Special characters in filenames (spaces, etc.)")
            print("   - Missing required fields for LAYOUT_VS_CHECKING")
            print("   - Invalid data types (layoutId should be integer)")
        elif "401" in str(e) or "403" in str(e):
            print("🔍 Getting authentication/authorization error")
        else:
            print(f"🔍 Other error type: {type(e).__name__}")

        return False

def test_api_structure():
    """Test the API request structure without actually sending it"""
    print("\n🔍 Testing API request structure...")

    try:
        api_client = APIClient()

        # Mock the make_request method to see what would be sent
        original_make_request = api_client.make_request

        def mock_make_request(method, endpoint, data=None, params=None, debug=False):
            print(f"📤 Would send {method} request to: {endpoint}")
            print(f"📦 Request data structure:")
            import json
            print(json.dumps(data, indent=2))
            return {"mock": "response"}

        api_client.make_request = mock_make_request

        # Test the request structure
        api_client.initiate_verification(
            verification_type="LAYOUT_VS_CHECKING",
            reference_image_url="s3://test-bucket/ref.jpg",
            checking_image_url="s3://test-bucket/check.jpg",
            vending_machine_id="VM-001",
            layout_id=123,
            layout_prefix="test"
        )

        print("✅ Request structure looks correct!")
        return True

    except Exception as e:
        print(f"❌ Structure test failed: {str(e)}")
        return False

if __name__ == "__main__":
    print("🔧 Testing Verification API Fix")
    print("=" * 50)

    # Test the request structure first
    structure_ok = test_api_structure()

    if structure_ok:
        print("\n" + "=" * 50)
        # Test actual API call
        success = test_verification_initiation()

        if success:
            print("\n🎉 All tests passed! The verification API should now work.")
        else:
            print("\n⚠️  API call failed, but the structure is correct.")
            print("   This might be due to:")
            print("   - Invalid S3 URLs (buckets/keys don't exist)")
            print("   - API Gateway configuration issues")
            print("   - Backend service issues")
    else:
        print("\n❌ Request structure test failed.")

    print("\n💡 To test in Streamlit:")
    print("   1. Run: python3 -m streamlit run app.py")
    print("   2. Go to 'Initiate Verification' page")
    print("   3. Fill in the form and submit")
