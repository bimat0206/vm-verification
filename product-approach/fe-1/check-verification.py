#!/usr/bin/env python3
"""
Script to check the status of a specific verification
"""

import os
import sys
import logging

# Add the current directory to Python path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from clients.api_client import APIClient

def check_verification_status(verification_id):
    """Check the status of a specific verification"""
    print(f"🔍 Checking verification status for ID: {verification_id}")
    
    try:
        # Initialize API client
        api_client = APIClient()
        print("✅ API client initialized successfully")
        
        # Get verification details
        print(f"\n📋 Fetching verification details...")
        response = api_client.get_verification_details(verification_id)
        
        print("✅ Verification details retrieved successfully!")
        print(f"📄 Response: {response}")
        
        # Parse and display key information
        if isinstance(response, dict):
            print(f"\n📊 Verification Summary:")
            print(f"   ID: {response.get('verificationId', 'N/A')}")
            print(f"   Status: {response.get('status', 'N/A')}")
            print(f"   Type: {response.get('verificationType', 'N/A')}")
            print(f"   Machine ID: {response.get('vendingMachineId', 'N/A')}")
            print(f"   Created: {response.get('verificationAt', 'N/A')}")
            print(f"   Updated: {response.get('updatedAt', 'N/A')}")
            
            if 'result' in response:
                result = response['result']
                print(f"   Result: {result.get('outcome', 'N/A')}")
                if 'confidence' in result:
                    print(f"   Confidence: {result['confidence']}")
        
        return True
        
    except Exception as e:
        print(f"❌ Failed to get verification details: {str(e)}")
        
        if "404" in str(e):
            print("🔍 Verification not found - ID may be invalid")
        elif "400" in str(e):
            print("🔍 Bad request - ID format may be invalid")
        elif "401" in str(e) or "403" in str(e):
            print("🔍 Authentication/authorization error")
        else:
            print(f"🔍 Other error type: {type(e).__name__}")
            
        return False

def check_verification_conversation(verification_id):
    """Check the conversation history for a verification"""
    print(f"\n💬 Checking conversation history...")
    
    try:
        api_client = APIClient()
        response = api_client.get_verification_conversation(verification_id)
        
        print("✅ Conversation history retrieved successfully!")
        print(f"📄 Response: {response}")
        
        if isinstance(response, dict) and 'history' in response:
            history = response['history']
            print(f"\n📝 Conversation Summary:")
            print(f"   Turn: {response.get('currentTurn', 'N/A')}/{response.get('maxTurns', 'N/A')}")
            print(f"   Status: {response.get('turnStatus', 'N/A')}")
            print(f"   Messages: {len(history)} entries")
            
            # Show recent messages
            for i, msg in enumerate(history[-3:]):  # Last 3 messages
                print(f"   [{i+1}] {msg.get('role', 'unknown')}: {msg.get('content', 'N/A')[:100]}...")
        
        return True
        
    except Exception as e:
        print(f"❌ Failed to get conversation: {str(e)}")
        return False

if __name__ == "__main__":
    # The verification ID you provided
    verification_id = "a041e458-3171-43e9-a149-f63c5916d3a2"
    
    print("🔍 Verification Status Check")
    print("=" * 50)
    
    # Check verification details
    details_success = check_verification_status(verification_id)
    
    if details_success:
        # Check conversation if details were successful
        check_verification_conversation(verification_id)
        
        print("\n🎉 Verification API is working correctly!")
        print("✅ The 400 Bad Request issue appears to be resolved.")
    else:
        print("\n⚠️  Could not retrieve verification details.")
        print("   This might indicate:")
        print("   - The verification ID is from a different environment")
        print("   - The verification was not successfully created")
        print("   - There are still API connectivity issues")
    
    print(f"\n💡 To check this verification in Streamlit:")
    print(f"   1. Run: python3 -m streamlit run app.py")
    print(f"   2. Go to 'Verification Details' page")
    print(f"   3. Enter verification ID: {verification_id}")
