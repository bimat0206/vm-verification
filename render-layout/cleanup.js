/**
 * Cleanup script for Node.js dependencies
 * This script removes node_modules, package-lock.json, and other generated files
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// Text colors for console
const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  red: '\x1b[31m',
  blue: '\x1b[34m'
};

function log(message, color = colors.reset) {
  console.log(`${color}${message}${colors.reset}`);
}

function cleanupDependencies() {
  log('Starting dependency cleanup...', colors.blue);

  // 1. Remove node_modules directory
  const nodeModulesPath = path.join(__dirname, 'node_modules');
  if (fs.existsSync(nodeModulesPath)) {
    log('Removing node_modules directory...', colors.yellow);
    try {
      // Use different removal methods depending on platform
      if (process.platform === 'win32') {
        execSync('rmdir /s /q node_modules', { stdio: 'inherit' });
      } else {
        execSync('rm -rf node_modules', { stdio: 'inherit' });
      }
      log('✅ node_modules removed successfully', colors.green);
    } catch (error) {
      log(`❌ Error removing node_modules: ${error.message}`, colors.red);
      log('You may need to remove it manually', colors.yellow);
    }
  } else {
    log('node_modules directory not found, skipping', colors.yellow);
  }

  // 2. Remove package-lock.json
  const packageLockPath = path.join(__dirname, 'package-lock.json');
  if (fs.existsSync(packageLockPath)) {
    log('Removing package-lock.json...', colors.yellow);
    try {
      fs.unlinkSync(packageLockPath);
      log('✅ package-lock.json removed successfully', colors.green);
    } catch (error) {
      log(`❌ Error removing package-lock.json: ${error.message}`, colors.red);
    }
  } else {
    log('package-lock.json not found, skipping', colors.yellow);
  }

  // 3. Remove npm cache
  log('Cleaning npm cache...', colors.yellow);
  try {
    execSync('npm cache clean --force', { stdio: 'inherit' });
    log('✅ npm cache cleaned successfully', colors.green);
  } catch (error) {
    log(`❌ Error cleaning npm cache: ${error.message}`, colors.red);
  }

  // 4. Remove any generated test files
  const testImagePath = path.join(__dirname, 'canvas-test.png');
  if (fs.existsSync(testImagePath)) {
    log('Removing test image file...', colors.yellow);
    try {
      fs.unlinkSync(testImagePath);
      log('✅ Test image removed successfully', colors.green);
    } catch (error) {
      log(`❌ Error removing test image: ${error.message}`, colors.red);
    }
  }

  // 5. Check and optionally update package.json
  const packageJsonPath = path.join(__dirname, 'package.json');
  if (fs.existsSync(packageJsonPath)) {
    log('\nPackage.json found. Do you want to:', colors.blue);
    log('1. Leave package.json as is', colors.yellow);
    log('2. Reset dependencies to default', colors.yellow);
    log('3. Remove canvas dependency only', colors.yellow);
    
    // Use readline for user input
    const readline = require('readline').createInterface({
      input: process.stdin,
      output: process.stdout
    });

    readline.question('\nEnter your choice (1-3): ', (choice) => {
      try {
        if (choice === '2') {
          // Reset dependencies to default
          log('Resetting dependencies to default...', colors.blue);
          const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
          packageJson.dependencies = {
            "canvas": "^2.11.2",
            "fs": "0.0.1-security",
            "path": "^0.12.7"
          };
          fs.writeFileSync(packageJsonPath, JSON.stringify(packageJson, null, 2));
          log('✅ Dependencies reset successfully', colors.green);
        } else if (choice === '3') {
          // Remove canvas dependency only
          log('Removing canvas dependency...', colors.blue);
          const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
          if (packageJson.dependencies && packageJson.dependencies.canvas) {
            delete packageJson.dependencies.canvas;
            fs.writeFileSync(packageJsonPath, JSON.stringify(packageJson, null, 2));
            log('✅ Canvas dependency removed successfully', colors.green);
          } else {
            log('Canvas dependency not found in package.json', colors.yellow);
          }
        } else {
          log('Leaving package.json unchanged', colors.blue);
        }
        
        log('\n✅ Cleanup completed successfully!', colors.green);
        log('\nTo reinstall dependencies:', colors.blue);
        log('1. Run: node install.js', colors.yellow);
        log('   OR', colors.blue);
        log('2. Run: npm install', colors.yellow);
        
        readline.close();
      } catch (error) {
        log(`\n❌ Error updating package.json: ${error.message}`, colors.red);
        readline.close();
      }
    });
  } else {
    log('\n✅ Cleanup completed successfully!', colors.green);
    log('\nNote: package.json not found', colors.yellow);
    log('To reinstall dependencies, first create a package.json file', colors.blue);
  }
}

// Run the cleanup
cleanupDependencies();