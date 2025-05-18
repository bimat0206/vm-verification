package internal

import (
	"fmt"
	"strings"
	"text/template"
)

// ProcessTemplate renders a template with the provided data
func ProcessTemplate(tmpl *template.Template, data TemplateData) (string, error) {
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	return buf.String(), nil
}
