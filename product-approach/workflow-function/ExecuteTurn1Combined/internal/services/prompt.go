// internal/services/prompt.go - CLEAN AND FOCUSED VERSION
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	"workflow-function/shared/templateloader"
)

// PromptService defines prompt-generation operations.
type PromptService interface {
	// GenerateTurn1Prompt renders the Turn-1 template given the verification context and system prompt.
	GenerateTurn1Prompt(ctx context.Context, vCtx models.VerificationContext, systemPrompt string) (string, error)
	
	// GenerateTurn1PromptWithMetrics generates prompt and returns processing metrics
	GenerateTurn1PromptWithMetrics(ctx context.Context, vCtx models.VerificationContext, systemPrompt string) (string, *schema.TemplateProcessor, error)
}

type promptService struct {
	templateLoader templateloader.TemplateLoader
	logger         logger.Logger
	version        string
}

// NewPromptService constructs a PromptService with template management
// In NewPromptService function, add logging:
func NewPromptService(cfg *config.Config, logger logger.Logger) (PromptService, error) {
    logger.Info("initializing_prompt_service", map[string]interface{}{
        "template_version":   cfg.Prompts.TemplateVersion,
        "template_base_path": cfg.Prompts.TemplateBasePath,
    })
    
    // Initialize template loader with configuration
    loaderConfig := templateloader.Config{
        BasePath:     cfg.Prompts.TemplateBasePath,
        CacheEnabled: true,
    }
    
    logger.Info("creating_template_loader", map[string]interface{}{
        "base_path":     loaderConfig.BasePath,
        "cache_enabled": loaderConfig.CacheEnabled,
    })
    
    templateLoader, err := templateloader.New(loaderConfig)
    if err != nil {
        logger.Error("template_loader_initialization_failed", map[string]interface{}{
            "error":              err.Error(),
            "template_base_path": cfg.Prompts.TemplateBasePath,
        })
        return nil, errors.WrapError(err, errors.ErrorTypeInternal, 
            "failed to initialize template loader", false).
            WithContext("template_version", cfg.Prompts.TemplateVersion).
            WithContext("template_base_path", cfg.Prompts.TemplateBasePath)
    }
    
    logger.Info("prompt_service_initialized_successfully", map[string]interface{}{
        "template_version":   cfg.Prompts.TemplateVersion,
        "template_base_path": cfg.Prompts.TemplateBasePath,
    })
    
    return &promptService{
        templateLoader: templateLoader,
        logger:         logger,
        version:        cfg.Prompts.TemplateVersion,
    }, nil
}

// GenerateTurn1Prompt renders the "turn1" template with the given context (legacy method)
func (p *promptService) GenerateTurn1Prompt(
	ctx context.Context,
	vCtx models.VerificationContext,
	systemPrompt string,
) (string, error) {
	prompt, _, err := p.GenerateTurn1PromptWithMetrics(ctx, vCtx, systemPrompt)
	return prompt, err
}

// GenerateTurn1PromptWithMetrics generates prompt and returns processing metrics
func (p *promptService) GenerateTurn1PromptWithMetrics(
	ctx context.Context,
	vCtx models.VerificationContext,
	systemPrompt string,
) (string, *schema.TemplateProcessor, error) {
	startTime := time.Now()
	p.logger.Info("generating_turn1_prompt", map[string]interface{}{
        "verification_type":    vCtx.VerificationType,
        "template_version":     p.version,
        "system_prompt_length": len(systemPrompt),
    })
	// Validate inputs
	if err := p.validateInputs(vCtx, systemPrompt); err != nil {
		return "", nil, err
	}
	
	// Validate and fix context structure
	if err := vCtx.Validate(); err != nil {
		return "", nil, errors.NewValidationError("invalid verification context", map[string]interface{}{
			"error": err.Error(),
		})
	}
	
	// Determine template type based on verification type
	templateType := p.getTemplateType(vCtx.VerificationType)
	p.logger.Info("loading_template", map[string]interface{}{
        "template_type":     templateType,
        "template_version":  p.version,
        "verification_type": vCtx.VerificationType,
    })
	
	// Build template context
	templateContext := p.buildTemplateContext(vCtx, systemPrompt)
	
	// Render template using the standardized template loader
	var processedPrompt string
	var err error
	if p.version != "" {
		processedPrompt, err = p.templateLoader.RenderTemplateWithVersion(templateType, p.version, templateContext)
	} else {
		processedPrompt, err = p.templateLoader.RenderTemplate(templateType, templateContext)
	}
	
	if err != nil {
		p.logger.Error("template_rendering_failed", map[string]interface{}{
			"error":             err.Error(),
			"template_type":     templateType,
			"template_version":  p.version,
			"verification_type": vCtx.VerificationType,
			"original_error":    err.Error(),
		})
		return "", nil, p.classifyTemplateError(err, vCtx, systemPrompt)
	}
    
    p.logger.Info("template_rendered_successfully", map[string]interface{}{
        "template_type":       templateType,
        "prompt_length":       len(processedPrompt),
        "processing_time_ms":  time.Since(startTime).Milliseconds(),
        "estimated_tokens":    len(processedPrompt) / 4,
    })

	// Quick token estimate check (not persisted, only for validation)
	estimate := len(processedPrompt) / 4 // Rough estimate: 4 chars per token
	if estimate > p.getMaxTokenBudget() {
		return "", nil, errors.NewValidationError(
			"prompt exceeds token budget",
			map[string]interface{}{
				"estimated_tokens": estimate,
				"max_budget":       p.getMaxTokenBudget(),
				"prompt_length":    len(processedPrompt),
			})
	}
	
	// Create processor info for metrics tracking
	processingTime := time.Since(startTime)
	processor := &schema.TemplateProcessor{
		Template: &schema.PromptTemplate{
			TemplateId:      templateType,
			TemplateVersion: p.version,
			TemplateType:    templateType,
			Content:         processedPrompt,
		},
		ContextData:     templateContext,
		ProcessedPrompt: processedPrompt,
		ProcessingTime:  processingTime.Milliseconds(),
		// InputTokens and OutputTokens will be populated later from Bedrock response
		InputTokens:     0,
		OutputTokens:    0,
	}
	
	return processedPrompt, processor, nil
}

// validateInputs performs proactive validation of inputs
func (p *promptService) validateInputs(vCtx models.VerificationContext, systemPrompt string) error {
	if vCtx.VerificationType == "" {
		return errors.NewValidationError(
			"verification type is required for prompt generation",
			map[string]interface{}{
				"available_fields": p.getAvailableContextFields(vCtx),
			})
	}
	
	if len(systemPrompt) == 0 {
		return errors.NewValidationError(
			"system prompt cannot be empty",
			map[string]interface{}{
				"verification_type": vCtx.VerificationType,
			})
	}
	
	// Check for reasonable prompt size limits
	if len(systemPrompt) > 50000 { // 50KB limit
		return errors.NewValidationError(
			"system prompt exceeds reasonable size limit",
			map[string]interface{}{
				"system_prompt_length": len(systemPrompt),
				"max_allowed_length":   50000,
			})
	}
	
	return nil
}

// getTemplateType determines the template type based on verification type
func (p *promptService) getTemplateType(verificationType string) string {
	switch verificationType {
	case schema.VerificationTypeLayoutVsChecking:
		return "turn1-layout-vs-checking"
	case schema.VerificationTypePreviousVsCurrent:
		return "turn1-previous-vs-current"
	default:
		return "turn1-default"
	}
}

// buildTemplateContext creates the context data for template processing
func (p *promptService) buildTemplateContext(vCtx models.VerificationContext, systemPrompt string) map[string]interface{} {
	context := map[string]interface{}{
		"VerificationType": vCtx.VerificationType,
		"SystemPrompt":     systemPrompt,
		"VendingMachineId": vCtx.VendingMachineId,
		"VendingMachineID": vCtx.VendingMachineId, // Also add with uppercase ID for template compatibility
		"TemplateVersion":  p.version,
		"CreatedAt":        schema.FormatISO8601(),
	}
	
	// Ensure essential fields are initialized with defaults if missing
	context["RowCount"] = p.getIntOrDefault(vCtx, "RowCount", 6)
	context["ColumnCount"] = p.getIntOrDefault(vCtx, "ColumnCount", 10)
	
	// Ensure RowLabels is properly initialized
	rowCount := context["RowCount"].(int)
	context["RowLabels"] = p.ensureRowLabels(vCtx, rowCount)
	
	// Add layout-specific context
	if vCtx.VerificationType == schema.VerificationTypeLayoutVsChecking {
		context["LayoutId"] = vCtx.LayoutId
		context["LayoutPrefix"] = vCtx.LayoutPrefix
		context["LayoutMetadata"] = vCtx.LayoutMetadata
		
		// Add Location if available in layout metadata
		if vCtx.LayoutMetadata != nil {
			if location, ok := vCtx.LayoutMetadata["location"].(string); ok {
				context["Location"] = location
			}
		}
	}
	
	// Add historical context - flatten it for template access
	if vCtx.VerificationType == schema.VerificationTypePreviousVsCurrent && vCtx.HistoricalContext != nil {
		// Flatten historical context fields for direct template access
		for key, value := range vCtx.HistoricalContext {
			context[key] = value
		}
		
		// Ensure required fields for previous-vs-current template
		context["PreviousVerificationAt"] = p.getStringOrDefault(vCtx.HistoricalContext, "PreviousVerificationAt", "unknown")
		context["HoursSinceLastVerification"] = p.getFloatOrDefault(vCtx.HistoricalContext, "HoursSinceLastVerification", 0.0)
		context["PreviousVerificationStatus"] = p.getStringOrDefault(vCtx.HistoricalContext, "PreviousVerificationStatus", "unknown")
		
		// Ensure VerificationSummary is properly structured
		if summary, ok := vCtx.HistoricalContext["VerificationSummary"].(map[string]interface{}); ok {
			context["VerificationSummary"] = p.ensureVerificationSummary(summary)
		}
	}
	
	return context
}

// classifyTemplateError analyzes template rendering failures and provides appropriate error classification
func (p *promptService) classifyTemplateError(err error, vCtx models.VerificationContext, systemPrompt string) error {
	baseContext := map[string]interface{}{
		"template_version":     p.version,
		"verification_type":    vCtx.VerificationType,
		"system_prompt_length": len(systemPrompt),
	}
	
	errMsg := err.Error()
	
	// Log the error with full context for debugging
	p.logger.Error("template_error_classification", map[string]interface{}{
		"original_error":       errMsg,
		"error_type":          fmt.Sprintf("%T", err),
		"template_version":    p.version,
		"verification_type":   vCtx.VerificationType,
		"has_layout_metadata": vCtx.LayoutMetadata != nil,
		"has_historical_ctx":  vCtx.HistoricalContext != nil,
	})
	
	if contains(errMsg, "template", "not found") {
		return errors.NewInternalError("prompt_service", err).
			WithContext("error_category", "missing_template").
			WithContext("template_version", p.version).
			WithContext("severity", "critical").
			WithContext("original_error", errMsg)
			
	} else if contains(errMsg, "parse", "syntax") {
		return errors.NewInternalError("prompt_service", err).
			WithContext("error_category", "template_syntax_error").
			WithContext("severity", "critical").
			WithContext("original_error", errMsg)
			
	} else if contains(errMsg, "execute", "field") || contains(errMsg, "function", "not defined") {
		return errors.NewValidationError(
			"template execution failed due to data structure mismatch or missing function",
			mergeMaps(baseContext, map[string]interface{}{
				"error_category": "template_data_mismatch",
				"original_error": errMsg,
			}))
			
	} else {
		return errors.NewInternalError("prompt_service", err).
			WithContext("error_category", "unknown_template_error").
			WithContext("original_error", errMsg).
			WithContext("debug_context", baseContext)
	}
}

// getAvailableContextFields returns available context fields for debugging
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

// Helper functions
func contains(msg string, keywords ...string) bool {
	lowerMsg := strings.ToLower(msg)
	for _, keyword := range keywords {
		if keyword != "" && strings.Contains(lowerMsg, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func mergeMaps(base map[string]interface{}, additional map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	for k, v := range base {
		result[k] = v
	}
	
	for k, v := range additional {
		result[k] = v
	}
	
	return result
}

// getMaxTokenBudget returns the maximum token budget for prompts
func (p *promptService) getMaxTokenBudget() int {
	// This should ideally come from configuration
	// For now, using a reasonable default
	return 16000 // Conservative budget to ensure we don't exceed Bedrock limits
}

// Helper methods for safe data extraction with defaults

func (p *promptService) getIntOrDefault(vCtx models.VerificationContext, key string, defaultValue int) int {
	// Check in LayoutMetadata first
	if vCtx.LayoutMetadata != nil {
		if val, ok := vCtx.LayoutMetadata[key]; ok {
			switch v := val.(type) {
			case int:
				return v
			case float64:
				return int(v)
			case int64:
				return int(v)
			}
		}
	}
	
	// Check in HistoricalContext if present
	if vCtx.HistoricalContext != nil {
		if val, ok := vCtx.HistoricalContext[key]; ok {
			switch v := val.(type) {
			case int:
				return v
			case float64:
				return int(v)
			case int64:
				return int(v)
			}
		}
	}
	
	return defaultValue
}

func (p *promptService) getFloatOrDefault(data map[string]interface{}, key string, defaultValue float64) float64 {
	if data == nil {
		return defaultValue
	}
	
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	
	return defaultValue
}

func (p *promptService) getStringOrDefault(data map[string]interface{}, key string, defaultValue string) string {
	if data == nil {
		return defaultValue
	}
	
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	
	return defaultValue
}

func (p *promptService) ensureRowLabels(vCtx models.VerificationContext, rowCount int) []string {
	// Check if RowLabels already exists in context
	if vCtx.LayoutMetadata != nil {
		if labels, ok := vCtx.LayoutMetadata["RowLabels"].([]string); ok && len(labels) >= rowCount {
			return labels
		}
		// Handle []interface{} case
		if labelsInterface, ok := vCtx.LayoutMetadata["RowLabels"].([]interface{}); ok {
			labels := make([]string, 0, len(labelsInterface))
			for _, l := range labelsInterface {
				if str, ok := l.(string); ok {
					labels = append(labels, str)
				}
			}
			if len(labels) >= rowCount {
				return labels
			}
		}
	}
	
	// Generate default row labels A, B, C, etc.
	labels := make([]string, rowCount)
	for i := 0; i < rowCount; i++ {
		labels[i] = string(rune('A' + i))
	}
	
	return labels
}

func (p *promptService) ensureVerificationSummary(summary map[string]interface{}) map[string]interface{} {
	// Ensure all required fields are present with proper types
	result := make(map[string]interface{})
	
	result["OverallAccuracy"] = p.getFloatOrDefault(summary, "OverallAccuracy", 0.0)
	result["MissingProducts"] = p.getIntOrDefault(models.VerificationContext{HistoricalContext: summary}, "MissingProducts", 0)
	result["IncorrectProductTypes"] = p.getIntOrDefault(models.VerificationContext{HistoricalContext: summary}, "IncorrectProductTypes", 0)
	result["EmptyPositionsCount"] = p.getIntOrDefault(models.VerificationContext{HistoricalContext: summary}, "EmptyPositionsCount", 0)
	
	return result
}