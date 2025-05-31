import streamlit as st
import logging
import requests

logger = logging.getLogger(__name__)

def app(api_client):
    st.title("Health Check")

    # Display configuration status
    try:
        # Get configuration from the API client's config loader
        config_loader = api_client.config_loader
        api_endpoint = config_loader.get("API_ENDPOINT", "")

        if config_loader.is_loaded_from_secret():
            st.success(f"✅ Configuration loaded from AWS Secrets Manager. API Endpoint: {api_endpoint[:30]}...")
            config_source = "AWS Secrets Manager"
        elif config_loader.is_loaded_from_streamlit():
            st.info(f"🔧 Configuration loaded from Streamlit secrets (local development). API Endpoint: {api_endpoint[:30]}...")
            config_source = "Streamlit Secrets (Local Development)"
        else:
            st.info(f"ℹ️ Configuration loaded from environment variables. API Endpoint: {api_endpoint[:30]}...")
            config_source = "Environment Variables"

        # Show configuration source
        st.write(f"**Configuration Source:** {config_source}")

    except Exception as e:
        st.error(f"❌ Error loading configuration: {str(e)}")

    # Add debug mode checkbox
    debug_mode = st.checkbox("Debug Mode")

    # Direct API call without using the client
    if st.button("Direct API Call"):
        try:
            # Get values from session state or API client configuration
            config_loader = api_client.config_loader
            api_endpoint = st.session_state.get('api_endpoint', config_loader.get("API_ENDPOINT", ""))
            api_key = st.session_state.get('api_key', api_client.api_key)

            if not api_endpoint or not api_key:
                st.error("API endpoint or API key not found.")
                return

            # Display the values being used (masked for security)
            st.write(f"Using endpoint: {api_endpoint}")
            st.write(f"Using API key: {'*' * len(api_key)}")

            # Make direct request
            st.info("Making direct request...")

            # Use the standard API v1 health endpoint with proper header case
            url = f"{api_endpoint.rstrip('/')}/api/health"
            headers = {
                'X-Api-Key': api_key,  # Using capital X to match API Gateway expectations
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            }

            st.write(f"Requesting: {url}")

            try:
                response = requests.get(url, headers=headers, timeout=10)
                st.write(f"Status code: {response.status_code}")

                # Display headers for debugging
                if debug_mode:
                    st.json(dict(response.headers))

                # Raise for bad status
                response.raise_for_status()

                # Display result
                result = response.json()
                st.success("Health check successful")
                st.json(result)

                # Update the working URL in session state
                st.session_state['working_health_url'] = url

            except requests.exceptions.HTTPError as e:
                st.error(f"HTTP Error: {str(e)}")
                if response.status_code == 403:
                    st.error("Possible API key authentication issue. Please verify your API key.")

                # Try to display response body even if status code is error
                try:
                    st.json(response.json())
                except:
                    st.text(response.text)

            except Exception as e:
                st.error(f"Request error: {str(e)}")

        except Exception as e:
            logger.error(f"Direct API call failed: {str(e)}")
            st.error(f"Direct API call error: {str(e)}")

    # Regular health check using API client
    if st.button("Check System Health"):
        try:
            # First try with debug mode if enabled
            if debug_mode:
                debug_info = api_client.health_check(debug=True)
                st.subheader("Debug Information")
                st.json(debug_info)

            # Then try actual health check
            health_status = api_client.health_check()
            st.write("System Health:")

            # Display overall status with color
            status = health_status.get('status', 'unknown')
            if status == 'healthy':
                st.success(f"Overall Status: {status}")
            elif status == 'degraded':
                st.warning(f"Overall Status: {status}")
            else:
                st.error(f"Overall Status: {status}")

            # Display service details in expandable sections
            services = health_status.get('services', {})
            if services:
                st.subheader("Services Status")
                for service_name, service_info in services.items():
                    service_status = service_info.get('status', 'unknown')
                    col1, col2, col3 = st.columns([2, 1, 1])
                    with col1:
                        st.write(f"**{service_name.capitalize()}**")
                    with col2:
                        if service_status == 'healthy':
                            st.success(service_status)
                        elif service_status == 'degraded':
                            st.warning(service_status)
                        else:
                            st.error(service_status)
                    with col3:
                        if 'latency' in service_info:
                            st.write(f"{service_info['latency']} ms")

            # Display version and timestamp
            st.text(f"Version: {health_status.get('version', 'unknown')}")
            st.text(f"Timestamp: {health_status.get('timestamp', 'unknown')}")

        except Exception as e:
            logger.error(f"Health check failed: {str(e)}")
            st.error(f"Health check error: {str(e)}")

    # Form to update API settings
    with st.expander("Update API Settings"):
        with st.form("api_settings"):
            config_loader = api_client.config_loader
            current_endpoint = st.session_state.get('api_endpoint', config_loader.get("API_ENDPOINT", ""))
            current_key = st.session_state.get('api_key', api_client.api_key)

            new_endpoint = st.text_input("API Endpoint", value=current_endpoint)
            new_key = st.text_input("API Key", value=current_key, type="password")

            # Add submit button
            submitted = st.form_submit_button("Update Settings")
            if submitted:
                st.session_state['api_endpoint'] = new_endpoint
                st.session_state['api_key'] = new_key
                st.success("Settings updated!")
