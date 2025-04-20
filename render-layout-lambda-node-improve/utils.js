/**
 * utils.js - Utility functions for the vending machine layout generator
 * 
 * Contains event parsing and text processing utilities.
 */

const { log } = require('./common');

/**
 * Parse the Lambda event to extract S3 bucket and key information
 * Supports both EventBridge and S3 event formats
 * 
 * @param {Object} event - Lambda event object
 * @returns {Object} - Parsed bucket and key or skipped status
 */
function parseEvent(event) {
  try {
    if (!event) {
      log('Event is undefined');
      throw new Error('Event is undefined');
    }

    log('Parsing event for bucket/key', event);
    let bucket, key;

    if (event.detail && event.detail.bucket && event.detail.object) {
      // EventBridge format (primary format for this application)
      bucket = event.detail.bucket.name;
      key = event.detail.object.key;
      log('Detected EventBridge S3 event structure', { bucket, key });
    } else if (event.Records && event.Records[0]) {
      // S3 notification format (fallback)
      bucket = event.Records[0].s3.bucket.name;
      key = decodeURIComponent(event.Records[0].s3.object.key.replace(/\+/g, ' '));
      log('Detected S3 Put event structure', { bucket, key });
    } else {
      // Handle direct invocation with bucket/key parameters
      if (event.bucket && event.key) {
        bucket = event.bucket;
        key = event.key;
        log('Detected direct invocation with bucket and key parameters', { bucket, key });
      } else {
        log('Event does not contain S3 object info', event);
        throw new Error('Event does not contain S3 object info');
      }
    }

    // Only process raw JSON files (but allow for testing bypasses with a force flag)
    if ((!key.startsWith('raw/') || !key.endsWith('.json')) && !event.force) {
      log('Not a raw JSON file, skipping.', { key });
      return { status: 'skipped', reason: 'Not a raw JSON file' };
    }

    return { bucket, key };
  } catch (err) {
    log('Failed to parse event:', err.message, err.stack);
    throw err;
  }
}

/**
 * Split text into multiple lines to fit within a maximum width
 * 
 * @param {CanvasRenderingContext2D} ctx - Canvas context for text measurement
 * @param {string} text - Text to split
 * @param {number} maxWidth - Maximum line width
 * @returns {string[]} - Array of text lines
 */
function splitTextToLines(ctx, text, maxWidth) {
  if (!text) return [''];
  if (ctx.measureText(text).width <= maxWidth) return [text];
  
  const words = text.split(' ');
  if (words.length === 0) return [''];
  
  const lines = [];
  let currentLine = words[0];
  
  for (let i = 1; i < words.length; i++) {
    const word = words[i];
    const testLine = `${currentLine} ${word}`;
    
    if (ctx.measureText(testLine).width <= maxWidth) {
      currentLine = testLine;
    } else {
      lines.push(currentLine);
      currentLine = word;
    }
  }
  
  if (currentLine) lines.push(currentLine);
  
  // Limit to 2 lines with ellipsis if needed
  if (lines.length > 2) {
    lines[1] = `${lines[1].slice(0, -3)}...`;
    return lines.slice(0, 2);
  }
  
  return lines;
}

module.exports = {
  parseEvent,
  splitTextToLines
};