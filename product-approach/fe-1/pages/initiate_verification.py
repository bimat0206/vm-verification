import streamlit as st
import logging
import time

logger = logging.getLogger(__name__)

def app(api_client):
    st.title("Initiate Verification")
    
    # Step 1: Select verification type
    verification_type = st.selectbox(
        "Verification Type", 
        ["LAYOUT_VS_CHECKING", "PREVIOUS_VS_CURRENT"],
        format_func=lambda x: "Layout vs Checking" if x == "LAYOUT_VS_CHECKING" else "Previous vs Current"
    )
    
    # Step 2: Select images and provide metadata
    with st.form(key='init_verification_form'):
        vending_machine_id = st.text_input("Vending Machine ID", help="Required - ID of the vending machine")
        
        col1, col2 = st.columns(2)
        
        with col1:
            st.subheader("Reference Image")
            # For simplicity, we'll use URL inputs instead of file upload
            reference_image_url = st.text_input(
                "Reference Image URL", 
                help="S3 URL to the reference image (s3://bucket-name/path/to/image)"
            )
            
            # Type-specific fields
            if verification_type == "LAYOUT_VS_CHECKING":
                layout_id = st.number_input("Layout ID", min_value=1, help="Required for layout verification")
                layout_prefix = st.text_input("Layout Prefix", help="Required for layout verification")
            else:  # PREVIOUS_VS_CURRENT
                previous_verification_id = st.text_input("Previous Verification ID", help="Optional - ID of previous verification")
        
        with col2:
            st.subheader("Checking Image")
            checking_image_url = st.text_input(
                "Checking Image URL", 
                help="S3 URL to the checking image (s3://bucket-name/path/to/image)"
            )
        
        submit_button = st.form_submit_button(label='Start Verification')
    
    if submit_button:
        required_fields = [vending_machine_id, reference_image_url, checking_image_url]
        if verification_type == "LAYOUT_VS_CHECKING":
            if not layout_id or not layout_prefix:
                st.error("Layout ID and Layout Prefix are required for Layout vs Checking verification.")
                return
        
        if not all(required_fields):
            st.error("All fields marked as required must be filled.")
            return
        
        try:
            # Different parameters based on verification type
            if verification_type == "LAYOUT_VS_CHECKING":
                response = api_client.initiate_verification(
                    verification_type=verification_type,
                    reference_image_url=reference_image_url,
                    checking_image_url=checking_image_url,
                    vending_machine_id=vending_machine_id,
                    layout_id=layout_id,
                    layout_prefix=layout_prefix
                )
            else:  # PREVIOUS_VS_CURRENT
                response = api_client.initiate_verification(
                    verification_type=verification_type,
                    reference_image_url=reference_image_url,
                    checking_image_url=checking_image_url,
                    vending_machine_id=vending_machine_id,
                    previous_verification_id=previous_verification_id if 'previous_verification_id' in locals() else None
                )
                
            st.success("Verification initiated successfully!")
            verification_id = response.get('verificationId', 'N/A')
            st.write(f"Verification ID: {verification_id}")
            st.write(f"Status: {response.get('status', 'N/A')}")
            st.write(f"Initiated at: {response.get('verificationAt', 'N/A')}")
            
            # Store the verification ID in session state and redirect to details page
            if verification_id != 'N/A':
                st.session_state['selected_verification'] = verification_id
                st.info("Redirecting to verification details page in 5 seconds...")
                time.sleep(5)
                st.session_state['page'] = 'Verification Details'
                st.rerun()
                
        except Exception as e:
            logger.error(f"Initiate verification failed: {str(e)}")
            st.error(str(e))