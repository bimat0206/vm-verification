import streamlit as st
import logging

logger = logging.getLogger(__name__)

def app(api_client):
    st.title("Vending Machine Verification System")
    st.write("Welcome to the verification system. Use the sidebar to navigate.")
    try:
        health_status = api_client.health_check()
        st.subheader("System Status")
        st.write(health_status)
    except Exception as e:
        logger.error(f"Health check failed: {str(e)}")
        st.warning("Unable to retrieve system status.")