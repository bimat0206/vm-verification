import streamlit as st
import logging
from .improved_image_selector import render_improved_s3_image_selector

logger = logging.getLogger(__name__)

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
    st.title("Vending Machine Verification System")
    st.write("Welcome! Start a new verification by selecting images and configuring the verification type below.")

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
    verification_type = st.selectbox(
        "Verification Type",
        ["LAYOUT_VS_CHECKING", "PREVIOUS_VS_CURRENT"],
        format_func=lambda x: "Layout vs Checking" if x == "LAYOUT_VS_CHECKING" else "Previous vs Current"
    )

    # Step 2: Select images (outside of form to allow interactive browsing)
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

    # Step 3: Additional metadata in form
    with st.form(key='init_verification_form'):
        # Only show Vending Machine ID for PREVIOUS_VS_CURRENT
        vending_machine_id = None
        if verification_type == "PREVIOUS_VS_CURRENT":
            vending_machine_id = st.text_input("Vending Machine ID", help="Required - ID of the vending machine")

            # Type-specific fields for PREVIOUS_VS_CURRENT only
            previous_verification_id = st.text_input("Previous Verification ID", help="Optional - ID of previous verification")
        else:
            previous_verification_id = None

        submit_button = st.form_submit_button(label='Start Verification')

    if submit_button:
        # Validation based on verification type
        if verification_type == "LAYOUT_VS_CHECKING":
            # For Layout vs Checking, only images are required
            required_fields = [reference_image_url, checking_image_url]
            if not all(required_fields):
                st.error("Both Reference Image and Checking Image must be selected.")
                return
        else:  # PREVIOUS_VS_CURRENT
            # For Previous vs Current, vending machine ID and images are required
            required_fields = [vending_machine_id, reference_image_url, checking_image_url]
            if not all(required_fields):
                st.error("Vending Machine ID, Reference Image, and Checking Image are all required.")
                return

        try:
            # Different parameters based on verification type
            if verification_type == "LAYOUT_VS_CHECKING":
                # For Layout vs Checking, we don't need layout_id, layout_prefix, or vending_machine_id
                response = api_client.initiate_verification(
                    verification_type=verification_type,
                    reference_image_url=reference_image_url,
                    checking_image_url=checking_image_url,
                    vending_machine_id="default",  # Use a default value since API might require it
                    layout_id=1,  # Use default values since these are removed from UI
                    layout_prefix="default"
                )
            else:  # PREVIOUS_VS_CURRENT
                response = api_client.initiate_verification(
                    verification_type=verification_type,
                    reference_image_url=reference_image_url,
                    checking_image_url=checking_image_url,
                    vending_machine_id=vending_machine_id,
                    previous_verification_id=previous_verification_id if previous_verification_id else None
                )

            st.success("Verification initiated successfully!")
            verification_id = response.get('verificationId', 'N/A')
            st.write(f"Verification ID: {verification_id}")
            st.write(f"Status: {response.get('status', 'N/A')}")
            st.write(f"Initiated at: {response.get('verificationAt', 'N/A')}")

            # Store the verification ID in session state
            if verification_id != 'N/A':
                st.session_state['selected_verification'] = verification_id
                st.info("✅ Verification initiated successfully! You can view the status in the Verification Results page.")

        except Exception as e:
            logger.error(f"Initiate verification failed: {str(e)}")
            st.error(str(e))