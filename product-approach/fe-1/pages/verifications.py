import streamlit as st
import logging

logger = logging.getLogger(__name__)

def app(api_client):
    st.title("Verifications")
    status_filter = st.selectbox("Filter by Status", ["all", "pending", "completed", "failed"])
    params = {"status": status_filter} if status_filter != "all" else {}
    try:
        verifications = api_client.list_verifications(params)
        if not verifications:
            st.write("No verifications found.")
            return
        for v in verifications:
            col1, col2 = st.columns([3, 1])
            with col1:
                st.write(f"ID: {v.get('verification_id', 'N/A')} | Status: {v.get('status', 'N/A')}")
            with col2:
                if st.button("Details", key=v.get('verification_id')):
                    st.session_state['selected_verification'] = v.get('verification_id')
                    st.rerun()
    except Exception as e:
        logger.error(f"List verifications failed: {str(e)}")
        st.error(str(e))