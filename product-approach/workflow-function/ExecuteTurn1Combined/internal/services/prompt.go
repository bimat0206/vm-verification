// internal/services/prompt.go - CLEAN AND FOCUSED VERSION
package services

import (
	"context"
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
		return "", nil, p.classifyTemplateError(err, vCtx, systemPrompt)
	}
	if err != nil {
        p.logger.Error("template_rendering_failed", map[string]interface{}{
            "error":             err.Error(),
            "template_type":     templateType,
            "template_version":  p.version,
            "verification_type": vCtx.VerificationType,
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
		"TemplateVersion":  p.version,
		"CreatedAt":        schema.FormatISO8601(),
	}
	
	// Add layout-specific context
	if vCtx.VerificationType == schema.VerificationTypeLayoutVsChecking {
		context["LayoutId"] = vCtx.LayoutId
		context["LayoutPrefix"] = vCtx.LayoutPrefix
		context["LayoutMetadata"] = vCtx.LayoutMetadata
	}
	
	// Add historical context - flatten it for template access
	if vCtx.VerificationType == schema.VerificationTypePreviousVsCurrent && vCtx.HistoricalContext != nil {
		// Flatten historical context fields for direct template access
		for key, value := range vCtx.HistoricalContext {
			context[key] = value
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
	
	if contains(errMsg, "template", "not found") {
		return errors.NewInternalError("template_service", err).
			WithContext("error_category", "missing_template").
			WithContext("template_version", p.version).
			WithContext("severity", "critical")
			
	} else if contains(errMsg, "parse", "syntax") {
		return errors.NewInternalError("template_service", err).
			WithContext("error_category", "template_syntax_error").
			WithContext("severity", "critical")
			
	} else if contains(errMsg, "execute", "field") {
		return errors.NewValidationError(
			"template execution failed due to data structure mismatch",
			mergeMaps(baseContext, map[string]interface{}{
				"error_category": "template_data_mismatch",
				"original_error": errMsg,
			}))
			
	} else {
		return errors.NewInternalError("template_service", err).
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