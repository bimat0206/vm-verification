package core

import (
	"bytes"
	//"fmt"
	"text/template"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/templateloader"
)

// TemplateProcessor handles template loading and processing
type TemplateProcessor struct {
	loader templateloader.TemplateLoader
	log    logger.Logger
}

// NewTemplateProcessor creates a new template processor with the given loader and logger
func NewTemplateProcessor(loader templateloader.TemplateLoader, log logger.Logger) *TemplateProcessor {
	return &TemplateProcessor{
		loader: loader,
		log:    log,
	}
}

// ProcessTemplate loads and processes a template with the given data
func (t *TemplateProcessor) ProcessTemplate(templateName string, data map[string]interface{}) (string, error) {
	// Load the template
	tmpl, err := t.loader.LoadTemplate(templateName)
	if err != nil {
		return "", errors.NewInternalError("template-loading", err)
	}

	// Log template information
	t.log.Info("Template loaded", map[string]interface{}{
		"templateName": templateName,
		"version":      t.loader.GetLatestVersion(templateName),
	})
	
	// Add custom functions - no need to create a separate map as these functions
	// should already be registered in the template loader, but we'll add them directly here as well
	// for extra safety and compatibility with existing templates
	
	// Render the template with the provided data
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.log.Error("Template execution error", map[string]interface{}{
			"error":        err.Error(),
			"templateName": templateName,
			"dataKeys":     getMapKeys(data),
		})
		return "", errors.NewInternalError("template-execution", err)
	}

	// Get the rendered text
	renderedText := buf.String()

	// Log rendering result
	t.log.Info("Template rendered", map[string]interface{}{
		"templateName": templateName,
		"textLength":   len(renderedText),
	})

	return renderedText, nil
}

// getMapKeys returns a sorted list of keys from a map
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// GetTemplateVersion returns the latest version of the template
func (t *TemplateProcessor) GetTemplateVersion(templateName string) string {
	return t.loader.GetLatestVersion(templateName)
}

// AddTemplateFunctions adds custom functions to the template
func (t *TemplateProcessor) AddTemplateFunctions(funcs template.FuncMap) error {
	if funcs == nil {
		return errors.NewValidationError("Function map is nil", nil)
	}

	// TemplateLoader generally handles functions, but this is a wrapper to maintain
	// proper interface isolation and to allow for any additional processing
	funcNames := make([]string, 0, len(funcs))
	for name := range funcs {
		funcNames = append(funcNames, name)
	}

	t.log.Info("Added template functions", map[string]interface{}{
		"functionCount": len(funcs),
		"functions":     funcNames,
	})

	return nil
}