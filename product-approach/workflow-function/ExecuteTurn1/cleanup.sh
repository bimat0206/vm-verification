#!/bin/bash
set -euo pipefail

echo "Cleaning up temporary build files..."

# Remove common temp directories
rm -rf ./docker_build
rm -rf ./temp_build*
rm -rf ./.temp*

# Remove any backup files
find . -name "*.bak" -delete

# Remove any Docker temp files
find . -name ".docker*" -delete

echo "✅ Cleanup completed"