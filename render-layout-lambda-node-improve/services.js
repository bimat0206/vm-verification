/**
 * services.js - Core services with improved ARM64 compatibility
 */

const { S3Client, GetObjectCommand, PutObjectCommand } = require('@aws-sdk/client-s3');
const fs = require('fs');
const { ENV, FONTS, NETWORK } = require('./config');
const { log, capturedLogs } = require('./common');

// Initialize variables
let fetch;
let AbortController;
let s3Client;

// Setup fetch safely with graceful fallbacks
try {
  // For Node.js versions with node-fetch v2 compatibility
  const nodeFetch = require('node-fetch');
  fetch = typeof nodeFetch === 'function' ? nodeFetch : nodeFetch.default;
  AbortController = nodeFetch.AbortController || global.AbortController;
  log('Using node-fetch module');
} catch (err) {
  // Fallback to global fetch if available (Node 18+)
  if (typeof global.fetch === 'function') {
    fetch = global.fetch;
    AbortController = global.AbortController;
    log('Using built-in fetch API');
  } else {
    log('Warning: fetch not available:', err.message);
  }
}

// Mock S3 for local testing with fixed path structure
function getMockS3Client() {
  // When running in SAM or Lambda, use /tmp for writable storage
  const tempDir = process.env.AWS_SAM_LOCAL === 'true' 
    ? '/tmp/s3-mock' 
    : (process.env.TEMP_DIR || '/tmp/s3-mock');
  
  log(`Using mock S3 with directory: ${tempDir}`);
  
  // Ensure temp directory exists
  try {
    if (!fs.existsSync(tempDir)) {
      fs.mkdirSync(tempDir, { recursive: true });
      log(`Created directory: ${tempDir}`);
    }
    
    // Create subdirectories
    ['raw', 'rendered-layout', 'logs'].forEach(subdir => {
      const dirPath = `${tempDir}/${subdir}`;
      if (!fs.existsSync(dirPath)) {
        fs.mkdirSync(dirPath, { recursive: true });
        log(`Created subdirectory: ${dirPath}`);
      }
    });
  } catch (err) {
    log('Warning: Failed to create temp directory:', err.message);
  }
  
  return {
    send: async (command) => {
      log(`Mock S3: ${command.constructor.name}`);
      
      if (command.constructor.name === 'GetObjectCommand') {
        const { Bucket, Key } = command.input;
        log(`Mock S3 GetObject: ${Bucket}/${Key}`);
        
        // Generate mock file path - don't include bucket name
        const filePath = `${tempDir}/${Key}`;
        const dir = require('path').dirname(filePath);
        
        // Create directory if it doesn't exist
        if (!fs.existsSync(dir)) {
          fs.mkdirSync(dir, { recursive: true });
        }
        
        // Create mock file if it doesn't exist
        if (!fs.existsSync(filePath)) {
          const mockData = {
            layoutId: "test123",
            subLayoutList: [{
              trayList: [
                {
                  trayCode: "A",
                  trayNo: 1,
                  slotList: [
                    {
                      slotNo: 1,
                      productTemplateName: "Test Product",
                      productTemplateImage: null
                    }
                  ]
                }
              ]
            }]
          };
          fs.writeFileSync(filePath, JSON.stringify(mockData));
          log(`Created mock file: ${filePath}`);
        }
        
        const content = fs.readFileSync(filePath);
        return {
          Body: {
            async *[Symbol.asyncIterator]() {
              yield content;
            }
          }
        };
      } else if (command.constructor.name === 'PutObjectCommand') {
        const { Bucket, Key, Body } = command.input;
        log(`Mock S3 PutObject: ${Bucket}/${Key}`);
        
        // Create output file - don't include bucket name in path
        const filePath = `${tempDir}/${Key}`;
        const dir = require('path').dirname(filePath);
        
        if (!fs.existsSync(dir)) {
          fs.mkdirSync(dir, { recursive: true });
          log(`Created directory: ${dir}`);
        }
        
        fs.writeFileSync(filePath, Body);
        log(`Wrote file to: ${filePath} (${Buffer.isBuffer(Body) ? Body.length : 'unknown'} bytes)`);
        
        // For debugging, list directory contents
        try {
          const files = fs.readdirSync(dir);
          log(`Files in ${dir}: ${files.join(', ')}`);
        } catch (err) {
          log(`Error listing directory ${dir}: ${err.message}`);
        }
        
        return { ETag: '"mock-etag"' };
      }
      
      return {};
    }
  };
}

// Initialize S3 client safely with regional endpoint support
try {
  // Check if we should use mock S3
  if (process.env.USE_MOCK_S3 === 'true' || process.env.AWS_SAM_LOCAL === 'true') {
    log('Using mock S3 client for local testing');
    s3Client = getMockS3Client();
  } else {
    const clientOptions = { 
      region: ENV.AWS_REGION,
      maxAttempts: 3,
      // Handle regional bucket endpoints
      useArnRegion: true,
      followRegionRedirects: true
    };
    
    // Add endpoint URL if specified (for LocalStack or testing)
    if (process.env.AWS_ENDPOINT_URL) {
      clientOptions.endpoint = process.env.AWS_ENDPOINT_URL;
      clientOptions.forcePathStyle = true;
    }
    
    s3Client = new S3Client(clientOptions);
    log('S3 client initialized');
  }
} catch (err) {
  log('Error initializing S3 client:', err.message);
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
    
    try {
      const layout = JSON.parse(contentBuffer.toString('utf-8'));
      log('Parsed layout JSON', { layoutId: layout.layoutId });
      return layout;
    } catch (parseErr) {
      log('Error parsing JSON:', parseErr.message);
      throw new Error(`Invalid JSON format: ${parseErr.message}`);
    }
  } catch (err) {
    log('Failed to download or parse JSON from S3:', err.message);
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
    
    log('Uploading to S3', { bucket, key, contentType });
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
    log('Failed to upload to S3:', err.message);
    throw err;
  }
}

/**
 * Handle errors with improved logging
 */
async function handleError(err) {
  log('Error in Lambda handler:', err.message);
  log('Error stack:', err.stack || 'No stack trace available');
  
  try {
    const targetBucket = ENV.LOG_BUCKET;
    const fallbackKey = `logs/error_${Date.now()}.log`;
    
    if (targetBucket && s3Client) {
      const logContent = capturedLogs.join('\n') + `\n${err.stack || err.message}`;
      await uploadToS3(targetBucket, fallbackKey, logContent, 'text/plain');
      log('Error logs uploaded to S3');
    } else {
      log('No S3 bucket specified or S3 client not initialized, skipping log upload');
    }
  } catch (uploadErr) {
    log('Failed to upload error log:', uploadErr.message);
  }
  
  return {
    status: 'error',
    message: err.message,
    timestamp: new Date().toISOString()
  };
}

/**
 * Set up font configuration with ARM64 compatibility
 */
async function setupFonts() {
  log('Starting font setup');
  
  try {
    // Set environment variables
    process.env.FONTCONFIG_PATH = FONTS.tempDir;
    process.env.FONTCONFIG_FILE = FONTS.configFile;
    
    log(`Setting font paths: ${FONTS.tempDir}, ${FONTS.configFile}`);

    // Create font config directory safely
    try {
      if (!fs.existsSync(FONTS.tempDir)) {
        log(`Creating font directory: ${FONTS.tempDir}`);
        fs.mkdirSync(FONTS.tempDir, { recursive: true });
        try {
          fs.chmodSync(FONTS.tempDir, '777');
        } catch (chmodErr) {
          log(`Warning: Could not set permissions on font directory: ${chmodErr.message}`);
        }
      }
    } catch (dirErr) {
      log(`Warning: Failed to create font directory: ${dirErr.message}`);
    }

    // Create font config file safely
    try {
      if (!fs.existsSync(FONTS.configFile)) {
        log(`Creating font config file: ${FONTS.configFile}`);
        fs.writeFileSync(FONTS.configFile, FONTS.configContent.trim());
      }
    } catch (fileErr) {
      log(`Warning: Failed to create font config file: ${fileErr.message}`);
    }

    // Check for canvas module without registering fonts yet
    try {
      log('Checking for canvas module availability');
      
      // Test-require canvas - don't store it yet
      require('canvas');
      
      log('Canvas module is available');
      
      // Now try registering fonts
      const { registerFont } = require('canvas');
      
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
          }
        } catch (fontErr) {
          log(`Failed to register font from ${fontPath}: ${fontErr.message}`);
        }
      }

      if (!fontRegistered) {
        log('Warning: Could not register custom fonts, will use system defaults');
      }
    } catch (canvasErr) {
      log('Canvas module unavailable:', canvasErr.message);
      log('Will use fallback rendering method');
    }
    
    log('Font setup completed');
  } catch (err) {
    log('Font setup encountered issues:', err.message);
    log('Continuing without custom fonts');
  }
}

/**
 * Fetch an image with improved error handling
 */
async function fetchImage(url) {
  if (!fetch) {
    log('Fetch API not available, cannot download images');
    return null;
  }
  
  const maxRetries = NETWORK.retryAttempts;
  let lastError = null;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      if (attempt > 0) {
        const delay = Math.min(100 * Math.pow(2, attempt - 1), 1000);
        log(`Retry attempt ${attempt} for ${url}, waiting ${delay}ms`);
        await new Promise(resolve => setTimeout(resolve, delay));
      }

      log(`Fetching image from ${url} (attempt ${attempt + 1}/${maxRetries + 1})`);
      
      let options = { headers: NETWORK.headers };
      let timeoutId = null;
      
      if (AbortController) {
        const controller = new AbortController();
        options.signal = controller.signal;
        timeoutId = setTimeout(() => controller.abort(), NETWORK.fetchTimeout);
      }
      
      try {
        const response = await fetch(url, options);
        
        if (timeoutId) clearTimeout(timeoutId);
        
        if (!response.ok) {
          throw new Error(`HTTP status ${response.status}`);
        }

        // Handle different node-fetch versions
        let buffer;
        if (typeof response.buffer === 'function') {
          buffer = await response.buffer();
        } else if (typeof response.arrayBuffer === 'function') {
          const arrayBuffer = await response.arrayBuffer();
          buffer = Buffer.from(arrayBuffer);
        } else {
          throw new Error('Cannot convert response to buffer');
        }
        
        if (!buffer || buffer.length === 0) {
          throw new Error('Empty image buffer received');
        }
        
        log(`Successfully fetched image (${buffer.length} bytes)`);
        return buffer;
      } catch (fetchErr) {
        if (timeoutId) clearTimeout(timeoutId);
        throw fetchErr;
      }
    } catch (err) {
      lastError = err;
      log(`Attempt ${attempt + 1} failed to fetch image:`, { url, error: err.message });
      
      if (err.message.includes('404') || err.message.includes('403')) {
        log('Permanent error detected, stopping retry attempts');
        break;
      }
    }
  }

  log(`All attempts to fetch image failed:`, { url, error: lastError?.message });
  return null;
}

/**
 * Validate image URL
 */
function isValidImageUrl(url) {
  if (!url) return false;
  
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