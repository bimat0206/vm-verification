import json
import logging
import os
import sys
import traceback
from utils import render_layout
import boto3

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s %(levelname)s %(message)s'
)
logger = logging.getLogger(__name__)

# AWS S3 client
s3_client = boto3.client('s3', region_name=os.getenv('AWS_REGION', 'us-east-1'))

# Import font functions but not the Flask app itself
from app import get_font

def lambda_handler(event, context):
    """
    AWS Lambda handler function to process S3 events
    """
    logger.info(f"Lambda handler started")
    logger.info(f"Received event: {json.dumps(event)}")
    
    try:
        # Extract S3 bucket and key information from different event types
        bucket = None
        key = None
        
        # EventBridge S3 event (newer format)
        if event.get('detail-type') == 'Object Created' and event.get('source') == 'aws.s3':
            bucket = event['detail']['bucket']['name']
            key = event['detail']['object']['key']
            logger.info(f"Processing EventBridge S3 event: bucket={bucket}, key={key}")
        
        # S3 event notification (older format)
        elif event.get('Records') and len(event['Records']) > 0:
            record = event['Records'][0]
            if record.get('eventSource') == 'aws:s3' or record.get('s3'):
                s3_info = record.get('s3', {})
                bucket = s3_info.get('bucket', {}).get('name')
                key = s3_info.get('object', {}).get('key')
                logger.info(f"Processing S3 notification event: bucket={bucket}, key={key}")
        
        # Direct invocation with bucket/key in the event
        else:
            bucket = event.get('bucket')
            key = event.get('key')
            logger.info(f"Processing direct invocation: bucket={bucket}, key={key}")
        
        if not bucket or not key:
            logger.error("Missing bucket or key in event")
            return {
                'statusCode': 400,
                'body': json.dumps({'error': 'Missing bucket or key', 'event': event})
            }
        
        # Validate the key follows the expected pattern
        if not key.startswith('raw/') or not key.endswith('.json'):
            logger.info(f"Not a raw JSON file, skipping: {key}")
            return {
                'statusCode': 200,
                'body': json.dumps({
                    'status': 'skipped',
                    'reason': 'Not a raw JSON file',
                    'key': key
                })
            }
            
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
            stack_trace = traceback.format_exc()
            logger.error(f"Stack trace: {stack_trace}")
            return {
                'statusCode': 500,
                'body': json.dumps({
                    'error': str(e),
                    'bucket': bucket,
                    'key': key,
                    'stackTrace': stack_trace
                })
            }
            
        # Render the layout to PNG
        try:
            logger.info(f"Rendering layout {layout_id} to PNG")
            png_buffer = render_layout(layout, get_font)
            logger.info(f"Rendered PNG buffer: size={len(png_buffer)} bytes")
        except Exception as e:
            logger.error(f"Error rendering layout: {e}")
            stack_trace = traceback.format_exc()
            logger.error(f"Stack trace: {stack_trace}")
            return {
                'statusCode': 500,
                'body': json.dumps({
                    'error': str(e),
                    'layoutId': layout_id,
                    'stackTrace': stack_trace
                })
            }
            
        # Upload PNG to S3 - one with a standardized name and one preserving the path structure
        try:
            # Main output with standardized name based on layout ID
            output_key = f"rendered-layout/{layout_id}.png"
            logger.info(f"Uploading PNG to S3: bucket={bucket}, key={output_key}")
            
            s3_client.put_object(
                Bucket=bucket,
                Key=output_key,
                Body=png_buffer,
                ContentType='image/png',
                Metadata={
                    'layoutId': str(layout_id),
                    'sourceFile': key,
                    'generatedBy': 'layout-renderer-lambda'
                }
            )
            logger.info(f"Successfully uploaded PNG to {output_key}")
            
            # Also save a copy with the original file path structure
            processed_key = key.replace('raw/', 'processed/').replace('.json', '.png')
            if processed_key != output_key:
                logger.info(f"Saving additional copy to {processed_key}")
                s3_client.put_object(
                    Bucket=bucket,
                    Key=processed_key,
                    Body=png_buffer,
                    ContentType='image/png'
                )
                logger.info(f"Successfully uploaded PNG to {processed_key}")
                
        except Exception as e:
            logger.error(f"Error uploading PNG: {e}")
            stack_trace = traceback.format_exc()
            logger.error(f"Stack trace: {stack_trace}")
            return {
                'statusCode': 500,
                'body': json.dumps({
                    'error': str(e),
                    'layoutId': layout_id,
                    'stackTrace': stack_trace
                })
            }
            
        # Return success response with both output paths
        logger.info("Lambda function completed successfully")
        return {
            'statusCode': 200,
            'body': json.dumps({
                'status': 'success',
                'layoutId': layout_id,
                'output_key': output_key,
                'processed_key': processed_key if 'processed_key' in locals() else None
            })
        }
        
    except Exception as e:
        logger.error(f"Unhandled exception in lambda_handler: {e}")
        stack_trace = traceback.format_exc()
        logger.error(f"Stack trace: {stack_trace}")
        return {
            'statusCode': 500,
            'body': json.dumps({
                'error': str(e),
                'stackTrace': stack_trace,
                'event': event
            })
        }