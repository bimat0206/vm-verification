package internal

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"
)

// RetryManager manages retry logic with exponential backoff
type RetryManager struct {
	maxAttempts int
	baseDelay   time.Duration
}

// NewRetryManager creates a new retry manager
func NewRetryManager(maxAttempts int, baseDelayMs int) *RetryManager {
	return &RetryManager{
		maxAttempts: maxAttempts,
		baseDelay:   time.Duration(baseDelayMs) * time.Millisecond,
	}
}

// ExecuteWithRetry executes a function with retry logic
func (rm *RetryManager) ExecuteWithRetry(ctx context.Context, fn func() error) bool {
	for attempt := 1; attempt <= rm.maxAttempts; attempt++ {
		log.Printf("Attempt %d of %d", attempt, rm.maxAttempts)
		
		err := fn()
		if err == nil {
			log.Printf("Operation succeeded on attempt %d", attempt)
			return true
		}

		// Check if the error is retryable
		if !IsRetryable(err) {
			log.Printf("Non-retryable error encountered: %v", err)
			return false
		}

		// Check if we should retry
		if attempt >= rm.maxAttempts {
			log.Printf("Max retry attempts (%d) reached", rm.maxAttempts)
			return false
		}

		// Calculate delay with exponential backoff
		delay := rm.calculateDelay(attempt)
		
		log.Printf("Attempt %d failed with retryable error: %v. Retrying in %v", 
			attempt, err, delay)

		// Wait for the delay or context cancellation
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled during retry wait")
			return false
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return false
}

// calculateDelay calculates the delay for a retry attempt using exponential backoff
func (rm *RetryManager) calculateDelay(attempt int) time.Duration {
	// Exponential backoff: delay = baseDelay * 2^(attempt-1)
	// With jitter to avoid thundering herd problem
	backoffMultiplier := math.Pow(2, float64(attempt-1))
	delay := time.Duration(float64(rm.baseDelay) * backoffMultiplier)
	
	// Add jitter (up to 20% of the delay)
	jitter := time.Duration(float64(delay) * 0.2 * (0.5 - float64(time.Now().UnixNano()%100)/100.0))
	delay = delay + jitter
	
	// Cap the maximum delay at 30 seconds
	maxDelay := 30 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}
	
	return delay
}

// RetryPolicy defines the retry policy for different error types
type RetryPolicy struct {
	ShouldRetry    bool
	MaxAttempts    int
	BackoffFactor  float64
	MaxBackoffTime time.Duration
}

// GetRetryPolicy returns the retry policy for different error types
func GetRetryPolicy(err error) RetryPolicy {
	if execErr, ok := err.(*ExecuteTurn1Error); ok {
		switch execErr.Type {
		case ErrorTypeBedrock:
			switch execErr.Code {
			case "THROTTLING":
				return RetryPolicy{
					ShouldRetry:    true,
					MaxAttempts:    5,
					BackoffFactor:  2.0,
					MaxBackoffTime: 60 * time.Second,
				}
			case "BEDROCK_INTERNAL_ERROR":
				return RetryPolicy{
					ShouldRetry:    true,
					MaxAttempts:    3,
					BackoffFactor:  2.0,
					MaxBackoffTime: 30 * time.Second,
				}
			case "TOKEN_LIMIT_EXCEEDED", "MODEL_UNAVAILABLE", "BEDROCK_ACCESS_DENIED":
				return RetryPolicy{
					ShouldRetry:    false,
					MaxAttempts:    0,
					BackoffFactor:  0,
					MaxBackoffTime: 0,
				}
			}
		case ErrorTypeS3:
			switch execErr.Code {
			case "S3_OPERATION_FAILED":
				return RetryPolicy{
					ShouldRetry:    true,
					MaxAttempts:    3,
					BackoffFactor:  1.5,
					MaxBackoffTime: 20 * time.Second,
				}
			case "S3_ACCESS_DENIED":
				return RetryPolicy{
					ShouldRetry:    false,
					MaxAttempts:    0,
					BackoffFactor:  0,
					MaxBackoffTime: 0,
				}
			}
		case ErrorTypeDynamoDB:
			return RetryPolicy{
				ShouldRetry:    true,
				MaxAttempts:    3,
				BackoffFactor:  1.5,
				MaxBackoffTime: 15 * time.Second,
			}
		case ErrorTypeTimeout:
			return RetryPolicy{
				ShouldRetry:    true,
				MaxAttempts:    2,
				BackoffFactor:  1.0,
				MaxBackoffTime: 10 * time.Second,
			}
		case ErrorTypeValidation, ErrorTypeInternal:
			return RetryPolicy{
				ShouldRetry:    false,
				MaxAttempts:    0,
				BackoffFactor:  0,
				MaxBackoffTime: 0,
			}
		}
	}

	// Default policy for unknown errors
	return RetryPolicy{
		ShouldRetry:    true,
		MaxAttempts:    2,
		BackoffFactor:  2.0,
		MaxBackoffTime: 10 * time.Second,
	}
}

// RetryableExecutor is a more advanced retry executor with configurable policies
type RetryableExecutor struct {
	defaultPolicy RetryPolicy
}

// NewRetryableExecutor creates a new retryable executor
func NewRetryableExecutor() *RetryableExecutor {
	return &RetryableExecutor{
		defaultPolicy: RetryPolicy{
			ShouldRetry:    true,
			MaxAttempts:    3,
			BackoffFactor:  2.0,
			MaxBackoffTime: 30 * time.Second,
		},
	}
}

// Execute executes a function with adaptive retry policy based on error type
func (re *RetryableExecutor) Execute(ctx context.Context, operation func() error) error {
	var lastErr error
	
	for attempt := 1; attempt <= re.defaultPolicy.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err
		policy := GetRetryPolicy(err)
		
		if !policy.ShouldRetry || attempt >= policy.MaxAttempts {
			return lastErr
		}

		delay := re.calculateAdaptiveDelay(attempt, policy)
		
		log.Printf("Attempt %d failed with error: %v. Retrying in %v", 
			attempt, err, delay)

		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", lastErr)
		case <-time.After(delay):
			continue
		}
	}

	return lastErr
}

// calculateAdaptiveDelay calculates delay based on the retry policy
func (re *RetryableExecutor) calculateAdaptiveDelay(attempt int, policy RetryPolicy) time.Duration {
	baseDelay := 1 * time.Second // Default base delay
	
	// Calculate backoff delay
	delay := time.Duration(float64(baseDelay) * math.Pow(policy.BackoffFactor, float64(attempt-1)))
	
	// Apply jitter to prevent thundering herd
	jitter := time.Duration(float64(delay) * 0.1 * (0.5 - float64(time.Now().UnixNano()%100)/100.0))
	delay = delay + jitter
	
	// Cap at maximum backoff time
	if delay > policy.MaxBackoffTime {
		delay = policy.MaxBackoffTime
	}
	
	return delay
}

// CircuitBreaker implements a circuit breaker pattern for retry management
type CircuitBreaker struct {
	maxFailures    int
	resetTimeout   time.Duration
	state          CircuitState
	failureCount   int
	lastFailureTime time.Time
}

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// Execute executes an operation with circuit breaker protection
func (cb *CircuitBreaker) Execute(operation func() error) error {
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
		} else {
			return NewBedrockError("Circuit breaker is open", "CIRCUIT_BREAKER_OPEN", false)
		}
	}

	err := operation()
	
	if err != nil {
		cb.onFailure()
		return err
	}
	
	cb.onSuccess()
	return nil
}

// onFailure handles failure cases
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	
	if cb.failureCount >= cb.maxFailures {
		cb.state = StateOpen
	}
}

// onSuccess handles success cases
func (cb *CircuitBreaker) onSuccess() {
	cb.failureCount = 0
	cb.state = StateClosed
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	return cb.state
}

// RetryMetrics tracks retry metrics for monitoring
type RetryMetrics struct {
	TotalAttempts    int
	SuccessfulAttempts int
	FailedAttempts   int
	TotalRetries     int
	TotalDelay       time.Duration
}

// NewRetryMetrics creates new retry metrics
func NewRetryMetrics() *RetryMetrics {
	return &RetryMetrics{}
}

// RecordAttempt records a retry attempt
func (rm *RetryMetrics) RecordAttempt(success bool, attempt int, delay time.Duration) {
	rm.TotalAttempts++
	if success {
		rm.SuccessfulAttempts++
	} else {
		rm.FailedAttempts++
	}
	
	if attempt > 1 {
		rm.TotalRetries++
		rm.TotalDelay += delay
	}
}

// GetRetryRate returns the retry rate
func (rm *RetryMetrics) GetRetryRate() float64 {
	if rm.TotalAttempts == 0 {
		return 0
	}
	return float64(rm.TotalRetries) / float64(rm.TotalAttempts)
}

// GetSuccessRate returns the success rate
func (rm *RetryMetrics) GetSuccessRate() float64 {
	if rm.TotalAttempts == 0 {
		return 0
	}
	return float64(rm.SuccessfulAttempts) / float64(rm.TotalAttempts)
}

// GetAverageDelay returns the average delay between retries
func (rm *RetryMetrics) GetAverageDelay() time.Duration {
	if rm.TotalRetries == 0 {
		return 0
	}
	return rm.TotalDelay / time.Duration(rm.TotalRetries)
}

// ToMap converts metrics to a map for logging/monitoring
func (rm *RetryMetrics) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"totalAttempts":      rm.TotalAttempts,
		"successfulAttempts": rm.SuccessfulAttempts,
		"failedAttempts":     rm.FailedAttempts,
		"totalRetries":       rm.TotalRetries,
		"retryRate":          rm.GetRetryRate(),
		"successRate":        rm.GetSuccessRate(),
		"averageDelay":       rm.GetAverageDelay().String(),
	}
}

// BackoffStrategy defines different backoff strategies
type BackoffStrategy string

const (
	BackoffExponential BackoffStrategy = "exponential"
	BackoffLinear      BackoffStrategy = "linear"
	BackoffFixed       BackoffStrategy = "fixed"
)

// CalculateDelay calculates delay based on strategy
func CalculateDelay(strategy BackoffStrategy, attempt int, baseDelay time.Duration) time.Duration {
	switch strategy {
	case BackoffExponential:
		return time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))
	case BackoffLinear:
		return time.Duration(float64(baseDelay) * float64(attempt))
	case BackoffFixed:
		return baseDelay
	default:
		return baseDelay
	}
}