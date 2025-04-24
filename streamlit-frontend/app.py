import streamlit as st
import requests
import pandas as pd
import json
import time
import io
import os
from PIL import Image
from datetime import datetime, timedelta
import plotly.graph_objects as go
import matplotlib.pyplot as plt
import matplotlib.patches as patches

# Configuration
API_URL = os.getenv("API_URL", "http://localhost:3000")  # Default to localhost

# Set page config
st.set_page_config(
    page_title="Vending Machine Verification System",
    page_icon="🔍",
    layout="wide",
    initial_sidebar_state="expanded"
)

# Custom CSS
st.markdown("""
<style>
    .main-header {
        font-size: 2.5rem;
        font-weight: 700;
        margin-bottom: 1rem;
    }
    .sub-header {
        font-size: 1.5rem;
        font-weight: 500;
        margin-bottom: 1rem;
    }
    .status-card {
        padding: 1rem;
        border-radius: 5px;
        margin-bottom: 1rem;
    }
    .status-correct {
        background-color: #d4edda;
        color: #155724;
    }
    .status-incorrect {
        background-color: #f8d7da;
        color: #721c24;
    }
    .status-processing {
        background-color: #fff3cd;
        color: #856404;
    }
    .metric-card {
        background-color: #f8f9fa;
        border-radius: 5px;
        padding: 1rem;
        text-align: center;
        margin: 0.5rem;
    }
    .metric-value {
        font-size: 2rem;
        font-weight: 700;
    }
    .metric-label {
        font-size: 1rem;
        color: #6c757d;
    }
    .discrepancy-card {
        background-color: #f1f1f1;
        border-radius: 5px;
        padding: 1rem;
        margin-bottom: 0.5rem;
    }
    .position-label {
        font-weight: 700;
        color: #212529;
    }
    .discrepancy-type {
        color: #dc3545;
        font-weight: 500;
    }
    .expected-value {
        color: #28a745;
    }
    .found-value {
        color: #dc3545;
    }
</style>
""", unsafe_allow_html=True)

# Helper Functions
def format_timestamp(timestamp_str):
    """Format timestamp string to a more readable format"""
    if not timestamp_str:
        return ""
    try:
        timestamp = datetime.fromisoformat(timestamp_str.replace('Z', '+00:00'))
        return timestamp.strftime('%Y-%m-%d %H:%M:%S')
    except:
        return timestamp_str

def get_verification(verification_id):
    """Get verification details from API"""
    try:
        response = requests.get(f"{API_URL}/api/v1/verification/{verification_id}")
        if response.status_code == 200:
            return response.json()
        elif response.status_code == 202:
            # Still processing
            return response.json()
        else:
            st.error(f"Error retrieving verification: {response.text}")
            return None
    except Exception as e:
        st.error(f"Error connecting to API: {str(e)}")
        return None

def list_verifications(filters=None, limit=20, offset=0):
    """List verifications with optional filters"""
    try:
        params = {"limit": limit, "offset": offset}
        if filters:
            params.update(filters)
        
        response = requests.get(f"{API_URL}/api/v1/verification", params=params)
        if response.status_code == 200:
            return response.json()
        else:
            st.error(f"Error listing verifications: {response.text}")
            return None
    except Exception as e:
        st.error(f"Error connecting to API: {str(e)}")
        return None

def initiate_verification(reference_img, checking_img, vm_id, layout_id, layout_prefix):
    """Initiate a new verification"""
    try:
        # First, upload both images to S3 (in a real implementation)
        # For this demo, we'll assume the S3 URLs are generated from filenames
        
        # In a real implementation, you would upload the files to S3 and get the URLs
        reference_filename = reference_img.name
        checking_filename = checking_img.name
        
        # Create mock S3 URLs for demo purposes
        timestamp = datetime.now().strftime("%Y-%m-%d/%H-%M-%S")
        reference_image_url = f"s3://kootoro-reference-bucket/processed/{timestamp}/{layout_id}_{layout_prefix}/image.png"
        checking_image_url = f"s3://kootoro-checking-bucket/{timestamp}/{vm_id}/check_{timestamp.replace('/', '_')}.jpg"
        
        # Prepare request payload
        payload = {
            "referenceImageUrl": reference_image_url,
            "checkingImageUrl": checking_image_url,
            "vendingMachineId": vm_id,
            "layoutId": layout_id,
            "layoutPrefix": layout_prefix,
            "notificationEnabled": True
        }
        
        # Send request to API
        response = requests.post(f"{API_URL}/api/v1/verification", json=payload)
        if response.status_code in [200, 201, 202]:
            return response.json()
        else:
            st.error(f"Error initiating verification: {response.text}")
            return None
    except Exception as e:
        st.error(f"Error connecting to API: {str(e)}")
        return None

def draw_discrepancy_visualization(reference_img, checking_img, discrepancies):
    """Create a simple visualization showing discrepancies"""
    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(12, 6))
    
    # Display reference image
    ax1.imshow(reference_img)
    ax1.set_title("Reference Image")
    ax1.axis('off')
    
    # Display checking image with discrepancy highlights
    ax2.imshow(checking_img)
    ax2.set_title("Checking Image with Discrepancies")
    ax2.axis('off')
    
    # Add rectangles for discrepancies
    for disc in discrepancies:
        if disc.get('imageCoordinates'):
            coords = disc['imageCoordinates']
            rect = patches.Rectangle(
                (coords['x'], coords['y']), 
                coords['width'], 
                coords['height'], 
                linewidth=2, 
                edgecolor='r', 
                facecolor='none'
            )
            ax2.add_patch(rect)
            ax2.text(
                coords['x'], 
                coords['y'] - 10, 
                f"{disc['position']}: {disc['issue']}", 
                color='red', 
                fontsize=8, 
                bbox=dict(facecolor='white', alpha=0.7)
            )
    
    st.pyplot(fig)

def status_badge(status):
    """Return HTML for status badge"""
    if status in ["CORRECT"]:
        return f'<span style="background-color: #28a745; color: white; padding: 3px 8px; border-radius: 3px;">{status}</span>'
    elif status in ["INCORRECT", "ERROR"]:
        return f'<span style="background-color: #dc3545; color: white; padding: 3px 8px; border-radius: 3px;">{status}</span>'
    elif status in ["PARTIAL_RESULTS"]:
        return f'<span style="background-color: #ffc107; color: black; padding: 3px 8px; border-radius: 3px;">{status}</span>'
    else:
        return f'<span style="background-color: #17a2b8; color: white; padding: 3px 8px; border-radius: 3px;">{status}</span>'

def severity_badge(severity):
    """Return HTML for severity badge"""
    if severity == "Low":
        return f'<span style="background-color: #28a745; color: white; padding: 3px 8px; border-radius: 3px;">{severity}</span>'
    elif severity == "Medium":
        return f'<span style="background-color: #ffc107; color: black; padding: 3px 8px; border-radius: 3px;">{severity}</span>'
    elif severity == "High":
        return f'<span style="background-color: #dc3545; color: white; padding: 3px 8px; border-radius: 3px;">{severity}</span>'
    else:
        return f'<span style="background-color: #6c757d; color: white; padding: 3px 8px; border-radius: 3px;">{severity}</span>'

def create_accuracy_gauge(accuracy):
    """Create a gauge chart for accuracy percentage"""
    fig = go.Figure(go.Indicator(
        mode = "gauge+number",
        value = accuracy,
        domain = {'x': [0, 1], 'y': [0, 1]},
        title = {'text': "Accuracy"},
        gauge = {
            'axis': {'range': [None, 100]},
            'steps': [
                {'range': [0, 60], 'color': "#ff5252"},
                {'range': [60, 80], 'color': "#ffeb3b"},
                {'range': [80, 100], 'color': "#4caf50"}
            ],
            'threshold': {
                'line': {'color': "red", 'width': 4},
                'thickness': 0.75,
                'value': 90
            }
        }
    ))
    
    fig.update_layout(height=250)
    return fig

# Sidebar Navigation
st.sidebar.markdown('<p class="main-header">Verification System</p>', unsafe_allow_html=True)
page = st.sidebar.radio("Navigation", ["New Verification", "Verification History", "Dashboard", "Settings"])

# Settings in sidebar
if page == "Settings":
    st.sidebar.markdown("---")
    sidebar_api_url = st.sidebar.text_input("API URL", value=API_URL)
    if sidebar_api_url != API_URL:
        API_URL = sidebar_api_url
        st.sidebar.success(f"API URL updated to {API_URL}")

# Main content
if page == "New Verification":
    st.markdown('<p class="main-header">New Verification</p>', unsafe_allow_html=True)
    
    # Form for new verification
    with st.form("verification_form"):
        col1, col2 = st.columns(2)
        
        with col1:
            st.markdown('<p class="sub-header">Reference Image</p>', unsafe_allow_html=True)
            reference_img = st.file_uploader("Upload Reference Image", type=["jpg", "jpeg", "png"])
            if reference_img:
                st.image(reference_img, caption="Reference Image", use_column_width=True)
        
        with col2:
            st.markdown('<p class="sub-header">Checking Image</p>', unsafe_allow_html=True)
            checking_img = st.file_uploader("Upload Checking Image", type=["jpg", "jpeg", "png"])
            if checking_img:
                st.image(checking_img, caption="Checking Image", use_column_width=True)
        
        # Metadata inputs
        st.markdown('<p class="sub-header">Verification Metadata</p>', unsafe_allow_html=True)
        col1, col2, col3 = st.columns(3)
        
        with col1:
            vm_id = st.text_input("Vending Machine ID", value="VM-3245")
        
        with col2:
            layout_id = st.number_input("Layout ID", min_value=1, value=23591)
        
        with col3:
            layout_prefix = st.text_input("Layout Prefix", value="1q2w3e")
        
        submit_button = st.form_submit_button("Start Verification")
    
    # Handle form submission
    if submit_button:
        if not (reference_img and checking_img):
            st.error("Please upload both reference and checking images")
        else:
            with st.spinner("Initiating verification..."):
                result = initiate_verification(reference_img, checking_img, vm_id, layout_id, layout_prefix)
                
                if result:
                    st.success(f"Verification initiated! Verification ID: {result['verificationId']}")
                    st.session_state.current_verification = result['verificationId']
                    st.session_state.verification_status = result['status']
                    
                    # Create expandable section for verification details
                    with st.expander("Verification Details"):
                        st.json(result)
                    
                    # Redirect to verification result
                    st.info("You will be redirected to the verification result page once processing is complete...")
                    
                    # Poll for status change
                    placeholder = st.empty()
                    progress_bar = st.progress(0)
                    
                    max_polls = 10  # Maximum number of polls
                    for i in range(max_polls):
                        time.sleep(2)  # Wait 2 seconds between polls
                        progress_bar.progress((i + 1) / max_polls)
                        
                        verification = get_verification(result['verificationId'])
                        if verification:
                            if verification['status'] not in ['INITIALIZED', 'IMAGES_FETCHED', 'SYSTEM_PROMPT_READY',
                                                            'TURN1_PROMPT_READY', 'TURN1_PROCESSING', 'TURN1_COMPLETED',
                                                            'TURN2_PROMPT_READY', 'TURN2_PROCESSING']:
                                placeholder.success("Verification complete! View the results below.")
                                break
                            else:
                                placeholder.info(f"Status: {verification['status']} - Still processing...")
                    
                    # Show mock result for demo
                    if 'verification' in locals() and verification and verification['status'] in ['TURN2_COMPLETED', 'RESULTS_FINALIZED', 'RESULTS_STORED', 'NOTIFICATION_SENT']:
                        st.markdown('<p class="sub-header">Verification Result</p>', unsafe_allow_html=True)
                        
                        # Show result status
                        if 'status' in verification:
                            status_class = "status-correct" if verification['status'] == "CORRECT" else "status-incorrect"
                            st.markdown(f'<div class="status-card {status_class}">Status: {verification["status"]}</div>', unsafe_allow_html=True)
                        
                        # Show mock verification summary
                        st.markdown('<p class="sub-header">Summary</p>', unsafe_allow_html=True)
                        
                        # Mock data for summary
                        mock_summary = {
                            "totalPositionsChecked": 60,
                            "correctPositions": 58,
                            "discrepantPositions": 2,
                            "missingProducts": 1,
                            "incorrectProductTypes": 1,
                            "unexpectedProducts": 0,
                            "emptyPositionsCount": 3,
                            "overallAccuracy": 96.7,
                            "overallConfidence": 95,
                            "verificationStatus": "INCORRECT",
                            "verificationOutcome": "Discrepancies Detected - Row E contains Incorrect Product Type and Row F contains Missing Product"
                        }
                        
                        # Display summary metrics
                        col1, col2, col3, col4 = st.columns(4)
                        
                        with col1:
                            st.markdown(f'''
                            <div class="metric-card">
                                <div class="metric-value">{mock_summary["totalPositionsChecked"]}</div>
                                <div class="metric-label">Positions Checked</div>
                            </div>
                            ''', unsafe_allow_html=True)
                        
                        with col2:
                            st.markdown(f'''
                            <div class="metric-card">
                                <div class="metric-value">{mock_summary["correctPositions"]}</div>
                                <div class="metric-label">Correct Positions</div>
                            </div>
                            ''', unsafe_allow_html=True)
                        
                        with col3:
                            st.markdown(f'''
                            <div class="metric-card">
                                <div class="metric-value">{mock_summary["discrepantPositions"]}</div>
                                <div class="metric-label">Discrepancies</div>
                            </div>
                            ''', unsafe_allow_html=True)
                        
                        with col4:
                            st.markdown(f'''
                            <div class="metric-card">
                                <div class="metric-value">{mock_summary["emptyPositionsCount"]}</div>
                                <div class="metric-label">Empty Positions</div>
                            </div>
                            ''', unsafe_allow_html=True)
                        
                        # Display accuracy gauge
                        st.plotly_chart(create_accuracy_gauge(mock_summary["overallAccuracy"]))
                        
                        # Mock discrepancies
                        st.markdown('<p class="sub-header">Discrepancies</p>', unsafe_allow_html=True)
                        
                        mock_discrepancies = [
                            {
                                "position": "E01",
                                "expected": 'Green "Mi Cung Đình" cup noodle',
                                "found": 'Red/white "Mi modern Lẩu thái" cup noodle',
                                "issue": "Incorrect Product Type",
                                "confidence": 95,
                                "evidence": "Different packaging color and branding visible",
                                "verificationResult": "INCORRECT",
                                "severity": "High",
                                "imageCoordinates": {
                                    "x": 100,
                                    "y": 150,
                                    "width": 100,
                                    "height": 80
                                }
                            },
                            {
                                "position": "F03",
                                "expected": "Mi Hảo Hảo",
                                "found": "",
                                "issue": "Missing Product",
                                "confidence": 98,
                                "evidence": "Empty slot where product should be",
                                "verificationResult": "INCORRECT",
                                "severity": "Medium",
                                "imageCoordinates": {
                                    "x": 300,
                                    "y": 350,
                                    "width": 90,
                                    "height": 70
                                }
                            }
                        ]
                        
                        for disc in mock_discrepancies:
                            st.markdown(f'''
                            <div class="discrepancy-card">
                                <span class="position-label">Position {disc["position"]}</span> - 
                                <span class="discrepancy-type">{disc["issue"]}</span> 
                                {severity_badge(disc["severity"])}
                                <br>
                                <span class="expected-value">Expected: {disc["expected"]}</span>
                                <br>
                                <span class="found-value">Found: {disc["found"] or "Nothing (Empty)"}</span>
                                <br>
                                Evidence: {disc["evidence"]} (Confidence: {disc["confidence"]}%)
                            </div>
                            ''', unsafe_allow_html=True)
                        
                        # Display mock visualization
                        st.markdown('<p class="sub-header">Visualization</p>', unsafe_allow_html=True)
                        
                        # For demo, we'll just show the uploaded images with mock highlights
                        if reference_img and checking_img:
                            ref_img = Image.open(reference_img)
                            check_img = Image.open(checking_img)
                            draw_discrepancy_visualization(ref_img, check_img, mock_discrepancies)

elif page == "Verification History":
    st.markdown('<p class="main-header">Verification History</p>', unsafe_allow_html=True)
    
    # Filters
    st.markdown('<p class="sub-header">Filters</p>', unsafe_allow_html=True)
    
    col1, col2, col3 = st.columns(3)
    
    with col1:
        vm_id_filter = st.text_input("Vending Machine ID")
    
    with col2:
        status_filter = st.selectbox(
            "Status", 
            ["", "CORRECT", "INCORRECT", "ERROR", "PARTIAL_RESULTS"]
        )
    
    with col3:
        date_range = st.date_input(
            "Date Range",
            value=(datetime.now() - timedelta(days=7), datetime.now()),
            max_value=datetime.now()
        )
    
    # Apply filters button
    filter_button = st.button("Apply Filters")
    
    # Create filters dict based on inputs
    filters = {}
    if vm_id_filter:
        filters["vendingMachineId"] = vm_id_filter
    if status_filter:
        filters["verificationStatus"] = status_filter
    if len(date_range) == 2:
        filters["fromDate"] = date_range[0].isoformat()
        filters["toDate"] = date_range[1].isoformat()
    
    # Get verification history
    if filter_button or page == "Verification History" and not st.session_state.get('history_loaded'):
        with st.spinner("Loading verification history..."):
            verifications = list_verifications(filters)
            st.session_state.history_loaded = True
            
            if not verifications:
                # Show mock data for demonstration
                verifications = {
                    "results": [
                        {
                            "verificationId": "verif-20250423120000",
                            "verificationAt": "2025-04-23T12:00:00Z",
                            "status": "CORRECT",
                            "vendingMachineId": "VM-3245",
                            "discrepancies": []
                        },
                        {
                            "verificationId": "verif-20250422150000",
                            "verificationAt": "2025-04-22T15:00:00Z",
                            "status": "INCORRECT",
                            "vendingMachineId": "VM-3245",
                            "discrepancies": [
                                {
                                    "position": "E01",
                                    "issue": "Incorrect Product Type",
                                    "severity": "High"
                                }
                            ]
                        },
                        {
                            "verificationId": "verif-20250421093000",
                            "verificationAt": "2025-04-21T09:30:00Z",
                            "status": "CORRECT",
                            "vendingMachineId": "VM-3246",
                            "discrepancies": []
                        },
                        {
                            "verificationId": "verif-20250420140000",
                            "verificationAt": "2025-04-20T14:00:00Z",
                            "status": "INCORRECT",
                            "vendingMachineId": "VM-3247",
                            "discrepancies": [
                                {
                                    "position": "A03",
                                    "issue": "Missing Product",
                                    "severity": "Medium"
                                },
                                {
                                    "position": "B02",
                                    "issue": "Incorrect Product Type",
                                    "severity": "Low"
                                }
                            ]
                        }
                    ],
                    "pagination": {
                        "total": 4,
                        "limit": 20,
                        "offset": 0,
                        "nextOffset": 0
                    }
                }
    
            # Display verification history table
            if verifications and "results" in verifications:
                # Create DataFrame
                data = []
                for v in verifications["results"]:
                    # Count discrepancies by severity
                    discrepancies = {
                        "High": 0,
                        "Medium": 0,
                        "Low": 0
                    }
                    
                    if "discrepancies" in v:
                        for disc in v["discrepancies"]:
                            if "severity" in disc:
                                discrepancies[disc["severity"]] += 1
                    
                    data.append({
                        "Verification ID": v["verificationId"],
                        "Date": format_timestamp(v.get("verificationAt", "")),
                        "Vending Machine": v.get("vendingMachineId", ""),
                        "Status": status_badge(v.get("status", "")),
                        "Discrepancies": sum(discrepancies.values()),
                        "High": discrepancies["High"],
                        "Medium": discrepancies["Medium"],
                        "Low": discrepancies["Low"],
                        "View": "📋"
                    })
                
                df = pd.DataFrame(data)
                
                # Convert to HTML with clickable links
                html_table = df.to_html(escape=False, index=False)
                html_table = html_table.replace('&lt;', '<').replace('&gt;', '>')
                
                # Add JavaScript to make rows clickable
                html_table += '''
                <script>
                const table = document.querySelector('table');
                const rows = table.querySelectorAll('tbody tr');
                rows.forEach(row => {
                    row.style.cursor = 'pointer';
                    row.addEventListener('click', () => {
                        const id = row.querySelector('td:first-child').textContent;
                        // In a real app, you'd use a proper router
                        window.location.href = `?verification_id=${id}`;
                    });
                });
                </script>
                '''
                
                st.markdown(html_table, unsafe_allow_html=True)
                
                # Pagination controls
                if "pagination" in verifications:
                    col1, col2, col3 = st.columns([1, 2, 1])
                    
                    with col2:
                        st.markdown(f"Showing {len(verifications['results'])} of {verifications['pagination']['total']} results")
                        
                        prev_disabled = verifications['pagination']['offset'] == 0
                        next_disabled = verifications['pagination']['nextOffset'] == 0
                        
                        col1, col2 = st.columns(2)
                        
                        with col1:
                            if not prev_disabled:
                                st.button("Previous Page")
                        
                        with col2:
                            if not next_disabled:
                                st.button("Next Page")

elif page == "Dashboard":
    st.markdown('<p class="main-header">Dashboard</p>', unsafe_allow_html=True)
    
    # Summary cards
    col1, col2, col3, col4 = st.columns(4)
    
    with col1:
        st.markdown('''
        <div class="metric-card">
            <div class="metric-value">152</div>
            <div class="metric-label">Total Verifications</div>
        </div>
        ''', unsafe_allow_html=True)
    
    with col2:
        st.markdown('''
        <div class="metric-card">
            <div class="metric-value">124</div>
            <div class="metric-label">Correct</div>
        </div>
        ''', unsafe_allow_html=True)
    
    with col3:
        st.markdown('''
        <div class="metric-card">
            <div class="metric-value">28</div>
            <div class="metric-label">Incorrect</div>
        </div>
        ''', unsafe_allow_html=True)
    
    with col4:
        st.markdown('''
        <div class="metric-card">
            <div class="metric-value">81.6%</div>
            <div class="metric-label">Accuracy Rate</div>
        </div>
        ''', unsafe_allow_html=True)
    
    # Charts
    st.markdown('<p class="sub-header">Verification Trends</p>', unsafe_allow_html=True)
    
    # Mock data for trends
    dates = pd.date_range(start=datetime.now() - timedelta(days=30), end=datetime.now(), freq='D')
    correct_counts = [8, 7, 5, 6, 8, 9, 7, 6, 5, 4, 7, 8, 9, 7, 6, 5, 8, 9, 7, 6, 8, 9, 7, 6, 5, 8, 9, 7, 6, 5, 4]
    incorrect_counts = [1, 2, 3, 1, 0, 1, 2, 1, 3, 2, 1, 0, 2, 1, 2, 3, 0, 1, 2, 1, 0, 1, 1, 2, 3, 0, 1, 0, 1, 2, 3]
    
    # Create DataFrame
    df_trends = pd.DataFrame({
        'Date': dates,
        'Correct': correct_counts,
        'Incorrect': incorrect_counts
    })
    
    # Create Plotly figure
    fig = go.Figure()
    
    fig.add_trace(go.Scatter(
        x=df_trends['Date'],
        y=df_trends['Correct'],
        mode='lines',
        name='Correct',
        line=dict(color='#4caf50', width=3)
    ))
    
    fig.add_trace(go.Scatter(
        x=df_trends['Date'],
        y=df_trends['Incorrect'],
        mode='lines',
        name='Incorrect',
        line=dict(color='#f44336', width=3)
    ))
    
    fig.update_layout(
        title='Daily Verification Results',
        xaxis_title='Date',
        yaxis_title='Number of Verifications',
        legend_title='Status',
        height=400
    )
    
    st.plotly_chart(fig, use_container_width=True)
    
    # Discrepancy types distribution
    st.markdown('<p class="sub-header">Discrepancy Types Distribution</p>', unsafe_allow_html=True)
    
    # Mock data for discrepancy types
    discrepancy_types = [
        "Incorrect Product Type", 
        "Missing Product", 
        "Unexpected Product", 
        "Incorrect Position", 
        "Incorrect Quantity", 
        "Incorrect Orientation", 
        "Label Not Visible"
    ]
    discrepancy_counts = [45, 32, 18, 12, 8, 5, 3]
    
    # Create Plotly figure
    fig = go.Figure(data=[
        go.Bar(
            x=discrepancy_types,
            y=discrepancy_counts,
            marker_color=['#f44336', '#ff9800', '#ffc107', '#4caf50', '#2196f3', '#9c27b0', '#795548']
        )
    ])
    
    fig.update_layout(
        title='Discrepancy Types Distribution',
        xaxis_title='Discrepancy Type',
        yaxis_title='Count',
        height=400
    )
    
    st.plotly_chart(fig, use_container_width=True)
    
    # Vending machine performance
    st.markdown('<p class="sub-header">Vending Machine Performance</p>', unsafe_allow_html=True)
    
    # Mock data for vending machine performance
    vm_ids = ["VM-3245", "VM-3246", "VM-3247", "VM-3248", "VM-3249", "VM-3250", "VM-3251", "VM-3252"]
    verification_counts = [28, 25, 22, 18, 20, 15, 12, 12]
    accuracy_rates = [92.8, 88.0, 86.4, 94.4, 90.0, 93.3, 91.7, 83.3]
    
    # Create DataFrame
    df_vm = pd.DataFrame({
        'VM ID': vm_ids,
        'Verifications': verification_counts,
        'Accuracy': accuracy_rates
    })
    
    # Sort by accuracy
    df_vm = df_vm.sort_values('Accuracy', ascending=False)
    
    # Create Plotly figure
    fig = go.Figure()
    
    fig.add_trace(go.Bar(
        x=df_vm['VM ID'],
        y=df_vm['Accuracy'],
        marker_color='#2196f3',
        name='Accuracy Rate (%)'
    ))
    
    fig.add_trace(go.Scatter(
        x=df_vm['VM ID'],
        y=df_vm['Verifications'],
        mode='markers',
        marker=dict(size=12, color='#f44336'),
        name='Verification Count',
        yaxis='y2'
    ))
    
    fig.update_layout(
        title='Vending Machine Performance',
        xaxis_title='Vending Machine ID',
        yaxis_title='Accuracy Rate (%)',
        yaxis2=dict(
            title='Verification Count',
            overlaying='y',
            side='right'
        ),
        legend=dict(
            x=0.01,
            y=0.99,
            bgcolor='rgba(255, 255, 255, 0.8)'
        ),
        height=400
    )
    
    st.plotly_chart(fig, use_container_width=True)

elif page == "Settings":
    st.markdown('<p class="main-header">Settings</p>', unsafe_allow_html=True)
    
    # General settings
    st.markdown('<p class="sub-header">General Settings</p>', unsafe_allow_html=True)
    
    col1, col2 = st.columns(2)
    
    with col1:
        st.text_input("API URL", value=API_URL)
    
    with col2:
        st.selectbox("Theme", ["Light", "Dark", "System Default"], index=0)
    
    # Notification settings
    st.markdown('<p class="sub-header">Notification Settings</p>', unsafe_allow_html=True)
    
    st.checkbox("Enable Email Notifications", value=True)
    st.checkbox("Enable Slack Notifications", value=False)
    
    email = st.text_input("Notification Email")
    webhook = st.text_input("Slack Webhook URL")
    
    # Thresholds
    st.markdown('<p class="sub-header">Alert Thresholds</p>', unsafe_allow_html=True)
    
    col1, col2, col3 = st.columns(3)
    
    with col1:
        st.number_input("Minimum Accuracy (%)", min_value=0, max_value=100, value=90)
    
    with col2:
        st.number_input("High Severity Threshold", min_value=0, value=1, 
                       help="Number of high severity discrepancies to trigger alert")
    
    with col3:
        st.number_input("Total Discrepancy Threshold", min_value=0, value=3,
                       help="Total number of discrepancies to trigger alert")
    
    # Save button
    if st.button("Save Settings"):
        st.success("Settings saved successfully!")

# Initialize session state
if 'history_loaded' not in st.session_state:
    st.session_state.history_loaded = False

if 'current_verification' not in st.session_state:
    st.session_state.current_verification = None

if 'verification_status' not in st.session_state:
    st.session_state.verification_status = None