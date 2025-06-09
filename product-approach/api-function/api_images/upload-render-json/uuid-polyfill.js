// UUID v4 polyfill for environments that don't support crypto.randomUUID()

function generateUUID() {
  // Check if crypto.randomUUID is available (Node.js 15.6.0+)
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  
  // Check if crypto.getRandomValues is available (browsers and newer Node.js)
  if (typeof crypto !== 'undefined' && crypto.getRandomValues) {
    return generateUUIDWithGetRandomValues();
  }
  
  // Fallback for older Node.js versions
  if (typeof require !== 'undefined') {
    try {
      const crypto = require('crypto');
      return generateUUIDWithNodeCrypto(crypto);
    } catch (e) {
      // If crypto module is not available, use Math.random fallback
      return generateUUIDWithMathRandom();
    }
  }
  
  // Final fallback using Math.random (less secure)
  return generateUUIDWithMathRandom();
}

function generateUUIDWithGetRandomValues() {
  const array = new Uint8Array(16);
  crypto.getRandomValues(array);
  
  // Set version (4) and variant bits
  array[6] = (array[6] & 0x0f) | 0x40; // Version 4
  array[8] = (array[8] & 0x3f) | 0x80; // Variant 10
  
  const hex = Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20, 32)}`;
}

function generateUUIDWithNodeCrypto(crypto) {
  const bytes = crypto.randomBytes(16);
  
  // Set version (4) and variant bits
  bytes[6] = (bytes[6] & 0x0f) | 0x40; // Version 4
  bytes[8] = (bytes[8] & 0x3f) | 0x80; // Variant 10
  
  const hex = bytes.toString('hex');
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20, 32)}`;
}

function generateUUIDWithMathRandom() {
  // Less secure fallback using Math.random
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

// Export for different module systems
if (typeof module !== 'undefined' && module.exports) {
  // CommonJS
  module.exports = { generateUUID };
} else if (typeof define === 'function' && define.amd) {
  // AMD
  define(function() {
    return { generateUUID };
  });
} else {
  // Browser global
  window.generateUUID = generateUUID;
}

// Usage examples:
// const { generateUUID } = require('./uuid-polyfill');
// const id = generateUUID();
// 
// Or in browser:
// const id = generateUUID();
