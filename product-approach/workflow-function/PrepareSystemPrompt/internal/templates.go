package internal

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

// NewTemplateManager creates a new TemplateManager
func NewTemplateManager(baseDir string) *TemplateManager {
	tm := &TemplateManager{
		baseDir:   baseDir,
		templates: make(map[string]*template.Template),
		versions:  make(map[string]string),
	}
	
	// Load version information
	tm.discoverTemplateVersions()
	
	return tm
}

// GetTemplate retrieves and caches a template by verification type
func (tm *TemplateManager) GetTemplate(verificationType string) (*template.Template, error) {
	// Normalize verification type for file system
	templateKey := strings.ToLower(verificationType)
	
	// Check if template is already cached
	if tmpl, exists := tm.templates[templateKey]; exists {
		return tmpl, nil
	}
	
	// Get latest version for this template type
	version := tm.getLatestVersion(verificationType)
	if version == "" {
		return nil, fmt.Errorf("no template version found for type: %s", verificationType)
	}
	
	// Build template path
	templateType := strings.ReplaceAll(strings.ToLower(verificationType), "_", "-")
	templatePath := filepath.Join(tm.baseDir, templateType, fmt.Sprintf("v%s.tmpl", version))
	
	// Read template file
	content, err := ioutil.ReadFile(templatePath)
	if err != nil {
		// Try alternate format if first attempt fails
		altTemplatePath := filepath.Join(tm.baseDir, templateType + ".tmpl")
		content, err = ioutil.ReadFile(altTemplatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read template file at %s or %s: %w", 
				templatePath, altTemplatePath, err)
		}
	}
	
	// Parse template
	tmpl, err := template.New(templateType).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	
	// Cache template
	tm.templates[templateKey] = tmpl
	
	log.Printf("Loaded template: %s v%s", templateType, version)
	return tmpl, nil
}

// getLatestVersion returns the latest version for a verification type
func (tm *TemplateManager) getLatestVersion(verificationType string) string {
	normalizedType := strings.ToLower(verificationType)
	
	// Check if we have a cached version
	if version, exists := tm.versions[normalizedType]; exists {
		return version
	}
	
	// Check environment variables
	envKey := "TEMPLATE_VERSION_" + strings.ToUpper(strings.ReplaceAll(verificationType, "-", "_"))
	version := os.Getenv(envKey)
	if version != "" {
		return version
	}
	
	// Fallback to default versions
	defaultVersions := map[string]string{
		"layout_vs_checking":  "1.2.3",
		"previous_vs_current": "1.1.0",
	}
	
	return defaultVersions[normalizedType]
}

// discoverTemplateVersions scans the template directory to find available versions
func (tm *TemplateManager) discoverTemplateVersions() {
	// Ensure the template directory exists
	if _, err := os.Stat(tm.baseDir); os.IsNotExist(err) {
		log.Printf("Warning: Template base directory does not exist: %s", tm.baseDir)
		return
	}
	
	// Read directory entries
	entries, err := ioutil.ReadDir(tm.baseDir)
	if err != nil {
		log.Printf("Warning: Failed to read template directory: %v", err)
		return
	}
	
	// Process each verification type directory
	for _, entry := range entries {
		if entry.IsDir() {
			templateType := entry.Name()
			normalizedType := strings.ReplaceAll(templateType, "-", "_")
			
			// Get all template files in this directory
			templateDir := filepath.Join(tm.baseDir, templateType)
			files, err := ioutil.ReadDir(templateDir)
			if err != nil {
				log.Printf("Warning: Failed to read template type directory %s: %v", templateType, err)
				continue
			}
			
			// Extract version numbers from filenames
			var versions []string
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".tmpl") {
					name := file.Name()
					if strings.HasPrefix(name, "v") && strings.HasSuffix(name, ".tmpl") {
						version := strings.TrimPrefix(strings.TrimSuffix(name, ".tmpl"), "v")
						versions = append(versions, version)
					}
				}
			}
			
			// Sort versions (simple string sort for semver-like versions)
			sort.Strings(versions)
			
			// Store the latest version if any were found
			if len(versions) > 0 {
				latestVersion := versions[len(versions)-1]
				tm.versions[normalizedType] = latestVersion
				log.Printf("Discovered template version: %s v%s", normalizedType, latestVersion)
			}
		}
	}
}

// ListAvailableTemplates returns a list of available templates and versions
func (tm *TemplateManager) ListAvailableTemplates() map[string]string {
	result := make(map[string]string)
	
	for templateType, version := range tm.versions {
		result[templateType] = version
	}
	
	return result
}