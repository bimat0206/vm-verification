import streamlit as st
import sys
import os

# Add parent directory to path to access clients module
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from clients.config_loader import ConfigLoader

st.title("Configuration Debug")

# Test config loader
config_loader = ConfigLoader()

st.subheader("Configuration Values")
config = config_loader.get_all()
for key, value in config.items():
    if 'KEY' in key.upper():
        st.write(f"**{key}**: {'*' * len(value) if value else 'Not set'}")
    else:
        st.write(f"**{key}**: {value if value else 'Not set'}")

st.subheader("Configuration Source")
if config_loader.is_loaded_from_secret():
    st.success("✅ Loaded from AWS Secrets Manager")
elif config_loader.is_loaded_from_streamlit():
    st.success("✅ Loaded from Streamlit Secrets")
else:
    st.warning("⚠️ Loaded from Environment Variables")

st.subheader("Streamlit Secrets")
if hasattr(st, 'secrets') and st.secrets:
    st.write("Available secrets:")
    for key in st.secrets.keys():
        if 'KEY' in key.upper():
            st.write(f"- {key}: {'*' * len(st.secrets[key]) if st.secrets[key] else 'Not set'}")
        else:
            st.write(f"- {key}: {st.secrets[key] if st.secrets[key] else 'Not set'}")
else:
    st.error("❌ No Streamlit secrets available")
