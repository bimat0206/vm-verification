package internal

import (
	"fmt"
	"log"

	//"os"
	"strings"
	"text/template"

	"workflow-function/shared/templateloader"
)

var tmplLoader *templateloader.Loader

// InitializeTemplateLoader creates and initializes a new template loader
func InitializeTemplateLoader(basePath string) error {
	// Get configuration for template loader
	config := templateloader.Config{
		BasePath:     basePath,
		CacheEnabled: true,
		// We'll use the default template functions provided by the loader
	}
	
	// Initialize template loader
	var err error
	tmplLoader, err = templateloader.New(config)
	if err != nil {
		return fmt.Errorf("failed to initialize template loader: %w", err)
	}
	
	// Log available templates
	templates := make(map[string]string)
	for _, templateType := range []string{"layout-vs-checking", "previous-vs-current"} {
		versions := tmplLoader.ListVersions(templateType)
		if len(versions) > 0 {
			templates[templateType] = versions[len(versions)-1]
		}
	}
	if len(templates) > 0 {
		log.Printf("Available templates: %v", templates)
	}
	
	return nil
}

// GetTemplate retrieves a template by verification type
func GetTemplate(verificationType string) (*template.Template, error) {
	if tmplLoader == nil {
		return nil, fmt.Errorf("template loader not initialized")
	}
	
	// Normalize verification type for template loader
	templateType := MapVerificationTypeToTemplateType(verificationType)
	
	// Get the template from the loader
	tmpl, err := tmplLoader.LoadTemplate(templateType)
	if err != nil {
		return nil, fmt.Errorf("failed to load template for type %s: %w", verificationType, err)
	}
	
	return tmpl, nil
}

// GetTemplateWithVersion retrieves a specific version of a template
func GetTemplateWithVersion(verificationType, version string) (*template.Template, error) {
	if tmplLoader == nil {
		return nil, fmt.Errorf("template loader not initialized")
	}
	
	// Normalize verification type for template loader
	templateType := MapVerificationTypeToTemplateType(verificationType)
	
	// Get the template from the loader
	tmpl, err := tmplLoader.LoadTemplateWithVersion(templateType, version)
	if err != nil {
		return nil, fmt.Errorf("failed to load template %s version %s: %w", 
			verificationType, version, err)
	}
	
	return tmpl, nil
}

// RenderTemplate renders a template with data
func RenderTemplate(verificationType string, data interface{}) (string, error) {
	if tmplLoader == nil {
		return "", fmt.Errorf("template loader not initialized")
	}
	
	// Normalize verification type for template loader
	templateType := MapVerificationTypeToTemplateType(verificationType)
	
	// Render the template with data
	result, err := tmplLoader.RenderTemplate(templateType, data)
	if err != nil {
		return "", fmt.Errorf("failed to render template for type %s: %w", verificationType, err)
	}
	
	return result, nil
}

// GetLatestVersion returns the latest version for a template type
func GetLatestVersion(verificationType string) string {
	if tmplLoader == nil {
		// Fallback to default versions if loader not initialized
		defaultVersions := map[string]string{
			"layout_vs_checking":  "1.2.3",
			"previous_vs_current": "1.1.0",
		}
		return defaultVersions[strings.ToLower(verificationType)]
	}
	
	// Normalize verification type for template loader
	templateType := MapVerificationTypeToTemplateType(verificationType)
	
	// Get the latest version from the loader
	version := tmplLoader.GetLatestVersion(templateType)
	if version == "" {
		// Fallback to default versions
		defaultVersions := map[string]string{
			"layout-vs-checking":  "1.2.3",
			"previous-vs-current": "1.1.0",
		}
		return defaultVersions[templateType]
	}
	
	return version
}

// ListAvailableTemplates returns a list of available templates and versions
func ListAvailableTemplates() map[string]string {
	if tmplLoader == nil {
		return map[string]string{}
	}
	
	// Get available templates and their latest versions
	result := make(map[string]string)
	for _, templateType := range []string{"layout-vs-checking", "previous-vs-current"} {
		versions := tmplLoader.ListVersions(templateType)
		if len(versions) > 0 {
			result[templateType] = versions[len(versions)-1] // Latest version
		}
	}
	
	return result
}

// MapVerificationTypeToTemplateType converts a verification type to a template type
func MapVerificationTypeToTemplateType(verificationType string) string {
	// Convert to lowercase and replace underscores with hyphens
	return strings.ReplaceAll(strings.ToLower(verificationType), "_", "-")
}

// ProcessTemplate renders a template with the provided data
// Compatibility function for existing code
func ProcessTemplate(tmpl *template.Template, data TemplateData) (string, error) {
	if tmpl == nil {
		return "", fmt.Errorf("template is nil")
	}
	
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	return buf.String(), nil
}