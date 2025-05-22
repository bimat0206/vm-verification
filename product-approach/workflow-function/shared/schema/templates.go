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
        TokenEstimate:   len(buf.String()) / 4, // Rough estimate
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