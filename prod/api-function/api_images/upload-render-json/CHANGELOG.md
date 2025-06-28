# API Images Upload Render JSON - Changelog

All notable changes to the API Images Upload Render JSON Lambda function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.1] - 2025-06-19

### Fixed
- **Dynamic Grid Rendering**: Fixed issue where layouts were rendering fixed 20×20 grids instead of dynamic sizing
  - Added dynamic column calculation based on actual JSON data (`actualColumns`)
  - Canvas width now adjusts to actual number of columns needed (e.g., 7 for DataInput.json, 10 for m_AAKD template)
  - Column headers now render only for actual columns present in data
  - Cell rendering loop now uses dynamic column count instead of fixed `cfg.NumColumns`
  - Prevents rendering empty slots when JSON data has fewer columns than maximum supported

### Changed
- **Grid Size Support**: Expanded maximum grid support from 7x7 to 20x20
  - Updated `NumColumns` configuration from 7 to 20 columns (acts as maximum limit)
  - System now dynamically sizes based on JSON data up to this maximum
  - System now supports vending machine layouts up to 20 rows × 20 columns
  - Maintains backward compatibility with smaller layouts
  - Enables support for larger vending machine configurations

### Fixed
- **Binary File Corruption Detection**: Added comprehensive binary file integrity validation
  - Added `isBinaryFile()` function to detect binary files by extension
  - Added `validateBinaryFileIntegrity()` function to check file signatures and detect corruption
  - Added validation for JPEG, PNG, GIF, and PDF file signatures
  - Added detection of UTF-8 replacement characters indicating binary corruption
  - Enhanced `parseMultipartForm()` with better logging and corruption detection
  - This prevents corrupted binary files from being uploaded to S3

### Enhanced
- **Improved Error Handling**: Enhanced multipart form parsing with better error messages
  - Added detailed logging for file parsing operations
  - Added warnings for non-base64 encoded bodies that may contain binary data
  - Added file information logging including size, encoding status, and content type

## [1.1.0] - 2025-01-03

### Fixed
- **Image Rendering Engine**: Fixed critical image rendering issues by aligning with proven render-layout-go-lambda implementation
  - **Image Drawing Logic**: Fixed improper image scaling and positioning
    - Replaced `DrawImageAnchored()` with proper scaling sequence: `Push()`, `Translate()`, `Scale()`, `DrawImage()`, `Pop()`
    - Images now correctly scale and position within vending machine layout cells
  - **Text Rendering**: Added missing `splitTextToLines` function for proper product name wrapping
    - Product names now wrap intelligently across maximum 2 lines with ellipsis for overflow
    - Fixed positioning to display below product images instead of at cell bottom
    - Added fallback to "Sản phẩm" for empty product names
  - **Placeholder Handling**: Improved image unavailable placeholders
    - Changed placeholder text from "No Image" to "Image Unavailable"
    - Fixed placeholder color to proper gray (0.588) and centered positioning
  - **Image Loading**: Enhanced image loading and caching reliability
    - Increased HTTP timeout from 10s to 20s for better image loading success
    - Added proper cache directory creation with `os.MkdirAll()`
    - Improved error handling and cache file management
  - **Footer Rendering**: Fixed footer font styling
    - Changed from regular font (14.0pt) to bold font (18.0pt) for consistency

### Changed
- **Code Modernization**: Updated deprecated functions for Go 1.16+ compatibility
  - Replaced `ioutil.ReadAll()` with `io.ReadAll()`
  - Replaced `ioutil.WriteFile()` with `os.WriteFile()`
  - Updated import statements to remove deprecated `io/ioutil` package

### Technical Details
- **Files Modified**:
  - `renderer/renderer.go` - Complete image rendering engine overhaul
    - Fixed image scaling and positioning logic
    - Added missing `splitTextToLines` function
    - Improved product name rendering with text wrapping
    - Enhanced image loading with better timeout and caching
    - Updated placeholder handling and footer styling
    - Modernized deprecated function calls

### Testing
- ✅ **Compilation**: Code compiles successfully without errors
- ✅ **Unit Tests**: All Go tests pass
- ✅ **Dependencies**: `go mod tidy` runs cleanly
- ✅ **Build Process**: Docker build process works correctly

### Benefits
- **Improved Reliability**: Images now render consistently without scaling issues
- **Better User Experience**: Product names display properly with intelligent text wrapping
- **Enhanced Error Handling**: Better fallbacks when images fail to load
- **Future-Proof Code**: Updated to use modern Go standard library functions
- **Consistent Styling**: Footer and text rendering matches reference implementation

### Migration Notes
- **No Breaking Changes**: API interface remains unchanged
- **Backward Compatibility**: All existing functionality preserved
- **Automatic Deployment**: Changes take effect immediately upon Lambda function update
- **No Configuration Changes**: Environment variables and settings remain the same

---

## [1.0.0] - 2024-12-15

### Added
- **Initial Release**: Combined file upload and JSON layout rendering functionality
- **File Upload API**: Support for multiple file types with configurable upload paths
- **JSON Layout Rendering**: Automatic rendering of vending machine layouts to PNG images
- **S3 Integration**: Organized file storage with date-based path structure
- **DynamoDB Integration**: Optional metadata storage for rendered layouts
- **CORS Support**: Cross-origin request support for web applications
- **Environment Configuration**: Flexible configuration via Lambda environment variables

### Features
- **Multi-Bucket Support**: Reference and checking bucket configurations
- **File Validation**: Size limits and type checking
- **Error Handling**: Comprehensive error responses and logging
- **Image Caching**: HTTP image caching for performance optimization
- **Layout Validation**: JSON structure validation for vending machine layouts
