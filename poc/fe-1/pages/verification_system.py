import streamlit as st
import logging
import time
from datetime import datetime
from .improved_image_selector import render_improved_s3_image_selector

logger = logging.getLogger(__name__)

def apply_verification_system_css():
    """Apply custom CSS for Verification System page LLM response sections"""
    st.markdown("""
    <style>
    .llm-response-card {
        background-color: #262730;
        border: 1px solid #444;
        border-radius: 8px;
        padding: 1.5rem;
        margin-bottom: 1.5rem;
        transition: all 0.2s ease-in-out;
    }
    .llm-response-card:hover {
        border-color: #6366F1;
        box-shadow: 0 0 10px rgba(99, 102, 241, 0.3);
    }
    .llm-card-expanded {
        border-left: 3px solid #6366F1;
    }

    .llm-status-badge {
        display: inline-flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 1rem;
        border-radius: 16px;
        font-weight: 600;
        font-size: 0.85rem;
        text-transform: uppercase;
        border: 1px solid transparent;
    }
    .status-processing {
        background-color: #1E3A8A;
        color: #93C5FD;
        border-color: #3B82F6;
    }
    .status-completed {
        background-color: #1C4A36;
        color: #A6E6A6;
        border-color: #2A6E49;
    }
    .status-failed {
        background-color: #5D1D23;
        color: #F4AAAA;
        border-color: #8B2C35;
    }

    .llm-metric-display {
        background-color: #31333F;
        border: 1px solid #4A4C5A;
        border-radius: 6px;
        padding: 1rem;
        text-align: center;
        margin-bottom: 0.5rem;
    }
    .llm-metric-value {
        font-size: 1.8rem !important;
        font-weight: 700 !important;
        color: #FAFAFA !important;
        margin-bottom: 0.25rem !important;
    }
    .llm-metric-label {
        font-size: 0.9rem !important;
        color: #A0A0A0 !important;
        font-weight: 500 !important;
        text-transform: uppercase;
    }

    .llm-analysis-section {
        margin-top: 0.5rem;
        padding: 0.8rem;
        background-color: #1E1E1E;
        border-radius: 6px;
        border: 1px solid #333;
    }
    .llm-analysis-section h6 {
         color: #E0E0E0;
         font-weight: 600;
         margin-top: 0.2rem;
         margin-bottom: 0.4rem;
         font-size: 0.95em;
    }
    </style>
    """, unsafe_allow_html=True)

def extract_llm_analysis(basic_data, detailed_data=None):
    """Extract LLM analysis from verification data."""
    # Try to find analysis in various possible fields
    analysis_fields = [
        'llmAnalysis', 'analysis', 'llm_analysis',
        'verificationAnalysis', 'aiAnalysis', 'description'
    ]

    # Check detailed data first, then basic data
    for data_source in [detailed_data, basic_data]:
        if not data_source:
            continue

        for field in analysis_fields:
            if field in data_source and data_source[field]:
                return data_source[field]

        # Check nested in result or summary
        for nested_field in ['result', 'verificationSummary']:
            if nested_field in data_source and isinstance(data_source[nested_field], dict):
                for field in analysis_fields:
                    if field in data_source[nested_field] and data_source[nested_field][field]:
                        return data_source[nested_field][field]

    return None

def render_llm_response_section(api_client, verification_id, verification_data):
    """Render LLM response data section with analysis results"""
    if not verification_id or not verification_data:
        return

    # Apply CSS
    apply_verification_system_css()

    st.markdown("---")
    st.markdown("## 🤖 LLM Analysis Results")

    # Main LLM response card
    card_class = "llm-response-card llm-card-expanded"
    st.markdown(f'<div class="{card_class}">', unsafe_allow_html=True)

    # Header with verification info
    col1, col2, col3 = st.columns([2, 2, 1])

    with col1:
        st.markdown(f"""
        <div style="background: transparent;">
            <strong style="font-size: 1.2em; color: #FAFAFA;">Verification Analysis</strong><br>
            <span style="font-size: 1.0em; color: #E0E0E0;">ID: {verification_id}</span><br>
            <small style="color: #A0A0A0; font-size: 0.9em;">📅 {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</small>
        </div>
        """, unsafe_allow_html=True)

    with col2:
        verification_type = verification_data.get('verificationType', 'N/A')
        st.markdown(f"""
        <div style="background: transparent;">
            <strong style="color: #E0E0E0;">Type:</strong> {verification_type}<br>
            <strong style="color: #E0E0E0;">Machine:</strong> {verification_data.get('vendingMachineId', 'N/A')}
        </div>
        """, unsafe_allow_html=True)

    with col3:
        status = verification_data.get('status', 'PROCESSING')
        if status in ['COMPLETED', 'CORRECT']:
            status_class = "status-completed"
            status_text = "✓ Completed"
        elif status in ['FAILED', 'INCORRECT']:
            status_class = "status-failed"
            status_text = "✗ Failed"
        else:
            status_class = "status-processing"
            status_text = "⏳ Processing"
        st.markdown(f'<div class="llm-status-badge {status_class}">{status_text}</div>', unsafe_allow_html=True)

    # LLM Analysis section
    st.markdown('<div class="llm-analysis-section">', unsafe_allow_html=True)

    # Try to get detailed verification data
    detailed_verification = None
    try:
        detailed_verification = api_client.get_verification_details(verification_id)
    except Exception as e:
        logger.warning(f"Could not fetch detailed verification data: {str(e)}")

    # Display LLM analysis
    llm_analysis = extract_llm_analysis(verification_data, detailed_verification)

    analysis_cols = st.columns([2, 1])

    with analysis_cols[0]:
        st.markdown("<h6>🧠 LLM Analysis Summary</h6>", unsafe_allow_html=True)
        if llm_analysis:
            st.info(llm_analysis)
        else:
            if status in ['PROCESSING', 'PENDING']:
                st.info("🔄 Analysis in progress... Please wait for the verification to complete.")
            else:
                st.warning("⚠️ Analysis not yet available or processing in progress.")

    with analysis_cols[1]:
        st.markdown("<h6>📊 Confidence Metrics</h6>", unsafe_allow_html=True)

        # Extract metrics from verification data
        accuracy = verification_data.get('overallAccuracy') or (detailed_verification.get('overallAccuracy') if detailed_verification else None)
        confidence = verification_data.get('confidence') or (detailed_verification.get('confidence') if detailed_verification else None)

        if accuracy is not None:
            st.markdown(f"""
            <div class="llm-metric-display">
                <div class="llm-metric-value">{accuracy}%</div>
                <div class="llm-metric-label">Accuracy</div>
            </div>
            """, unsafe_allow_html=True)

        if confidence is not None:
            st.markdown(f"""
            <div class="llm-metric-display">
                <div class="llm-metric-value">{confidence}%</div>
                <div class="llm-metric-label">Confidence</div>
            </div>
            """, unsafe_allow_html=True)

        if accuracy is None and confidence is None:
            st.markdown("<p><small>Metrics will appear when analysis completes.</small></p>", unsafe_allow_html=True)

    st.markdown('</div>', unsafe_allow_html=True)  # Close analysis section
    st.markdown('</div>', unsafe_allow_html=True)  # Close card

    # Detailed breakdown section (collapsible)
    with st.expander("🔍 Detailed Analysis Breakdown", expanded=False):
        if detailed_verification:
            # Show additional details
            additional_fields = {
                'processingTime': '⏱️ Processing Time',
                'modelVersion': '🤖 Model Version',
                'correctPositions': '✅ Correct Positions',
                'discrepantPositions': '❌ Discrepant Positions'
            }

            details_shown = False
            for field, label in additional_fields.items():
                value = detailed_verification.get(field) or verification_data.get(field)
                if value is not None:
                    if not details_shown:
                        details_shown = True
                    st.write(f"**{label}:** {value}")

            # Show raw data if available
            raw_result = detailed_verification.get('result') or verification_data.get('result')
            raw_summary = detailed_verification.get('verificationSummary') or verification_data.get('verificationSummary')

            if raw_result or raw_summary:
                st.markdown("**📋 Raw Response Data:**")
                if raw_result:
                    st.json(raw_result)
                if raw_summary:
                    st.json(raw_summary)

            if not details_shown and not raw_result and not raw_summary:
                st.info("Detailed breakdown will be available once the verification completes.")
        else:
            st.info("Detailed analysis data will be available once the verification completes.")

def render_s3_image_selector(api_client, bucket_type, label, session_key):
    """Render S3 image selector with browse functionality"""
    st.subheader(label)

    # Initialize session state for this selector
    if f"{session_key}_path" not in st.session_state:
        st.session_state[f"{session_key}_path"] = ""
    if f"{session_key}_selected_image" not in st.session_state:
        st.session_state[f"{session_key}_selected_image"] = None
    if f"{session_key}_selected_url" not in st.session_state:
        st.session_state[f"{session_key}_selected_url"] = ""

    # Path input and browse button
    col1, col2 = st.columns([3, 1])
    with col1:
        current_path = st.text_input(
            f"Path in {bucket_type} bucket:",
            value=st.session_state[f"{session_key}_path"],
            key=f"{session_key}_path_input",
            placeholder="Leave empty for root, or enter folder path"
        )
        st.session_state[f"{session_key}_path"] = current_path

    with col2:
        browse_button = st.button(f"Browse {bucket_type.title()}", key=f"{session_key}_browse")

    # Load items if browse button clicked or no items cached
    should_load = browse_button or f"{session_key}_items" not in st.session_state
    if should_load:
        try:
            with st.spinner("Loading images..."):
                browser_response = api_client.browse_images(current_path, bucket_type)
                st.session_state[f"{session_key}_items"] = browser_response.get('items', [])
                st.session_state[f"{session_key}_current_path"] = browser_response.get('currentPath', '')
                st.session_state[f"{session_key}_parent_path"] = browser_response.get('parentPath', '')
        except Exception as e:
            st.error(f"Failed to browse {bucket_type} bucket: {str(e)}")
            return None

    # Display items if available
    if f"{session_key}_items" in st.session_state:
        items = st.session_state[f"{session_key}_items"]
        current_path_display = st.session_state.get(f"{session_key}_current_path", "")
        parent_path = st.session_state.get(f"{session_key}_parent_path", "")

        if current_path_display:
            st.info(f"📁 Current path: `{current_path_display}`")

        # Parent directory navigation
        if parent_path is not None and current_path_display:
            if st.button("⬆️ Go to Parent Directory", key=f"{session_key}_parent"):
                st.session_state[f"{session_key}_path"] = parent_path
                st.rerun()

        if items:
            # Separate folders and files
            folders = [item for item in items if item.get('type') == 'folder']
            files = [item for item in items if item.get('type') == 'file']

            # Display folders first
            if folders:
                st.write("📁 **Folders:**")
                for folder in folders:
                    folder_name = folder.get('name', '')
                    folder_path = folder.get('path', '')
                    col1, col2 = st.columns([3, 1])
                    with col1:
                        st.write(f"📁 {folder_name}")
                    with col2:
                        if st.button("Open", key=f"{session_key}_folder_{folder_name}"):
                            st.session_state[f"{session_key}_path"] = folder_path
                            st.rerun()

            # Display files
            if files:
                st.write("🖼️ **Images:**")
                for idx, file in enumerate(files):
                    file_name = file.get('name', '')
                    file_path = file.get('path', '')

                    col1, col2, col3 = st.columns([2, 1, 1])
                    with col1:
                        st.write(f"🖼️ {file_name}")
                    with col2:
                        # Preview button
                        if st.button("👁️ Preview", key=f"{session_key}_preview_{idx}"):
                            try:
                                url_response = api_client.get_image_url(file_path, bucket_type)
                                if 'url' in url_response:
                                    st.image(url_response['url'], caption=file_name, width=200)
                            except Exception as e:
                                st.error(f"Failed to load preview: {str(e)}")
                    with col3:
                        # Select button
                        if st.button("✅ Select", key=f"{session_key}_select_{idx}"):
                            try:
                                # Get bucket name from config
                                if bucket_type == "reference":
                                    bucket_name = api_client.config_loader.get('REFERENCE_BUCKET', '')
                                elif bucket_type == "checking":
                                    bucket_name = api_client.config_loader.get('CHECKING_BUCKET', '')
                                else:
                                    bucket_name = f"{bucket_type}-bucket"

                                st.session_state[f"{session_key}_selected_image"] = file_name
                                st.session_state[f"{session_key}_selected_url"] = f"s3://{bucket_name}/{file_path}"
                                st.success(f"Selected: {file_name}")
                            except Exception as e:
                                st.error(f"Error selecting image: {str(e)}")
        else:
            st.info("No items found in this location.")

    # Display selected image
    if st.session_state[f"{session_key}_selected_image"]:
        st.success(f"✅ Selected: {st.session_state[f'{session_key}_selected_image']}")
        st.code(st.session_state[f'{session_key}_selected_url'])

    return st.session_state[f"{session_key}_selected_url"]

def app(api_client):
    # Apply CSS styling
    apply_verification_system_css()

    st.title("🔍 Vending Machine Verification System")
    st.write("Start a new verification by selecting your verification type and choosing reference and checking images.")

    # Initialize session state for LLM response tracking
    if 'current_verification_id' not in st.session_state:
        st.session_state['current_verification_id'] = None
    if 'current_verification_data' not in st.session_state:
        st.session_state['current_verification_data'] = None
    if 'show_llm_response' not in st.session_state:
        st.session_state['show_llm_response'] = False

    # Debug section (collapsible)
    with st.expander("🔧 Debug Tools"):
        st.write("**Test Image URL Generation**")
        col1, col2 = st.columns(2)
        with col1:
            test_path = st.text_input("Test image path:", placeholder="e.g., AACZ_3.png")
        with col2:
            test_bucket = st.selectbox("Bucket type:", ["reference", "checking"])

        if st.button("Test Image URL") and test_path:
            try:
                url_response = api_client.get_image_url(test_path, test_bucket)
                st.success("✅ API call successful!")
                st.json(url_response)
            except Exception as e:
                st.error(f"❌ API call failed: {str(e)}")

    # Step 1: Select verification type
    st.subheader("Step 1: Choose Verification Type")
    verification_type = st.selectbox(
        "Verification Type",
        ["LAYOUT_VS_CHECKING", "PREVIOUS_VS_CURRENT"],
        format_func=lambda x: "Layout vs Checking" if x == "LAYOUT_VS_CHECKING" else "Previous vs Current",
        help="Select the type of verification to perform"
    )

    # Step 2: Select images (outside of form to allow interactive browsing)
    st.subheader("Step 2: Select Images")
    col1, col2 = st.columns(2)

    with col1:
        # Reference Image S3 Selector
        reference_image_url = render_improved_s3_image_selector(
            api_client, "reference", "Reference Image", "ref_img"
        )

    with col2:
        # Checking Image S3 Selector
        checking_image_url = render_improved_s3_image_selector(
            api_client, "checking", "Checking Image", "check_img"
        )

    # Step 3: Submit verification
    st.subheader("Step 3: Submit Verification")
    with st.form(key='init_verification_form'):
        submit_button = st.form_submit_button(label='Start Verification')

    if submit_button:
        # Validation - only images are required for both verification types
        required_fields = [reference_image_url, checking_image_url]
        if not all(required_fields):
            st.error("Both Reference Image and Checking Image must be selected.")
            return

        try:
            # Simplified API call with only the three required fields
            response = api_client.initiate_verification(
                verificationType=verification_type,
                referenceImageUrl=reference_image_url,
                checkingImageUrl=checking_image_url
            )

            st.success("Verification initiated successfully!")
            verification_id = response.get('verificationId', 'N/A')
            st.write(f"Verification ID: {verification_id}")
            st.write(f"Status: {response.get('status', 'N/A')}")
            st.write(f"Initiated at: {response.get('verificationAt', 'N/A')}")

            # Store the verification data in session state for LLM response display
            if verification_id != 'N/A':
                st.session_state['selected_verification'] = verification_id
                st.session_state['current_verification_id'] = verification_id
                st.session_state['current_verification_data'] = response
                st.session_state['show_llm_response'] = True
                st.info("✅ Verification initiated successfully! You can view the status in the Verification Results page.")

                # Auto-refresh to show LLM response section
                time.sleep(1)
                st.rerun()

        except Exception as e:
            logger.error(f"Initiate verification failed: {str(e)}")
            st.error(str(e))

    # Display LLM Response Section if verification has been initiated
    if st.session_state.get('show_llm_response', False) and st.session_state.get('current_verification_id'):
        verification_id = st.session_state['current_verification_id']
        verification_data = st.session_state.get('current_verification_data', {})

        # Add refresh button for real-time updates
        col1, col2, col3 = st.columns([1, 1, 2])
        with col1:
            if st.button("🔄 Refresh Status", help="Check for updated verification results"):
                try:
                    # Fetch latest verification data
                    updated_data = api_client.get_verification_details(verification_id)
                    if updated_data:
                        st.session_state['current_verification_data'] = updated_data
                        st.rerun()
                except Exception as e:
                    st.warning(f"Could not refresh status: {str(e)}")

        with col2:
            if st.button("❌ Clear Results", help="Hide LLM response section"):
                st.session_state['show_llm_response'] = False
                st.session_state['current_verification_id'] = None
                st.session_state['current_verification_data'] = None
                st.rerun()

        with col3:
            st.markdown(f"<small style='color: #A0A0A0;'>Last updated: {datetime.now().strftime('%H:%M:%S')}</small>", unsafe_allow_html=True)

        # Render the LLM response section
        render_llm_response_section(api_client, verification_id, verification_data)
