// Universal UUID generator that works in all environments

(function(global) {
  'use strict';

  function createUUIDGenerator() {
    // Check environment and available APIs
    const isNode = typeof process !== 'undefined' && process.versions && process.versions.node;
    const isBrowser = typeof window !== 'undefined';
    const hasWebCrypto = typeof crypto !== 'undefined' && crypto.getRandomValues;
    const hasNodeCrypto = isNode && (() => {
      try {
        require('crypto');
        return true;
      } catch (e) {
        return false;
      }
    })();

    function generateUUID() {
      // Modern environments with crypto.randomUUID
      if (typeof crypto !== 'undefined' && crypto.randomUUID) {
        return crypto.randomUUID();
      }

      // Node.js with crypto module
      if (hasNodeCrypto) {
        const crypto = require('crypto');
        const bytes = crypto.randomBytes(16);
        
        // Set version (4) and variant bits
        bytes[6] = (bytes[6] & 0x0f) | 0x40;
        bytes[8] = (bytes[8] & 0x3f) | 0x80;
        
        const hex = bytes.toString('hex');
        return formatUUID(hex);
      }

      // Browser with Web Crypto API
      if (hasWebCrypto) {
        const array = new Uint8Array(16);
        crypto.getRandomValues(array);
        
        // Set version (4) and variant bits
        array[6] = (array[6] & 0x0f) | 0x40;
        array[8] = (array[8] & 0x3f) | 0x80;
        
        const hex = Array.from(array, byte => 
          byte.toString(16).padStart(2, '0')
        ).join('');
        return formatUUID(hex);
      }

      // Fallback using Math.random
      return generateUUIDFallback();
    }

    function formatUUID(hex) {
      return [
        hex.slice(0, 8),
        hex.slice(8, 12),
        hex.slice(12, 16),
        hex.slice(16, 20),
        hex.slice(20, 32)
      ].join('-');
    }

    function generateUUIDFallback() {
      // RFC 4122 version 4 UUID using Math.random
      return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        const r = Math.random() * 16 | 0;
        const v = c === 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
      });
    }

    function getCapabilities() {
      return {
        environment: isNode ? 'node' : (isBrowser ? 'browser' : 'unknown'),
        hasNativeCryptoUUID: typeof crypto !== 'undefined' && !!crypto.randomUUID,
        hasWebCrypto: hasWebCrypto,
        hasNodeCrypto: hasNodeCrypto,
        isSecure: hasNodeCrypto || hasWebCrypto || (typeof crypto !== 'undefined' && !!crypto.randomUUID)
      };
    }

    return {
      generate: generateUUID,
      capabilities: getCapabilities,
      // Alias for compatibility
      v4: generateUUID
    };
  }

  const UUIDGenerator = createUUIDGenerator();

  // Install polyfill if crypto.randomUUID is not available
  if (typeof crypto !== 'undefined' && !crypto.randomUUID) {
    crypto.randomUUID = UUIDGenerator.generate;
  }

  // Export for different module systems
  if (typeof module !== 'undefined' && module.exports) {
    // CommonJS (Node.js)
    module.exports = UUIDGenerator;
  } else if (typeof define === 'function' && define.amd) {
    // AMD
    define(function() {
      return UUIDGenerator;
    });
  } else {
    // Browser global
    global.UUIDGenerator = UUIDGenerator;
    
    // Also add to window if in browser
    if (typeof window !== 'undefined') {
      window.UUIDGenerator = UUIDGenerator;
    }
  }

})(typeof globalThis !== 'undefined' ? globalThis : 
   typeof window !== 'undefined' ? window : 
   typeof global !== 'undefined' ? global : this);

// Usage examples:
//
// CommonJS:
// const { generate } = require('./universal-uuid');
// const id = generate();
//
// ES6 Modules:
// import UUIDGenerator from './universal-uuid.js';
// const id = UUIDGenerator.generate();
//
// Browser:
// const id = UUIDGenerator.generate();
// or if polyfilled:
// const id = crypto.randomUUID();
//
// Check capabilities:
// console.log(UUIDGenerator.capabilities());
