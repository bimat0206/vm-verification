import streamlit as st
import logging
import time

logger = logging.getLogger(__name__)

def app(api_client):
    st.title("Initiate Verification")
    with st.form(key='init_verification_form'):
        vending_machine_id = st.text_input("Vending Machine ID")
        reference_image = st.file_uploader("Reference Image", type=['png', 'jpg'])
        checking_image = st.file_uploader("Checking Image", type=['png', 'jpg'])
        submit_button = st.form_submit_button(label='Start Verification')

    if submit_button:
        if not all([vending_machine_id, reference_image, checking_image]):
            st.error("All fields are required.")
            return
        data = {
            "vending_machine_id": vending_machine_id,
            # Note: For simplicity, assuming API handles image upload via pre-signed URLs in real implementation
            "reference_image": reference_image.name,
            "checking_image": checking_image.name
        }
        try:
            response = api_client.initiate_verification(data)
            st.success("Verification initiated successfully!")
            verification_id = response.get('verification_id', 'N/A')
            st.write(f"Verification ID: {verification_id}")
            # Simple polling for status
            with st.spinner("Checking status..."):
                for _ in range(5):
                    time.sleep(2)
                    details = api_client.get_verification_details(verification_id)
                    status = details.get('status', 'pending')
                    st.write(f"Current Status: {status}")
                    if status in ['completed', 'failed']:
                        break
        except Exception as e:
            logger.error(f"Initiate verification failed: {str(e)}")
            st.error(str(e))