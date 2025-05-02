"""
Utility Functions for Vending Machine Verification System

This module provides common utility functions used throughout the application.
Enhanced with better formatting and UI helpers.
"""

import os
import logging
from datetime import datetime
from typing import Optional, Dict, Any, List

def setup_logging() -> logging.Logger:
    """
    Configure logging for the application.
    
    Returns:
        logging.Logger: Configured logger instance
    """
    # Configure logging
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    logger = logging.getLogger("streamlit-app")
    
    # Add file handler if in production environment
    if os.environ.get('ENVIRONMENT') == 'production':
        log_dir = os.environ.get('LOG_DIR', '/tmp/logs')
        os.makedirs(log_dir, exist_ok=True)
        
        file_handler = logging.FileHandler(
            f"{log_dir}/streamlit_{datetime.now().strftime('%Y%m%d')}.log"
        )
        file_handler.setFormatter(logging.Formatter(
            '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        ))
        logger.addHandler(file_handler)
    
    return logger

def format_timestamp(timestamp_str: str) -> str:
    """
    Format ISO timestamp string to human-readable format.
    
    Args:
        timestamp_str: ISO format timestamp string
        
    Returns:
        str: Human-readable timestamp
    """
    try:
        dt = datetime.fromisoformat(timestamp_str.replace('Z', '+00:00'))
        return dt.strftime("%Y-%m-%d %H:%M:%S")
    except Exception:
        return timestamp_str

def parse_s3_url(s3_url: str) -> Dict[str, str]:
    """
    Parse S3 URL into components.
    
    Args:
        s3_url: S3 URL in format s3://bucket/key
        
    Returns:
        Dict with bucket and key
    """
    if not s3_url:
        return {"error": "Empty S3 URL"}
        
    if not s3_url.startswith('s3://'):
        return {"error": "Invalid S3 URL format"}
    
    try:
        # Remove s3:// prefix
        path = s3_url[5:]
        
        # Split into bucket and key
        parts = path.split('/', 1)
        if len(parts) < 2:
            return {"bucket": parts[0], "key": ""}
        
        return {"bucket": parts[0], "key": parts[1]}
    except Exception as e:
        return {"error": str(e)}

def format_file_size(size_bytes: int) -> str:
    """
    Format file size in bytes to human-readable format.
    
    Args:
        size_bytes: Size in bytes
        
    Returns:
        str: Human-readable size
    """
    if not isinstance(size_bytes, (int, float)):
        return "0 B"
        
    for unit in ['B', 'KB', 'MB', 'GB', 'TB']:
        if size_bytes < 1024 or unit == 'TB':
            return f"{size_bytes:.2f} {unit}" if unit != 'B' else f"{size_bytes} {unit}"
        size_bytes /= 1024

def calculate_accuracy_percentage(correct: int, total: int) -> float:
    """
    Calculate accuracy percentage.
    
    Args:
        correct: Number of correct items
        total: Total number of items
        
    Returns:
        float: Accuracy percentage
    """
    if total == 0:
        return 0.0
    return (correct / total) * 100

def generate_color_for_status(status: str) -> str:
    """
    Generate color hex code based on verification status.
    
    Args:
        status: Verification status string
        
    Returns:
        str: Hex color code
    """
    status_colors = {
        "CORRECT": "#28a745",  # Green
        "INCORRECT": "#dc3545",  # Red
        "PROCESSING": "#ffc107",  # Yellow
        "FAILED": "#6c757d",  # Gray
        "INITIALIZED": "#17a2b8"  # Blue
    }
    
    return status_colors.get(status, "#6c757d")  # Default to gray

def group_discrepancies_by_row(discrepancies: List[Dict[str, Any]]) -> Dict[str, List[Dict[str, Any]]]:
    """
    Group discrepancies by row for better display.
    
    Args:
        discrepancies: List of discrepancy objects
        
    Returns:
        Dict mapping row labels to lists of discrepancies
    """
    grouped = {}
    
    for disc in discrepancies:
        position = disc.get("position", "")
        if position and len(position) >= 1:
            row = position[0]  # Assume first character is row (A, B, C, etc.)
            if row not in grouped:
                grouped[row] = []
            grouped[row].append(disc)
    
    return grouped

def safe_get(data: Dict[str, Any], *keys, default=None) -> Any:
    """
    Safely get a value from a nested dictionary.
    
    Args:
        data: Dictionary to extract value from
        *keys: Keys to traverse in the dictionary
        default: Default value if path doesn't exist
        
    Returns:
        Value at the specified path or default
    """
    if data is None:
        return default
        
    for key in keys:
        if isinstance(data, dict) and key in data:
            data = data[key]
        else:
            return default
    return data

def get_verification_type_options() -> List[Dict[str, Any]]:
    """
    Get verification type options with details for UI.
    
    Returns:
        List of verification type options
    """
    return [
        {
            "key": "LAYOUT_VS_CHECKING",
            "display": "Layout vs Checking",
            "description": "Compare against reference layout",
            "icon": "📋",
            "reference_label": "Reference Layout Image",
            "checking_label": "Current Checking Image"
        },
        {
            "key": "PREVIOUS_VS_CURRENT",
            "display": "Previous vs Current",
            "description": "Compare against previous state",
            "icon": "🔄",
            "reference_label": "Previous Checking Image",
            "checking_label": "Current Checking Image"
        }
    ]

def create_image_card_html(image: Dict[str, Any], image_url: str, index: int) -> str:
    """
    Create HTML for an image card.
    
    Args:
        image: Image metadata
        image_url: URL to the image
        index: Index for button IDs
        
    Returns:
        str: HTML code for image card
    """
    return f"""
    <div style="border: 1px solid #ddd; border-radius: 8px; padding: 10px; margin-bottom: 15px;">
        <img src="{image_url}" style="width: 100%; border-radius: 4px;">
        <h4 style="margin-top: 10px; margin-bottom: 5px;">{image.get('name', 'Image')}</h4>
        <p style="color: #666; font-size: 0.8em; margin-bottom: 5px;">
            Size: {format_file_size(image.get('size', 0))}<br>
            Modified: {format_timestamp(image.get('lastModified', ''))}
        </p>
    </div>
    """