# Detailed Modular Architecture Design: ECS-Based Vending Machine Verification Service

## 1. System Architecture Details

### 1.1 Complete Component Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                                  Client Applications                                     │
└─────────────────────────────────────────┬───────────────────────────────────────────────┘
                                          │
                                          ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                                  Application Load Balancer                               │
└─────────────────────────────────────────┬───────────────────────────────────────────────┘
                                          │
                                          ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                                   ECS Service (Fargate)                                  │
│                                                                                         │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐    │
│  │                               Presentation Layer                                 │    │
│  │  ┌──────────────┐  ┌───────────────┐  ┌──────────────┐  ┌───────────────────┐   │    │
│  │  │HTTP Server   │  │API Router     │  │Controllers   │  │Request/Response    │   │    │
│  │  │& Middleware  │  │               │  │              │  │Handlers            │   │    │
│  │  └──────────────┘  └───────────────┘  └──────────────┘  └───────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────────────────────┘    │
│                                          │                                               │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐    │
│  │                               Application Layer                                  │    │
│  │  ┌──────────────┐  ┌───────────────┐  ┌──────────────┐  ┌───────────────────┐   │    │
│  │  │Verification  │  │Process        │  │Result        │  │Notification       │   │    │
│  │  │Orchestrator  │  │Manager        │  │Processor     │  │Service            │   │    │
│  │  └──────────────┘  └───────────────┘  └──────────────┘  └───────────────────┘   │    │
│  │  ┌──────────────┐  ┌───────────────┐  ┌──────────────┐  ┌───────────────────┐   │    │
│  │  │Image Service │  │Workflow       │  │Error         │  │Configuration      │   │    │
│  │  │              │  │Service        │  │Handler       │  │Service            │   │    │
│  │  └──────────────┘  └───────────────┘  └──────────────┘  └───────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────────────────────┘    │
│                                          │                                               │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐    │
│  │                                 Domain Layer                                     │    │
│  │  ┌──────────────┐  ┌───────────────┐  ┌──────────────┐  ┌───────────────────┐   │    │
│  │  │Verification  │  │Reference      │  │Discrepancy   │  │Visualization      │   │    │
│  │  │Engine        │  │Analyzer       │  │Detector      │  │Generator          │   │    │
│  │  └──────────────┘  └───────────────┘  └──────────────┘  └───────────────────┘   │    │
│  │  ┌──────────────┐  ┌───────────────┐  ┌──────────────┐  ┌───────────────────┐   │    │
│  │  │Prompt        │  │Response       │  │Domain Models │  │Validation Rules   │   │    │
│  │  │Generator     │  │Analyzer       │  │              │  │                   │   │    │
│  │  └──────────────┘  └───────────────┘  └──────────────┘  └───────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────────────────────┘    │
│                                          │                                               │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐    │
│  │                              Infrastructure Layer                                │    │
│  │  ┌──────────────┐  ┌───────────────┐  ┌──────────────┐  ┌───────────────────┐   │    │
│  │  │Bedrock       │  │DynamoDB       │  │S3            │  │Image Processing   │   │    │
│  │  │Provider      │  │Repository     │  │Repository    │  │Service            │   │    │
│  │  └──────────────┘  └───────────────┘  └──────────────┘  └───────────────────┘   │    │
│  │  ┌──────────────┐  ┌───────────────┐  ┌──────────────┐  ┌───────────────────┐   │    │
│  │  │Metrics       │  │Logging        │  │Cache         │  │Secret             │   │    │
│  │  │Collector     │  │Provider       │  │Provider      │  │Manager            │   │    │
│  │  └──────────────┘  └───────────────┘  └──────────────┘  └───────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────────────────────┘    │
│                                                                                         │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                │                     │                    │                     │
                ▼                     ▼                    ▼                     ▼
┌───────────────────┐  ┌───────────────────┐  ┌───────────────────┐  ┌───────────────────┐
│   Amazon Bedrock  │  │  Amazon DynamoDB  │  │   Amazon S3       │  │  Amazon CloudWatch│
└───────────────────┘  └───────────────────┘  └───────────────────┘  └───────────────────┘
```

### 1.2 Inter-Layer Communication Rules

- **Presentation Layer**: Can only communicate with Application Layer
- **Application Layer**: Can communicate with Presentation Layer and Domain Layer
- **Domain Layer**: Can communicate with Application Layer and defines interfaces for Infrastructure Layer
- **Infrastructure Layer**: Implements interfaces defined by Domain Layer

### 1.3 Dependency Injection Strategy

All components use dependency injection to maintain loose coupling:
- Constructor-based injection for required dependencies
- Interface-based dependencies rather than concrete implementations
- Centralized dependency registration and resolution
- Environment-specific dependency configurations

## 2. Detailed Module Specifications

### 2.1 Presentation Layer Modules

#### 2.1.1 HTTP Server Module

**Responsibilities:**
- Configure HTTP server with appropriate middleware
- Set up CORS, compression, and security headers
- Initialize request parsing and body handling
- Configure rate limiting and request timeout handling
- Register routes and controllers
- Set up error handling middleware

**Key Components:**
- ServerConfiguration: Configures server parameters
- MiddlewareRegistry: Registers middleware in correct order
- ErrorMiddleware: Centralizes HTTP error handling
- SecurityMiddleware: Implements security best practices

#### 2.1.2 API Router Module

**Responsibilities:**
- Define API route structure and versioning
- Map HTTP routes to controller methods
- Apply route-specific middleware
- Handle route parameter validation
- Define API documentation endpoints

**Key Components:**
- RouteRegistry: Registers and manages API routes
- VersionManager: Handles API versioning
- RouteValidator: Validates route parameters
- MiddlewareApplicator: Applies middleware to specific routes

#### 2.1.3 Controllers Module

**Responsibilities:**
- Handle incoming HTTP requests
- Validate request inputs
- Invoke appropriate application services
- Format and return responses
- Handle controller-level errors

**Key Components:**
- **VerificationController**:
  - POST /api/v1/verification: Initiate verification
  - GET /api/v1/verification/:id: Get verification status/results
  - GET /api/v1/verification: List verifications with filtering

- **HealthController**:
  - GET /health: Basic health check for ALB
  - GET /health/details: Detailed system health status

- **ResultsController**:
  - GET /api/v1/verification/:id/image: Get result visualization
  - GET /api/v1/verification/:id/discrepancies: Get detailed discrepancies

#### 2.1.4 Request/Response Handlers Module

**Responsibilities:**
- Standardize request validation
- Implement consistent response formatting
- Handle content negotiation
- Implement pagination for list endpoints
- Format error responses

**Key Components:**
- RequestValidator: Validates incoming request data
- ResponseFormatter: Formats API responses
- PaginationHandler: Implements pagination logic
- ErrorResponseFormatter: Formats error responses

### 2.2 Application Layer Modules

#### 2.2.1 Verification Orchestrator Module

**Responsibilities:**
- Coordinate the end-to-end verification workflow
- Manage state transitions between verification stages
- Track verification progress
- Handle retry logic for failed steps
- Coordinate between different services

**Key Components:**
- VerificationService: Primary service for verification operations
- VerificationInitializer: Handles verification setup
- VerificationStateManager: Manages verification state
- VerificationCompleter: Finalizes verification results

#### 2.2.2 Process Manager Module

**Responsibilities:**
- Implement state machine for verification workflow
- Track verification process state
- Persist state transitions in DynamoDB
- Handle process resumption after failures
- Implement timeout handling

**Key Components:**
- StateMachine: Defines verification workflow states
- StateTransitionManager: Manages state transitions
- StateRepository: Persists state information
- TimeoutManager: Handles process timeouts

#### 2.2.3 Result Processor Module

**Responsibilities:**
- Process and analyze results from each verification turn
- Merge Turn 1 and Turn 2 results
- Calculate verification statistics
- Format finalized verification results
- Generate result summaries

**Key Components:**
- Turn1Processor: Processes reference layout analysis
- Turn2Processor: Processes checking image analysis
- ResultMerger: Combines results from both turns
- StatisticsCalculator: Computes verification metrics

#### 2.2.4 Image Service Module

**Responsibilities:**
- Manage image retrieval from S3
- Prepare images for AI processing
- Validate image content and format
- Optimize images for Bedrock processing
- Handle image metadata

**Key Components:**
- ImageRetriever: Fetches images from S3
- ImageValidator: Validates image properties
- ImageOptimizer: Optimizes images for AI processing
- ImageMetadataExtractor: Extracts metadata from images

#### 2.2.5 Workflow Service Module

**Responsibilities:**
- Define workflow step sequences
- Track workflow progress
- Implement workflow retries
- Handle parallel workflow steps
- Provide workflow audit trail

**Key Components:**
- WorkflowDefinition: Defines workflow steps
- WorkflowExecutor: Executes workflow steps
- WorkflowMonitor: Monitors workflow execution
- RetryManager: Manages step retries

#### 2.2.6 Error Handler Module

**Responsibilities:**
- Implement application-wide error handling
- Categorize and normalize errors
- Implement error recovery strategies
- Log detailed error information
- Transform technical errors to domain errors

**Key Components:**
- ErrorHandler: Centralized error handling
- ErrorClassifier: Categorizes errors by type
- ErrorRecoveryStrategyFactory: Creates error recovery strategies
- ErrorLogger: Logs detailed error information

#### 2.2.7 Configuration Service Module

**Responsibilities:**
- Manage application configuration
- Handle environment-specific settings
- Implement configuration validation
- Support dynamic configuration updates
- Securely manage sensitive configuration

**Key Components:**
- ConfigurationProvider: Provides configuration values
- ConfigurationValidator: Validates configuration
- ConfigurationRefresher: Handles configuration updates
- SecureConfigurationManager: Manages sensitive configuration

### 2.3 Domain Layer Modules

#### 2.3.1 Verification Engine Module

**Responsibilities:**
- Implement core two-turn verification logic
- Define verification business rules
- Execute verification algorithms
- Handle verification edge cases
- Define verification outcomes

**Key Components:**
- TwoTurnVerificationEngine: Implements two-turn approach
- VerificationRules: Defines verification business rules
- VerificationAlgorithm: Implements verification algorithms
- OutcomeEvaluator: Evaluates verification outcomes

#### 2.3.2 Reference Analyzer Module

**Responsibilities:**
- Analyze reference layout images
- Extract product information from reference images
- Map products to machine positions
- Identify expected product configurations
- Create structured reference layout model

**Key Components:**
- ReferenceImageAnalyzer: Analyzes reference images
- ProductExtractor: Extracts product information
- PositionMapper: Maps products to positions
- LayoutStructureAnalyzer: Analyzes layout structure

#### 2.3.3 Discrepancy Detector Module

**Responsibilities:**
- Detect discrepancies between reference and checking images
- Classify discrepancy types (position, identity, quantity, visibility)
- Calculate confidence scores for discrepancies
- Generate discrepancy evidence
- Prioritize discrepancies by severity

**Key Components:**
- DiscrepancyDetector: Detects discrepancies
- DiscrepancyClassifier: Classifies discrepancy types
- ConfidenceCalculator: Calculates confidence scores
- EvidenceGenerator: Generates discrepancy evidence
- SeverityEvaluator: Evaluates discrepancy severity

#### 2.3.4 Visualization Generator Module

**Responsibilities:**
- Generate visual representations of verification results
- Highlight discrepancies in images
- Create side-by-side comparisons
- Generate visual annotations
- Format visualization for different outputs (web, PDF, etc.)

**Key Components:**
- VisualizationGenerator: Generates visualizations
- DiscrepancyHighlighter: Highlights discrepancies
- ComparisonGenerator: Creates side-by-side comparisons
- AnnotationGenerator: Generates annotations
- OutputFormatter: Formats for different outputs

#### 2.3.5 Prompt Generator Module

**Responsibilities:**
- Generate optimized prompts for Bedrock
- Create system prompts with verification instructions
- Create turn-specific user prompts
- Incorporate layout and product information
- Optimize prompts for token efficiency

**Key Components:**
- SystemPromptGenerator: Generates system prompts
- Turn1PromptGenerator: Generates Turn 1 prompts
- Turn2PromptGenerator: Generates Turn 2 prompts
- LayoutFormatter: Formats layout information for prompts
- TokenOptimizer: Optimizes prompts for token efficiency

#### 2.3.6 Response Analyzer Module

**Responsibilities:**
- Parse and analyze Bedrock responses
- Extract structured data from AI responses
- Validate response against expected schema
- Handle malformed or unexpected responses
- Extract confidence information

**Key Components:**
- ResponseParser: Parses AI responses
- StructuredDataExtractor: Extracts structured data
- ResponseValidator: Validates response schema
- MalformedResponseHandler: Handles unexpected responses
- ConfidenceExtractor: Extracts confidence information

#### 2.3.7 Domain Models Module

**Responsibilities:**
- Define core domain entities and value objects
- Implement domain-specific validation rules
- Define relationships between domain objects
- Implement domain event definitions
- Support domain-driven design principles

**Key Components:**
- See Detailed Domain Models section below

#### 2.3.8 Validation Rules Module

**Responsibilities:**
- Define domain-specific validation rules
- Implement validation logic
- Define validation error messages
- Support complex validation scenarios
- Provide validation context

**Key Components:**
- ValidationRule: Base for all validation rules
- ValidationContext: Provides validation context
- ValidationRuleSet: Groups related validation rules
- ValidationResult: Represents validation outcome
- ValidationErrorFactory: Creates validation errors

### 2.4 Infrastructure Layer Modules

#### 2.4.1 Bedrock Provider Module

**Responsibilities:**
- Interface with Amazon Bedrock API
- Handle Bedrock request/response formatting
- Manage Bedrock authentication
- Implement retry logic for Bedrock API
- Monitor and optimize Bedrock usage

**Key Components:**
- BedrockClient: Primary client for Bedrock API
- BedrockRequestFormatter: Formats Bedrock requests
- BedrockResponseParser: Parses Bedrock responses
- BedrockRetryStrategy: Implements retry logic
- BedrockUsageMonitor: Monitors API usage

#### 2.4.2 DynamoDB Repository Module

**Responsibilities:**
- Implement data access layer for DynamoDB
- Map domain objects to DynamoDB items
- Implement query patterns for efficient access
- Handle DynamoDB pagination
- Implement optimistic concurrency control

**Key Components:**
- VerificationRepository: Manages verification records
- LayoutRepository: Manages layout records
- QueryBuilder: Builds DynamoDB queries
- ItemMapper: Maps between domain objects and DynamoDB items
- IndexSelector: Selects appropriate DynamoDB indexes

#### 2.4.3 S3 Repository Module

**Responsibilities:**
- Manage S3 object storage and retrieval
- Handle S3 permissions and access control
- Generate presigned URLs for access
- Implement object lifecycle management
- Optimize S3 operations for performance

**Key Components:**
- ImageRepository: Manages image objects
- ResultRepository: Manages result objects
- S3Client: Interfaces with S3 API
- PresignedUrlGenerator: Generates presigned URLs
- ObjectLifecycleManager: Manages object lifecycle

#### 2.4.4 Image Processing Service Module

**Responsibilities:**
- Process and transform images
- Resize and crop images as needed
- Optimize images for different use cases
- Extract image metadata
- Create image annotations

**Key Components:**
- ImageProcessor: Processes images
- ImageResizer: Resizes images
- ImageOptimizer: Optimizes images
- MetadataExtractor: Extracts image metadata
- AnnotationCreator: Creates image annotations

#### 2.4.5 Metrics Collector Module

**Responsibilities:**
- Collect application performance metrics
- Monitor resource utilization
- Track business metrics
- Implement custom metrics
- Publish metrics to CloudWatch

**Key Components:**
- MetricsCollector: Collects metrics
- MetricsPublisher: Publishes metrics
- MetricsFormatter: Formats metrics data
- DimensionBuilder: Builds metric dimensions
- CustomMetricsRegistry: Registers custom metrics

#### 2.4.6 Logging Provider Module

**Responsibilities:**
- Implement structured logging
- Configure log levels and destinations
- Add context information to logs
- Handle sensitive data in logs
- Format logs for analysis

**Key Components:**
- Logger: Primary logging interface
- LogFormatter: Formats log entries
- LogContext: Provides logging context
- SensitiveDataMasker: Masks sensitive data
- LogDestinationManager: Manages log destinations

#### 2.4.7 Cache Provider Module

**Responsibilities:**
- Implement caching for performance optimization
- Manage cache invalidation
- Handle distributed caching
- Implement cache eviction policies
- Monitor cache performance

**Key Components:**
- CacheProvider: Primary caching interface
- CacheKeyGenerator: Generates cache keys
- CacheInvalidator: Invalidates cache entries
- EvictionPolicyManager: Manages eviction policies
- CacheMonitor: Monitors cache performance

#### 2.4.8 Secret Manager Module

**Responsibilities:**
- Securely manage sensitive credentials
- Rotate credentials as needed
- Interface with AWS Secrets Manager
- Encrypt sensitive data
- Control access to secrets

**Key Components:**
- SecretProvider: Provides access to secrets
- SecretRotator: Rotates credentials
- SecretsManagerClient: Interfaces with AWS Secrets Manager
- SecretEncryptor: Encrypts sensitive data
- SecretAccessController: Controls secret access

## 3. Detailed Domain Models

### 3.1 Primary Domain Entities

#### 3.1.1 Verification

```
VerificationEntity {
  verificationId: string                    // Unique identifier
  status: VerificationStatus                // Current status
  createdAt: Date                           // Creation timestamp
  updatedAt: Date                           // Last update timestamp
  vendingMachineId: string                  // Vending machine identifier
  layoutId: number                          // Layout identifier
  layoutPrefix: string                      // Layout version indicator
  referenceImageUrl: string                 // S3 URL for reference image
  checkingImageUrl: string                  // S3 URL for checking image
  resultImageUrl?: string                   // S3 URL for result image
  turnResults: TurnResult[]                 // Results for each turn
  discrepancies: Discrepancy[]              // Detected discrepancies
  verificationSummary: VerificationSummary  // Summary of results
  metadata: VerificationMetadata            // Additional metadata
}
```

#### 3.1.2 VerificationContext

```
VerificationContext {
  verificationId: string                    // Unique identifier
  status: VerificationStatus                // Current status
  turnConfig: TurnConfig                    // Turn configuration
  currentTurn: number                       // Current turn number
  vendingMachineId: string                  // Vending machine identifier
  layoutId: number                          // Layout identifier
  layoutPrefix: string                      // Layout version indicator
  referenceImageUrl: string                 // S3 URL for reference image
  checkingImageUrl: string                  // S3 URL for checking image
  turnTimestamps: {                         // Timestamps for each turn
    initialized: Date
    turn1?: Date
    turn2?: Date
    completed?: Date
  }
  notificationEnabled: boolean              // Whether to send notifications
  processingMetadata: {                     // Processing metadata
    requestId: string                       // Original request ID
    startTime: Date                         // Processing start time
    retryCount: number                      // Number of retries
    timeout: number                         // Processing timeout
  }
}
```

#### 3.1.3 ReferenceAnalysis (Turn 1 Result)

```
ReferenceAnalysis {
  turnNumber: number                        // Always 1 for reference analysis
  machineStructure: MachineStructure        // Machine structure details
  rowAnalysis: {                            // Analysis by row
    [rowId: string]: {
      description: string                   // Human-readable description
      productsFound: ProductPosition[]      // Products found in this row
    }
  }
  productPositions: {                       // Products by position
    [positionId: string]: {
      productName: string                   // Product name
      visible: boolean                      // Whether product is visible
    }
  }
  emptyPositions: string[]                  // Empty positions in reference
  confidence: number                        // Confidence score (0-100)
  initialConfirmation: string               // Confirmation of structure
  originalResponse: string                  // Original AI response text
  completedAt: Date                         // Completion timestamp
}
```

#### 3.1.4 CheckingAnalysis (Turn 2 Result)

```
CheckingAnalysis {
  turnNumber: number                        // Always 2 for checking analysis
  verificationStatus: VerificationStatus    // Overall status (CORRECT/INCORRECT)
  discrepancies: Discrepancy[]              // Detected discrepancies
  totalDiscrepancies: number                // Total number of discrepancies
  severity: DiscrepancySeverity             // Overall severity
  rowAnalysis: {                            // Analysis by row
    [rowId: string]: {
      description: string                   // Human-readable description
      status: RowStatus                     // Row status
    }
  }
  emptySlotReport: {                        // Empty slot analysis
    referenceEmptyRows: string[]            // Rows expected to be empty
    checkingEmptyRows: string[]             // Rows found empty
    checkingPartiallyEmptyRows: string[]    // Rows partially empty
    checkingEmptyPositions: string[]        // Positions found empty
    totalEmpty: number                      // Total empty positions
  }
  confidence: number                        // Confidence score (0-100)
  originalResponse: string                  // Original AI response text
  completedAt: Date                         // Completion timestamp
}
```

#### 3.1.5 Discrepancy

```
Discrepancy {
  position: string                          // Position ID (e.g., "A01")
  expected: string                          // Expected product
  found: string                             // Actual product found
  issue: DiscrepancyType                    // Type of discrepancy
  confidence: number                        // Confidence score (0-100)
  evidence: string                          // Evidence description
  verificationResult: VerificationStatus    // Result for this position
  severity: DiscrepancySeverity             // Severity of this discrepancy
  imageCoordinates?: {                      // Visual coordinates
    x: number                               // X coordinate
    y: number                               // Y coordinate
    width: number                           // Width
    height: number                          // Height
  }
}
```

#### 3.1.6 MachineStructure

```
MachineStructure {
  rowCount: number                          // Number of rows
  columnsPerRow: number                     // Number of columns per row
  rowOrder: string[]                        // Row identifiers in order
  columnOrder: string[]                     // Column identifiers in order
  physicalOrientation: {                    // Physical orientation
    topRow: string                          // Identifier for top row
    leftColumn: string                      // Identifier for leftmost column
    rowDirection: 'topToBottom' | 'bottomToTop' // Row direction
    columnDirection: 'leftToRight' | 'rightToLeft' // Column direction
  }
}
```

#### 3.1.7 VerificationResult

```
VerificationResult {
  verificationId: string                    // Unique identifier
  verificationAt: Date                      // Verification timestamp
  status: VerificationStatus                // Overall status
  vendingMachineId: string                  // Vending machine identifier
  layoutId: number                          // Layout identifier
  layoutPrefix: string                      // Layout version indicator
  referenceImageUrl: string                 // S3 URL for reference image
  checkingImageUrl: string                  // S3 URL for checking image
  resultImageUrl: string                    // S3 URL for result image
  machineStructure: MachineStructure        // Machine structure
  initialConfirmation: string               // Structure confirmation
  correctedRows: string[]                   // Rows with no discrepancies
  emptySlotReport: {                        // Empty slot report
    referenceEmptyRows: string[]            // Rows expected to be empty
    checkingEmptyRows: string[]             // Rows found empty
    checkingPartiallyEmptyRows: string[]    // Rows partially empty
    checkingEmptyPositions: string[]        // Positions found empty
    totalEmpty: number                      // Total empty positions
  }
  referenceStatus: {                        // Reference layout status
    [rowId: string]: string                 // Description by row
  }
  checkingStatus: {                         // Checking image status
    [rowId: string]: string                 // Description by row
  }
  discrepancies: Discrepancy[]              // Detected discrepancies
  verificationSummary: {                    // Summary of results
    totalPositionsChecked: number           // Total positions checked
    correctPositions: number                // Correct positions
    discrepantPositions: number             // Positions with discrepancies
    missingProducts: number                 // Missing products
    incorrectProductTypes: number           // Incorrect product types
    unexpectedProducts: number              // Unexpected products
    emptyPositionsCount: number             // Empty positions
    overallAccuracy: number                 // Accuracy percentage
    overallConfidence: number               // Confidence score
    verificationStatus: VerificationStatus  // Overall status
    verificationOutcome: string             // Human-readable outcome
  }
  metadata: {                               // Additional metadata
    bedrockModel: string                    // Bedrock model used
    completedAt: Date                       // Completion timestamp
    processingTime: number                  // Processing time in ms
    tokenUsage: {                           // Token usage
      input: number                         // Input tokens
      output: number                        // Output tokens
      total: number                         // Total tokens
    }
  }
}
```

### 3.2 Value Objects and Enumerations

#### 3.2.1 VerificationStatus Enum

```
enum VerificationStatus {
  INITIALIZED = 'INITIALIZED',
  IMAGES_FETCHED = 'IMAGES_FETCHED',
  SYSTEM_PROMPT_READY = 'SYSTEM_PROMPT_READY',
  TURN1_PROMPT_READY = 'TURN1_PROMPT_READY',
  TURN1_PROCESSING = 'TURN1_PROCESSING',
  TURN1_COMPLETED = 'TURN1_COMPLETED',
  TURN2_PROMPT_READY = 'TURN2_PROMPT_READY',
  TURN2_PROCESSING = 'TURN2_PROCESSING',
  TURN2_COMPLETED = 'TURN2_COMPLETED',
  RESULTS_FINALIZED = 'RESULTS_FINALIZED',
  RESULTS_STORED = 'RESULTS_STORED',
  NOTIFICATION_SENT = 'NOTIFICATION_SENT',
  ERROR = 'ERROR',
  PARTIAL_RESULTS = 'PARTIAL_RESULTS',
  CORRECT = 'CORRECT',           // For position/verification result
  INCORRECT = 'INCORRECT'        // For position/verification result
}
```

#### 3.2.2 DiscrepancyType Enum

```
enum DiscrepancyType {
  INCORRECT_PRODUCT_TYPE = 'Incorrect Product Type',
  MISSING_PRODUCT = 'Missing Product',
  UNEXPECTED_PRODUCT = 'Unexpected Product',
  INCORRECT_POSITION = 'Incorrect Position',
  INCORRECT_QUANTITY = 'Incorrect Quantity',
  INCORRECT_ORIENTATION = 'Incorrect Orientation',
  LABEL_NOT_VISIBLE = 'Label Not Visible'
}
```

#### 3.2.3 DiscrepancySeverity Enum

```
enum DiscrepancySeverity {
  LOW = 'Low',
  MEDIUM = 'Medium',
  HIGH = 'High'
}
```

#### 3.2.4 RowStatus Enum

```
enum RowStatus {
  CORRECT = 'Correct',
  INCORRECT = 'Incorrect',
  PARTIALLY_CORRECT = 'Partially Correct',
  EMPTY = 'Empty',
  PARTIALLY_EMPTY = 'Partially Empty',
  NOT_VISIBLE = 'Not Visible'
}
```

#### 3.2.5 TurnConfig Value Object

```
TurnConfig {
  maxTurns: number                          // Maximum number of turns (2)
  referenceImageTurn: number                // Turn for reference image (1)
  checkingImageTurn: number                 // Turn for checking image (2)
  turnTimeout: number                       // Timeout per turn in ms
}
```

#### 3.2.6 ProductPosition Value Object

```
ProductPosition {
  position: string                          // Position ID (e.g., "A01")
  product: string                           // Product name
  isPresent: boolean                        // Whether product is present
  quantity?: number                         // Quantity if specified
  confidence?: number                       // Confidence score
}
```

## 4. Detailed Workflow Sequences

### 4.1 Complete Two-Turn Verification Workflow

```
┌─────────────┐      ┌─────────────────┐      ┌───────────────┐      ┌─────────────┐      ┌─────────────┐
│ Controller  │      │ Verification    │      │ Image         │      │ DynamoDB    │      │ S3          │
│             │      │ Service         │      │ Service       │      │ Repository  │      │ Repository  │
└──────┬──────┘      └────────┬────────┘      └───────┬───────┘      └──────┬──────┘      └──────┬──────┘
       │                      │                        │                     │                    │
       │ initiateVerification │                        │                     │                    │
       │─────────────────────>│                        │                     │                    │
       │                      │                        │                     │                    │
       │                      │ generateVerificationId │                     │                    │
       │                      │◀────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
       │                      │  createInitialRecord   │                     │                    │
       │                      │─────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
       │  return verification │                        │                     │                    │
       │<─────────────────────│                        │                     │                    │
       │      context         │                        │                     │                    │
       │                      │                        │                     │                    │
       │                      │       fetchImages      │                     │                    │
       │                      │───────────────────────>│                     │                    │
       │                      │                        │                     │                    │
       │                      │                        │     fetchImages     │                    │
       │                      │                        │────────────────────────────────────────>│
       │                      │                        │                     │                    │
       │                      │                        │                     │                    │
       │                      │                        │    return images    │                    │
       │                      │                        │<───────────────────────────────────────┐│
       │                      │                        │                     │                  ││
       │                      │                        │                     │                  ││
       │                      │                        │   fetchMetadata     │                  ││
       │                      │                        │────────────────────>│                  ││
       │                      │                        │                     │                  ││
       │                      │                        │  return metadata    │                  ││
       │                      │                        │<────────────────────│                  ││
       │                      │                        │                     │                  ││
       │                      │  return images+metadata│                     │                  ││
       │                      │<──────────────────────│                     │                  ││
       │                      │                        │                     │                  ││
       │                      │   updateStatus         │                     │                  ││
       │                      │    IMAGES_FETCHED      │                     │                  ││
       │                      │─────────────────────────────────────────────>│                  ││
       │                      │                        │                     │                  ││
       │                      │                        │                     │                  ││
       │                      │                        │                     │                  ││
┌──────┴──────┐      ┌────────┴────────┐      ┌───────┴───────┐      ┌──────┴──────┐      ┌──────┴──────┐
│ Controller  │      │ Verification    │      │ Image         │      │ DynamoDB    │      │ S3          │
│             │      │ Service         │      │ Service       │      │ Repository  │      │ Repository  │
└─────────────┘      └────────┬────────┘      └───────────────┘      └─────────────┘      └─────────────┘
                              │                                              │
                              │                                              │
┌─────────────┐      ┌────────┴────────┐      ┌───────────────┐      ┌──────┴──────┐      ┌─────────────┐
│ Prompt      │      │ Verification    │      │ Bedrock       │      │ DynamoDB    │      │ Reference   │
│ Generator   │      │ Service         │      │ Provider      │      │ Repository  │      │ Analyzer    │
└──────┬──────┘      └────────┬────────┘      └───────┬───────┘      └──────┬──────┘      └──────┬──────┘
       │                      │                        │                     │                    │
       │   generateSystemPrompt                        │                     │                    │
       │<─────────────────────│                        │                     │                    │
       │                      │                        │                     │                    │
       │  return systemPrompt │                        │                     │                    │
       │─────────────────────>│                        │                     │                    │
       │                      │                        │                     │                    │
       │                      │   updateStatus         │                     │                    │
       │                      │  SYSTEM_PROMPT_READY   │                     │                    │
       │                      │─────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
       │  generateTurn1Prompt │                        │                     │                    │
       │<─────────────────────│                        │                     │                    │
       │                      │                        │                     │                    │
       │   return turn1Prompt │                        │                     │                    │
       │─────────────────────>│                        │                     │                    │
       │                      │                        │                     │                    │
       │                      │   updateStatus         │                     │                    │
       │                      │  TURN1_PROMPT_READY    │                     │                    │
       │                      │─────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
       │                      │    invokeModel         │                     │                    │
       │                      │───────────────────────>│                     │                    │
       │                      │                        │                     │                    │
       │                      │                        │                     │                    │
       │                      │   return turn1Response │                     │                    │
       │                      │<──────────────────────│                     │                    │
       │                      │                        │                     │                    │
       │                      │   updateStatus         │                     │                    │
       │                      │  TURN1_PROCESSING      │                     │                    │
       │                      │─────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
       │                      │                      processTurn1Response    │                    │
       │                      │────────────────────────────────────────────────────────────────>│
       │                      │                        │                     │                  ││
       │                      │                        │                     │                  ││
       │                      │                       return referenceAnalysis                  ││
       │                      │<───────────────────────────────────────────────────────────────┘│
       │                      │                        │                     │                    │
       │                      │   storeReferenceAnalysis                     │                    │
       │                      │─────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
       │                      │   updateStatus         │                     │                    │
       │                      │  TURN1_COMPLETED       │                     │                    │
       │                      │─────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
┌──────┴──────┐      ┌────────┴────────┐      ┌───────┴───────┐      ┌──────┴──────┐      ┌──────┴──────┐
│ Prompt      │      │ Verification    │      │ Bedrock       │      │ DynamoDB    │      │ Reference   │
│ Generator   │      │ Service         │      │ Provider      │      │ Repository  │      │ Analyzer    │
└─────────────┘      └────────┬────────┘      └───────────────┘      └─────────────┘      └─────────────┘
                              │                                              │
                              │                                              │
┌─────────────┐      ┌────────┴────────┐      ┌───────────────┐      ┌──────┴──────┐      ┌─────────────┐
│ Prompt      │      │ Verification    │      │ Bedrock       │      │ DynamoDB    │      │ Discrepancy │
│ Generator   │      │ Service         │      │ Provider      │      │ Repository  │      │ Detector    │
└──────┬──────┘      └────────┬────────┘      └───────┬───────┘      └──────┬──────┘      └──────┬──────┘
       │                      │                        │                     │                    │
       │  generateTurn2Prompt │                        │                     │                    │
       │<─────────────────────│                        │                     │                    │
       │                      │                        │                     │                    │
       │   return turn2Prompt │                        │                     │                    │
       │─────────────────────>│                        │                     │                    │
       │                      │                        │                     │                    │
       │                      │   updateStatus         │                     │                    │
       │                      │  TURN2_PROMPT_READY    │                     │                    │
       │                      │─────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
       │                      │    invokeModel         │                     │                    │
       │                      │───────────────────────>│                     │                    │
       │                      │                        │                     │                    │
       │                      │                        │                     │                    │
       │                      │   return turn2Response │                     │                    │
       │                      │<──────────────────────│                     │                    │
       │                      │                        │                     │                    │
       │                      │   updateStatus         │                     │                    │
       │                      │  TURN2_PROCESSING      │                     │                    │
       │                      │─────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
       │                      │                      processTurn2Response    │                    │
       │                      │────────────────────────────────────────────────────────────────>│
       │                      │                        │                     │                  ││
       │                      │                        │                     │                  ││
       │                      │                       return checkingAnalysis                   ││
       │                      │<───────────────────────────────────────────────────────────────┘│
       │                      │                        │                     │                    │
       │                      │   storeCheckingAnalysis                      │                    │
       │                      │─────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
       │                      │   updateStatus         │                     │                    │
       │                      │  TURN2_COMPLETED       │                     │                    │
       │                      │─────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
┌──────┴──────┐      ┌────────┴────────┐      ┌───────┴───────┐      ┌──────┴──────┐      ┌──────┴──────┐
│ Prompt      │      │ Verification    │      │ Bedrock       │      │ DynamoDB    │      │ Discrepancy │
│ Generator   │      │ Service         │      │ Provider      │      │ Repository  │      │ Detector    │
└─────────────┘      └────────┬────────┘      └───────────────┘      └─────────────┘      └─────────────┘
                              │                                              │
                              │                                              │
┌─────────────┐      ┌────────┴────────┐      ┌───────────────┐      ┌──────┴──────┐      ┌─────────────┐
│ Result      │      │ Verification    │      │ Visualization │      │ DynamoDB    │      │ S3          │
│ Processor   │      │ Service         │      │ Generator     │      │ Repository  │      │ Repository  │
└──────┬──────┘      └────────┬────────┘      └───────┬───────┘      └──────┬──────┘      └──────┬──────┘
       │                      │                        │                     │                    │
       │   finalizeResults    │                        │                     │                    │
       │<─────────────────────│                        │                     │                    │
       │                      │                        │                     │                    │
       │  return finalResults │                        │                     │                    │
       │─────────────────────>│                        │                     │                    │
       │                      │                        │                     │                    │
       │                      │  storeFinalResults     │                     │                    │
       │                      │─────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
       │                      │   updateStatus         │                     │                    │
       │                      │  RESULTS_FINALIZED     │                     │                    │
       │                      │─────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
       │                      │  generateVisualization │                     │                    │
       │                      │───────────────────────>│                     │                    │
       │                      │                        │                     │                    │
       │                      │                        │    storeResultImage │                    │
       │                      │                        │────────────────────────────────────────>│
       │                      │                        │                     │                    │
       │                      │                        │                     │                    │
       │                      │                        │   return resultImageUrl                  │
       │                      │                        │<───────────────────────────────────────┐│
       │                      │                        │                     │                  ││
       │                      │                        │                     │                  ││
       │                      │  return resultImageUrl │                     │                  ││
       │                      │<──────────────────────│                     │                  ││
       │                      │                        │                     │                  ││
       │                      │  storeResultImageUrl   │                     │                  ││
       │                      │─────────────────────────────────────────────>│                  ││
       │                      │                        │                     │                  ││
       │                      │   updateStatus         │                     │                  ││
       │                      │  RESULTS_STORED        │                     │                  ││
       │                      │─────────────────────────────────────────────>│                  ││
       │                      │                        │                     │                  ││
┌──────┴──────┐      ┌────────┴────────┐      ┌───────┴───────┐      ┌──────┴──────┐      ┌──────┴──────┐
│ Result      │      │ Verification    │      │ Visualization │      │ DynamoDB    │      │ S3          │
│ Processor   │      │ Service         │      │ Generator     │      │ Repository  │      │ Repository  │
└─────────────┘      └────────┬────────┘      └───────────────┘      └─────────────┘      └─────────────┘
                              │                                                                  │
                              │                                                                  │
┌─────────────┐      ┌────────┴────────┐                                                         │
│ Notification│      │ Verification    │                                                         │
│ Service     │      │ Service         │                                                         │
└──────┬──────┘      └────────┬────────┘                                                         │
       │                      │                                                                  │
       │   sendNotifications  │                                                                  │
       │<─────────────────────│                                                                  │
       │                      │                                                                  │
       │  notificationResults │                                                                  │
       │─────────────────────>│                                                                  │
       │                      │                                                                  │
       │                      │ updateStatus           │                     │                    │
       │                      │ NOTIFICATION_SENT      │                     │                    │
       │                      │─────────────────────────────────────────────>│                    │
       │                      │                        │                     │                    │
┌──────┴──────┐      ┌────────┴────────┐                                                         │
│ Notification│      │ Verification    │                                                         │
│ Service     │      │ Service         │                                                         │
└─────────────┘      └─────────────────┘                                                         │
```

### 4.2 Error Handling Sequence for Bedrock Failure

```
┌─────────────┐      ┌─────────────────┐      ┌───────────────┐      ┌─────────────┐      ┌─────────────┐
│ Verification│      │ Bedrock         │      │ Error         │      │ DynamoDB    │      │ Finalize    │
│ Service     │      │ Provider        │      │ Handler       │      │ Repository  │      │ With Error  │
└──────┬──────┘      └────────┬────────┘      └───────┬───────┘      └──────┬──────┘      └──────┬──────┘
       │                      │                        │                     │                    │
       │    invokeModel       │                        │                     │                    │
       │─────────────────────>│                        │                     │                    │
       │                      │                        │                     │                    │
       │                      │                        │                     │                    │
       │      Error Response  │                        │                     │                    │
       │<─────────────────────│                        │                     │                    │
       │                      │                        │                     │                    │
       │                      │                        │                     │                    │
       │   handleBedrockError │                        │                     │                    │
       │───────────────────────────────────────────────>                     │                    │
       │                      │                        │                     │                    │
       │                      │                        │   logError          │                    │
       │                      │                        │────────────────────>│                    │
       │                      │                        │                     │                    │
       │                      │                        │ determineRecoveryStrategy               │
       │                      │                        │────────────────────────────────────────>│
       │                      │                        │                     │                  ││
       │                      │                        │                     │                  ││
       │                      │                      recovery recommendation │                  ││
       │                      │                      <─────────────────────────────────────────┘│
       │                      │                        │                     │                    │
       │ error handling result│                        │                     │                    │
       │<──────────────────────────────────────────────│                     │                    │
       │                      │                        │                     │                    │
       │                      │                        │                     │                    │
       │  finalizeWithError   │                        │                     │                    │
       │─────────────────────────────────────────────────────────────────────────────────────────>
       │                      │                        │                     │                    │
       │                      │                        │                     │                    │
       │                      │                        │                     │                    │
       │   partial results    │                        │                     │                    │
       │<─────────────────────────────────────────────────────────────────────────────────────────
       │                      │                        │                     │                    │
       │                      │                        │                     │                    │
       │   storePartialResults│                        │                     │                    │
       │─────────────────────────────────────────────────────────────────────>                    │
       │                      │                        │                     │                    │
       │   updateStatus       │                        │                     │                    │
       │  PARTIAL_RESULTS     │                        │                     │                    │
       │─────────────────────────────────────────────────────────────────────>                    │
       │                      │                        │                     │                    │
┌──────┴──────┐      ┌────────┴────────┐      ┌───────┴───────┐      ┌──────┴──────┐      ┌──────┴──────┐
│ Verification│      │ Bedrock         │      │ Error         │      │ DynamoDB    │      │ Finalize    │
│ Service     │      │ Provider        │      │ Handler       │      │ Repository  │      │ With Error  │
└─────────────┘      └─────────────────┘      └───────────────┘      └─────────────┘      └─────────────┘
```

## 5. Database Design and Persistence Strategy

### 5.1 DynamoDB Table Design

#### 5.1.1 VerificationResults Table

**Primary Key Structure:**
- Partition Key: `verificationId` (String)
- Sort Key: `verificationAt` (String) - ISO 8601 timestamp

**Secondary Indexes:**
- GSI1:
  - Partition Key: `layoutId` (Number)
  - Sort Key: `verificationAt` (String)
  - Projection: ALL
  - Purpose: Retrieve history of verifications for a specific layoutID

- GSI2:
  - Partition Key: `verificationStatus` (String)
  - Sort Key: `verificationAt` (String)
  - Projection: INCLUDE (vendingMachineId, location, verificationSummary)
  - Purpose: Identify all incorrect verifications within a time period

**Time-To-Live (TTL) Attribute:**
- `expirationTime` (Number): Epoch timestamp for record expiration

**Item Size Optimization:**
- Large attributes stored in S3 with references in DynamoDB
- Compressed text attributes for responses
- Separate verification context from verification results

#### 5.1.2 LayoutMetadata Table

**Primary Key Structure:**
- Partition Key: `layoutId` (Number)
- Sort Key: `layoutPrefix` (String)

**Secondary Indexes:**
- GSI1:
  - Partition Key: `createdAt` (String)
  - Sort Key: `layoutId` (Number)
  - Projection: KEYS_ONLY
  - Purpose: Time-based analysis of layout creation

- GSI2:
  - Partition Key: `vendingMachineId` (String)
  - Sort Key: `createdAt` (String)
  - Projection: ALL
  - Purpose: Retrieve all layouts for a specific vending machine

### 5.2 S3 Bucket Structure

#### 5.2.1 Reference Bucket

**Structure:**
```
/raw
  /{year}/{month}/{day}/
    layout_{layoutId}.json
/processed
  /{year}/{month}/{day}/{timestamp}/
    {layoutId}_{layoutPrefix}/
      image.png
      metadata.json
/logs
  /{year}/{month}/{day}/
    {layoutId}_{layoutPrefix}.log
```

**Lifecycle Policies:**
- Raw files: 12-month retention
- Processed files: 12-month retention
- Logs: 3-month retention

#### 5.2.2 Checking Bucket

**Structure:**
```
/{year}/{month}/{day}/
  {vendingMachineId}/
    check_{timestamp}.jpg
    metadata.json
```

**Lifecycle Policies:**
- Checking images: 12-month retention
- Metadata files: 12-month retention

#### 5.2.3 Results Bucket

**Structure:**
```
/{year}/{month}/{day}/
  {verificationId}/
    result.jpg
    discrepancies.json
    reference_checking_comparison.jpg
    highlighted_discrepancies.jpg
```

**Lifecycle Policies:**
- Result files: 6-month retention
- Detailed discrepancy files: 6-month retention

### 5.3 Caching Strategy

#### 5.3.1 Layout Metadata Caching

- In-memory cache for frequently accessed layouts
- TTL: 1 hour
- Invalidation on layout update
- Cache capacity: 1000 layouts

#### 5.3.2 Image Caching

- In-memory LRU cache for recently accessed images
- TTL: 30 minutes
- Maximum cached image size: 5MB
- Cache capacity: 200 images

#### 5.3.3 Verification Result Caching

- In-memory cache for recent verification results
- TTL: 15 minutes
- Invalidation on result update
- Cache capacity: 500 verification results

## 6. Comprehensive Error Handling

### 6.1 Error Taxonomy

**Validation Errors:**
- InvalidInputError: Invalid input parameters
- MissingResourceError: Required resource not found
- ValidationError: Input fails validation rules

**Service Errors:**
- BedrockAPIError: Error from Bedrock API
- DynamoDBError: Error from DynamoDB operations
- S3Error: Error from S3 operations
- TimeoutError: Operation timeout

**Processing Errors:**
- ImageProcessingError: Error processing images
- PromptGenerationError: Error generating prompts
- ResponseParsingError: Error parsing AI responses
- VisualizationError: Error generating visualizations

**System Errors:**
- ConfigurationError: Error in system configuration
- DependencyError: Error in system dependencies
- ResourceExhaustionError: System resource exhaustion

### 6.2 Error Recovery Strategies

**Strategy Types:**
- RetryStrategy: Retry the operation with backoff
- FallbackStrategy: Use alternative approach
- CompensationStrategy: Undo previous operations
- PartialResultStrategy: Return partial results
- NotificationStrategy: Notify admins of error

**Strategy Selection Criteria:**
- Error type and severity
- Current process state
- Available alternatives
- Impact on user experience
- Resource availability

### 6.3 Error Logging

**Log Structure:**
```json
{
  "timestamp": "2025-04-21T15:30:45Z",
  "level": "ERROR",
  "service": "VerificationService",
  "module": "BedrockProvider",
  "operation": "invokeModel",
  "verificationId": "verif-2025042115302500",
  "error": {
    "type": "BedrockAPIError",
    "message": "Bedrock throttling exception",
    "code": "ThrottlingException",
    "details": {
      "retryable": true,
      "recommendedBackoffMs": 1000
    }
  },
  "context": {
    "turn": 2,
    "status": "TURN2_PROCESSING",
    "elapsedTimeMs": 3245
  },
  "stack": "Error: Bedrock throttling exception\n    at BedrockProvider.invokeModel (/app/src/infrastructure/bedrock/bedrock-provider.js:125:23)\n    ..."
}
```

## 7. Security Design

### 7.1 Network Security

**VPC Configuration:**
- Private subnets for ECS tasks
- VPC endpoints for AWS services
- Security groups with minimal access
- Network ACLs for additional security

**ALB Security:**
- HTTPS only (TLS 1.2+)
- Security headers (HSTS, CSP, X-XSS-Protection)
- WAF integration for request filtering
- Rate limiting and throttling

### 7.2 Data Security

**Data at Rest:**
- S3 bucket encryption (SSE-S3)
- DynamoDB encryption at rest
- ECS task encryption
- Secrets encrypted with KMS

**Data in Transit:**
- TLS 1.2+ for all communications
- HTTPS for API endpoints
- VPC endpoints for AWS services
- Private subnets for internal traffic

**Data Classification:**
- Public: Non-sensitive data (e.g., public layouts)
- Internal: Business data (e.g., verification results)
- Confidential: Sensitive business data (e.g., vending machine locations)
- Restricted: Credentials and secrets

### 7.3 Access Control

**IAM Roles:**
- ECS Task Role: Minimum permissions for task execution
- S3 Access Role: Read/write access to specific buckets
- DynamoDB Access Role: Access to specific tables
- Bedrock Access Role: Access to specific models

**Access Policies:**
- Least privilege principle
- Resource-based policies
- Condition-based restrictions
- Time-bound access

### 7.4 Secrets Management

**Secret Types:**
- API credentials
- Database credentials
- Encryption keys
- Service account credentials

**Secret Handling:**
- No secrets in code or environment variables
- AWS Secrets Manager integration
- Secret rotation policies
- Secure secret distribution

## 8. Performance Optimization

### 8.1 Image Optimization

**Size Optimization:**
- Resize large images before processing
- JPEG compression for checking images
- WebP format for result visualizations
- Thumbnail generation for UI display

**Content Optimization:**
- Crop to relevant areas
- Normalize lighting and contrast
- Remove unnecessary metadata
- Focus on product-containing regions

### 8.2 Bedrock Token Optimization

**Prompt Optimization:**
- Concise, focused prompts
- Minimal repetition of information
- Structured format for consistent parsing
- Clear instructions for specific outputs

**Image Token Optimization:**
- Optimize image resolution for Bedrock
- Use appropriate image format
- Crop images to relevant areas
- Compress images without losing detail

**Response Optimization:**
- Request structured JSON responses
- Limit response scope to necessary information
- Specify output format clearly
- Set appropriate max token limits

### 8.3 Parallel Processing

**Parallel Operations:**
- Concurrent image fetching
- Parallel metadata retrieval
- Asynchronous result storage
- Concurrent notification sending

**Resource Limits:**
- Configurable concurrency limits
- Resource-aware scheduling
- Dynamic throttling based on load
- Circuit breakers for dependent services

### 8.4 Caching Strategy

**Cache Levels:**
- In-memory cache for frequent access
- Distributed cache for shared data
- S3 for persistent cache
- Client-side caching with ETags

**Cache Invalidation:**
- Time-based expiration
- Event-based invalidation
- Selective cache updates
- Versioned cache keys

## 9. Testing Strategy

### 9.1 Unit Testing

**Test Areas:**
- Individual components
- Business logic
- Data transformations
- Validation rules

**Testing Approach:**
- Isolated component testing
- Mock dependencies
- Focus on edge cases
- High code coverage

### 9.2 Integration Testing

**Test Areas:**
- Component interactions
- External service integration
- Data persistence
- Workflow execution

**Testing Approach:**
- Local integration environment
- Service mocks for external dependencies
- Real data structures
- Focus on component boundaries

### 9.3 Functional Testing

**Test Areas:**
- End-to-end verification workflows
- API contracts
- User scenarios
- Error handling

**Testing Approach:**
- Automated API tests
- Scenario-based testing
- Regression test suite
- Performance validations

### 9.4 AI Model Testing

**Test Areas:**
- Prompt effectiveness
- Response parsing
- Handling of edge cases
- Model performance metrics

**Testing Approach:**
- Synthetic test cases
- Golden dataset comparisons
- A/B testing of prompts
- Response quality metrics

## 10. Deployment and Operations

### 10.1 Deployment Pipeline

**Pipeline Stages:**
1. Build container image
2. Run unit tests
3. Deploy to staging environment
4. Run integration tests
5. Run performance tests
6. Approval gate
7. Deploy to production
8. Post-deployment validation

**Deployment Artifacts:**
- Docker image
- Infrastructure as Code templates
- Configuration files
- Migration scripts

### 10.2 Monitoring and Alerting

**Metrics Categories:**
- Application health
- Performance metrics
- Business metrics
- Resource utilization

**Alerting Thresholds:**
- Error rate > 5%
- Verification latency > 30 seconds
- API latency > 500ms
- Resource utilization > 80%
- Bedrock throttling events

**Alert Channels:**
- PagerDuty for critical alerts
- Slack for operational alerts
- Email for summary reports
- Dashboard for real-time monitoring

### 10.3 Scaling Strategy

**Scaling Dimensions:**
- ECS task count
- Container memory and CPU
- DynamoDB capacity
- Bedrock API quota

**Scaling Triggers:**
- CPU utilization > 70%
- Memory utilization > 70%
- Queue depth > 50
- Processing latency > 15 seconds

**Scaling Policies:**
- Target tracking for predictable loads
- Step scaling for rapid changes
- Scheduled scaling for known patterns
- Manual scaling for exceptional cases

## Conclusion

This detailed design provides a comprehensive foundation for implementing the ECS-based vending machine verification service. The modular architecture with clean separation of concerns enables:

1. **Maintainability**: Each component has a single responsibility
2. **Testability**: Components can be tested in isolation
3. **Scalability**: Components can scale independently
4. **Flexibility**: Components can be replaced or upgraded individually
5. **Reliability**: Robust error handling and recovery strategies

The implementation follows industry best practices for cloud-native applications, ensuring security, performance, and operational excellence while delivering the core two-turn verification capability for vending machine products.