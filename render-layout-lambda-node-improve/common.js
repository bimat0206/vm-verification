/**
 * common.js - Shared utilities to break circular dependencies
 * 
 * Contains logging and error handling utilities used across modules.
 */

// Logging service
const capturedLogs = [];

/**
 * Log a message to console and capture it for later upload
 * 
 * @param {...any} args - Log message and parameters
 */
function log(...args) {
  try {
    const msg = args
      .map((a) => (typeof a === 'string' ? a : JSON.stringify(a, null, 2)))
      .join(' ');
    const entry = `${new Date().toISOString()} ${msg}`;
    console.log(entry);
    capturedLogs.push(entry);
  } catch (err) {
    console.error('Error in log function:', err);
  }
}

module.exports = {
  log,
  capturedLogs
};