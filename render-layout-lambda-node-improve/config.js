/**
 * config.js - Configuration settings for the vending machine layout generator
 * 
 * Contains all configuration parameters, constants, and environment variables
 * used throughout the application.
 */

// Environment variables with defaults
const ENV = {
    NODE_ENV: process.env.NODE_ENV || 'production',
    AWS_REGION: process.env.AWS_REGION || 'us-east-1',
    LOG_BUCKET: process.env.LOG_BUCKET || process.env.BUCKET,
    FONTCONFIG_PATH: '/tmp/fontconfig',
    FONTCONFIG_FILE: '/tmp/fonts.conf'
  };
  
  // Canvas rendering settings
  const CANVAS = {
    scale: 2.0,
    maxTrays: 50,
    cell: {
      width: 150,
      height: 180,
      spacing: 10
    },
    row: {
      spacing: 60
    },
    header: {
      height: 40
    },
    footer: {
      height: 30
    },
    image: {
      size: 100
    },
    padding: 20,
    titlePadding: 40,
    textPadding: 5,
    metadataHeight: 20,
    fallbackWidth: 800,
    fallbackHeight: 600
  };
  
  // Font configuration
  const FONTS = {
    paths: [
      '/app/fonts/dejavu-sans.ttf',
      '/app/fonts/arial.ttf',
      '/var/task/fonts/arial.ttf',
      `${process.env.LAMBDA_TASK_ROOT}/fonts/arial.ttf`,
      `${__dirname}/fonts/arial.ttf`,
      '/opt/fonts/arial.ttf'
    ],
    tempDir: '/tmp/fontconfig',
    configFile: '/tmp/fonts.conf',
    configContent: `<?xml version="1.0"?>
  <!DOCTYPE fontconfig SYSTEM "fonts.dtd">
  <fontconfig>
    <dir>/app/fonts</dir>
    <cachedir>/tmp/fontconfig</cachedir>
  </fontconfig>`
  };
  
  // Network settings
  const NETWORK = {
    fetchTimeout: 5000,
    retryAttempts: 2,
    headers: {
      'User-Agent': 'Mozilla/5.0 Vending Machine Layout Generator',
      'Accept': 'image/*'
    }
  };
  
  // Kootoro specific settings
  const KOOTORO = {
    company: 'Kootoro',
    brandColor: 'rgb(0, 0, 150)',
    defaultProductName: 'Sản phẩm',  // Default Vietnamese product name
    supportedImageTypes: ['png', 'jpg', 'jpeg', 'gif', 'webp'],
    layoutTypes: {
      standard: 7,  // 7 columns
      compact: 5    // 5 columns (for future use)
    },
    maxProducts: 200
  };
  
  // Changed from "export {}" to CommonJS syntax to fix the error
  module.exports = {
    ENV,
    CANVAS,
    FONTS,
    NETWORK,
    KOOTORO
  };