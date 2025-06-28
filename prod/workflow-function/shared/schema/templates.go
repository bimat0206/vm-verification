// Package schema provides template utilities for combined functions
package schema

import (
    "fmt"
    "strings"
    "text/template"
)

// TemplateManager handles template operations
type TemplateManager struct {
    templates map[string]*PromptTemplate
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
    return &TemplateManager{
        templates: make(map[string]*PromptTemplate),
    }
}

// RegisterTemplate registers a new template
func (tm *TemplateManager) RegisterTemplate(tmpl *PromptTemplate) error {
    if tmpl.TemplateId == "" {
        return fmt.Errorf("template ID cannot be empty")
    }
    
    tm.templates[tmpl.TemplateId] = tmpl
    return nil
}

// ProcessTemplate processes a template with given context
func (tm *TemplateManager) ProcessTemplate(templateId string, context map[string]interface{}) (*TemplateProcessor, error) {
    tmpl, exists := tm.templates[templateId]
    if !exists {
        return nil, fmt.Errorf("template not found: %s", templateId)
    }
    
    // Process template with context
    t, err := template.New(templateId).Parse(tmpl.Content)
    if err != nil {
        return nil, fmt.Errorf("failed to parse template: %w", err)
    }
    
    var buf strings.Builder
    if err := t.Execute(&buf, context); err != nil {
        return nil, fmt.Errorf("failed to execute template: %w", err)
    }
    
    return &TemplateProcessor{
        Template:        tmpl,
        ContextData:     context,
        ProcessedPrompt: buf.String(),
        ProcessingTime:  0, // Set by caller
        InputTokens:     0, // Set by caller after Bedrock response
        OutputTokens:    0, // Set by caller after Bedrock response
    }, nil
}

// GetTemplateTypes returns available template types
func GetTemplateTypes() []string {
    return []string{
        "turn1-layout-vs-checking",
        "turn1-previous-vs-current",
        "turn2-layout-vs-checking", 
        "turn2-previous-vs-current",
        "turn2-historical-enhancement",
    }
}

// ADD: Use case specific template functions
func GetTemplateForUseCase(verificationType string, turnNumber int) (string, error) {
    switch verificationType {
    case VerificationTypeLayoutVsChecking:
        if turnNumber == 1 {
            return "turn1-layout-vs-checking", nil
        }
        return "turn2-layout-vs-checking", nil
    case VerificationTypePreviousVsCurrent:
        if turnNumber == 1 {
            return "turn1-previous-vs-current", nil
        }
        return "turn2-previous-vs-current", nil
    default:
        return "", fmt.Errorf("unknown verification type: %s", verificationType)
    }
}

// ADD: Enhanced template context builder
func BuildTemplateContext(verificationContext *VerificationContext, layoutMetadata *LayoutMetadata, historicalContext map[string]interface{}) *TemplateContext {
    ctx := &TemplateContext{
        VerificationType: verificationContext.VerificationType,
    }
    
    if layoutMetadata != nil {
        ctx.LayoutMetadata = map[string]interface{}{
            "layoutId":           layoutMetadata.LayoutId,
            "layoutPrefix":       layoutMetadata.LayoutPrefix,
            "machineStructure":   layoutMetadata.MachineStructure,
            "productPositionMap": layoutMetadata.ProductPositionMap,
        }
        ctx.MachineStructure = layoutMetadata.MachineStructure
    }
    
    if historicalContext != nil {
        ctx.HistoricalContext = historicalContext
    }
    
    return ctx
}