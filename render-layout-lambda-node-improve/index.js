/**
 * index.js - Main entry point for the vending machine layout generator
 * 
 * Lambda handler that orchestrates the workflow for generating and storing
 * vending machine layout images.
 */

const crypto = require('crypto');
const { log, capturedLogs } = require('./common');
const { downloadJsonFromS3, uploadToS3, setupFonts, handleError } = require('./services');
const { parseEvent } = require('./utils');
const { renderLayout } = require('./renderer');
const { KOOTORO } = require('./config');

/**
 * AWS Lambda handler function
 * 
 * Processes events from S3 or EventBridge, downloads JSON layout data,
 * renders the layout as an image, and uploads the result back to S3.
 * 
 * @param {Object} event - Lambda event object
 * @returns {Object} - Response object with status and output information
 */
exports.handler = async (event) => {
  log('Lambda handler started');
  
  try {
    // Initialize fonts early to catch font issues before processing
    await setupFonts();
    log('Font setup completed');

    // Parse event to get bucket and key
    const parseResult = parseEvent(event);
    if (parseResult.status === 'skipped') {
      log('Skipping processing: ' + parseResult.reason);
      return parseResult;
    }
    const { bucket, key } = parseResult;

    // Download and parse JSON layout
    const layout = await downloadJsonFromS3(bucket, key);
    
    // Validate layout before proceeding
    validateLayout(layout);

    // Render layout to PNG
    const pngBuffer = await renderLayout(layout);

    // Generate output key with layout ID or random hex
    const layoutId = layout.layoutId || crypto.randomBytes(4).toString('hex');
    const outputKey = `rendered-layout/layout_${layoutId}.png`;
    
    // Upload PNG to S3
    await uploadToS3(bucket, outputKey, pngBuffer, 'image/png');

    // Upload logs to S3
    const logKey = `logs/layout_${layoutId}_${Date.now()}.log`;
    await uploadToS3(bucket, logKey, capturedLogs.join('\n'), 'text/plain');

    log('Lambda handler completed successfully');
    
    // Return success response
    return { 
      status: 'success', 
      outputKey,
      layoutId,
      timestamp: new Date().toISOString()
    };
  } catch (err) {
    // Handle errors and upload logs
    return await handleError(err);
  }
};

/**
 * Validate layout data to ensure it conforms to expected structure
 * 
 * @param {Object} layout - Layout data to validate
 * @throws {Error} If layout is invalid
 */
function validateLayout(layout) {
  // Check essential fields
  if (!layout) {
    throw new Error('Layout data is empty or null');
  }
  
  // Check for required fields
  if (!layout.layoutId) {
    log('Warning: Layout missing layoutId, will generate one');
  }
  
  // Check sublayout structure
  if (!layout.subLayoutList || !Array.isArray(layout.subLayoutList) || layout.subLayoutList.length === 0) {
    throw new Error('Layout missing valid subLayoutList array');
  }
  
  // Check tray structure
  const trays = layout.subLayoutList[0]?.trayList;
  if (!trays || !Array.isArray(trays) || trays.length === 0) {
    throw new Error('Layout missing valid trayList array');
  }
  
  // Ensure we don't exceed maximum trays
  if (trays.length > KOOTORO.maxProducts) {
    throw new Error(`Layout exceeds maximum product limit of ${KOOTORO.maxProducts}`);
  }
  
  // Validate each tray has expected properties
  for (const tray of trays) {
    if (!tray.trayCode) {
      log(`Warning: Tray missing trayCode, will generate one for trayNo ${tray.trayNo}`);
    }
    
    // Check slots
    if (!tray.slotList || !Array.isArray(tray.slotList)) {
      log(`Warning: Tray ${tray.trayCode || tray.trayNo} has no valid slotList`);
    }
  }
  
  log('Layout validation passed');
}