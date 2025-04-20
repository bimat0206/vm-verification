# 5-File Architecture for Vending Machine Layout Generator

## File Structure
```
render-layout-lambda-node/
├── index.js                   # Main entry point and handler
├── services.js                # Core services (S3, logging)
├── renderer.js                # Canvas and rendering functionality
├── utils.js                   # Utilities (error handling, parsers)
├── config.js                  # Configuration and constants
├── package.json               # Dependencies
└── Dockerfile                 # Container configuration
```

## File Responsibilities

### 1. index.js
- Lambda handler function
- Main workflow orchestration
- Event handling
- High-level error handling

### 2. services.js
- S3 client and operations
- Logging service
- Font setup service
- Image fetching service

### 3. renderer.js
- Canvas initialization
- Layout rendering logic
- Component rendering (cells, trays, headers, footers)
- Text and image processing for the canvas

### 4. utils.js
- Event parsing
- Error handling utilities
- Text manipulation utilities
- Helper functions

### 5. config.js
- Environment variables
- Canvas and layout settings
- AWS configuration
- Default values and constants