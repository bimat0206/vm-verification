import os
import io
import json
import logging
import boto3
import requests
from flask import Flask, request, jsonify, send_file
from PIL import Image, ImageDraw, ImageFont
from datetime import datetime

from utils import render_layout, write_file_local

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s %(levelname)s %(message)s'
)
logger = logging.getLogger(__name__)

# AWS S3 client
s3_client = boto3.client('s3', region_name=os.getenv('AWS_REGION', 'us-east-1'))

# Font registration with expanded paths
FONT_PATHS = [
    '/app/fonts/arial.ttf',
    os.path.join(os.path.dirname(__file__), 'fonts', 'arial.ttf'),
    '/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf',
    '/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf',
    '/usr/share/fonts/truetype/freefont/FreeSans.ttf',
    '/var/task/fonts/arial.ttf',  # Lambda layer path
    '/opt/fonts/arial.ttf',       # Lambda layer alternative path
    '/System/Library/Fonts/Helvetica.ttc',  # macOS
    'C:/Windows/Fonts/arial.ttf'  # Windows
]

def get_font(size=20):
    """Get font of specified size from available system fonts with improved fallbacks"""
    # Try the predefined paths first
    for path in FONT_PATHS:
        try:
            if os.path.exists(path):
                logger.info(f"Found font at path: {path}")
                font = ImageFont.truetype(path, size)
                logger.info(f"Successfully loaded font from {path}")
                return font
        except Exception as e:
            logger.info(f"Font not found at {path}: {e}")
    
    # Try using system font by name with additional fallbacks
    font_names = ["Arial", "Helvetica", "DejaVuSans", "FreeSans", "LiberationSans"]
    for font_name in font_names:
        try:
            logger.info(f"Attempting to load system font by name: {font_name}")
            font = ImageFont.truetype(font_name, size)
            logger.info(f"Successfully loaded {font_name} font")
            return font
        except Exception as e:
            logger.info(f"Could not load {font_name} font: {e}")
    
    # Last resort: use default font
    logger.warning("Falling back to default font")
    return ImageFont.load_default()

# Create Flask app for local testing
app = Flask(__name__)

@app.route('/process', methods=['POST'])
def process_layout():
    """
    Endpoint to process layout JSON from S3 and generate PNG
    
    Expects JSON:
    {
        "bucket": "bucket-name",
        "key": "raw/layout.json"
    }
    """
    try:
        data = request.get_json()
        bucket = data.get('bucket')
        key = data.get('key')

        if not bucket or not key:
            logger.error("Missing bucket or key in request")
            return jsonify({'error': 'Missing bucket or key'}), 400
            
        if not key.startswith('raw/') or not key.endswith('.json'):
            logger.info(f"Not a raw JSON file, skipping: {key}")
            return jsonify({
                'status': 'skipped',
                'reason': 'Not a raw JSON file'
            })

        # Download JSON from S3
        try:
            logger.info(f"Downloading JSON from S3: bucket={bucket}, key={key}")
            obj = s3_client.get_object(Bucket=bucket, Key=key)
            layout_json = obj['Body'].read()
            layout = json.loads(layout_json)
            layout_id = layout.get('layoutId', 'unknown')
            logger.info(f"Parsed layout JSON: layoutId={layout_id}")
        except Exception as e:
            logger.error(f"Error downloading or parsing layout JSON: {e}")
            return jsonify({'error': str(e)}), 500

        # Render the layout to PNG
        try:
            logger.info(f"Rendering layout {layout_id} to PNG")
            png_buffer = render_layout(layout, get_font)
            logger.info(f"Rendered PNG buffer: size={len(png_buffer)} bytes")
        except Exception as e:
            logger.error(f"Error rendering layout: {e}")
            return jsonify({'error': str(e)}), 500

        # Upload PNG to S3
        output_key = key.replace('raw/', 'processed/').replace('.json', '.png')
        try:
            logger.info(f"Uploading PNG to S3: bucket={bucket}, key={output_key}")
            s3_client.put_object(
                Bucket=bucket,
                Key=output_key,
                Body=png_buffer,
                ContentType='image/png'
            )
            logger.info(f"Upload successful to {output_key}")
        except Exception as e:
            logger.error(f"Error uploading PNG to S3: {e}")
            return jsonify({'error': str(e)}), 500

        # Also save locally if in development environment
        if os.getenv('FLASK_ENV') == 'development':
            local_path = os.path.join('output', os.path.basename(output_key))
            os.makedirs(os.path.dirname(local_path), exist_ok=True)
            write_file_local(local_path, png_buffer)
            logger.info(f"Saved local copy to {local_path}")

        return jsonify({
            'status': 'success',
            'outputKey': output_key,
            'layoutId': layout_id
        })

    except Exception as e:
        logger.error(f"Unexpected error in process_layout: {e}")
        return jsonify({'error': str(e)}), 500

@app.route('/health', methods=['GET'])
def health_check():
    """Simple health check endpoint"""
    return jsonify({'status': 'ok', 'timestamp': datetime.now().isoformat()})

@app.route('/preview', methods=['GET'])
def preview():
    """Endpoint to preview the latest processed layout"""
    try:
        preview_dir = 'output'
        if not os.path.exists(preview_dir):
            return jsonify({'error': 'No preview available'}), 404
            
        # Find the latest PNG file
        files = [f for f in os.listdir(preview_dir) if f.endswith('.png')]
        if not files:
            return jsonify({'error': 'No preview available'}), 404
            
        latest_file = max(files, key=lambda f: os.path.getmtime(os.path.join(preview_dir, f)))
        file_path = os.path.join(preview_dir, latest_file)
        
        return send_file(file_path, mimetype='image/png')
    except Exception as e:
        logger.error(f"Error serving preview: {e}")
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    port = int(os.getenv('PORT', 5000))
    debug = os.getenv('FLASK_ENV') == 'development'
    
    # Ensure output directory exists
    if not os.path.exists('output'):
        os.makedirs('output', exist_ok=True)
        
    app.run(host='0.0.0.0', port=port, debug=debug)