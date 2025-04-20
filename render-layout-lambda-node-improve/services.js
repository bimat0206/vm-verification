/**
 * services.js - Core services for the vending machine layout generator
 * 
 * Contains S3 operations, logging service, font setup, and image fetching.
 */

const { S3Client, GetObjectCommand, PutObjectCommand } = require('@aws-sdk/client-s3');
const fs = require('fs');
const { ENV, FONTS, NETWORK } = require('./config');
const { log, capturedLogs } = require('./common');

// Initialize variables
let fetch;
let AbortController;
let s3Client;

// Setup fetch safely
try {
  // For Node.js versions with node-fetch v2 compatibility
  const nodeFetch = require('node-fetch');
  fetch = nodeFetch;
  AbortController = nodeFetch.AbortController || global.AbortController;
  console.log('Using node-fetch module');
} catch (err) {
  console.error('Warning: fetch not available:', err.message);
}

// Initialize S3 client safely
try {
  s3Client = new S3Client({ region: ENV.AWS_REGION });
} catch (err) {
  console.error('Error initializing S3 client:', err.message);
}

/**
 * Download and parse JSON from S3
 * 
 * @param {string} bucket - S3 bucket name
 * @param {string} key - S3 object key
 * @returns {Object} - Parsed JSON object
 */
async function downloadJsonFromS3(bucket, key) {
  try {
    if (!s3Client) {
      log('S3 client not initialized');
      throw new Error('S3 client not initialized');
    }
    
    log('Downloading JSON from S3', { bucket, key });
    const response = await s3Client.send(new GetObjectCommand({ Bucket: bucket, Key: key }));
    log('Received S3 object response');

    const streamToBuffer = async (stream) => {
      log('Converting S3 stream to buffer');
      const chunks = [];
      for await (const chunk of stream) {
        chunks.push(chunk);
      }
      const buffer = Buffer.concat(chunks);
      log('Stream converted to buffer', { bufferLength: buffer.length });
      return buffer;
    };

    const contentBuffer = await streamToBuffer(response.Body);
    const layout = JSON.parse(contentBuffer.toString('utf-8'));
    log('Parsed layout JSON', { layoutId: layout.layoutId });
    return layout;
  } catch (err) {
    log('Failed to download or parse JSON from S3:', err.message, err.stack);
    throw err;
  }
}

/**
 * Upload content to S3
 * 
 * @param {string} bucket - S3 bucket name
 * @param {string} key - S3 object key
 * @param {Buffer|string} body - Content to upload
 * @param {string} contentType - Content MIME type
 */
async function uploadToS3(bucket, key, body, contentType) {
  try {
    if (!s3Client) {
      log('S3 client not initialized');
      throw new Error('S3 client not initialized');
    }
    
    log('Uploading to S3', { bucket, key });
    await s3Client.send(
      new PutObjectCommand({
        Bucket: bucket,
        Key: key,
        Body: body,
        ContentType: contentType,
      })
    );
    log(`Uploaded to ${key}`);
  } catch (err) {
    log('Failed to upload to S3:', err.message, err.stack);
    throw err;
  }
}

/**
 * Handle errors, upload logs to S3, and return formatted error response
 * 
 * @param {Error} err - Error object
 * @returns {Object} - Formatted error response
 */
async function handleError(err) {
  log('Error in Lambda handler:', err.message, err.stack);
  
  try {
    const targetBucket = ENV.LOG_BUCKET;
    const fallbackKey = `logs/error_${Date.now()}.log`;
    
    if (targetBucket && s3Client) {
      await uploadToS3(targetBucket, fallbackKey, capturedLogs.join('\n') + `\n${err.stack}`, 'text/plain');
    } else {
      log('No S3 bucket specified for error log upload or S3 client not initialized, skipping');
    }
  } catch (uploadErr) {
    log('Failed to upload error log:', uploadErr.message);
  }
  
  // Return formatted error
  return {
    status: 'error',
    message: err.message,
    timestamp: new Date().toISOString()
  };
}

/**
 * Set up font configuration for canvas
 */
async function setupFonts() {
  try {
    // Set environment variables
    process.env.FONTCONFIG_PATH = FONTS.tempDir;
    process.env.FONTCONFIG_FILE = FONTS.configFile;

    log(`Font directories: ${FONTS.tempDir}, ${FONTS.configFile}`);
    
    // Create font config directory if it doesn't exist
    try {
      if (!fs.existsSync(FONTS.tempDir)) {
        log(`Creating font directory: ${FONTS.tempDir}`);
        fs.mkdirSync(FONTS.tempDir, { recursive: true });
        fs.chmodSync(FONTS.tempDir, '777');
      }
    } catch (dirErr) {
      log(`Error creating font directory: ${dirErr.message}`);
      // Continue even if directory creation fails - Lambda might already have this set up
    }

    // Create font config file if it doesn't exist
    try {
      if (!fs.existsSync(FONTS.configFile)) {
        log(`Creating font config file: ${FONTS.configFile}`);
        fs.writeFileSync(FONTS.configFile, FONTS.configContent.trim());
      }
    } catch (fileErr) {
      log(`Error creating font config file: ${fileErr.message}`);
      // Continue even if file creation fails
    }

    // Check for canvas module without registering fonts yet
    let canvasModule;
    try {
      log('Attempting to load canvas module');
      // Just require it but don't use it yet
      canvasModule = require('canvas');
      log('Canvas module loaded successfully');
    } catch (err) {
      log('Failed to load canvas module:', err.message, err.stack);
      throw new Error(`Canvas module initialization failed: ${err.message}`);
    }
    
    // Only attempt font registration if canvas loaded successfully
    if (canvasModule) {
      const { registerFont } = canvasModule;
      
      let fontRegistered = false;
      for (const fontPath of FONTS.paths) {
        try {
          log(`Checking font path: ${fontPath}`);
          if (fs.existsSync(fontPath)) {
            log(`Registering font from path: ${fontPath}`);
            registerFont(fontPath, { family: 'Arial' });
            log(`Successfully registered font from ${fontPath}`);
            fontRegistered = true;
            break;
          } else {
            log(`Font file not found at ${fontPath}`);
          }
        } catch (fontErr) {
          log(`Failed to register font from ${fontPath}: ${fontErr.message}`);
          // Continue to next font path
        }
      }

      if (!fontRegistered) {
        log('Warning: Could not register Arial font, will use default system fonts');
      }
    }
    
    log('Font setup completed successfully');
  } catch (err) {
    log('Font setup failed:', err.message, err.stack);
    // Don't throw, try to continue without custom fonts
    log('Continuing without custom fonts');
  }
}

/**
 * Fetch an image from a URL with timeout and retries
 * 
 * @param {string} url - Image URL
 * @returns {Promise<Buffer>} - Image buffer or null if failed
 */
async function fetchImage(url) {
  // Check if fetch is available
  if (!fetch) {
    log('Fetch API not available, cannot download images');
    return null;
  }
  
  // Maximum number of retry attempts
  const maxRetries = NETWORK.retryAttempts;
  let lastError = null;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      // If not the first attempt, add a small delay with exponential backoff
      if (attempt > 0) {
        const delay = Math.min(100 * Math.pow(2, attempt - 1), 1000);
        log(`Retry attempt ${attempt} for ${url}, waiting ${delay}ms`);
        await new Promise(resolve => setTimeout(resolve, delay));
      }

      // Attempt to fetch the image with timeout
      log(`Fetching image from ${url} (attempt ${attempt + 1}/${maxRetries + 1})`);
      
      // Create a controller for timeout
      const controller = new AbortController ? new AbortController() : null;
      let timeoutId = null;
      
      try {
        const options = {
          headers: NETWORK.headers
        };
        
        // Add signal if AbortController is available
        if (controller) {
          options.signal = controller.signal;
          timeoutId = setTimeout(() => controller.abort(), NETWORK.fetchTimeout);
        }
        
        const response = await fetch(url, options);
        
        if (timeoutId) {
          clearTimeout(timeoutId);
        }
        
        if (!response.ok) {
          throw new Error(`HTTP status ${response.status}`);
        }

        // Convert response to buffer - handle node-fetch v2
        const buffer = await response.buffer();
        
        // Check if the buffer is valid (non-empty)
        if (!buffer || buffer.length === 0) {
          throw new Error('Empty image buffer received');
        }
        
        log(`Successfully fetched image (${buffer.length} bytes)`);
        return buffer;
      } catch (fetchErr) {
        if (timeoutId) {
          clearTimeout(timeoutId);
        }
        throw fetchErr;
      }
    } catch (err) {
      lastError = err;
      log(`Attempt ${attempt + 1} failed to fetch image:`, { url, error: err.message });
      
      // If this is a network error that might be temporary, continue to retry
      // Otherwise, for permanent errors like 404, stop retrying
      if (err.message.includes('404') || err.message.includes('403')) {
        log('Permanent error detected, stopping retry attempts');
        break;
      }
    }
  }

  // All attempts failed
  log(`All ${maxRetries + 1} attempts to fetch image failed:`, { url, error: lastError?.message });
  return null;
}

/**
 * Check if image URL is valid (basic validation)
 * 
 * @param {string} url - URL to validate
 * @returns {boolean} - Whether URL appears valid
 */
function isValidImageUrl(url) {
  if (!url) return false;
  
  // Very basic URL validation
  try {
    const parsedUrl = new URL(url);
    return ['http:', 'https:'].includes(parsedUrl.protocol);
  } catch (e) {
    return false;
  }
}

module.exports = {
  downloadJsonFromS3,
  uploadToS3,
  setupFonts,
  fetchImage,
  isValidImageUrl,
  handleError,
  capturedLogs
};