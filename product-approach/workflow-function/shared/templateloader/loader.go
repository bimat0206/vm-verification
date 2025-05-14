package templateloader

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

// Config holds configuration for the template loader
type Config struct {
	BasePath     string           `yaml:"base_path" json:"base_path"`
	CacheEnabled bool             `yaml:"cache_enabled" json:"cache_enabled"`
	CustomFuncs  template.FuncMap `yaml:"-" json:"-"`
}

// Loader handles template loading and rendering
type Loader struct {
	basePath  string
	cache     map[string]*template.Template
	functions template.FuncMap
	versions  map[string]string
	mu        sync.RWMutex
}

// TemplateLoader defines the interface for template operations
type TemplateLoader interface {
	LoadTemplate(templateType string) (*template.Template, error)
	LoadTemplateWithVersion(templateType, version string) (*template.Template, error)
	RenderTemplate(templateType string, data interface{}) (string, error)
	RenderTemplateWithVersion(templateType, version string, data interface{}) (string, error)
	GetLatestVersion(templateType string) string
	ListVersions(templateType string) []string
	ClearCache() error
	RefreshVersions() error
}

// New creates a new template loader with the given configuration
func New(config Config) (*Loader, error) {
	if config.BasePath == "" {
		config.BasePath = "/opt/templates"
	}

	loader := &Loader{
		basePath:  config.BasePath,
		functions: make(template.FuncMap),
		versions:  make(map[string]string),
	}

	// Initialize cache if enabled
	if config.CacheEnabled {
		loader.cache = make(map[string]*template.Template)
	}

	// Add default functions
	for name, fn := range DefaultFunctions {
		loader.functions[name] = fn
	}

	// Add custom functions if provided
	if config.CustomFuncs != nil {
		for name, fn := range config.CustomFuncs {
			loader.functions[name] = fn
		}
	}

	// Discover available template versions
	if err := loader.discoverVersions(); err != nil {
		return nil, fmt.Errorf("failed to discover template versions: %w", err)
	}

	return loader, nil
}

// LoadTemplate loads the latest version of a template
func (l *Loader) LoadTemplate(templateType string) (*template.Template, error) {
	version := l.GetLatestVersion(templateType)
	if version == "" {
		return nil, fmt.Errorf("no template found for type: %s", templateType)
	}
	return l.LoadTemplateWithVersion(templateType, version)
}

// LoadTemplateWithVersion loads a specific version of a template
func (l *Loader) LoadTemplateWithVersion(templateType, version string) (*template.Template, error) {
	// Check cache first if enabled
	if l.cache != nil {
		cacheKey := fmt.Sprintf("%s:%s", templateType, version)
		l.mu.RLock()
		if tmpl, exists := l.cache[cacheKey]; exists {
			l.mu.RUnlock()
			return tmpl, nil
		}
		l.mu.RUnlock()
	}

	// Try versioned template first
	templatePath := l.getVersionedTemplatePath(templateType, version)
	content, err := ioutil.ReadFile(templatePath)
	if err != nil {
		// Try flat file structure as fallback
		templatePath = l.getFlatTemplatePath(templateType)
		content, err = ioutil.ReadFile(templatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read template file at %s: %w", templatePath, err)
		}
	}

	// Parse template with custom functions
	tmpl, err := template.New(templateType).Funcs(l.functions).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Cache the template if caching is enabled
	if l.cache != nil {
		cacheKey := fmt.Sprintf("%s:%s", templateType, version)
		l.mu.Lock()
		l.cache[cacheKey] = tmpl
		l.mu.Unlock()
	}

	return tmpl, nil
}

// RenderTemplate renders the latest version of a template with data
func (l *Loader) RenderTemplate(templateType string, data interface{}) (string, error) {
	version := l.GetLatestVersion(templateType)
	if version == "" {
		return "", fmt.Errorf("no template found for type: %s", templateType)
	}
	return l.RenderTemplateWithVersion(templateType, version, data)
}

// RenderTemplateWithVersion renders a specific version of a template with data
func (l *Loader) RenderTemplateWithVersion(templateType, version string, data interface{}) (string, error) {
	tmpl, err := l.LoadTemplateWithVersion(templateType, version)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return buf.String(), nil
}

// GetLatestVersion returns the latest version for a template type
func (l *Loader) GetLatestVersion(templateType string) string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.versions[templateType]
}

// ListVersions returns all available versions for a template type, sorted
func (l *Loader) ListVersions(templateType string) []string {
	templateDir := filepath.Join(l.basePath, l.normalizeTemplateType(templateType))

	files, err := ioutil.ReadDir(templateDir)
	if err != nil {
		return []string{}
	}

	var versions []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".tmpl") {
			if strings.HasPrefix(file.Name(), "v") {
				version := strings.TrimSuffix(strings.TrimPrefix(file.Name(), "v"), ".tmpl")
				versions = append(versions, version)
			}
		}
	}

	// Sort versions semantically
	sort.Sort(semVerSlice(versions))
	return versions
}

// ClearCache clears the template cache
func (l *Loader) ClearCache() error {
	if l.cache == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache = make(map[string]*template.Template)
	return nil
}

// RefreshVersions re-discovers available template versions
func (l *Loader) RefreshVersions() error {
	return l.discoverVersions()
}

// Private methods

// discoverVersions scans the template directory to find available versions
func (l *Loader) discoverVersions() error {
	entries, err := ioutil.ReadDir(l.basePath)
	if err != nil {
		return fmt.Errorf("failed to read template base directory %s: %w", l.basePath, err)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for _, entry := range entries {
		if entry.IsDir() {
			templateType := entry.Name()
			versions := l.ListVersions(templateType)
			if len(versions) > 0 {
				l.versions[templateType] = versions[len(versions)-1] // Latest version
			}
		}
	}

	return nil
}

// normalizeTemplateType converts template type to filesystem format
func (l *Loader) normalizeTemplateType(templateType string) string {
	return strings.ReplaceAll(strings.ToLower(templateType), "_", "-")
}

// getVersionedTemplatePath returns the path for a versioned template
func (l *Loader) getVersionedTemplatePath(templateType, version string) string {
	templateDir := l.normalizeTemplateType(templateType)
	filename := fmt.Sprintf("v%s.tmpl", version)
	return filepath.Join(l.basePath, templateDir, filename)
}

// getFlatTemplatePath returns the path for a flat template structure
func (l *Loader) getFlatTemplatePath(templateType string) string {
	templateDir := l.normalizeTemplateType(templateType)
	filename := templateDir + ".tmpl"
	return filepath.Join(l.basePath, filename)
}

// semVerSlice implements sort.Interface for semantic version sorting
type semVerSlice []string

func (s semVerSlice) Len() int      { return len(s) }
func (s semVerSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s semVerSlice) Less(i, j int) bool {
	return compareVersions(s[i], s[j]) < 0
}

// compareVersions compares two semantic version strings
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			n1 = parseVersionPart(parts1[i])
		}
		if i < len(parts2) {
			n2 = parseVersionPart(parts2[i])
		}

		if n1 < n2 {
			return -1
		} else if n1 > n2 {
			return 1
		}
	}

	return 0
}

// parseVersionPart parses a version part to integer, handling non-numeric parts gracefully
func parseVersionPart(s string) int {
	num, err := strconv.Atoi(s)
	if err != nil {
		// For non-numeric parts, use ASCII value sum as fallback
		sum := 0
		for _, r := range s {
			sum += int(r)
		}
		return sum
	}
	return num
}