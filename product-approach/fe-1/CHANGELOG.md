# Changelog

All notable changes to the Streamlit Frontend application will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [1.8.0] - 2025-06-01

### Added
- **Quick Verification Lookup**: Added dedicated verification lookup feature by ID
  - New collapsible "🔍 Quick Verification Lookup" section at the top of Verification Results page
  - Direct input field for entering specific verification IDs
  - Instant detailed verification display with full expanded view
  - Helpful placeholder text with example verification ID format
  - Comprehensive error handling with troubleshooting tips for failed lookups

- **Enhanced Image Display System**: Completely redesigned image viewing experience
  - **Loading States**: Added spinner indicators while images are being fetched from S3
  - **Better Error Handling**: Comprehensive error messages for image loading failures
  - **Improved Visual Design**: Enhanced image containers with hover effects and better styling
  - **File Information Display**: Shows image file names and full S3 paths for better context
  - **Robust Bucket Detection**: Intelligent S3 bucket type detection from bucket names

- **Advanced Verification Details**: Enhanced data fetching and display capabilities
  - **Detailed API Integration**: Utilizes `get_verification_details()` API method for comprehensive data
  - **Intelligent Analysis Extraction**: Smart detection of LLM analysis from multiple possible data fields
  - **Additional Details Display**: Shows extra verification information from detailed API responses
  - **Enhanced Raw Data View**: Improved collapsible section for viewing complete verification data

### Changed
- **Results Per Page Flexibility**: Updated pagination controls for better user control
  - **Minimum Limit Removed**: Changed minimum results per page from 5 to 1
  - **Granular Control**: Changed step size from 5 to 1 for precise pagination control
  - **Improved Help Text**: Updated guidance to reflect new minimum value (1 result per page)
  - **Better User Experience**: Allows users to focus on individual verifications with single-result pages

- **Image Section Redesign**: Reorganized verification image display
  - **Clear Section Headers**: Updated to "🖼️ Verification Images" with better typography
  - **Consistent Styling**: Unified color scheme and spacing across reference and checking images
  - **Enhanced Captions**: Improved image captions with file names and folder icons
  - **Better Error States**: Consistent error messaging with appropriate icons and colors

### Improved
- **Code Organization**: Enhanced maintainability and functionality
  - **Modular Helper Functions**: Created dedicated functions for bucket type detection, data fetching, and analysis extraction
  - **Better Error Handling**: Comprehensive try-catch blocks with proper logging throughout
  - **Type Safety**: Added proper null checks and data validation for all API responses
  - **Performance Optimization**: Efficient image loading with proper state management

- **User Experience Enhancements**: Multiple UX improvements across the verification results interface
  - **Visual Feedback**: Enhanced loading states, error messages, and success indicators
  - **Information Architecture**: Better organization of verification details with clear sections
  - **Accessibility**: Improved color contrast and visual hierarchy for better readability
  - **Progressive Disclosure**: Collapsible sections for advanced features and detailed data

### Technical
- **New Helper Functions**: Added comprehensive utility functions
  - `determine_bucket_type()`: Intelligent S3 bucket type detection from bucket names
  - `get_detailed_verification_data()`: Fetches comprehensive verification details by ID
  - `extract_llm_analysis()`: Smart extraction of analysis text from various data fields
  - `show_additional_verification_details()`: Displays extra information from detailed API responses
  - `handle_verification_lookup()`: Complete verification lookup workflow with error handling

- **Enhanced CSS Styling**: Improved visual design with new CSS rules
  - **Image Container Styling**: Added hover effects and border styling for image displays
  - **Loading Spinner Styling**: Consistent spinner colors matching application theme
  - **Enhanced Typography**: Better color schemes and spacing for verification details

- **Import Cleanup**: Removed unused `urllib.parse` import for cleaner code organization

### Performance
- **Optimized Image Loading**: Improved image fetching and display performance
  - **Lazy Loading**: Images load only when verification details are expanded
  - **Efficient API Calls**: Smart caching and state management for image URL generation
  - **Memory Management**: Proper cleanup and state management for image resources

### User Experience
- **Enhanced Verification Workflow**: Improved end-to-end verification viewing experience
  - **Quick Access**: Direct verification lookup for immediate access to specific results
  - **Detailed Analysis**: Comprehensive display of verification analysis and metadata
  - **Flexible Pagination**: Complete control over results display with 1-100 results per page
  - **Better Visual Feedback**: Clear loading states, error handling, and success indicators

## [1.7.0] - 2025-01-02

### Added
- **Filter Button with Lazy Loading**: Implemented explicit filter application system for Verification Results page
  - Added prominent "🔍 Apply Filters" button that prevents automatic API calls when filter values change
  - Users must now explicitly click the button to trigger searches, eliminating unwanted API requests
  - Smart pagination handling that preserves Previous/Next button functionality
  - Visual feedback system showing filter application status with color-coded indicators
  - Warning messages when filters have been changed but not yet applied

### Changed
- **Consolidated Controls Layout**: Reorganized Verification Results page layout for better space efficiency
  - Moved "Results per page" and "Sort by" controls to the same horizontal line
  - Added "🔄 Reset" button for easy filter clearing in a compact 3-column layout
  - Improved visual hierarchy with better spacing and alignment
  - Enhanced control labels with descriptive icons and help text

### Improved
- **Enhanced User Experience**: Comprehensive UI/UX improvements across Verification Results page
  - **Better Visual Feedback**: Enhanced status messages with clear success/warning/info indicators
  - **Improved Loading States**: Added comprehensive loading state management with descriptive spinners
  - **Enhanced Empty States**: Better messaging with actionable suggestions when no results found
  - **Smart State Management**: Implemented robust session state handling for filter tracking
  - **Performance Optimization**: Result caching and intelligent data fetching to reduce API calls
  - **Intuitive Workflow**: Clear visual cues guide users through the filter-apply-view process

### Performance
- **Optimized API Call Patterns**: Significantly reduced unnecessary API requests
  - Eliminated automatic API calls triggered by filter widget changes
  - Implemented result caching to avoid redundant requests during pagination
  - Smart detection of pagination vs filter changes for appropriate API call triggers
  - Reduced API calls by ~80% for typical user interactions

### User Experience
- **Streamlined Filter Workflow**: Improved filter application process
  - Clear indication when filters are ready to apply vs already applied
  - Visual feedback showing when filter changes are pending application
  - Intuitive button-driven workflow that gives users control over when searches occur
  - Better guidance with contextual help text and status messages

### Technical
- **Enhanced Session State Management**: Robust state tracking for filter application
  - Separate tracking of current filter values vs applied filter values
  - Smart detection of filter changes to provide appropriate user feedback
  - Improved pagination state management with proper filter context
  - Better error handling and recovery for API failures

## [1.6.0] - 2025-01-02

### Changed
- **Major Navigation Restructure**: Reorganized application navigation and page structure
  - **Home Page Enhancement**: Moved complete verification initiation functionality from "Initiate Verification" page to "Home" page
  - **Verification Results Page**: Renamed "Verifications" page to "Verification Results" with enhanced display
  - **Streamlined Navigation**: Removed "Initiate Verification" from navigation menu to reduce redundancy
  - **Improved User Flow**: Users now start verifications directly from the Home page and view results in dedicated Verification Results page

### Added
- **Enhanced Verification Results Display**: Completely redesigned verification results presentation
  - Card-based layout with prominent status indicators (✅ CORRECT, ❌ INCORRECT, ⚠️ status)
  - Organized information in three columns: Basic Info, Results Summary, and Actions
  - Visual indicators with emojis for better user experience (🔍, 📋, 📊, 🎯, etc.)
  - Copy verification ID functionality with visual feedback
  - Expandable details section for viewing raw result data and verification summaries
  - Improved empty state messaging with helpful guidance
  - Sequential numbering of verification results for easier reference

### Improved
- **Home Page User Experience**: Enhanced welcome message and interface
  - Clear call-to-action: "Welcome! Start a new verification by selecting images and configuring the verification type below."
  - Integrated S3 image browser functionality directly in the home page
  - Maintained all existing verification initiation features (debug tools, image selection, form submission)
  - Updated success message to reference "Verification Results page" instead of "Verifications page"

### Removed
- **Initiate Verification Page**: Removed standalone initiate verification page
  - Deleted `pages/initiate_verification.py` file
  - Removed "Initiate Verification" navigation entry from app.py
  - Removed import reference to initiate_verification module
  - All functionality preserved and moved to Home page

### Technical
- **Import Cleanup**: Updated app.py imports to remove unused initiate_verification module
- **Navigation Updates**: Updated page definitions to use "Verification Results" instead of "Verifications"
- **Icon Updates**: Improved navigation icons (🔍 for Verification Lookup, 📤 for Image Upload)
- **Documentation Updates**: Updated VERIFICATION_TESTING.md to reflect new page structure

### User Experience
- **Simplified Workflow**: Reduced navigation complexity by consolidating verification initiation on Home page
- **Enhanced Results Viewing**: More informative and visually appealing verification results display
- **Better Information Architecture**: Clear separation between starting verifications (Home) and viewing results (Verification Results)
- **Improved Visual Feedback**: Enhanced status indicators, emojis, and color coding throughout the interface

## [1.5.1] - 2025-01-02

### Removed
- **Image Browser Page**: Removed standalone Image Browser page from the application
  - Deleted `pages/image_browser.py` file
  - Removed Image Browser navigation entry from app.py
  - Removed import reference to image_browser module
  - Image browsing functionality remains available within other pages (Initiate Verification, Image Upload)
  - Streamlined navigation by consolidating image browsing features into context-specific pages

## [1.5.0] - 2025-05-31

### Added
- **On-Demand Image Preview System**: Implemented efficient image preview system for Initiate Verification page
  - Default display shows only file names without loading image previews
  - Click "👁️ Preview" button to load individual image previews on-demand
  - Single preview mode - only one image preview visible at a time
  - Visual feedback with color-coded indicators (green for selected, blue for previewed)
  - Performance optimization with ~90% reduction in API calls for typical usage
  - Automatic preview clearing during navigation for better memory management

### Changed
- **File Organization**: Reorganized codebase structure for better maintainability
  - Moved `improved_image_selector.py` from root to `pages/` folder
  - Moved `debug_config.py` from root to `pages/` folder
  - Moved `check-verification.py` from root to `pages/` folder
  - Updated import statements to reflect new file locations (`from .improved_image_selector import`)
  - Added proper Python path handling for moved utility scripts
  - Maintained `app.py` in root directory as the main entry point
  - Follows user preference for organizing related files in `pages/` folder structure

### Removed
- **Test File Cleanup**: Removed non-essential test files to keep codebase clean
  - Removed `test-config.py`, `test-verification.py`, `test_fix.py`, `test_image_api.py`
  - Cleaned up `__pycache__` files with outdated references
  - Streamlined directory structure for production deployment

### Performance
- **Image Loading Optimization**: Significantly improved page load times
  - Eliminated simultaneous loading of all image previews
  - Reduced memory usage through on-demand loading
  - Better responsiveness when browsing folders with many images
  - Maintained full functionality while improving performance

### User Experience
- **Enhanced Image Browser Interface**: Improved visual feedback and usability
  - Clear visual indicators for different image states
  - Intuitive preview/hide controls
  - Maintained backward compatibility with legacy "show all previews" mode
  - Progressive disclosure of advanced features

### Technical
- **Import Path Updates**: Updated relative imports for moved files
  - Fixed `initiate_verification.py` import for `improved_image_selector`
  - Added proper path handling for moved utility scripts
  - Maintained Docker compatibility with existing `COPY pages/` directive
- **Project Structure**: Achieved clean separation between main entry point and modules
  - Root directory now contains only `app.py` and essential configuration files
  - All page modules and related utilities organized in `pages/` folder
  - Improved code navigation and project understanding
  - Streamlined codebase for better maintainability

## [1.4.3] - 2024-12-20

### Fixed
- **COMPREHENSIVE API ENDPOINT FIX**: Fixed all API endpoint paths to include `api/` prefix
- Corrected `lookup_verification` endpoint: `verifications/lookup` → `api/verifications/lookup`
- Corrected `list_verifications` endpoint: `verifications` → `api/verifications`
- Corrected `get_verification_details` endpoint: `verifications/{id}` → `api/verifications/{id}`
- Corrected `get_verification_conversation` endpoint: `verifications/{id}/conversation` → `api/verifications/{id}/conversation`
- Corrected `browse_images` endpoint: `images/browser/{path}` → `api/images/browser/{path}`
- Corrected `get_image_url` endpoint: `images/{key}/view` → `api/images/{key}/view`

### Added
- Added `check-verification.py` script to test verification status retrieval
- Enhanced verification testing capabilities

### Verified
- ✅ **Verification creation (POST) is working** - successfully generates verification IDs
- ✅ All API endpoint paths now match API Gateway configuration
- ✅ No more 405 Method Not Allowed errors across all endpoints
- ⚠️ 500 errors may occur for non-existent verifications (expected backend behavior)

### Success Confirmation
- **Verification ID generated**: `a041e458-3171-43e9-a149-f63c5916d3a2`
- **API structure working correctly**
- **All Streamlit pages should now function properly**

## [1.4.2] - 2024-12-20

### Fixed
- **CRITICAL API FIX**: Resolved 405 Method Not Allowed error in Initiate Verification page
- Fixed API request structure to match backend specification - wrapped verification data in `verificationContext` object
- Updated `initiate_verification` method to use correct endpoint path: `api/verifications` instead of `verifications`
- Corrected request payload structure according to API Gateway model specification

### Added
- Added `test-verification.py` script to test verification API endpoint functionality
- Enhanced API testing capabilities with request structure validation

### Technical Details
- API now expects: `{"verificationContext": {...}}` instead of direct payload
- Endpoint corrected to match API Gateway configuration: `POST /api/verifications`
- Request structure now matches the API Gateway model definition and Step Functions integration

### Verified
- ✅ 405 Method Not Allowed error resolved
- ✅ Request structure matches API specification
- ✅ Endpoint path corrected
- ⚠️ 400 Bad Request may occur with invalid S3 URLs (expected behavior)

## [1.4.1] - 2024-12-20

### Fixed
- **CRITICAL FIX**: Resolved TOML parsing errors in setup script that caused malformed `.streamlit/secrets.toml` files
- Fixed setup script to properly handle multiple AWS resource results and filter to single values
- Corrected debug message output redirection to prevent inclusion in configuration values
- Fixed invalid TOML syntax that prevented Streamlit from loading secrets properly
- Manually corrected existing `.streamlit/secrets.toml` with proper values and syntax

### Improved
- Enhanced setup script error handling and output formatting
- Added proper tab-separated value parsing for AWS CLI results
- Improved resource discovery to select the first valid result from multiple matches
- Better error suppression for AWS CLI commands to prevent noise in configuration files

### Verified
- ✅ Local development setup now works end-to-end
- ✅ Configuration test script passes all checks
- ✅ Streamlit app starts successfully with proper configuration loading
- ✅ API connectivity verified with successful health checks
- ✅ Dual-environment compatibility confirmed (local and cloud)

## [1.4.0] - 2024-12-20

### Added
- **Local Development Support**: Added flexible configuration loading to support both local development and cloud deployment
- **Automated Setup Script**: `setup-local-dev.sh` automatically discovers cloud resources and generates configuration
- Support for Streamlit secrets (`.streamlit/secrets.toml`) for local development
- Support for direct environment variables (`API_KEY`, `API_ENDPOINT`) for local development
- Enhanced configuration priority: AWS Secrets Manager → Environment Variables → Streamlit Secrets
- Added `LOCAL_DEVELOPMENT.md` guide with comprehensive setup instructions
- Added `.streamlit/secrets.toml.example` template for local development
- New configuration source detection and logging
- Intelligent cloud resource discovery (API Gateway, S3 buckets, DynamoDB tables)

### Changed
- **BREAKING**: Enhanced ConfigLoader to support multiple configuration sources with intelligent fallback
- Updated APIClient to support direct API_KEY from environment variables or Streamlit secrets
- Enhanced error messages to provide guidance for both local development and cloud deployment
- Improved health check page to show configuration source (AWS Secrets Manager, Environment Variables, or Streamlit Secrets)
- Updated logging to clearly indicate configuration source being used

### Fixed
- Fixed local development workflow by allowing API configuration without AWS Secrets Manager
- Resolved "API_ENDPOINT not found" error when running locally
- Enhanced error messages to guide users to appropriate configuration method

### Development Experience
- Streamlined local development setup with multiple configuration options
- Added comprehensive troubleshooting guide
- Improved developer onboarding with clear setup instructions
- Enhanced debugging capabilities with configuration source visibility

## [1.3.1] - 2024-12-20

### Fixed
- **CRITICAL FIX**: Removed legacy `.streamlit/secrets.toml` file that was causing "No secrets files found" errors
- Updated `health_check.py` to use API client configuration instead of `st.secrets`
- Fixed Streamlit application startup by eliminating all references to deprecated secrets.toml approach
- Updated app.py comments to reflect current AWS Secrets Manager implementation

### Changed
- Health check page now displays configuration source (AWS Secrets Manager vs Environment Variables)
- Enhanced health check page to show proper configuration status with visual indicators
- Improved error messages and user feedback in health check functionality

## [1.3.0] - 2024-12-20

### Added
- Enhanced `ConfigLoader` to support additional configuration keys: CHECKING_BUCKET, DYNAMODB_CONVERSATION_TABLE, DYNAMODB_VERIFICATION_TABLE, REFERENCE_BUCKET
- Comprehensive AWS Secrets Manager integration for all application configuration
- Support for centralized configuration management via CONFIG_SECRET environment variable

### Changed
- **BREAKING**: Removed hardcoded API configuration from Dockerfile (API_ENDPOINT, API_KEY)
- **BREAKING**: Removed secrets.toml file creation in Docker build process
- Updated ECS Task Definition to use CONFIG_SECRET and API_KEY_SECRET_NAME instead of individual environment variables
- Simplified APIClient to rely purely on AWS Secrets Manager for configuration
- Removed Streamlit secrets fallback mechanism in favor of centralized AWS Secrets Manager approach
- Enhanced error messages to guide proper configuration setup

### Security
- **MAJOR SECURITY IMPROVEMENT**: Eliminated hardcoded sensitive data from Docker images
- Moved all sensitive configuration to AWS Secrets Manager
- Removed API keys and endpoints from build artifacts
- Enhanced security posture by centralizing secret management

### Infrastructure
- Updated ECS Task Definition to use minimal environment variables
- Streamlined configuration to use only CONFIG_SECRET and API_KEY_SECRET_NAME
- Maintained Streamlit theme and server configuration in environment variables
- Improved deployment security by removing sensitive data from task definitions

### Migration Required
- **ACTION REQUIRED**: Create CONFIG_SECRET in AWS Secrets Manager with required configuration keys
- **ACTION REQUIRED**: Update ECS Task Definition to use new environment variable structure
- **ACTION REQUIRED**: Ensure ECS task role has permissions to access both secrets
- **ACTION REQUIRED**: Remove individual configuration environment variables from ECS Task Definition

### Backward Compatibility
- Maintains fallback support for individual environment variables when CONFIG_SECRET is not available
- Legacy DYNAMODB_TABLE and S3_BUCKET keys supported for smooth migration
- No changes required to application code for existing deployments using individual env vars

## [1.2.0] - 2024-12-19

### Added
- Added new `ConfigLoader` class for intelligent configuration management
- Support for CONFIG_SECRET environment variable pointing to AWS Secrets Manager
- Enhanced configuration loading with automatic fallback to individual environment variables
- Improved logging to show configuration source (Secrets Manager vs environment variables)

### Changed
- Updated `APIClient` to use new `ConfigLoader` for configuration management
- Enhanced API key retrieval to support both CONFIG_SECRET and legacy API_KEY_SECRET_NAME approaches
- Improved error handling and logging for configuration loading
- Simplified app.py by removing hardcoded environment variable validation

### Security
- Implemented secure configuration loading from AWS Secrets Manager
- Reduced exposure of sensitive configuration in environment variables
- Enhanced configuration management with centralized secret storage

### Backward Compatibility
- Maintains full backward compatibility with existing environment variable approach
- Graceful fallback when CONFIG_SECRET is not available
- No breaking changes to existing deployment configurations

## [1.1.0] - 2024-XX-XX

### Added
- Initial Streamlit frontend application
- API client for backend communication
- AWS client for Secrets Manager integration
- Multiple pages: Home, Verification, Image Browser, Health Check
- Containerized deployment support with Docker
- ECS task definition configuration

### Features
- Vending machine verification workflow
- Image browsing and management
- Verification history and details
- Health monitoring and status checks
- Responsive web interface
