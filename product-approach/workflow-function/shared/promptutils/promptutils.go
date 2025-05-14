package promptutils

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"
	"time"

	"shared/promptutils/bedrock"
	"shared/promptutils/processor"
	"shared/promptutils/templates"
	"shared/promptutils/validator"
)

// PromptProcessor manages the generation of system prompts
type PromptProcessor struct {
	templateManager *templates.TemplateManager
	templateBasePath string
}

// NewPromptProcessor creates a new PromptProcessor
func NewPromptProcessor(templateBasePath string) *PromptProcessor {
	// If template base path is not provided, use default
	if templateBasePath == "" {
		templateBasePath = os.Getenv("TEMPLATE_BASE_PATH")
		if templateBasePath == "" {
			templateBasePath = "/opt/templates" // Default in container
		}
	}

	return &PromptProcessor{
		templateManager: templates.NewTemplateManager(templateBasePath),
		templateBasePath: templateBasePath,
	}
}

// ProcessInput handles the main workflow of validating input, preparing data, and rendering templates
func (p *PromptProcessor) ProcessInput(event json.RawMessage) (Response, error) {
	// Parse and validate input
	var input Input
	if err := json.Unmarshal(event, &input); err != nil {
		return Response{}, fmt.Errorf("invalid input format: %w", err)
	}
	
	// Validate input
	if err := validator.ValidateInput(&input); err != nil {
		return Response{}, fmt.Errorf("input validation failed: %w", err)
	}
	
	// Extract verification type
	verificationType := input.VerificationContext.VerificationType
	
	// Get appropriate template
	tmpl, err := p.templateManager.GetTemplate(verificationType)
	if err != nil {
		return Response{}, fmt.Errorf("template error: %w", err)
	}
	
	// Create template data context
	templateData, err := processor.BuildTemplateData(&input)
	if err != nil {
		return Response{}, fmt.Errorf("context preparation failed: %w", err)
	}
	
	// Generate system prompt
	systemPrompt, err := p.renderTemplate(tmpl, templateData)
	if err != nil {
		return Response{}, fmt.Errorf("prompt generation failed: %w", err)
	}
	
	// Configure Bedrock
	bedrockConfig := bedrock.ConfigureBedrockSettings()
	
	// Update verification context status
	input.VerificationContext.Status = "SYSTEM_PROMPT_READY"
	
	// Prepare response
	response := Response{
		VerificationContext: input.VerificationContext,
		SystemPrompt: SystemPrompt{
			Content:       systemPrompt,
			PromptID:      fmt.Sprintf("prompt-%s-system", input.VerificationContext.VerificationID),
			CreatedAt:     time.Now().UTC().Format(time.RFC3339),
			PromptVersion: p.templateManager.GetLatestVersion(verificationType),
		},
		BedrockConfig: bedrockConfig,
	}
	
	// Include appropriate metadata based on verification type
	if verificationType == "LAYOUT_VS_CHECKING" && input.LayoutMetadata != nil {
		response.LayoutMetadata = input.LayoutMetadata
	} else if verificationType == "PREVIOUS_VS_CURRENT" && input.HistoricalContext != nil {
		response.HistoricalContext = input.HistoricalContext
	}
	
	return response, nil
}

// renderTemplate renders a template with the provided data
func (p *PromptProcessor) renderTemplate(tmpl *template.Template, data TemplateData) (string, error) {
	return templates.ProcessTemplate(tmpl, data)
}

// PrepareBedrockRequest prepares a Bedrock API request with a system prompt
func (p *PromptProcessor) PrepareBedrockRequest(systemPrompt string, bedrockConfig bedrock.BedrockConfig) (bedrock.BedrockRequest, error) {
	return bedrock.PrepareBedrockRequest(systemPrompt, bedrockConfig)
}

// PrepareBedrockTurnRequest prepares a Bedrock API request for a specific turn
func (p *PromptProcessor) PrepareBedrockTurnRequest(systemPrompt string, userMessage bedrock.BedrockMessage, bedrockConfig bedrock.BedrockConfig) (bedrock.BedrockRequest, error) {
	return bedrock.PrepareBedrockTurnRequest(systemPrompt, userMessage, bedrockConfig)
}

// ListAvailableTemplates returns a list of all available templates and their versions
func (p *PromptProcessor) ListAvailableTemplates() map[string]string {
	return p.templateManager.ListAvailableTemplates()
}

// GetTemplateBasePath returns the base path for templates
func (p *PromptProcessor) GetTemplateBasePath() string {
	return p.templateBasePath
}