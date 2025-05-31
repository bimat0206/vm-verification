import streamlit as st
import logging
import boto3
import uuid
from datetime import datetime
from PIL import Image

logger = logging.getLogger(__name__)

def get_bucket_name(api_client, bucket_type):
    """Get the actual bucket name from the CONFIG_SECRET in AWS Secrets Manager"""
    try:
        # Get bucket name from the API client's config loader (which loads from CONFIG_SECRET)
        if bucket_type == "reference":
            bucket_name = api_client.config_loader.get('REFERENCE_BUCKET', '')
        elif bucket_type == "checking":
            bucket_name = api_client.config_loader.get('CHECKING_BUCKET', '')
        else:
            logger.warning(f"Unknown bucket type: {bucket_type}")
            raise ValueError(f"Unknown bucket type: {bucket_type}")

        # If we don't have the bucket name from configuration, this is an error
        if not bucket_name:
            error_msg = f"Bucket name for {bucket_type} not found in configuration. Please ensure REFERENCE_BUCKET and CHECKING_BUCKET are properly configured."
            logger.error(error_msg)
            raise ValueError(error_msg)

        logger.info(f"Retrieved {bucket_type} bucket name from configuration: {bucket_name}")
        return bucket_name

    except Exception as e:
        logger.error(f"Failed to get bucket name for {bucket_type}: {str(e)}")
        # Re-raise the exception so the UI can show the proper error
        raise Exception(f"Configuration error: Could not retrieve {bucket_type} bucket name. Please check your configuration.")

def validate_file(uploaded_file, bucket_type):
    """Validate uploaded file based on bucket type"""
    if uploaded_file is None:
        return False, "No file uploaded"

    # Check file size (limit to 10MB)
    if uploaded_file.size > 10 * 1024 * 1024:
        return False, "File size must be less than 10MB"

    if bucket_type == "reference":
        # Reference bucket only allows JSON files
        # Check if it's a JSON file first (by extension, as MIME type detection can be unreliable)
        if uploaded_file.name.lower().endswith('.json'):
            try:
                # Try to parse JSON to validate it's valid JSON
                import json
                uploaded_file.seek(0)
                content = uploaded_file.read().decode('utf-8')
                json.loads(content)
                uploaded_file.seek(0)  # Reset file pointer
                return True, "Valid JSON file"
            except json.JSONDecodeError as e:
                return False, f"Invalid JSON file: {str(e)}"
            except Exception as e:
                return False, f"Error reading JSON file: {str(e)}"

        # If MIME type detection failed but it's a JSON file by extension, try JSON validation
        if uploaded_file.type in ['application/json', 'text/json', 'text/plain'] and uploaded_file.name.lower().endswith('.json'):
            try:
                import json
                uploaded_file.seek(0)
                content = uploaded_file.read().decode('utf-8')
                json.loads(content)
                uploaded_file.seek(0)  # Reset file pointer
                return True, "Valid JSON file"
            except json.JSONDecodeError as e:
                return False, f"Invalid JSON file: {str(e)}"
            except Exception as e:
                return False, f"Error reading JSON file: {str(e)}"

        # If we get here, the file type is not supported
        return False, f"File type '{uploaded_file.type}' not supported. Reference bucket only accepts JSON files (.json)"

    elif bucket_type == "checking":
        # Checking bucket only allows image files
        allowed_types = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/bmp']
        if uploaded_file.type not in allowed_types:
            return False, f"File type {uploaded_file.type} not allowed. Checking bucket only accepts image files: {', '.join(allowed_types)}"

        try:
            # Try to open the image to validate it's a valid image file
            image = Image.open(uploaded_file)
            image.verify()
            return True, "Valid image file"
        except Exception as e:
            return False, f"Invalid image file: {str(e)}"

    return False, "Unknown bucket type"

def upload_to_s3(uploaded_file, bucket_name, key):
    """Upload file directly to S3"""
    try:
        # Initialize S3 client
        s3_client = boto3.client('s3')
        
        # Reset file pointer to beginning
        uploaded_file.seek(0)
        
        # Upload file
        s3_client.upload_fileobj(
            uploaded_file,
            bucket_name,
            key,
            ExtraArgs={
                'ContentType': uploaded_file.type,
                'Metadata': {
                    'uploaded_by': 'streamlit_app',
                    'upload_timestamp': datetime.now().isoformat()
                }
            }
        )
        
        return True, f"s3://{bucket_name}/{key}"
    except Exception as e:
        logger.error(f"S3 upload failed: {str(e)}")
        return False, str(e)

def render_s3_path_browser(api_client, bucket_type, session_key):
    """Render an S3 path browser component for selecting upload location"""
    st.write("**📁 Browse and Select Upload Path:**")

    # Initialize session state for this browser
    if f"{session_key}_path" not in st.session_state:
        st.session_state[f"{session_key}_path"] = ""
    if f"{session_key}_selected_path" not in st.session_state:
        st.session_state[f"{session_key}_selected_path"] = ""

    # Path input and browse button
    col_path, col_browse = st.columns([3, 1])
    with col_path:
        current_path = st.text_input(
            f"Path in {bucket_type} bucket",
            value=st.session_state[f"{session_key}_path"],
            key=f"{session_key}_path_input",
            help="Leave blank for root directory"
        )
    with col_browse:
        browse_button = st.button("Browse", key=f"{session_key}_browse")

    # Update path if user typed in the input
    if current_path != st.session_state[f"{session_key}_path"]:
        st.session_state[f"{session_key}_path"] = current_path
        # Clear items to force refresh when path changes
        if f"{session_key}_items" in st.session_state:
            del st.session_state[f"{session_key}_items"]

    # Load folder structure if browse button clicked or no items cached
    should_load = browse_button or f"{session_key}_items" not in st.session_state
    if should_load:
        try:
            with st.spinner("Loading folder structure..."):
                browser_response = api_client.browse_images(current_path, bucket_type)
                st.session_state[f"{session_key}_items"] = browser_response.get('items', [])
                st.session_state[f"{session_key}_current_path"] = browser_response.get('currentPath', '')
                st.session_state[f"{session_key}_parent_path"] = browser_response.get('parentPath', '')
        except Exception as e:
            st.error(f"Failed to browse {bucket_type} bucket: {str(e)}")
            return None

    # Display current path and navigation
    if f"{session_key}_current_path" in st.session_state:
        current_display_path = st.session_state[f"{session_key}_current_path"]
        st.write(f"**Current Path:** {current_display_path if current_display_path else '(root)'}")

        # Navigation buttons
        col_up, col_select = st.columns([1, 2])

        with col_up:
            # Go up button
            if st.session_state.get(f"{session_key}_parent_path") is not None:
                if st.button("⬆️ Go Up", key=f"{session_key}_go_up"):
                    new_path = st.session_state[f"{session_key}_parent_path"]
                    st.session_state[f"{session_key}_path"] = new_path
                    # Clear items to force refresh
                    if f"{session_key}_items" in st.session_state:
                        del st.session_state[f"{session_key}_items"]
                    st.rerun()

        with col_select:
            # Select current path button
            if st.button("📍 Select Current Path", key=f"{session_key}_select_current"):
                st.session_state[f"{session_key}_selected_path"] = current_display_path
                # Don't rerun immediately, let the user see the selection

    # Display items (folders only for path selection)
    if f"{session_key}_items" in st.session_state:
        items = st.session_state[f"{session_key}_items"]
        folders = [item for item in items if item.get('type') == 'folder']

        if not folders:
            st.info("No folders found in this location. You can upload to the current path or go up.")
        else:
            st.write("**Available Folders:**")
            # Create grid layout for folders
            cols = st.columns(3)
            for idx, item in enumerate(folders):
                col = cols[idx % 3]
                with col:
                    name = item.get('name', 'Unknown')
                    item_path = item.get('path', '')

                    if st.button(f"📁 {name}", key=f"{session_key}_folder_{idx}"):
                        st.session_state[f"{session_key}_path"] = item_path
                        # Clear items to force refresh when navigating to new folder
                        if f"{session_key}_items" in st.session_state:
                            del st.session_state[f"{session_key}_items"]
                        st.rerun()

    # Display selected path
    selected_path = st.session_state.get(f"{session_key}_selected_path", "")
    if selected_path:
        st.success(f"**Selected Upload Path:** `{selected_path if selected_path else '(root)'}`")

    return selected_path

def render_upload_section(api_client, bucket_type, title):
    """Render an upload section for a specific bucket type"""
    st.subheader(f"{title} Upload")

    # File uploader with different file types based on bucket type
    if bucket_type == "reference":
        uploaded_file = st.file_uploader(
            f"Choose {title.lower()} JSON file",
            type=['json'],
            key=f"{bucket_type}_uploader",
            help=f"Upload JSON files to the {bucket_type} bucket"
        )
    else:  # checking bucket
        uploaded_file = st.file_uploader(
            f"Choose {title.lower()} image",
            type=['png', 'jpg', 'jpeg', 'gif', 'bmp'],
            key=f"{bucket_type}_uploader",
            help=f"Upload images to the {bucket_type} bucket"
        )
    
    if uploaded_file is not None:
        # Display file info
        st.write(f"**File name:** {uploaded_file.name}")
        st.write(f"**File size:** {uploaded_file.size:,} bytes")
        st.write(f"**File type:** {uploaded_file.type}")
        
        # Show preview based on bucket type and file type
        if bucket_type == "reference":
            # Reference bucket only has JSON files
            try:
                # Show JSON content preview
                uploaded_file.seek(0)
                content = uploaded_file.read().decode('utf-8')
                uploaded_file.seek(0)  # Reset file pointer
                st.write("**JSON Preview:**")
                st.code(content[:500] + "..." if len(content) > 500 else content, language='json')
            except Exception as e:
                st.warning(f"Could not display JSON preview: {str(e)}")
        elif bucket_type == "checking":
            # Checking bucket only has images
            try:
                image = Image.open(uploaded_file)
                st.image(image, caption=f"Preview: {uploaded_file.name}", width=300)
            except Exception as e:
                st.warning(f"Could not display image preview: {str(e)}")

        # Validate file
        is_valid, validation_message = validate_file(uploaded_file, bucket_type)
        
        if is_valid:
            st.success(validation_message)

            # Upload options - different for reference vs checking buckets
            if bucket_type == "reference":
                # Reference bucket: Use interactive S3 browser
                selected_path = render_s3_path_browser(api_client, bucket_type, f"{bucket_type}_browser")

                # Custom filename input
                st.write("**📝 File Name:**")
                custom_filename = st.text_input(
                    "Custom filename (optional)",
                    value="",
                    key=f"{bucket_type}_custom_filename",
                    help="Leave empty to use original filename"
                )

                # Show final path
                final_filename = custom_filename.strip() if custom_filename.strip() else uploaded_file.name
                if not final_filename.lower().endswith('.json'):
                    final_filename += '.json'

                # Generate final path - recalculate at upload time to ensure we have the latest selected path
                # Store the filename in session state for upload
                st.session_state[f"{bucket_type}_upload_filename"] = final_filename

                # Show preview of what the path will be
                if selected_path:
                    # Remove any trailing slashes and add proper separator
                    clean_path = selected_path.rstrip('/')
                    preview_path = f"{clean_path}/{final_filename}" if clean_path else final_filename
                else:
                    preview_path = final_filename

                st.write(f"**Final Upload Path:** `{preview_path}`")

                # Show path selection status
                if selected_path:
                    st.info(f"📁 **Selected folder:** `{selected_path}`")
                else:
                    st.info("📁 **Upload location:** Root directory (no folder selected)")

            else:
                # Checking bucket: Use automatic path generation
                col1, col2 = st.columns(2)

                with col1:
                    # Custom path input
                    custom_path = st.text_input(
                        f"Custom path (optional)",
                        value="",
                        key=f"{bucket_type}_custom_path",
                        help="Leave empty for automatic path generation"
                    )

                with col2:
                    # Generate automatic path
                    timestamp = datetime.now().strftime("%Y/%m/%d")
                    unique_id = str(uuid.uuid4())[:8]
                    auto_path = f"uploads/{timestamp}/{unique_id}_{uploaded_file.name}"
                    st.write(f"**Auto path:** {auto_path}")

                # Determine final path
                final_path = custom_path.strip() if custom_path.strip() else auto_path
            
            # Upload button
            if st.button(f"Upload to {title} Bucket", key=f"{bucket_type}_upload_btn"):
                try:
                    # Get bucket name
                    bucket_name = get_bucket_name(api_client, bucket_type)

                    # Recalculate final path at upload time for reference bucket
                    if bucket_type == "reference":
                        # Get the current selected path from session state
                        current_selected_path = st.session_state.get(f"{bucket_type}_browser_selected_path", "")
                        stored_filename = st.session_state.get(f"{bucket_type}_upload_filename", uploaded_file.name)

                        logger.info(f"Upload - Selected path: '{current_selected_path}', Filename: '{stored_filename}'")

                        if current_selected_path:
                            clean_path = current_selected_path.rstrip('/')
                            final_path = f"{clean_path}/{stored_filename}" if clean_path else stored_filename
                        else:
                            final_path = stored_filename

                        st.write(f"**Uploading to path:** `{final_path}`")
                        logger.info(f"Final upload path: '{final_path}'")

                    # Show upload progress
                    with st.spinner(f"Uploading to {bucket_name}..."):
                        success, result = upload_to_s3(uploaded_file, bucket_name, final_path)
                    
                    if success:
                        st.success(f"✅ Upload successful!")
                        st.write(f"**S3 URL:** {result}")
                        st.write(f"**Bucket:** {bucket_name}")
                        st.write(f"**Key:** {final_path}")
                        
                        # Store in session state for reference
                        if f"{bucket_type}_uploads" not in st.session_state:
                            st.session_state[f"{bucket_type}_uploads"] = []
                        
                        st.session_state[f"{bucket_type}_uploads"].append({
                            'filename': uploaded_file.name,
                            'url': result,
                            'bucket': bucket_name,
                            'key': final_path,
                            'timestamp': datetime.now().isoformat()
                        })
                        
                    else:
                        st.error(f"❌ Upload failed: {result}")
                        
                except Exception as e:
                    st.error(f"❌ Upload error: {str(e)}")
                    logger.error(f"Upload error for {bucket_type}: {str(e)}")
        else:
            st.error(f"❌ {validation_message}")

def app(api_client):
    st.title("🔄 Image Upload")
    st.write("Upload images to Reference and Checking buckets for verification processing.")

    # Display S3 bucket addresses from ECS environment
    st.subheader("📍 S3 Bucket Configuration")

    col1, col2 = st.columns(2)

    with col1:
        try:
            reference_bucket = get_bucket_name(api_client, "reference")
            st.success(f"**Reference Bucket:**")
            st.code(f"s3://{reference_bucket}", language='text')
        except Exception as e:
            st.error(f"**Reference Bucket:** Configuration Error")
            st.code(f"Error: {str(e)}", language='text')

    with col2:
        try:
            checking_bucket = get_bucket_name(api_client, "checking")
            st.success(f"**Checking Bucket:**")
            st.code(f"s3://{checking_bucket}", language='text')
        except Exception as e:
            st.error(f"**Checking Bucket:** Configuration Error")
            st.code(f"Error: {str(e)}", language='text')

    # Configuration source info
    config_source = "Unknown"
    if api_client.config_loader.is_loaded_from_secret():
        config_source = "AWS Secrets Manager (CONFIG_SECRET)"
    elif hasattr(api_client.config_loader, 'is_loaded_from_streamlit') and api_client.config_loader.is_loaded_from_streamlit():
        config_source = "Streamlit Secrets (Local Development)"
    else:
        config_source = "Environment Variables"

    st.info(f"📝 **Configuration Source:** {config_source}")
    st.markdown("---")

    # Warning about direct S3 access
    st.info("📝 **Note:** This tool uploads images directly to S3 buckets. Make sure you have proper AWS credentials configured.")
    
    # Create two columns for the upload sections
    col1, col2 = st.columns(2)
    
    with col1:
        render_upload_section(api_client, "reference", "Reference")
    
    with col2:
        render_upload_section(api_client, "checking", "Checking")
    
    # Display upload history
    st.markdown("---")
    st.subheader("📋 Upload History")
    
    # Tabs for different bucket histories
    tab1, tab2 = st.tabs(["Reference Uploads", "Checking Uploads"])
    
    with tab1:
        if "reference_uploads" in st.session_state and st.session_state["reference_uploads"]:
            for upload in reversed(st.session_state["reference_uploads"]):
                with st.expander(f"📁 {upload['filename']} - {upload['timestamp'][:19]}"):
                    st.write(f"**URL:** {upload['url']}")
                    st.write(f"**Bucket:** {upload['bucket']}")
                    st.write(f"**Key:** {upload['key']}")
                    st.code(upload['url'], language='text')
        else:
            st.info("No reference uploads yet.")

    with tab2:
        if "checking_uploads" in st.session_state and st.session_state["checking_uploads"]:
            for upload in reversed(st.session_state["checking_uploads"]):
                with st.expander(f"📁 {upload['filename']} - {upload['timestamp'][:19]}"):
                    st.write(f"**URL:** {upload['url']}")
                    st.write(f"**Bucket:** {upload['bucket']}")
                    st.write(f"**Key:** {upload['key']}")
                    st.code(upload['url'], language='text')
        else:
            st.info("No checking uploads yet.")
    
    # Clear history buttons
    st.markdown("---")
    col1, col2, col3 = st.columns([1, 1, 1])
    
    with col1:
        if st.button("🗑️ Clear Reference History"):
            st.session_state["reference_uploads"] = []
            st.success("Reference upload history cleared!")
            st.rerun()
    
    with col2:
        if st.button("🗑️ Clear Checking History"):
            st.session_state["checking_uploads"] = []
            st.success("Checking upload history cleared!")
            st.rerun()
    
    with col3:
        if st.button("🗑️ Clear All History"):
            st.session_state["reference_uploads"] = []
            st.session_state["checking_uploads"] = []
            st.success("All upload history cleared!")
            st.rerun()
