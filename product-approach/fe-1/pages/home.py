import streamlit as st

def apply_home_css():
    """Apply custom CSS for Home page styling"""
    st.markdown("""
    <style>
    .home-card {
        background-color: #262730;
        border: 1px solid #444;
        border-radius: 8px;
        padding: 1.5rem;
        margin-bottom: 1.5rem;
        transition: all 0.2s ease-in-out;
    }
    .home-card:hover {
        border-color: #6366F1;
        box-shadow: 0 0 10px rgba(99, 102, 241, 0.3);
    }

    .feature-card {
        background-color: #31333F;
        border: 1px solid #4A4C5A;
        border-radius: 6px;
        padding: 1rem;
        margin-bottom: 1rem;
        text-align: center;
    }

    .feature-icon {
        font-size: 2rem;
        margin-bottom: 0.5rem;
        display: block;
    }

    .feature-title {
        font-size: 1.1rem !important;
        font-weight: 600 !important;
        color: #FAFAFA !important;
        margin-bottom: 0.5rem !important;
    }

    .feature-description {
        font-size: 0.9rem !important;
        color: #A0A0A0 !important;
        line-height: 1.4;
    }

    .navigation-hint {
        background-color: #1E3A8A;
        border: 1px solid #3B82F6;
        border-radius: 6px;
        padding: 1rem;
        margin: 1rem 0;
        color: #93C5FD;
    }

    .welcome-header {
        text-align: center;
        margin-bottom: 2rem;
    }

    .welcome-title {
        font-size: 2.5rem !important;
        font-weight: 700 !important;
        color: #FAFAFA !important;
        margin-bottom: 0.5rem !important;
    }

    .welcome-subtitle {
        font-size: 1.2rem !important;
        color: #A0A0A0 !important;
        margin-bottom: 1rem !important;
    }
    </style>
    """, unsafe_allow_html=True)

def app(api_client):
    # Apply CSS styling
    apply_home_css()

    # Welcome header
    st.markdown("""
    <div class="welcome-header">
        <div class="welcome-title">🏠 Welcome to the Vending Machine Verification System</div>
        <div class="welcome-subtitle">Your comprehensive solution for automated vending machine verification and analysis</div>
    </div>
    """, unsafe_allow_html=True)

    # Introduction section
    st.markdown('<div class="home-card">', unsafe_allow_html=True)
    st.markdown("## 🎯 About This System")
    st.markdown("""
    The Vending Machine Verification System is an advanced AI-powered platform designed to automate
    the verification and analysis of vending machine layouts and inventory. Using cutting-edge computer
    vision and machine learning technologies, our system provides accurate, reliable verification results
    to help maintain optimal vending machine operations.

    **Key Benefits:**
    - **Automated Analysis**: Reduce manual inspection time and human error
    - **Real-time Results**: Get instant verification feedback with detailed analysis
    - **Comprehensive Reporting**: Access detailed metrics and performance data
    - **Easy Integration**: Simple interface for seamless workflow integration
    """)
    st.markdown('</div>', unsafe_allow_html=True)

    # Features overview
    st.markdown("## ✨ System Features")

    col1, col2, col3 = st.columns(3)

    with col1:
        st.markdown("""
        <div class="feature-card">
            <div class="feature-icon">🔍</div>
            <div class="feature-title">Verification System</div>
            <div class="feature-description">
                Start new verifications by comparing reference and checking images.
                Support for multiple verification types including layout comparison
                and temporal analysis.
            </div>
        </div>
        """, unsafe_allow_html=True)

    with col2:
        st.markdown("""
        <div class="feature-card">
            <div class="feature-icon">📊</div>
            <div class="feature-title">Results Analysis</div>
            <div class="feature-description">
                Browse and analyze verification results with advanced filtering,
                detailed metrics, and comprehensive reporting capabilities.
            </div>
        </div>
        """, unsafe_allow_html=True)

    with col3:
        st.markdown("""
        <div class="feature-card">
            <div class="feature-icon">🔧</div>
            <div class="feature-title">Management Tools</div>
            <div class="feature-description">
                Access image upload tools, health monitoring, and system
                diagnostics to maintain optimal performance.
            </div>
        </div>
        """, unsafe_allow_html=True)

    # Getting started guide
    st.markdown('<div class="home-card">', unsafe_allow_html=True)
    st.markdown("## 🚀 Getting Started")
    st.markdown("""
    ### Step 1: Start a Verification
    Navigate to the **Verification System** to begin a new verification process:
    1. Select your verification type (Layout vs Checking or Previous vs Current)
    2. Choose reference and checking images from S3 buckets
    3. Configure verification parameters
    4. Submit for analysis

    ### Step 2: Monitor Progress
    Track your verification progress in real-time:
    - View processing status and estimated completion time
    - Access preliminary results as they become available
    - Receive notifications when analysis is complete

    ### Step 3: Review Results
    Analyze detailed verification results in the **Verification Results** section:
    - Browse all completed verifications with advanced filtering
    - View detailed accuracy metrics and confidence scores
    - Access comprehensive LLM analysis and recommendations
    - Export results for reporting and documentation
    """)
    st.markdown('</div>', unsafe_allow_html=True)

    # Navigation guidance
    st.markdown("""
    <div class="navigation-hint">
        <strong>💡 Quick Navigation Tips:</strong><br>
        • Use the sidebar menu to navigate between different sections<br>
        • Start with <strong>Verification System</strong> to run new verifications<br>
        • Check <strong>Verification Results</strong> to review completed analyses<br>
        • Access <strong>Tools</strong> for image management and system diagnostics
    </div>
    """, unsafe_allow_html=True)

    # System status and information
    st.markdown('<div class="home-card">', unsafe_allow_html=True)
    st.markdown("## 📋 System Information")

    col1, col2 = st.columns(2)

    with col1:
        st.markdown("### 🔗 System Status")
        if api_client:
            try:
                # Test API connection
                test_response = api_client.list_verifications({'limit': 1})
                total_verifications = test_response.get('pagination', {}).get('total', 0)
                st.success(f"✅ System Online - {total_verifications} total verifications")
            except Exception as e:
                st.error(f"❌ System Error: {str(e)}")
        else:
            st.warning("⚠️ API client not available")

    with col2:
        st.markdown("### 📈 Quick Stats")
        if api_client:
            try:
                # Get recent verification stats
                recent_response = api_client.list_verifications({'limit': 10})
                recent_count = len(recent_response.get('results', []))
                st.info(f"📊 {recent_count} recent verifications available")
            except Exception:
                st.info("📊 Statistics temporarily unavailable")
        else:
            st.info("📊 Connect to view statistics")

    st.markdown('</div>', unsafe_allow_html=True)

    # Help and support
    with st.expander("❓ Help & Support", expanded=False):
        st.markdown("""
        ### Frequently Asked Questions

        **Q: What image formats are supported?**
        A: The system supports common image formats including PNG, JPG, and JPEG files stored in S3 buckets.

        **Q: How long does verification take?**
        A: Processing time varies based on image complexity, typically ranging from 30 seconds to 5 minutes.

        **Q: Can I verify multiple machines at once?**
        A: Currently, the system processes one verification at a time. You can queue multiple verifications sequentially.

        **Q: How accurate are the results?**
        A: Our AI models achieve high accuracy rates, with detailed confidence scores provided for each verification.

        ### Troubleshooting
        - Ensure images are properly uploaded to the correct S3 buckets
        - Check that verification parameters are correctly configured
        - Monitor system status for any service interruptions
        - Contact support if issues persist
        """)

    st.markdown("---")
    st.markdown("""
    <div style="text-align: center; color: #A0A0A0; font-size: 0.9rem;">
        <strong>Vending Machine Verification System</strong> | Powered by Advanced AI Technology
    </div>
    """, unsafe_allow_html=True)