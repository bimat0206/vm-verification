/**
 * Installation helper script
 * This script helps ensure all dependencies are properly installed
 * and sets up the project structure.
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const os = require('os');

// Canvas specific version
const CANVAS_VERSION = '3.1.0';

// Color codes for console output
const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  red: '\x1b[31m',
  blue: '\x1b[34m'
};

// Print a styled message
function log(message, color = colors.reset) {
  console.log(`${color}${message}${colors.reset}`);
}

// Create necessary directories
function createDirectories() {
  log('Creating project directories...', colors.blue);
  
  const directories = ['input', 'output'];
  
  directories.forEach(dir => {
    const dirPath = path.join(__dirname, dir);
    if (!fs.existsSync(dirPath)) {
      fs.mkdirSync(dirPath, { recursive: true });
      log(` - Created ${dir}/ directory`, colors.green);
    } else {
      log(` - ${dir}/ directory already exists`, colors.yellow);
    }
  });
}

// Create or update package.json
function setupPackageJson() {
  log('\nSetting up package.json...', colors.blue);
  
  const packageJsonPath = path.join(__dirname, 'package.json');
  let existingPackageJson = null;
  
  // Try to load existing package.json
  if (fs.existsSync(packageJsonPath)) {
    try {
      existingPackageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
      log(' - Found existing package.json', colors.yellow);
    } catch (error) {
      log(` - Error reading existing package.json: ${error.message}`, colors.red);
      existingPackageJson = null;
    }
  }
  
  // Create comprehensive package.json template
  const packageJsonTemplate = {
    "name": "vending-machine-layout-generator",
    "version": "1.0.0",
    "description": "Node.js application that generates vending machine layouts from JSON data",
    "main": "main.js",
    "scripts": {
      "start": "node main.js",
      "clean": "node cleanup.js",
      "install-deps": "node install.js"
    },
    "dependencies": {
      "canvas": CANVAS_VERSION,
      "fs": "0.0.1-security",
      "path": "^0.12.7"
    },
    "devDependencies": {},
    "engines": {
      "node": ">=14.0.0"
    },
    "author": "",
    "license": "ISC",
    "repository": {
      "type": "git",
      "url": ""
    },
    "keywords": [
      "vending-machine",
      "layout",
      "generator",
      "canvas",
      "node"
    ]
  };
  
  // If existing package.json, merge with template, prioritizing our dependencies
  let finalPackageJson = packageJsonTemplate;
  if (existingPackageJson) {
    // Merge existing package.json with our template, preserving existing fields
    finalPackageJson = {
      ...existingPackageJson,
      // Ensure these fields are taken from template
      scripts: {
        ...(existingPackageJson.scripts || {}),
        ...packageJsonTemplate.scripts
      },
      dependencies: {
        ...(existingPackageJson.dependencies || {}),
        // Force specific canvas version
        canvas: CANVAS_VERSION,
        // Ensure fs and path are included
        fs: packageJsonTemplate.dependencies.fs,
        path: packageJsonTemplate.dependencies.path
      },
      engines: packageJsonTemplate.engines
    };
    
    log(' - Merged with existing package.json, preserving custom settings', colors.green);
  } else {
    log(' - Created new package.json file', colors.green);
  }
  
  // Write the package.json file
  fs.writeFileSync(packageJsonPath, JSON.stringify(finalPackageJson, null, 2));
  log(' - Successfully wrote package.json', colors.green);
  
  return finalPackageJson;
}

// Install npm dependencies
function installDependencies() {
  log('\nInstalling dependencies...', colors.blue);
  
  try {
    // Set up package.json first
    setupPackageJson();
    
    // Remove package-lock.json to prevent version conflicts
    const packageLockPath = path.join(__dirname, 'package-lock.json');
    if (fs.existsSync(packageLockPath)) {
      log(' - Removing existing package-lock.json to avoid version conflicts...', colors.yellow);
      fs.unlinkSync(packageLockPath);
    }
    
    // Install dependencies with npm install instead of npm ci
    log(' - Running npm install...', colors.yellow);
    execSync('npm install', { stdio: 'inherit' });
    log(' - Base dependencies installed successfully', colors.green);
    
    // Install specific canvas version with force flag
    log(`\nInstalling canvas version ${CANVAS_VERSION}...`, colors.blue);
    try {
      // Force install the specific canvas version
      log(` - Installing canvas@${CANVAS_VERSION}...`, colors.yellow);
      execSync(`npm install canvas@${CANVAS_VERSION} --save --force`, { stdio: 'inherit' });
      log(` - Canvas ${CANVAS_VERSION} installed successfully`, colors.green);
    } catch (error) {
      log(`\nError installing canvas: ${error.message}`, colors.red);
      log(' - Trying alternative installation method...', colors.yellow);
      
      // Try with --build-from-source flag as a backup
      try {
        execSync(`npm install canvas@${CANVAS_VERSION} --build-from-source --save --force`, { stdio: 'inherit' });
        log(' - Alternative installation method succeeded', colors.green);
      } catch (altError) {
        log(`\nAlternative installation also failed: ${altError.message}`, colors.red);
        throw new Error('Canvas installation failed');
      }
    }
    
    // Verify canvas installation specifically
    log('\nVerifying canvas installation...', colors.blue);
    const canvasPath = path.join(__dirname, 'node_modules', 'canvas');
    if (fs.existsSync(canvasPath)) {
      log(' - Canvas library found', colors.green);
      
      // Check canvas version
      try {
        const packageJson = require(path.join(canvasPath, 'package.json'));
        log(` - Canvas version: ${packageJson.version}`, colors.green);
        
        if (packageJson.version !== CANVAS_VERSION) {
          log(` - Warning: Installed canvas version (${packageJson.version}) doesn't match requested version (${CANVAS_VERSION})`, colors.yellow);
        }
      } catch (err) {
        log(' - Could not determine canvas version', colors.yellow);
      }
    } else {
      log(' - Canvas library not found, installation may have failed', colors.red);
      throw new Error('Canvas not found after installation');
    }
    
    // Check for canvas prebuilt binaries
    const prebuildPath = path.join(canvasPath, 'build', 'Release');
    if (fs.existsSync(prebuildPath)) {
      log(' - Canvas prebuilt binaries found', colors.green);
    } else {
      log(' - Canvas may need to be built from source', colors.yellow);
      
      // Check system requirements for canvas
      if (os.platform() === 'linux') {
        log('\nOn Linux, canvas requires additional system packages:', colors.yellow);
        log(' - Debian/Ubuntu: sudo apt-get install build-essential libcairo2-dev libpango1.0-dev libjpeg-dev libgif-dev librsvg2-dev', colors.blue);
        log(' - Fedora: sudo yum install gcc-c++ cairo-devel pango-devel libjpeg-turbo-devel giflib-devel', colors.blue);
        log(' - After installing these packages, run this script again.', colors.blue);
      }
      else if (os.platform() === 'darwin') {
        log('\nOn macOS, canvas requires additional tools:', colors.yellow);
        log(' - Install XCode Command Line Tools: xcode-select --install', colors.blue);
        log(' - With Homebrew: brew install pkg-config cairo pango libpng jpeg giflib librsvg', colors.blue);
        log(' - After installing these packages, run this script again.', colors.blue);
      }
      else if (os.platform() === 'win32') {
        log('\nOn Windows, canvas should automatically use prebuilt binaries, but you may need:', colors.yellow);
        log(' - Visual Studio Build Tools with C++ development workload', colors.blue);
        log(' - GTK 2 (for advanced features)', colors.blue);
      }
    }
    
  } catch (error) {
    log(`\nError installing dependencies: ${error.message}`, colors.red);
    process.exit(1);
  }
}

// Check and report Node.js version
function checkNodeVersion() {
  const nodeVersion = process.version;
  log(`\nDetected Node.js ${nodeVersion}`, colors.blue);
  
  // Parse version number
  const versionMatch = nodeVersion.match(/v(\d+)\.\d+\.\d+/);
  if (versionMatch && parseInt(versionMatch[1]) < 14) {
    log('Warning: This application requires Node.js v14 or higher', colors.red);
    log('Please update your Node.js version and try again', colors.red);
    process.exit(1);
  } else {
    log('Node.js version is compatible', colors.green);
  }
}

// Additional checks for canvas library
function verifyCanvasInstallation() {
  log('\nVerifying canvas functionality...', colors.blue);
  
  try {
    // Try loading the canvas module
    const canvas = require('canvas');
    log(' - Canvas module loaded successfully', colors.green);
    log(` - Canvas version: ${canvas.version || 'unknown'}`, colors.green);
    
    // Try creating a simple canvas to verify it works
    const { createCanvas } = require('canvas');
    const testCanvas = createCanvas(100, 100);
    const ctx = testCanvas.getContext('2d');
    
    // Draw something simple with explicit white background
    ctx.fillStyle = 'white';
    ctx.fillRect(0, 0, 100, 100);
    ctx.fillStyle = 'black';
    ctx.font = '16px Arial';
    ctx.fillText('Test', 30, 50);
    
    // Convert to buffer (this would fail if canvas isn't working properly)
    const buffer = testCanvas.toBuffer('image/png', { 
      compressionLevel: 6,
      backgroundColor: '#ffffff'
    });
    
    log(' - Canvas rendering test passed', colors.green);
    
    // Save test image to verify transparency handling
    const testImagePath = path.join(__dirname, 'canvas-test.png');
    fs.writeFileSync(testImagePath, buffer);
    log(` - Test image saved to: ${testImagePath}`, colors.green);
    log(' - Please check this image to verify there are no transparency issues', colors.yellow);
  } catch (error) {
    log(`\nCanvas verification failed: ${error.message}`, colors.red);
    log('Please check the system requirements for canvas library:', colors.yellow);
    log('https://github.com/Automattic/node-canvas#compiling', colors.blue);
    process.exit(1);
  }
}

// Create sample cleanup script if it doesn't exist
function createCleanupScript() {
  const cleanupPath = path.join(__dirname, 'cleanup.js');
  
  if (!fs.existsSync(cleanupPath)) {
    log('\nCreating cleanup script...', colors.blue);
    
    const cleanupContent = `/**
 * Cleanup script for Node.js dependencies
 * This script removes node_modules, package-lock.json, and other generated files
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

console.log('Starting dependency cleanup...');

// 1. Remove node_modules directory
const nodeModulesPath = path.join(__dirname, 'node_modules');
if (fs.existsSync(nodeModulesPath)) {
  console.log('Removing node_modules directory...');
  try {
    // Use different removal methods depending on platform
    if (process.platform === 'win32') {
      execSync('rmdir /s /q node_modules', { stdio: 'inherit' });
    } else {
      execSync('rm -rf node_modules', { stdio: 'inherit' });
    }
    console.log('✅ node_modules removed successfully');
  } catch (error) {
    console.error(\`❌ Error removing node_modules: \${error.message}\`);
    console.log('You may need to remove it manually');
  }
} else {
  console.log('node_modules directory not found, skipping');
}

// 2. Remove package-lock.json
const packageLockPath = path.join(__dirname, 'package-lock.json');
if (fs.existsSync(packageLockPath)) {
  console.log('Removing package-lock.json...');
  try {
    fs.unlinkSync(packageLockPath);
    console.log('✅ package-lock.json removed successfully');
  } catch (error) {
    console.error(\`❌ Error removing package-lock.json: \${error.message}\`);
  }
} else {
  console.log('package-lock.json not found, skipping');
}

// 3. Remove generated test files
const testImagePath = path.join(__dirname, 'canvas-test.png');
if (fs.existsSync(testImagePath)) {
  console.log('Removing test image file...');
  try {
    fs.unlinkSync(testImagePath);
    console.log('✅ Test image removed successfully');
  } catch (error) {
    console.error(\`❌ Error removing test image: \${error.message}\`);
  }
}

console.log('\\n✅ Cleanup completed successfully!');
console.log('\\nTo reinstall dependencies, run:');
console.log('node install.js');
`;
    
    fs.writeFileSync(cleanupPath, cleanupContent);
    log(' - Created cleanup.js script', colors.green);
  }
}

// Run the full installation process
function runInstallation() {
  log('Starting installation process...', colors.green);
  log(`Target canvas version: ${CANVAS_VERSION}`, colors.blue);
  
  checkNodeVersion();
  createDirectories();
  installDependencies();
  verifyCanvasInstallation();
  createCleanupScript();
  
  log('\n✅ Installation completed successfully!', colors.green);
  log('\nYou can now run the application with:', colors.blue);
  log(' - npm start', colors.yellow);
  log(' or', colors.blue);
  log(' - node main.js', colors.yellow);
  log('\nTo clean up dependencies:', colors.blue);
  log(' - npm run clean', colors.yellow);
  log(' or', colors.blue);
  log(' - node cleanup.js', colors.yellow);
}

// Start the installation process
runInstallation();