#!/usr/bin/env python3
"""
Test script to isolate issues with the verifications page
"""

import sys
import traceback
from clients.api_client import APIClient

def test_api_client():
    """Test API client initialization and basic functionality"""
    print("🔧 Testing API Client...")
    try:
        api_client = APIClient()
        print("✅ API client initialized successfully")
        return api_client
    except Exception as e:
        print(f"❌ API client initialization failed: {str(e)}")
        traceback.print_exc()
        return None

def test_list_verifications(api_client):
    """Test the list_verifications endpoint"""
    print("\n🔍 Testing list_verifications endpoint...")
    try:
        # Test with minimal parameters
        response = api_client.list_verifications({'limit': 3})
        print("✅ list_verifications call successful")
        
        if isinstance(response, dict):
            print(f"📄 Response structure: {list(response.keys())}")
            
            if 'results' in response:
                results = response['results']
                print(f"📊 Found {len(results)} results")
                
                # Check first result structure
                if results:
                    first_result = results[0]
                    print(f"🔍 First result keys: {list(first_result.keys())}")
                    print(f"🆔 First result ID: {first_result.get('verificationId', 'N/A')}")
                    print(f"📅 First result date: {first_result.get('verificationAt', 'N/A')}")
                    print(f"✅ First result status: {first_result.get('verificationStatus', 'N/A')}")
            
            if 'pagination' in response:
                pagination = response['pagination']
                print(f"📄 Pagination: {pagination}")
        
        return True
        
    except Exception as e:
        print(f"❌ list_verifications failed: {str(e)}")
        traceback.print_exc()
        return False

def test_verifications_module():
    """Test importing the verifications module"""
    print("\n📦 Testing verifications module import...")
    try:
        from pages import verifications
        print("✅ verifications module imported successfully")
        
        # Check if the app function exists
        if hasattr(verifications, 'app'):
            print("✅ app function found in verifications module")
        else:
            print("❌ app function not found in verifications module")
            
        # Check if apply_custom_css function exists
        if hasattr(verifications, 'apply_custom_css'):
            print("✅ apply_custom_css function found")
        else:
            print("❌ apply_custom_css function not found")
            
        # Check if render_verification_card function exists
        if hasattr(verifications, 'render_verification_card'):
            print("✅ render_verification_card function found")
        else:
            print("❌ render_verification_card function not found")
            
        return True
        
    except Exception as e:
        print(f"❌ verifications module import failed: {str(e)}")
        traceback.print_exc()
        return False

def main():
    """Main test function"""
    print("🧪 Starting Verification Page Tests")
    print("=" * 50)
    
    # Test 1: API Client
    api_client = test_api_client()
    if not api_client:
        print("\n❌ Cannot proceed without working API client")
        return False
    
    # Test 2: List Verifications API
    api_success = test_list_verifications(api_client)
    if not api_success:
        print("\n❌ API endpoint test failed")
        return False
    
    # Test 3: Module Import
    module_success = test_verifications_module()
    if not module_success:
        print("\n❌ Module import test failed")
        return False
    
    print("\n" + "=" * 50)
    print("🎉 All tests passed! The verifications page should work correctly.")
    print("If you're still seeing errors in the browser, it might be:")
    print("1. A browser caching issue - try hard refresh (Ctrl+F5)")
    print("2. A session state issue - try clearing browser data")
    print("3. A specific UI interaction issue")
    
    return True

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
