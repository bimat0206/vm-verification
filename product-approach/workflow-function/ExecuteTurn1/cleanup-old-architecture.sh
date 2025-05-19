#!/bin/bash
# Script to clean up files from the old architecture approach

echo "Cleaning up files from the old architecture..."

# Remove the old models package entirely
echo "Removing old models package..."
rm -rf internal/models

# Remove old handler files
echo "Removing old handler files..."

# Keep our new handler.go but remove the old ones
find internal/handler -name "*.go" ! -name "handler.go" -exec rm -f {} \;

# Done
echo "Cleanup complete."