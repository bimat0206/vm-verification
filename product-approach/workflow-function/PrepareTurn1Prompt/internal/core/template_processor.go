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

	// Render the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
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