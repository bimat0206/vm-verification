// Browser-compatible UUID generator

class UUIDGenerator {
  static generate() {
    // Modern browsers with crypto.randomUUID support
    if (window.crypto && window.crypto.randomUUID) {
      return window.crypto.randomUUID();
    }
    
    // Browsers with crypto.getRandomValues support
    if (window.crypto && window.crypto.getRandomValues) {
      return UUIDGenerator.generateWithCrypto();
    }
    
    // Fallback for older browsers
    return UUIDGenerator.generateWithMathRandom();
  }
  
  static generateWithCrypto() {
    const array = new Uint8Array(16);
    window.crypto.getRandomValues(array);
    
    // Set version (4) and variant bits
    array[6] = (array[6] & 0x0f) | 0x40; // Version 4
    array[8] = (array[8] & 0x3f) | 0x80; // Variant 10
    
    const hex = Array.from(array, byte => 
      byte.toString(16).padStart(2, '0')
    ).join('');
    
    return [
      hex.slice(0, 8),
      hex.slice(8, 12),
      hex.slice(12, 16),
      hex.slice(16, 20),
      hex.slice(20, 32)
    ].join('-');
  }
  
  static generateWithMathRandom() {
    // Fallback using Math.random (less secure)
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      const r = Math.random() * 16 | 0;
      const v = c === 'x' ? r : (r & 0x3 | 0x8);
      return v.toString(16);
    });
  }
  
  static isSupported() {
    return !!(window.crypto && (window.crypto.randomUUID || window.crypto.getRandomValues));
  }
}

// Polyfill crypto.randomUUID if not available
if (window.crypto && !window.crypto.randomUUID) {
  window.crypto.randomUUID = function() {
    return UUIDGenerator.generate();
  };
}

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = UUIDGenerator;
} else {
  window.UUIDGenerator = UUIDGenerator;
}

// Usage:
// const id = UUIDGenerator.generate();
// or if polyfilled:
// const id = crypto.randomUUID();
