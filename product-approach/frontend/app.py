import streamlit as st
import requests
import json
import boto3
import os
import logging
import time
from datetime import datetime

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger("streamlit-app")

# Function to load config from AWS Secrets Manager with better error handling
def load_config():
    logger.info("Starting to load configuration")
    
    # First try environment variables as a fallback
    api_endpoint = os.environ.get("API_ENDPOINT", "")
    dynamodb_table = os.environ.get("DYNAMODB_TABLE", "")
    s3_bucket = os.environ.get("S3_BUCKET", "")
    
    # Log the startup environment
    logger.info(f"Environment variables: API_ENDPOINT={api_endpoint}, DYNAMODB_TABLE={dynamodb_table}, S3_BUCKET={s3_bucket}")
    logger.info(f"SECRET_ARN environment variable present: {'SECRET_ARN' in os.environ}")
    
    # If SECRET_ARN is defined, try to load from Secrets Manager
    if "SECRET_ARN" in os.environ:
        secret_arn = os.environ.get("SECRET_ARN")
        region = os.environ.get("REGION", "us-east-1")
        
        logger.info(f"Attempting to load config from Secrets Manager: {secret_arn} in {region}")
        
        try:
            # Create a Secrets Manager client
            session = boto3.session.Session()
            client = session.client(service_name='secretsmanager', region_name=region)
            
            # Try multiple times with backoff
            max_retries = 3
            for attempt in range(max_retries):
                try:
                    # Get the secret
                    logger.info(f"Attempt {attempt+1} to get secret")
                    response = client.get_secret_value(SecretId=secret_arn)
                    
                    if 'SecretString' in response:
                        logger.info("Successfully retrieved secret")
                        config = json.loads(response['SecretString'])
                        
                        # Override environment variables if secrets are available
                        if config.get("api_endpoint"):
                            api_endpoint = config.get("api_endpoint")
                        if config.get("dynamodb_table_name"):
                            dynamodb_table = config.get("dynamodb_table_name")
                        if config.get("s3_bucket_name"):
                            s3_bucket = config.get("s3_bucket_name")
                        
                        logger.info("Configuration loaded from Secrets Manager")
                        break
                    else:
                        logger.warning("Secret has no SecretString field")
                except Exception as e:
                    logger.warning(f"Attempt {attempt+1} failed: {str(e)}")
                    if attempt < max_retries - 1:  # Don't sleep on the last attempt
                        time.sleep(2 ** attempt)  # Exponential backoff
        except Exception as e:
            logger.error(f"Error retrieving configuration from Secrets Manager: {str(e)}")
    
    # Fallback configuration
    if not api_endpoint:
        api_endpoint = "https://am0ncga8rk.execute-api.us-east-1.amazonaws.com/v1"
        logger.info(f"Using fallback API endpoint: {api_endpoint}")
    
    if not dynamodb_table:
        dynamodb_table = "VerificationResults"
        logger.info(f"Using fallback DynamoDB table: {dynamodb_table}")
    
    if not s3_bucket:
        s3_bucket = "vending-verification-images"
        logger.info(f"Using fallback S3 bucket: {s3_bucket}")
    
    return {
        "api_endpoint": api_endpoint,
        "dynamodb_table_name": dynamodb_table,
        "s3_bucket_name": s3_bucket
    }

try:
    # Load configuration
    logger.info("Application startup: loading configuration")
    config = load_config()
    API_ENDPOINT = config.get("api_endpoint", "")
    logger.info(f"Using API endpoint: {API_ENDPOINT}")

    # Set page config
    st.set_page_config(
        page_title="Vending Machine Verification",
        page_icon="🏪",
        layout="wide",
        initial_sidebar_state="expanded"
    )

    # Header and description
    st.title("Vending Machine Verification System")
    st.markdown("""
    This application helps you verify the stock and state of vending machines using image comparisons.
    Upload images, request comparisons, and view results.
    """)

    # Sidebar for navigation
    st.sidebar.title("Navigation")
    page = st.sidebar.radio("Go to", ["Dashboard", "New Comparison", "View Results", "Images", "System Info"])

    # Dashboard page
    if page == "Dashboard":
        st.header("Dashboard")
        
        col1, col2 = st.columns(2)
        
        with col1:
            st.subheader("Recent Comparisons")
            try:
                # This would need a specific API endpoint to list recent comparisons
                # For now, we'll just show a placeholder
                st.info("API endpoint for listing recent comparisons not yet implemented")
                st.markdown("""
                | Comparison ID | Machine ID | Status | Date |
                | ------------- | ---------- | ------ | ---- |
                | comp-001 | vm-123 | Completed | 2023-05-12 |
                | comp-002 | vm-456 | In Progress | 2023-05-11 |
                | comp-003 | vm-123 | Failed | 2023-05-10 |
                """)
            except Exception as e:
                st.error(f"Error fetching recent comparisons: {str(e)}")
        
        with col2:
            st.subheader("Machines Overview")
            try:
                # This would need a specific API endpoint to get machines overview
                # For now, we'll just show a placeholder
                st.info("API endpoint for machines overview not yet implemented")
                st.markdown("""
                | Machine ID | Location | Last Verified | Status |
                | ---------- | -------- | ------------- | ------ |
                | vm-123 | Building A | 2023-05-12 | OK |
                | vm-456 | Building B | 2023-05-11 | Pending |
                | vm-789 | Building C | 2023-05-09 | Issue Detected |
                """)
            except Exception as e:
                st.error(f"Error fetching machines overview: {str(e)}")

    # New Comparison page
    elif page == "New Comparison":
        st.header("Create New Comparison")
        
        # Function to start a new comparison
        def create_comparison(reference_img, checking_img, machine_id, location=None):
            data = {
                "referenceImageKey": reference_img,
                "checkingImageKey": checking_img,
                "vendingMachineId": machine_id
            }
            if location:
                data["location"] = location
            
            try:
                response = requests.post(f"{API_ENDPOINT}/comparisons", json=data)
                if response.status_code == 200:
                    return response.json()
                else:
                    st.error(f"API Error: {response.status_code} - {response.text}")
                    return None
            except Exception as e:
                st.error(f"Error creating comparison: {str(e)}")
                return None
        
        # Form for creating a new comparison
        with st.form("new_comparison_form"):
            col1, col2 = st.columns(2)
            
            with col1:
                reference_img = st.text_input("Reference Image Key", 
                                            help="S3 key for the reference image (e.g., 'machine123/reference.jpg')")
                machine_id = st.text_input("Vending Machine ID", 
                                        help="Unique identifier for the vending machine")
            
            with col2:
                checking_img = st.text_input("Checking Image Key", 
                                            help="S3 key for the image to be checked (e.g., 'machine123/check.jpg')")
                location = st.text_input("Location (Optional)", 
                                        help="Physical location of the vending machine")
            
            submitted = st.form_submit_button("Start Comparison")
            
            if submitted:
                if not reference_img or not checking_img or not machine_id:
                    st.warning("Please fill all required fields")
                else:
                    with st.spinner("Starting comparison..."):
                        result = create_comparison(reference_img, checking_img, machine_id, location)
                        if result:
                            st.success("Comparison started successfully!")
                            st.json(result)
                            
                            # Store comparison ID in session state for easy access
                            if "executionArn" in result:
                                comparison_id = result["executionArn"].split(":")[-1]
                                st.session_state["last_comparison_id"] = comparison_id
                                
                                # Add button to view results
                                if st.button("View Results"):
                                    st.session_state["page"] = "View Results"
                                    st.session_state["comparison_id"] = comparison_id
                                    st.experimental_rerun()

    # View Results page
    elif page == "View Results":
        st.header("View Verification Results")
        
        # Function to get comparison results
        def get_comparison(comparison_id):
            try:
                response = requests.get(f"{API_ENDPOINT}/comparisons/{comparison_id}")
                if response.status_code == 200:
                    return response.json()
                else:
                    st.error(f"API Error: {response.status_code} - {response.text}")
                    return None
            except Exception as e:
                st.error(f"Error fetching comparison: {str(e)}")
                return None
        
        # Get comparison ID from URL or input
        comparison_id = st.text_input("Enter Comparison ID",
                                    value=st.session_state.get("comparison_id", ""),
                                    help="ID of the comparison to view")
        
        if st.button("Fetch Results") and comparison_id:
            with st.spinner("Fetching results..."):
                results = get_comparison(comparison_id)
                if results:
                    st.success("Results retrieved successfully!")
                    
                    # Display comparison details
                    st.subheader("Comparison Details")
                    if isinstance(results, dict):
                        cols = st.columns(3)
                        cols[0].metric("Machine ID", results.get("vendingMachineId", "N/A"))
                        cols[1].metric("Status", results.get("status", "N/A"))
                        cols[2].metric("Timestamp", results.get("timestamp", "N/A"))
                        
                        # Display comparison results
                        st.subheader("Verification Results")
                        
                        # If there are detailed results
                        if "verificationResults" in results:
                            verification = results["verificationResults"]
                            st.json(verification)
                        else:
                            st.write("No detailed verification results available")
                        
                        # Full JSON response (expandable)
                        with st.expander("View Raw JSON Response"):
                            st.json(results)
                    else:
                        st.write(results)
                else:
                    st.warning("No results found for this comparison ID")

    # Images page
    elif page == "Images":
        st.header("Available Images")
        
        # Function to list images
        def get_images(machine_id=None):
            params = {}
            if machine_id:
                params["machineId"] = machine_id
            
            try:
                response = requests.get(f"{API_ENDPOINT}/images", params=params)
                if response.status_code == 200:
                    return response.json()
                else:
                    st.error(f"API Error: {response.status_code} - {response.text}")
                    return []
            except Exception as e:
                st.error(f"Error fetching images: {str(e)}")
                return []
        
        col1, col2 = st.columns([3, 1])
        
        with col1:
            machine_id = st.text_input("Filter by Machine ID (optional)")
        
        with col2:
            st.write("")
            st.write("")
            fetch_button = st.button("List Images")
        
        if fetch_button:
            with st.spinner("Fetching images..."):
                images = get_images(machine_id)
                if images:
                    st.success(f"Found {len(images) if isinstance(images, list) else 'some'} images!")
                    
                    # Display images in a table or grid
                    if isinstance(images, list):
                        # Create a dataframe if it's a list of images
                        import pandas as pd
                        if len(images) > 0 and isinstance(images[0], dict):
                            df = pd.DataFrame(images)
                            st.dataframe(df)
                        else:
                            st.write(images)
                    else:
                        # Just display the JSON if it's not a list
                        st.json(images)
                else:
                    st.info("No images found matching the criteria")
                    
    # System Info page (new)
    elif page == "System Info":
        st.header("System Information")
        
        # Display environment info
        st.subheader("Environment")
        env_info = {
            "API Endpoint": API_ENDPOINT,
            "DynamoDB Table": config.get("dynamodb_table_name", "Not configured"),
            "S3 Bucket": config.get("s3_bucket_name", "Not configured"),
            "Region": os.environ.get("REGION", "Not set"),
            "Secret ARN": os.environ.get("SECRET_ARN", "Not set"),
            "Deployment Type": "ECS Fargate",
            "Instance ID": os.environ.get("ECS_CONTAINER_METADATA_URI_V4", "Not available")
        }
        
        for key, value in env_info.items():
            st.text(f"{key}: {value}")
        
        # Test API connection
        st.subheader("API Connection Test")
        if st.button("Test API Connection"):
            try:
                response = requests.get(f"{API_ENDPOINT}/health", timeout=5)
                if response.status_code == 200:
                    st.success(f"API connection successful! Status code: {response.status_code}")
                    st.json(response.json() if response.content else {"status": "OK"})
                else:
                    st.error(f"API connection failed with status code: {response.status_code}")
                    st.text(response.text if response.content else "No response body")
            except requests.RequestException as e:
                st.error(f"API connection error: {str(e)}")
                
        # Test Secrets Manager
        if "SECRET_ARN" in os.environ:
            st.subheader("Secrets Manager Test")
            if st.button("Test Secrets Manager Connection"):
                try:
                    session = boto3.session.Session()
                    client = session.client(
                        service_name='secretsmanager', 
                        region_name=os.environ.get("REGION", "us-east-1")
                    )
                    response = client.describe_secret(SecretId=os.environ.get("SECRET_ARN"))
                    st.success("Secrets Manager connection successful!")
                    
                    # Display non-sensitive metadata
                    st.json({
                        "Name": response.get("Name"),
                        "LastChangedDate": str(response.get("LastChangedDate")),
                        "LastAccessedDate": str(response.get("LastAccessedDate")),
                        "ARN": response.get("ARN")
                    })
                except Exception as e:
                    st.error(f"Secrets Manager connection error: {str(e)}")
        
        # ECS Container metadata
        st.subheader("Container Metadata")
        if "ECS_CONTAINER_METADATA_URI_V4" in os.environ:
            if st.button("Fetch Container Metadata"):
                try:
                    metadata_url = os.environ.get("ECS_CONTAINER_METADATA_URI_V4")
                    response = requests.get(metadata_url, timeout=2)
                    if response.status_code == 200:
                        st.success("Container metadata retrieved successfully")
                        st.json(response.json())
                    else:
                        st.error(f"Failed to retrieve container metadata: {response.status_code}")
                except Exception as e:
                    st.error(f"Error fetching container metadata: {str(e)}")
        else:
            st.info("ECS container metadata URI not available")

    # Add footer
    st.markdown("---")
    st.markdown(f"Vending Machine Verification System | {datetime.now().year}")

except Exception as e:
    logger.error(f"Unhandled exception in main app: {str(e)}", exc_info=True)
    st.error(f"An unexpected error occurred: {str(e)}")
    st.info("Please check the system logs for more information.")
    
    # Display traceback in development environments
    import traceback
    st.expander("Error Details").code(traceback.format_exc())