import streamlit as st
import logging

logger = logging.getLogger(__name__)

def app(api_client):
    st.title("Verification Details")
    verification_id = st.session_state.get('selected_verification', '')
    if not verification_id:
        st.write("Please select a verification from the Verifications page.")
        return
    try:
        details = api_client.get_verification_details(verification_id)
        st.subheader(f"Details for Verification ID: {verification_id}")
        st.json(details)
        conversation = api_client.get_verification_conversation(verification_id)
        st.subheader("Conversation History")
        st.json(conversation)
    except Exception as e:
        logger.error(f"Get verification details failed: {str(e)}")
        st.error(str(e))