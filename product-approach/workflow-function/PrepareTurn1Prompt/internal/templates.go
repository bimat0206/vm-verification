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

// Turn1TemplatePrefix is the prefix for Turn 1 templates
const Turn1TemplatePrefix = "turn1-"

// Custom template functions
var templateFuncs = template.FuncMap{
	"split": strings.Split,
	"join":  strings.Join,
	"add": func(a, b int) int {
		return a + b
	},
	"sub": func(a, b int) int {
		return a - b
	},
	"mul": func(a, b int) int {
		return a * b
	},
	"div": func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a / b
	},
	"gt": func(a, b int) bool {
		return a > b
	},
	"lt": func(a, b int) bool {
		return a < b
	},
	"eq": func(a, b interface{}) bool {
		return a == b
	},
	"index": func(arr []string, i int) string {
		if i < 0 || i >= len(arr) {
			return ""
		}
		return arr[i]
	},
	"ordinal": func(num int) string {
		switch num {
		case 1:
			return "first"
		case 2:
			return "second"
		case 3:
			return "third"
		case 4:
			return "fourth"
		case 5:
			return "fifth"
		case 6:
			return "sixth"
		case 7:
			return "seventh"
		case 8:
			return "eighth"
		case 9:
			return "ninth"
		case 10:
			return "tenth"
		default:
			suffix := "th"
			if num%10 == 1 && num%100 != 11 {
				suffix = "st"
			} else if num%10 == 2 && num%100 != 12 {
				suffix = "nd"
			} else if num%10 == 3 && num%100 != 13 {
				suffix = "rd"
			}
			return fmt.Sprintf("%d%s", num, suffix)
		}
	},
}

// NewTemplateManager creates a new TemplateManager
func NewTemplateManager(baseDir string) *TemplateManager {
	tm := &TemplateManager{
		baseDir:        baseDir,
		templates:      make(map[string]*template.Template), // For system prompts
		turn1Templates: make(map[string]*template.Template), // For Turn 1 prompts
		versions:       make(map[string]string),             // For system prompt versions
		turn1Versions:  make(map[string]string),             // For Turn 1 template versions
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
	
	// Parse template with custom functions
	tmpl, err := template.New(templateType).Funcs(templateFuncs).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	
	// Cache template
	tm.templates[templateKey] = tmpl
	
	log.Printf("Loaded template: %s v%s", templateType, version)
	return tmpl, nil
}

// GetTurn1Template retrieves and caches a Turn 1 template by verification type
func (tm *TemplateManager) GetTurn1Template(verificationType string) (*template.Template, error) {
	// Normalize verification type for file system
	templateKey := strings.ToLower(verificationType)
	
	// Check if template is already cached
	if tmpl, exists := tm.turn1Templates[templateKey]; exists {
		return tmpl, nil
	}
	
	// Get latest version for this template type
	version := tm.getLatestTurn1Version(verificationType)
	if version == "" {
		return nil, fmt.Errorf("no Turn 1 template version found for type: %s", verificationType)
	}
	
	// Build template path
	templateType := strings.ReplaceAll(strings.ToLower(verificationType), "_", "-")
	templatePath := filepath.Join(tm.baseDir, templateType, fmt.Sprintf("%sv%s.tmpl", Turn1TemplatePrefix, version))
	
	// Read template file
	content, err := ioutil.ReadFile(templatePath)
	if err != nil {
		// Try alternate format if first attempt fails
		altTemplatePath := filepath.Join(tm.baseDir, fmt.Sprintf("%s-%s.tmpl", Turn1TemplatePrefix, templateType))
		content, err = ioutil.ReadFile(altTemplatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read Turn 1 template file at %s or %s: %w", 
				templatePath, altTemplatePath, err)
		}
	}
	
	// Parse template with custom functions
	tmpl, err := template.New(templateType + "-turn1").Funcs(templateFuncs).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Turn 1 template: %w", err)
	}
	
	// Cache template
	tm.turn1Templates[templateKey] = tmpl
	
	log.Printf("Loaded Turn 1 template: %s v%s", templateType, version)
	return tmpl, nil
}

// GetLatestVersion returns the latest version for a verification type (exported version)
func (tm *TemplateManager) GetLatestVersion(verificationType string) string {
    return tm.getLatestVersion(verificationType)
}

// GetLatestTurn1Version returns the latest Turn 1 version for a verification type
func (tm *TemplateManager) GetLatestTurn1Version(verificationType string) string {
    return tm.getLatestTurn1Version(verificationType)
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

// getLatestTurn1Version returns the latest Turn 1 version for a verification type
func (tm *TemplateManager) getLatestTurn1Version(verificationType string) string {
	normalizedType := strings.ToLower(verificationType)
	
	// Check if we have a cached version
	if version, exists := tm.turn1Versions[normalizedType]; exists {
		return version
	}
	
	// Check environment variables
	envKey := "TURN1_TEMPLATE_VERSION_" + strings.ToUpper(strings.ReplaceAll(verificationType, "-", "_"))
	version := os.Getenv(envKey)
	if version != "" {
		return version
	}
	
	// Fallback to default versions
	defaultVersions := map[string]string{
		"layout_vs_checking":  "1.0.0",
		"previous_vs_current": "1.0.0",
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
			
			// Extract version numbers from filenames - system templates and turn1 templates
			var versions []string
			var turn1Versions []string
			
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".tmpl") {
					name := file.Name()
					// Check for Turn 1 templates
					if strings.HasPrefix(name, Turn1TemplatePrefix) && strings.Contains(name, "v") {
						// Extract version for Turn 1 template (turn1-v1.0.0.tmpl -> 1.0.0)
						versionPart := strings.TrimPrefix(name, Turn1TemplatePrefix)
						version := strings.TrimPrefix(strings.TrimSuffix(versionPart, ".tmpl"), "v")
						turn1Versions = append(turn1Versions, version)
					} else if strings.HasPrefix(name, "v") && strings.HasSuffix(name, ".tmpl") {
						// Regular system template (v1.0.0.tmpl -> 1.0.0)
						version := strings.TrimPrefix(strings.TrimSuffix(name, ".tmpl"), "v")
						versions = append(versions, version)
					}
				}
			}
			
			// Sort versions (simple string sort for semver-like versions)
			sort.Strings(versions)
			sort.Strings(turn1Versions)
			
			// Store the latest system template version if any were found
			if len(versions) > 0 {
				latestVersion := versions[len(versions)-1]
				tm.versions[normalizedType] = latestVersion
				log.Printf("Discovered system template version: %s v%s", normalizedType, latestVersion)
			}
			
			// Store the latest Turn 1 template version if any were found
			if len(turn1Versions) > 0 {
				latestVersion := turn1Versions[len(turn1Versions)-1]
				tm.turn1Versions[normalizedType] = latestVersion
				log.Printf("Discovered Turn 1 template version: %s v%s", normalizedType, latestVersion)
			}
		}
	}
}

// ListTemplateVersions returns the discovered template versions
func (tm *TemplateManager) ListTemplateVersions() map[string]map[string]string {
	result := make(map[string]map[string]string)
	
	systemVersions := make(map[string]string)
	for templateType, version := range tm.versions {
		systemVersions[templateType] = version
	}
	
	turn1Versions := make(map[string]string)
	for templateType, version := range tm.turn1Versions {
		turn1Versions[templateType] = version
	}
	
	result["system"] = systemVersions
	result["turn1"] = turn1Versions
	
	return result
}

// ProcessTemplate renders a template with the provided data
func ProcessTemplate(tmpl *template.Template, data TemplateData) (string, error) {
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	return buf.String(), nil
}