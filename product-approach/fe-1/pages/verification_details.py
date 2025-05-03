import streamlit as st
import logging
import json

logger = logging.getLogger(__name__)

def app(api_client):
    st.title("Verification Details")
    verification_id = st.session_state.get('selected_verification', '')
    
    if not verification_id:
        st.warning("Please select a verification from the Verifications page.")
        if st.button("Go to Verifications"):
            st.session_state['page'] = 'Verifications'
            st.rerun()
        return
    
    # Add refresh button
    col1, col2 = st.columns([6, 1])
    with col2:
        if st.button("🔄 Refresh"):
            st.rerun()
    
    try:
        # Get verification details
        details = api_client.get_verification_details(verification_id)
        status = details.get('status', 'UNKNOWN')
        
        # Show status with appropriate color
        if status == "COMPLETED":
            st.success(f"Status: {status}")
        elif status == "PROCESSING":
            st.info(f"Status: {status}")
            st.warning("Verification is still processing. Please check back later.")
            # Add auto-refresh option for processing verifications
            if st.checkbox("Auto-refresh every 10 seconds", value=True):
                st.rerun()
        else:
            st.error(f"Status: {status}")
        
        # Basic information
        st.subheader("Basic Information")
        col1, col2 = st.columns(2)
        
        with col1:
            st.write(f"**ID:** {verification_id}")
            st.write(f"**Date:** {details.get('verificationAt', 'N/A')}")
            if 'result' in details and 'vendingMachineId' in details['result']:
                st.write(f"**Machine ID:** {details['result']['vendingMachineId']}")
            if 'result' in details and 'location' in details['result']:
                st.write(f"**Location:** {details['result']['location']}")
        
        with col2:
            if 'result' in details and 'verificationType' in details['result']:
                verification_type = details['result']['verificationType']
                st.write(f"**Type:** {verification_type}")
                
                if verification_type == "LAYOUT_VS_CHECKING":
                    if 'layoutId' in details['result']:
                        st.write(f"**Layout ID:** {details['result']['layoutId']}")
                    if 'layoutPrefix' in details['result']:
                        st.write(f"**Layout Prefix:** {details['result']['layoutPrefix']}")
                elif verification_type == "PREVIOUS_VS_CURRENT":
                    if 'previousVerificationId' in details['result']:
                        prev_id = details['result']['previousVerificationId']
                        st.write(f"**Previous Verification:** {prev_id}")
                        if st.button("View Previous Verification"):
                            st.session_state['selected_verification'] = prev_id
                            st.rerun()
        
        # If verification is completed, show results
        if status == "COMPLETED" and 'result' in details:
            result = details['result']
            
            # Summary stats
            if 'verificationSummary' in result:
                summary = result['verificationSummary']
                st.subheader("Verification Summary")
                
                col1, col2, col3 = st.columns(3)
                with col1:
                    accuracy = summary.get('overallAccuracy', 0)
                    st.metric("Accuracy", f"{accuracy}%")
                
                with col2:
                    total = summary.get('totalPositionsChecked', 0)
                    correct = summary.get('correctPositions', 0)
                    st.metric("Correct Positions", f"{correct}/{total}")
                
                with col3:
                    discrepant = summary.get('discrepantPositions', 0)
                    st.metric("Discrepancies", discrepant)
                
                # Show outcome
                if 'verificationOutcome' in summary:
                    st.info(summary['verificationOutcome'])
            
            # Show images
            st.subheader("Images")
            col1, col2 = st.columns(2)
            
            with col1:
                st.write("**Reference Image**")
                if 'referenceImageUrl' in result:
                    ref_url = result['referenceImageUrl']
                    st.write(f"URL: {ref_url}")
                    try:
                        # Attempt to get presigned URL for image display
                        if '/' in ref_url:
                            key = ref_url.split('/')[-1]
                            url_response = api_client.get_image_url(key)
                            image_url = url_response.get('presignedUrl', '')
                            if image_url:
                                st.image(image_url, width=300)
                    except Exception as e:
                        st.warning("Could not load reference image preview")
            
            with col2:
                st.write("**Checking Image**")
                if 'checkingImageUrl' in result:
                    check_url = result['checkingImageUrl']
                    st.write(f"URL: {check_url}")
                    try:
                        # Attempt to get presigned URL for image display
                        if '/' in check_url:
                            key = check_url.split('/')[-1]
                            url_response = api_client.get_image_url(key)
                            image_url = url_response.get('presignedUrl', '')
                            if image_url:
                                st.image(image_url, width=300)
                    except Exception as e:
                        st.warning("Could not load checking image preview")
            
            # Result image if available
            if 'resultImageUrl' in details:
                st.write("**Result Visualization**")
                result_url = details['resultImageUrl']
                st.write(f"URL: {result_url}")
                try:
                    if '/' in result_url:
                        key = result_url.split('/')[-1]
                        url_response = api_client.get_image_url(key)
                        image_url = url_response.get('presignedUrl', '')
                        if image_url:
                            st.image(image_url, width=600)
                except Exception as e:
                    st.warning("Could not load result image preview")
            
            # Show discrepancies
            if 'discrepancies' in result:
                discrepancies = result['discrepancies']
                st.subheader(f"Discrepancies ({len(discrepancies)})")
                
                if discrepancies:
                    st.dataframe(
                        data=discrepancies,
                        column_config={
                            "position": "Position",
                            "expected": "Expected Product",
                            "found": "Found Product",
                            "issue": "Issue Type",
                            "confidence": "Confidence (%)",
                            "verificationResult": "Result"
                        },
                        use_container_width=True
                    )
                else:
                    st.success("No discrepancies found!")
        
        # Get conversation history
        try:
            conversation = api_client.get_verification_conversation(verification_id)
            
            st.subheader("Conversation History")
            if 'history' in conversation and conversation['history']:
                history = conversation['history']
                
                for turn in history:
                    turn_id = turn.get('turnId', 0)
                    timestamp = turn.get('timestamp', '')
                    analysis_stage = turn.get('analysisStage', '')
                    
                    with st.expander(f"Turn {turn_id} - {analysis_stage} - {timestamp}"):
                        st.write("**Prompt:**")
                        st.code(turn.get('prompt', 'No prompt available'), language='markdown')
                        
                        st.write("**Response:**")
                        st.write(turn.get('response', 'No response available'))
                        
                        if 'tokenUsage' in turn:
                            usage = turn['tokenUsage']
                            st.write(f"Token usage: Input: {usage.get('input', 0)}, " +
                                    f"Output: {usage.get('output', 0)}, " +
                                    f"Total: {usage.get('total', 0)}")
                        
                        if 'latencyMs' in turn:
                            st.write(f"Latency: {turn['latencyMs']} ms")
            else:
                st.info("No conversation history available.")
                
        except Exception as e:
            logger.warning(f"Failed to load conversation history: {str(e)}")
            st.warning("Could not load conversation history.")
        
        # Add raw JSON option for developers
        if st.checkbox("Show raw JSON (for developers)"):
            st.json(details)
            
    except Exception as e:
        logger.error(f"Get verification details failed: {str(e)}")
        st.error(str(e))