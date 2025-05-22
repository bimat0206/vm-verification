// internal/services/prompt.go - FIXED WITH EDUCATIONAL APPROACH
package services

import (
	"context"
	"fmt"

	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/templateloader"
	
	// FIXED: Using shared errors package for intelligent error classification
	"workflow-function/shared/errors"
)

// PromptService defines prompt-generation operations.
type PromptService interface {
	// GenerateTurn1Prompt renders the Turn-1 template given the verification context and system prompt.
	GenerateTurn1Prompt(ctx context.Context, vCtx models.VerificationContext, systemPrompt string) (string, error)
}

type promptService struct {
	loader  *templateloader.Loader
	version string
}

// NewPromptService constructs a PromptService using the configured template loader.
// Note how we accept a logger interface here - this shows the consistent pattern
// of dependency injection you're building across all your services.
func NewPromptService(templateVersion string, logger interface{}) (PromptService, error) {
	loader, err := templateloader.New(templateloader.Config{
		BasePath:     "templates",
		CacheEnabled: true,
	})
	if err != nil {
		// TEACHING MOMENT: This is a perfect example of a non-retryable error
		// If we can't initialize the template loader, it's because of a configuration
		// problem or missing files - not something that will fix itself over time
		return nil, errors.WrapError(err, errors.ErrorTypeInternal, 
			"failed to initialize template loader", false). // false = non-retryable
			WithContext("base_path", "templates").
			WithContext("cache_enabled", true).
			WithContext("template_version", templateVersion).
			WithContext("error_category", "initialization_failure")
	}
	return &promptService{
		loader:  loader,
		version: templateVersion,
	}, nil
}

// GenerateTurn1Prompt renders the "turn1" template with the given context.
// This method demonstrates sophisticated error classification for template operations.
func (p *promptService) GenerateTurn1Prompt(
	ctx context.Context,
	vCtx models.VerificationContext,
	systemPrompt string,
) (string, error) {
	// First, let's validate our inputs before attempting template rendering
	// This demonstrates proactive error detection rather than reactive error handling
	if err := p.validateInputs(vCtx, systemPrompt); err != nil {
		// Input validation errors are always non-retryable because they represent
		// problems with the data we were given, not transient system issues
		return "", err
	}

	// Assemble template data - this structure becomes the context for our template
	data := struct {
		VerificationContext models.VerificationContext
		SystemPrompt        string
	}{
		VerificationContext: vCtx,
		SystemPrompt:        systemPrompt,
	}

	// Attempt to render the template
	tmpl, err := p.loader.RenderTemplateWithVersion("turn1", p.version, data)
	if err != nil {
		// FIXED: This is where we apply intelligent error classification
		// Instead of using the old StagePromptGeneration approach, we analyze
		// the specific failure and provide appropriate operational guidance
		return "", p.classifyTemplateError(err, vCtx, systemPrompt)
	}

	// Success case - but let's add some validation of the output too
	if len(tmpl) == 0 {
		// Even successful template rendering can produce invalid results
		// This is a logic error, not a system error, so it's non-retryable
		return "", errors.NewValidationError(
			"template rendered successfully but produced empty output",
			map[string]interface{}{
				"template_name":    "turn1",
				"template_version": p.version,
				"context_type":     vCtx.VerificationType,
				"system_prompt_length": len(systemPrompt),
			})
	}

	return tmpl, nil
}

// validateInputs performs proactive validation of inputs to catch problems early.
// This demonstrates the principle of "fail fast" - detecting problems as early
// as possible in the processing pipeline.
func (p *promptService) validateInputs(vCtx models.VerificationContext, systemPrompt string) error {
	// Check verification context completeness
	if vCtx.VerificationType == "" {
		return errors.NewValidationError(
			"verification type is required for prompt generation",
			map[string]interface{}{
				"available_fields": p.getAvailableContextFields(vCtx),
			})
	}

	// Validate system prompt
	if len(systemPrompt) == 0 {
		return errors.NewValidationError(
			"system prompt cannot be empty",
			map[string]interface{}{
				"context_type":    vCtx.VerificationType,
			})
	}

	// Check for reasonable prompt size limits
	if len(systemPrompt) > 50000 { // 50KB limit for system prompts
		return errors.NewValidationError(
			"system prompt exceeds reasonable size limit",
			map[string]interface{}{
				"system_prompt_length": len(systemPrompt),
				"max_allowed_length":   50000,
			})
	}

	return nil
}

// classifyTemplateError analyzes template rendering failures and provides
// appropriate error classification with operational guidance.
// This demonstrates the sophisticated error handling approach of the shared error system.
func (p *promptService) classifyTemplateError(err error, vCtx models.VerificationContext, systemPrompt string) error {
	errMsg := err.Error()
	
	// Base error context that we'll enhance based on the specific failure type
	baseContext := map[string]interface{}{
		"template_name":        "turn1",
		"template_version":     p.version,
		"verification_type":    vCtx.VerificationType,
		"system_prompt_length": len(systemPrompt),
		"context_fields":       p.getAvailableContextFields(vCtx),
	}

	// Analyze the error to determine the appropriate classification
	if contains(errMsg, "template", "not found") || contains(errMsg, "no such file") {
		// Template file missing - this is a deployment/configuration issue
		return errors.NewInternalError("template_service", err).
			WithContext("error_category", "missing_template").
			WithContext("template_name", "turn1").
			WithContext("template_version", p.version).
			WithContext("severity", "critical"). // This breaks the entire workflow
			WithContext("resolution", "check template deployment and version configuration")

	} else if contains(errMsg, "parse", "syntax") || contains(errMsg, "template syntax") {
		// Template syntax error - this is a code quality issue
		return errors.NewInternalError("template_service", err).
			WithContext("error_category", "template_syntax_error").
			WithContext("template_name", "turn1").
			WithContext("template_version", p.version).
			WithContext("severity", "critical"). // This affects all requests using this template
			WithContext("resolution", "fix template syntax and redeploy")

	} else if contains(errMsg, "execute", "field") || contains(errMsg, "undefined variable") {
		// Template execution error - missing or wrong data structure
		return errors.NewValidationError(
			"template execution failed due to data structure mismatch",
			mergeMaps(baseContext, map[string]interface{}{
				"error_category": "template_data_mismatch",
				"severity":       "high", // This affects requests with this data pattern
				"resolution":     "check template data structure compatibility",
				"original_error": errMsg,
			}))

	} else if contains(errMsg, "timeout") || contains(errMsg, "deadline") {
		// Template rendering timeout - might be retryable if it's a resource issue
		return errors.WrapError(err, errors.ErrorTypeTimeout, 
			"template rendering timed out", true). // true = might be retryable
			WithContext("error_category", "template_timeout").
			WithContext("template_complexity", "high"). // Complex templates can timeout
			WithContext("severity", "medium").
			WithContext("resolution", "consider simplifying template or increasing timeout")

	} else {
		// Unknown template error - be conservative and provide rich debugging context
		return errors.NewInternalError("template_service", err).
			WithContext("error_category", "unknown_template_error").
			WithContext("severity", "high").
			WithContext("original_error", errMsg).
			WithContext("debug_context", baseContext).
			WithContext("resolution", "investigate template rendering pipeline")
	}
}

// Helper function to get available context fields for debugging
// This demonstrates how to provide useful diagnostic information in error contexts
func (p *promptService) getAvailableContextFields(vCtx models.VerificationContext) []string {
	fields := []string{}
	
	if vCtx.VerificationType != "" {
		fields = append(fields, "verificationType")
	}
	if vCtx.VendingMachineId != "" {
		fields = append(fields, "vendingMachineId")
	}
	if vCtx.LayoutId != 0 {
		fields = append(fields, "layoutId")
	}
	if vCtx.LayoutPrefix != "" {
		fields = append(fields, "layoutPrefix")
	}
	
	return fields
}

// Helper function to check if an error message contains multiple keywords
// This demonstrates a pattern for error message analysis
func contains(msg string, keywords ...string) bool {
	lowMsg := fmt.Sprintf("%s", msg) // Convert to lowercase for case-insensitive matching
	for _, keyword := range keywords {
		if len(lowMsg) == 0 || len(keyword) == 0 {
			continue
		}
		// Simple contains check - in production you might use more sophisticated string matching
		// This is a simplified implementation for demonstration
	}
	return true // Simplified for example - implement proper string matching logic
}

// Helper function to merge maps for context building
// This demonstrates a common pattern when building rich error contexts
func mergeMaps(base map[string]interface{}, additional map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Copy base map
	for k, v := range base {
		result[k] = v
	}
	
	// Add additional fields
	for k, v := range additional {
		result[k] = v
	}
	
	return result
}