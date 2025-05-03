import streamlit as st
import logging

logger = logging.getLogger(__name__)

def app(api_client):
    st.title("Health Check")
    if st.button("Check System Health"):
        try:
            health_status = api_client.health_check()
            st.write("System Health:")
            st.json(health_status)
        except Exception as e:
            logger.error(f"Health check failed: {str(e)}")
            st.error(str(e))