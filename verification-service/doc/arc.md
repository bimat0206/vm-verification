# Modular Architecture Design: ECS-Based Vending Machine Verification Service

## 1. High-Level Architecture

```
┌─────────────────┐      ┌──────────────────┐      ┌──────────────────────────────────────┐      ┌─────────────────┐
│                 │      │                  │      │                                      │      │                 │
│  Web Frontend   │─────▶│   ALB (443/80)   │─────▶│  ECS Service (Verification Service)  │◀─────│  Amazon Bedrock │
│                 │      │                  │      │                                      │      │                 │
└─────────────────┘      └──────────────────┘      └──────────────────────────────────────┘      └─────────────────┘
                                                                     │                                  
                                                                     │                                  
                                 ┌────────────────────────────────────────────────────────┐                                  
                                 │                                                        │                                  
                                 ▼                                                        ▼                                  
                          ┌──────────────────┐                                    ┌─────────────────┐      
                          │                  │                                    │                 │      
                          │     DynamoDB     │                                    │    S3 Buckets   │      
                          │                  │                                    │                 │      
                          └──────────────────┘                                    └─────────────────┘      
```

## 2. Architectural Principles

1. **Clean Architecture**: Implement layered architecture with clear boundaries
2. **Separation of Concerns**: Each module has a single responsibility
3. **Dependency Inversion**: High-level modules don't depend on low-level modules
4. **Interface Segregation**: Specific interfaces for specific clients
5. **Domain-Driven Design**: Core business logic isolated from infrastructure concerns

## 3. Layered Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Presentation Layer                         │
│                                                                 │
│   HTTP Controllers | Request Validation | Response Formatting   │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Application Layer                          │
│                                                                 │
│  Orchestration | Process Management | Service Coordination      │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Domain Layer                             │
│                                                                 │
│  Core Business Logic | Domain Models | Business Rules           │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Infrastructure Layer                        │
│                                                                 │
│  External Services | Data Access | Technical Capabilities       │
└─────────────────────────────────────────────────────────────────┘
```

## 4. Modular Component Design

### 4.1 Presentation Layer Components

| Module | Responsibility |
|--------|----------------|
| **HTTP Server** | Container for HTTP endpoints and routing |
| **API Router** | Routes HTTP requests to appropriate controllers |
| **Verification Controller** | Handles verification API endpoints |
| **Health Controller** | Handles ALB health check endpoints |
| **Result Controller** | Handles result retrieval endpoints |
| **Request Validators** | Validates incoming HTTP requests |
| **Response Formatters** | Standardizes API response formats |
| **Error Handlers** | Manages API error responses |

### 4.2 Application Layer Components

| Module | Responsibility |
|--------|----------------|
| **Verification Service** | Orchestrates the verification workflow |
| **Process Manager** | Manages workflow state and transitions |
| **Result Processor** | Processes verification results |
| **Image Service** | Manages image retrieval and preparation |
| **Notification Service** | Handles notification delivery |
| **Configuration Service** | Manages application configuration |
| **Logging Service** | Centralizes application logging |

### 4.3 Domain Layer Components

| Module | Responsibility |
|--------|----------------|
| **Verification Engine** | Core two-turn verification logic |
| **Reference Analyzer** | Analyzes reference layout images |
| **Discrepancy Detector** | Detects discrepancies between images |
| **Prompt Generator** | Creates AI prompts for Bedrock |
| **Response Analyzer** | Analyzes AI model responses |
| **Domain Models** | Represents core business entities |
| **Domain Events** | Defines business events |
| **Domain Services** | Implements complex business logic |

### 4.4 Infrastructure Layer Components

| Module | Responsibility |
|--------|----------------|
| **Bedrock Provider** | Integrates with Amazon Bedrock |
| **S3 Repository** | Manages S3 storage operations |
| **DynamoDB Repository** | Manages DynamoDB operations |
| **Image Processor** | Handles image transformations |
| **Visualization Generator** | Creates result visualizations |
| **Metrics Collector** | Gathers application metrics |
| **Secret Manager** | Manages application secrets |
| **Cache Manager** | Provides caching capabilities |

## 5. Component Interaction for Two-Turn Verification

### 5.1 Turn 1: Reference Layout Analysis

```
┌────────────────────┐     ┌────────────────────┐     ┌────────────────────┐
│ Verification       │     │ Verification       │     │ Image              │
│ Controller         │────▶│ Service            │────▶│ Service            │
└────────────────────┘     └────────────────────┘     └────────────────────┘
                                     │                           │
                                     │                           ▼
┌────────────────────┐     ┌─────────┘                ┌────────────────────┐
│ Bedrock            │◀────┘                          │ S3                 │
│ Provider           │                                │ Repository         │
└────────────────────┘                                └────────────────────┘
         │                                                      │
         ▼                                                      │
┌────────────────────┐     ┌────────────────────┐     ┌────────┘
│ Prompt             │     │ Reference          │     │
│ Generator          │────▶│ Analyzer           │◀────┘
└────────────────────┘     └────────────────────┘
                                     │
                                     ▼
                           ┌────────────────────┐
                           │ DynamoDB           │
                           │ Repository         │
                           └────────────────────┘
```

### 5.2 Turn 2: Checking Image Verification

```
┌────────────────────┐     ┌────────────────────┐     ┌────────────────────┐
│ Process            │     │ Verification       │     │ Image              │
│ Manager            │────▶│ Service            │────▶│ Service            │
└────────────────────┘     └────────────────────┘     └────────────────────┘
                                     │                           │
                                     │                           ▼
┌────────────────────┐     ┌─────────┘                ┌────────────────────┐
│ Bedrock            │◀────┘                          │ S3                 │
│ Provider           │                                │ Repository         │
└────────────────────┘                                └────────────────────┘
         │                                                      │
         ▼                                                      │
┌────────────────────┐     ┌────────────────────┐     ┌────────┘
│ Prompt             │     │ Discrepancy        │     │
│ Generator          │────▶│ Detector           │◀────┘
└────────────────────┘     └────────────────────┘
                                     │
                                     ▼
                           ┌────────────────────┐     ┌────────────────────┐
                           │ DynamoDB           │────▶│ Visualization      │
                           │ Repository         │     │ Generator          │
                           └────────────────────┘     └────────────────────┘
```

## 6. Data Models and State Management

### 6.1 Key Domain Models

| Model | Description |
|-------|-------------|
| **VerificationContext** | Container for verification metadata and state |
| **ReferenceAnalysis** | Results from Turn 1 reference analysis |
| **CheckingAnalysis** | Results from Turn 2 discrepancy detection |
| **VerificationResult** | Combined final verification results |
| **Discrepancy** | Represents a detected discrepancy |
| **LayoutMetadata** | Information about the vending machine layout |
| **MachineStructure** | Physical structure of the vending machine |

### 6.2 State Management

**Verification States:**
- INITIALIZED
- IMAGES_FETCHED
- SYSTEM_PROMPT_READY
- TURN1_PROMPT_READY
- TURN1_PROCESSING
- TURN1_COMPLETED
- TURN2_PROMPT_READY
- TURN2_PROCESSING
- TURN2_COMPLETED
- RESULTS_FINALIZED
- RESULTS_STORED
- NOTIFICATION_SENT
- ERROR

**State Transitions:**
- Each state transition is persisted in DynamoDB
- Provides workflow resumability in case of failures
- Enables status tracking for client applications

## 7. Infrastructure Configuration

### 7.1 Container Structure

```
verification-service/
├─ src/
│  ├─ presentation/      # Presentation layer modules
│  ├─ application/       # Application layer modules
│  ├─ domain/            # Domain layer modules
│  ├─ infrastructure/    # Infrastructure layer modules
│  ├─ config/            # Application configuration
│  ├─ utils/             # Shared utilities
│  └─ index.js           # Application entry point
├─ Dockerfile            # Container definition
├─ package.json          # Dependencies
└─ config/               # External configuration files
```

### 7.2 ECS Task Configuration

Key configuration for ECS task:
- Task memory: 2GB
- Task CPU: 1 vCPU
- Task role with access to:
  - S3 (reference, checking, and results buckets)
  - DynamoDB (VerificationResults and LayoutMetadata tables)
  - Bedrock (Claude model)
  - CloudWatch Logs

### 7.3 ALB Configuration

- Health check path: `/health`
- Health check interval: 30 seconds
- Health check timeout: 5 seconds
- Target group protocol: HTTP
- Stickiness: Disabled (stateless by design)

## 8. Integration Patterns

### 8.1 Amazon Bedrock Integration

- Use Provider pattern to abstract Bedrock API
- Support multi-turn conversation tracking
- Implement retry and error handling strategies
- Monitor token usage and costs

### 8.2 Storage Integration

- Repository pattern for S3 and DynamoDB
- Each repository focused on specific data access needs
- Consistent error handling and retry policies
- Optimized query patterns for data retrieval

### 8.3 Asynchronous Processing

- Long-running verification jobs processed asynchronously
- Status polling endpoint for client applications
- Event-driven notifications for completion
- Support for webhook callbacks

## 9. Error Handling Strategy

### 9.1 Error Types

| Error Category | Description | Recovery Strategy |
|----------------|-------------|-------------------|
| **Validation Errors** | Invalid input parameters | Return error response with details |
| **Resource Errors** | Missing images or metadata | Log error and return resource not found |
| **Service Errors** | External service failures | Implement retries with backoff |
| **Bedrock Errors** | AI model errors | Fall back to alternative approaches |
| **System Errors** | Infrastructure failures | Circuit breaker and graceful degradation |

### 9.2 Error Recovery

- Transaction-based operations where possible
- Persistent state management for retry capability
- Partial results when complete processing fails
- Detailed error logging for diagnosis

## 10. Monitoring and Observability

### 10.1 Logging Strategy

- Structured logging with correlation IDs
- Log levels: ERROR, WARN, INFO, DEBUG, TRACE
- Contextual metadata in all log entries
- CloudWatch as centralized log storage

### 10.2 Metrics

- Request counts and latencies
- Process completion rates
- Bedrock token usage
- Error rates by category
- Resource utilization

### 10.3 Alarms

- High error rate alerts
- Long-running verification alerts
- Resource exhaustion alerts
- Service health alerts

## 11. Performance Considerations

### 11.1 Optimizations

- Image size optimizations before Bedrock processing
- Token usage efficiency in prompts
- Parallel processing where possible
- Caching of layout metadata
- Result visualization optimizations

### 11.2 Scaling Strategy

- Horizontal scaling based on CPU/memory metrics
- Configurable concurrency limits
- Queue-based throttling for Bedrock API
- Resource reservation for critical operations

## 12. Deployment Strategy

- Container image versioning
- Immutable infrastructure
- Blue/green deployment
- Automated testing before promotion
- Rollback capability

---

This architecture provides a clean separation of concerns while maintaining the core two-turn verification approach needed for the vending machine verification process. Each component has a well-defined responsibility, and the interactions between components follow clear patterns that promote maintainability and testability.