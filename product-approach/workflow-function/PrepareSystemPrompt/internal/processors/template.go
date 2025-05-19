package processors

import (
	"fmt"
	"strings"
	"text/template"
	
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	"workflow-function/shared/templateloader"
	
	"workflow-function/PrepareSystemPrompt/internal/config"
	"workflow-function/PrepareSystemPrompt/internal/models"
)

// TemplateProcessor handles template loading and rendering
type TemplateProcessor struct {
	templateLoader *templateloader.Loader
	config         *config.Config
	logger         logger.Logger
}

// NewTemplateProcessor creates a new template processor
func NewTemplateProcessor(cfg *config.Config, log logger.Logger) (*TemplateProcessor, error) {
	// Configure template loader
	tlConfig := templateloader.Config{
		BasePath:     cfg.TemplateBasePath,
		CacheEnabled: true,
	}
	
	tmplLoader, err := templateloader.New(tlConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize template loader: %w", err)
	}
	
	return &TemplateProcessor{
		templateLoader: tmplLoader,
		config:         cfg,
		logger:         log,
	}, nil
}

// MapVerificationTypeToTemplateType converts a verification type to a template type
func MapVerificationTypeToTemplateType(verificationType string) string {
	return strings.ReplaceAll(strings.ToLower(verificationType), "_", "-")
}

// GetTemplate gets a template for a verification type
func (p *TemplateProcessor) GetTemplate(verificationType string) (*template.Template, error) {
	// Map verification type to template type
	templateType := MapVerificationTypeToTemplateType(verificationType)
	
	// Get template from loader
	tmpl, err := p.templateLoader.LoadTemplate(templateType)
	if err != nil {
		return nil, fmt.Errorf("failed to load template for %s: %w", verificationType, err)
	}
	
	return tmpl, nil
}

// GetTemplateWithVersion gets a specific version of a template
func (p *TemplateProcessor) GetTemplateWithVersion(verificationType, version string) (*template.Template, error) {
	// Map verification type to template type
	templateType := MapVerificationTypeToTemplateType(verificationType)
	
	// Get template from loader
	tmpl, err := p.templateLoader.LoadTemplateWithVersion(templateType, version)
	if err != nil {
		return nil, fmt.Errorf("failed to load template %s version %s: %w", verificationType, version, err)
	}
	
	return tmpl, nil
}

// GetLatestVersion gets the latest template version for a verification type
func (p *TemplateProcessor) GetLatestVersion(verificationType string) string {
	// Map verification type to template type
	templateType := MapVerificationTypeToTemplateType(verificationType)
	
	// Get latest version from loader
	version := p.templateLoader.GetLatestVersion(templateType)
	if version == "" {
		// Fallback to default versions
		return p.config.PromptVersion
	}
	
	return version
}

// RenderTemplate renders a template with data
func (p *TemplateProcessor) RenderTemplate(tmpl *template.Template, data *models.TemplateData) (string, error) {
	if tmpl == nil {
		return "", fmt.Errorf("template is nil")
	}
	
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	
	return buf.String(), nil
}

// BuildTemplateData builds template data from input context
func (p *TemplateProcessor) BuildTemplateData(vCtx *schema.VerificationContext, 
	layoutMetadata, historicalContext map[string]interface{}) (*models.TemplateData, error) {
	
	// Check if verification context is nil
	if vCtx == nil {
		return nil, fmt.Errorf("verification context is nil")
	}
	
	// Initialize template data with verification context
	data := &models.TemplateData{
		VerificationType: vCtx.VerificationType,
		VerificationID:   vCtx.VerificationId,
		VerificationAt:   vCtx.VerificationAt,
		VendingMachineID: vCtx.VendingMachineId,
	}
	
	// Handle verification type-specific data
	switch vCtx.VerificationType {
	case schema.VerificationTypeLayoutVsChecking:
		if err := p.processLayoutVsCheckingData(data, layoutMetadata); err != nil {
			return nil, err
		}
	case schema.VerificationTypePreviousVsCurrent:
		if err := p.processPreviousVsCurrentData(data, historicalContext); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported verification type: %s", vCtx.VerificationType)
	}
	
	return data, nil
}

// processLayoutVsCheckingData processes data for LAYOUT_VS_CHECKING verification type
func (p *TemplateProcessor) processLayoutVsCheckingData(data *models.TemplateData, 
	layoutMetadata map[string]interface{}) error {
	
	if layoutMetadata == nil {
		return fmt.Errorf("layout metadata is required for LAYOUT_VS_CHECKING")
	}
	
	// Extract machine structure from layout metadata
	ms, err := models.ExtractMachineStructure(layoutMetadata)
	if err != nil {
		return fmt.Errorf("failed to extract machine structure: %w", err)
	}
	
	if ms == nil {
		return fmt.Errorf("machine structure is required for LAYOUT_VS_CHECKING")
	}
	
	// Populate machine structure fields
	data.MachineStructure = ms
	data.RowCount = ms.RowCount
	data.ColumnCount = ms.ColumnsPerRow
	data.RowLabels = models.FormatArrayToString(ms.RowOrder)
	data.ColumnLabels = models.FormatArrayToString(ms.ColumnOrder)
	data.TotalPositions = ms.RowCount * ms.ColumnsPerRow
	
	// Extract product mappings if available
	productMap, err := models.ExtractProductPositionMap(layoutMetadata)
	if err == nil && productMap != nil {
		data.ProductMappings = models.FormatProductMappings(productMap)
	}
	
	// Extract location if available
	if locVal, ok := layoutMetadata["location"]; ok {
		if loc, ok := locVal.(string); ok {
			data.Location = loc
		}
	}
	
	return nil
}

// processPreviousVsCurrentData processes data for PREVIOUS_VS_CURRENT verification type
func (p *TemplateProcessor) processPreviousVsCurrentData(data *models.TemplateData, 
	historicalContext map[string]interface{}) error {
	
	if historicalContext == nil {
		return fmt.Errorf("historical context is required for PREVIOUS_VS_CURRENT")
	}
	
	// Extract previous verification details
	if prevId, ok := historicalContext["previousVerificationId"]; ok {
		if idStr, ok := prevId.(string); ok {
			data.PreviousVerificationID = idStr
		}
	}
	
	if prevAt, ok := historicalContext["previousVerificationAt"]; ok {
		if atStr, ok := prevAt.(string); ok {
			data.PreviousVerificationAt = atStr
		}
	}
	
	if prevStatus, ok := historicalContext["previousVerificationStatus"]; ok {
		if statusStr, ok := prevStatus.(string); ok {
			data.PreviousVerificationStatus = statusStr
		}
	}
	
	if hoursVal, ok := historicalContext["hoursSinceLastVerification"]; ok {
		if hours, ok := hoursVal.(float64); ok {
			data.HoursSinceLastVerification = hours
		}
	}
	
	// Extract machine structure from historical context if available
	if msVal, ok := historicalContext["machineStructure"]; ok {
		msData, err := models.ExtractMachineStructure(map[string]interface{}{"machineStructure": msVal})
		if err == nil && msData != nil {
			data.MachineStructure = msData
			data.RowCount = msData.RowCount
			data.ColumnCount = msData.ColumnsPerRow
			data.RowLabels = models.FormatArrayToString(msData.RowOrder)
			data.ColumnLabels = models.FormatArrayToString(msData.ColumnOrder)
			data.TotalPositions = msData.RowCount * msData.ColumnsPerRow
		}
	}
	
	// Extract verification summary if available
	vs, err := models.ExtractVerificationSummary(historicalContext)
	if err == nil && vs != nil {
		data.VerificationSummary = vs
	}
	
	return nil
}

// GenerateSystemPromptContent generates a system prompt from a verification context
func (p *TemplateProcessor) GenerateSystemPromptContent(vCtx *schema.VerificationContext, 
	layoutMetadata, historicalContext map[string]interface{}) (string, string, error) {
	
	// Build template data
	templateData, err := p.BuildTemplateData(vCtx, layoutMetadata, historicalContext)
	if err != nil {
		return "", "", fmt.Errorf("failed to build template data: %w", err)
	}
	
	// Get template version (from config or latest available)
	promptVersion := p.GetLatestVersion(vCtx.VerificationType)
	
	// Load template
	tmpl, err := p.GetTemplateWithVersion(vCtx.VerificationType, promptVersion)
	if err != nil {
		return "", "", fmt.Errorf("failed to get template: %w", err)
	}
	
	// Render template
	promptContent, err := p.RenderTemplate(tmpl, templateData)
	if err != nil {
		return "", "", fmt.Errorf("failed to render template: %w", err)
	}
	
	return promptContent, promptVersion, nil
}