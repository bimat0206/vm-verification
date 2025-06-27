package templateloader

import (
	"fmt"
	"text/template"
	"time"
)

// TemplateMetadata holds metadata about a template
type TemplateMetadata struct {
	Type     string                 `json:"type"`
	Version  string                 `json:"version"`
	Path     string                 `json:"path"`
	LoadedAt time.Time              `json:"loaded_at"`
	Size     int64                  `json:"size"`
	Checksum string                 `json:"checksum,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TemplateInfo represents information about an available template
type TemplateInfo struct {
	Type              string            `json:"type"`
	AvailableVersions []string          `json:"available_versions"`
	LatestVersion     string            `json:"latest_version"`
	Metadata          *TemplateMetadata `json:"metadata,omitempty"`
}

// LoaderStats provides statistics about the template loader
type LoaderStats struct {
	CacheEnabled    bool      `json:"cache_enabled"`
	CacheSize       int       `json:"cache_size"`
	CacheHits       int64     `json:"cache_hits"`
	CacheMisses     int64     `json:"cache_misses"`
	TemplatesLoaded int       `json:"templates_loaded"`
	TotalRenders    int64     `json:"total_renders"`
	LastRefresh     time.Time `json:"last_refresh"`
	TemplateTypes   []string  `json:"template_types"`
}

// RenderOptions provides options for template rendering
type RenderOptions struct {
	// IncludeMetadata adds metadata to the rendered output
	IncludeMetadata bool `json:"include_metadata"`

	// StrictMode enables strict template parsing (fails on undefined variables)
	StrictMode bool `json:"strict_mode"`

	// MaxExecutionTime sets a timeout for template execution
	MaxExecutionTime time.Duration `json:"max_execution_time"`

	// CustomFunctions allows passing additional functions for this render
	CustomFunctions template.FuncMap `json:"-"`
}

// RenderResult contains the result of a template render operation
type RenderResult struct {
	Output     string            `json:"output"`
	Metadata   *TemplateMetadata `json:"metadata,omitempty"`
	RenderTime time.Duration     `json:"render_time"`
	Size       int               `json:"size"`
	Error      error             `json:"error,omitempty"`
}

// ValidationRule defines a rule for template validation
type ValidationRule struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Required    bool                    `json:"required"`
	Pattern     string                  `json:"pattern,omitempty"`
	Validator   func(interface{}) error `json:"-"`
}

// ValidationResult contains the result of template validation
type ValidationResult struct {
	Valid    bool             `json:"valid"`
	Errors   []string         `json:"errors,omitempty"`
	Warnings []string         `json:"warnings,omitempty"`
	Rules    []ValidationRule `json:"rules,omitempty"`
}

// TemplateSource defines where templates are loaded from
type TemplateSource int

const (
	// SourceUnknown represents an unknown template source
	SourceUnknown TemplateSource = iota

	// SourceFilesystem represents templates loaded from filesystem
	SourceFilesystem

	// SourceVersioned represents templates loaded from versioned directories
	SourceVersioned

	// SourceFlat represents templates loaded from flat file structure
	SourceFlat
)

// String returns string representation of TemplateSource
func (ts TemplateSource) String() string {
	switch ts {
	case SourceFilesystem:
		return "filesystem"
	case SourceVersioned:
		return "versioned"
	case SourceFlat:
		return "flat"
	default:
		return "unknown"
	}
}

// ErrorType represents different types of template errors
type ErrorType string

const (
	// ErrorTypeNotFound indicates template was not found
	ErrorTypeNotFound ErrorType = "not_found"

	// ErrorTypeParsing indicates template parsing failed
	ErrorTypeParsing ErrorType = "parsing"

	// ErrorTypeExecution indicates template execution failed
	ErrorTypeExecution ErrorType = "execution"

	// ErrorTypeValidation indicates template validation failed
	ErrorTypeValidation ErrorType = "validation"

	// ErrorTypeConfiguration indicates configuration error
	ErrorTypeConfiguration ErrorType = "configuration"
)

// TemplateError represents a template-related error with additional context
type TemplateError struct {
	Type         ErrorType `json:"type"`
	Message      string    `json:"message"`
	TemplateType string    `json:"template_type,omitempty"`
	Version      string    `json:"version,omitempty"`
	Path         string    `json:"path,omitempty"`
	Line         int       `json:"line,omitempty"`
	Column       int       `json:"column,omitempty"`
	Cause        error     `json:"-"`
}

// Error implements the error interface
func (te *TemplateError) Error() string {
	if te.TemplateType != "" && te.Version != "" {
		return fmt.Sprintf("%s error in template %s:%s: %s", te.Type, te.TemplateType, te.Version, te.Message)
	}
	return fmt.Sprintf("%s error: %s", te.Type, te.Message)
}

// Unwrap returns the underlying error
func (te *TemplateError) Unwrap() error {
	return te.Cause
}

// ConfigDefaults provides default configuration values
type ConfigDefaults struct {
	BasePath        string
	CacheEnabled    bool
	MaxCacheSize    int
	RefreshInterval time.Duration
}

// GetDefaults returns default configuration values
func GetDefaults() ConfigDefaults {
	return ConfigDefaults{
		BasePath:        "/opt/templates",
		CacheEnabled:    true,
		MaxCacheSize:    100,
		RefreshInterval: 5 * time.Minute,
	}
}

// TemplateFunc represents a custom template function
type TemplateFunc struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Function    interface{} `json:"-"`
	Example     string      `json:"example,omitempty"`
}

// TemplateFuncRegistry holds registered template functions
type TemplateFuncRegistry struct {
	functions map[string]TemplateFunc
}

// NewTemplateFuncRegistry creates a new function registry
func NewTemplateFuncRegistry() *TemplateFuncRegistry {
	return &TemplateFuncRegistry{
		functions: make(map[string]TemplateFunc),
	}
}

// Register adds a new template function
func (tfr *TemplateFuncRegistry) Register(name string, fn TemplateFunc) {
	tfr.functions[name] = fn
}

// Get retrieves a template function by name
func (tfr *TemplateFuncRegistry) Get(name string) (TemplateFunc, bool) {
	fn, exists := tfr.functions[name]
	return fn, exists
}

// List returns all registered function names
func (tfr *TemplateFuncRegistry) List() []string {
	names := make([]string, 0, len(tfr.functions))
	for name := range tfr.functions {
		names = append(names, name)
	}
	return names
}

// ToFuncMap converts the registry to a template.FuncMap
func (tfr *TemplateFuncRegistry) ToFuncMap() template.FuncMap {
	funcMap := make(template.FuncMap)
	for name, fn := range tfr.functions {
		funcMap[name] = fn.Function
	}
	return funcMap
}

// VersionInfo contains information about a specific template version
type VersionInfo struct {
	Version  string    `json:"version"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"mod_time"`
	IsLatest bool      `json:"is_latest"`
}
