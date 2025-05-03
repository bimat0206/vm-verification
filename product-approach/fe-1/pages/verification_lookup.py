import streamlit as st
import logging

logger = logging.getLogger(__name__)

def app(api_client):
    st.title("Verification Lookup")
    st.write("Look up verification history by checking image URL")
    
    with st.form(key='lookup_form'):
        checking_image_url = st.text_input(
            "Checking Image URL", 
            help="S3 URL of the checking image to look up"
        )
        
        vending_machine_id = st.text_input(
            "Vending Machine ID (optional)",
            help="Filter results by vending machine ID"
        )
        
        limit = st.slider(
            "Max Results", 
            min_value=1, 
            max_value=10, 
            value=1,
            help="Maximum number of verification results to return"
        )
        
        submit_button = st.form_submit_button(label='Lookup')
    
    if submit_button:
        if not checking_image_url:
            st.error("Checking Image URL is required")
            return
        
        try:
            results = api_client.lookup_verification(
                checking_image_url=checking_image_url,
                vending_machine_id=vending_machine_id if vending_machine_id else None,
                limit=limit
            )
            
            if not results.get('results', []):
                st.info("No verification results found for this image.")
                return
                
            st.success(f"Found {len(results.get('results', []))} verification(s)")
            
            for idx, verification in enumerate(results.get('results', [])):
                with st.container():
                    col1, col2, col3 = st.columns([3, 2, 1])
                    
                    with col1:
                        verification_id = verification.get('verificationId', 'N/A')
                        st.subheader(f"ID: {verification_id}")
                        st.write(f"Machine: {verification.get('vendingMachineId', 'N/A')}")
                        st.write(f"Date: {verification.get('verificationAt', 'N/A')}")
                    
                    with col2:
                        status = verification.get('verificationStatus', 'UNKNOWN')
                        if status == 'CORRECT':
                            st.success(f"Status: {status}")
                        else:
                            st.error(f"Status: {status}")
                            
                        if 'verificationSummary' in verification:
                            summary = verification['verificationSummary']
                            accuracy = summary.get('overallAccuracy', 0)
                            st.metric("Accuracy", f"{accuracy}%")
                    
                    with col3:
                        if st.button("View Details", key=f"view_{idx}"):
                            st.session_state['selected_verification'] = verification_id
                            st.session_state['page'] = 'Verification Details'
                            st.rerun()
                
                st.divider()
                
        except Exception as e:
            logger.error(f"Verification lookup failed: {str(e)}")
            st.error(str(e))