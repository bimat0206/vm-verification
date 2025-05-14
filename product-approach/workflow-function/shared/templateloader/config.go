package templateloader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigManager handles configuration loading and validation
type ConfigManager struct {
	config    *Config
	validator *ConfigValidator
}

// ConfigValidator validates configuration values
type ConfigValidator struct {
	rules []ValidationRule
}

// AdvancedConfig extends the basic Config with more options
type AdvancedConfig struct {
	Config `yaml:",inline" json:",inline"`
	
	// Template discovery options
	Discovery DiscoveryConfig `yaml:"discovery" json:"discovery"`
	
	// Cache configuration
	CacheConfig CacheConfig `yaml:"cache" json:"cache"`
	
	// Performance settings
	Performance PerformanceConfig `yaml:"performance" json:"performance"`
	
	// Logging configuration
	Logging LoggingConfig `yaml:"logging" json:"logging"`
	
	// Security settings
	Security SecurityConfig `yaml:"security" json:"security"`
	
	// Environment overrides
	Environment map[string]string `yaml:"environment" json:"environment"`
}

// DiscoveryConfig configures template discovery behavior
type DiscoveryConfig struct {
	// Recursive search in subdirectories
	Recursive bool `yaml:"recursive" json:"recursive"`
	
	// Patterns to include/exclude
	IncludePatterns []string `yaml:"include_patterns" json:"include_patterns"`
	ExcludePatterns []string `yaml:"exclude_patterns" json:"exclude_patterns"`
	
	// File extensions to consider
	Extensions []string `yaml:"extensions" json:"extensions"`
	
	// Watch for file changes
	WatchEnabled bool `yaml:"watch_enabled" json:"watch_enabled"`
	WatchInterval time.Duration `yaml:"watch_interval" json:"watch_interval"`
}

// CacheConfig configures caching behavior
type CacheConfig struct {
	// Cache type (memory, redis, etc.)
	Type string `yaml:"type" json:"type"`
	
	// Memory cache settings
	Memory MemoryCacheConfig `yaml:"memory" json:"memory"`
	
	// Redis cache settings (for future extension)
	Redis RedisCacheConfig `yaml:"redis" json:"redis"`
	
	// General cache settings
	TTL time.Duration `yaml:"ttl" json:"ttl"`
	DefaultTTL time.Duration `yaml:"default_ttl" json:"default_ttl"`
}

// MemoryCacheConfig configures in-memory cache
type MemoryCacheConfig struct {
	MaxSize int `yaml:"max_size" json:"max_size"`
	EvictionPolicy string `yaml:"eviction_policy" json:"eviction_policy"` // LRU, LFU, FIFO
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
}

// RedisCacheConfig configures Redis cache (placeholder for future)
type RedisCacheConfig struct {
	Addr string `yaml:"addr" json:"addr"`
	Password string `yaml:"password" json:"password"`
	DB int `yaml:"db" json:"db"`
	KeyPrefix string `yaml:"key_prefix" json:"key_prefix"`
}

// PerformanceConfig configures performance settings
type PerformanceConfig struct {
	MaxConcurrentLoads int `yaml:"max_concurrent_loads" json:"max_concurrent_loads"`
	MaxMemoryUsage string `yaml:"max_memory_usage" json:"max_memory_usage"`
	GCInterval time.Duration `yaml:"gc_interval" json:"gc_interval"`
	PrecompileTemplates bool `yaml:"precompile_templates" json:"precompile_templates"`
}

// LoggingConfig configures logging
type LoggingConfig struct {
	Level string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"` // json, text
	Output string `yaml:"output" json:"output"` // stdout, stderr, file
	File string `yaml:"file" json:"file"`
}

// SecurityConfig configures security settings
type SecurityConfig struct {
	MaxTemplateSize string `yaml:"max_template_size" json:"max_template_size"`
	AllowedFunctions []string `yaml:"allowed_functions" json:"allowed_functions"`
	RestrictedPaths []string `yaml:"restricted_paths" json:"restricted_paths"`
	SandboxMode bool `yaml:"sandbox_mode" json:"sandbox_mode"`
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		validator: NewConfigValidator(),
	}
}

// LoadFromFile loads configuration from a file (YAML or JSON)
func (cm *ConfigManager) LoadFromFile(path string) (*AdvancedConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	// Determine file format by extension
	ext := strings.ToLower(filepath.Ext(path))
	var config AdvancedConfig

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	// Apply environment overrides
	if err := cm.applyEnvironmentOverrides(&config); err != nil {
		return nil, fmt.Errorf("failed to apply environment overrides: %w", err)
	}

	// Validate configuration
	if err := cm.validator.Validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	cm.config = &config.Config
	return &config, nil
}

// LoadFromEnvironment loads configuration from environment variables
func (cm *ConfigManager) LoadFromEnvironment(prefix string) (*Config, error) {
	config := GetDefaultConfig()

	// Map of environment variables to config fields
	envMap := map[string]func(string) error{
		prefix + "_BASE_PATH": func(val string) error {
			config.BasePath = val
			return nil
		},
		prefix + "_CACHE_ENABLED": func(val string) error {
			enabled, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			config.CacheEnabled = enabled
			return nil
		},
	}

	// Apply environment variables
	for envVar, setter := range envMap {
		if val := os.Getenv(envVar); val != "" {
			if err := setter(val); err != nil {
				return nil, fmt.Errorf("invalid value for %s: %w", envVar, err)
			}
		}
	}

	cm.config = config
	return config, nil
}

// GetDefaultConfig returns default configuration
func GetDefaultConfig() *Config {
	defaults := GetDefaults()
	return &Config{
		BasePath:     defaults.BasePath,
		CacheEnabled: defaults.CacheEnabled,
		CustomFuncs:  make(template.FuncMap),
	}
}

// GetDefaultAdvancedConfig returns default advanced configuration
func GetDefaultAdvancedConfig() *AdvancedConfig {
	return &AdvancedConfig{
		Config: *GetDefaultConfig(),
		Discovery: DiscoveryConfig{
			Recursive:       false,
			IncludePatterns: []string{"*.tmpl"},
			ExcludePatterns: []string{},
			Extensions:      []string{".tmpl", ".template"},
			WatchEnabled:    false,
			WatchInterval:   30 * time.Second,
		},
		CacheConfig: CacheConfig{
			Type: "memory",
			Memory: MemoryCacheConfig{
				MaxSize:        100,
				EvictionPolicy: "LRU",
				CleanupInterval: 5 * time.Minute,
			},
			TTL:        1 * time.Hour,
			DefaultTTL: 1 * time.Hour,
		},
		Performance: PerformanceConfig{
			MaxConcurrentLoads:  10,
			MaxMemoryUsage:      "100MB",
			GCInterval:          10 * time.Minute,
			PrecompileTemplates: true,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Security: SecurityConfig{
			MaxTemplateSize: "1MB",
			AllowedFunctions: []string{},
			RestrictedPaths:  []string{},
			SandboxMode:     false,
		},
		Environment: make(map[string]string),
	}
}

// applyEnvironmentOverrides applies environment variable overrides
func (cm *ConfigManager) applyEnvironmentOverrides(config *AdvancedConfig) error {
	// Apply custom environment overrides
	for key, envVar := range config.Environment {
		if val := os.Getenv(envVar); val != "" {
			if err := cm.setConfigValue(config, key, val); err != nil {
				return fmt.Errorf("failed to set %s from %s: %w", key, envVar, err)
			}
		}
	}

	// Apply standard environment overrides
	standardOverrides := map[string]string{
		"TEMPLATE_BASE_PATH":     "Config.BasePath",
		"TEMPLATE_CACHE_ENABLED": "Config.CacheEnabled",
		"TEMPLATE_CACHE_SIZE":    "CacheConfig.Memory.MaxSize",
		"TEMPLATE_CACHE_TTL":     "CacheConfig.TTL",
		"TEMPLATE_LOG_LEVEL":     "Logging.Level",
	}

	for envVar, configPath := range standardOverrides {
		if val := os.Getenv(envVar); val != "" {
			if err := cm.setConfigValue(config, configPath, val); err != nil {
				return fmt.Errorf("failed to set %s from %s: %w", configPath, envVar, err)
			}
		}
	}

	return nil
}

// setConfigValue sets a configuration value using dot notation
func (cm *ConfigManager) setConfigValue(config *AdvancedConfig, path, value string) error {
	parts := strings.Split(path, ".")
	
	switch parts[0] {
	case "Config":
		if len(parts) >= 2 {
			switch parts[1] {
			case "BasePath":
				config.Config.BasePath = value
			case "CacheEnabled":
				enabled, err := strconv.ParseBool(value)
				if err != nil {
					return err
				}
				config.Config.CacheEnabled = enabled
			}
		}
	case "CacheConfig":
		if len(parts) >= 2 {
			switch parts[1] {
			case "TTL":
				duration, err := time.ParseDuration(value)
				if err != nil {
					return err
				}
				config.CacheConfig.TTL = duration
			case "Memory":
				if len(parts) >= 3 && parts[2] == "MaxSize" {
					size, err := strconv.Atoi(value)
					if err != nil {
						return err
					}
					config.CacheConfig.Memory.MaxSize = size
				}
			}
		}
	case "Logging":
		if len(parts) >= 2 && parts[1] == "Level" {
			config.Logging.Level = value
		}
	}

	return nil
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		rules: []ValidationRule{
			{
				Name:        "base_path_exists",
				Description: "Base path must exist and be readable",
				Required:    true,
				Validator: func(v interface{}) error {
					config, ok := v.(*AdvancedConfig)
					if !ok {
						return fmt.Errorf("invalid config type")
					}
					if config.Config.BasePath == "" {
						return fmt.Errorf("base path cannot be empty")
					}
					if _, err := os.Stat(config.Config.BasePath); os.IsNotExist(err) {
						return fmt.Errorf("base path does not exist: %s", config.Config.BasePath)
					}
					return nil
				},
			},
			{
				Name:        "cache_size_positive",
				Description: "Cache size must be positive",
				Required:    false,
				Validator: func(v interface{}) error {
					config, ok := v.(*AdvancedConfig)
					if !ok {
						return fmt.Errorf("invalid config type")
					}
					if config.CacheConfig.Memory.MaxSize <= 0 {
						return fmt.Errorf("cache max size must be positive, got %d", config.CacheConfig.Memory.MaxSize)
					}
					return nil
				},
			},
			{
				Name:        "log_level_valid",
				Description: "Log level must be valid",
				Required:    false,
				Validator: func(v interface{}) error {
					config, ok := v.(*AdvancedConfig)
					if !ok {
						return fmt.Errorf("invalid config type")
					}
					validLevels := []string{"debug", "info", "warn", "error"}
					level := strings.ToLower(config.Logging.Level)
					for _, valid := range validLevels {
						if level == valid {
							return nil
						}
					}
					return fmt.Errorf("invalid log level %s, must be one of: %v", config.Logging.Level, validLevels)
				},
			},
		},
	}
}

// Validate validates the configuration
func (cv *ConfigValidator) Validate(config *AdvancedConfig) error {
	var errors []string

	for _, rule := range cv.rules {
		if err := rule.Validator(config); err != nil {
			if rule.Required {
				return fmt.Errorf("validation failed for required rule %s: %w", rule.Name, err)
			}
			errors = append(errors, fmt.Sprintf("%s: %v", rule.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation warnings: %s", strings.Join(errors, "; "))
	}

	return nil
}

// SaveToFile saves configuration to a file
func (cm *ConfigManager) SaveToFile(config *AdvancedConfig, path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	var data []byte
	var err error

	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config to YAML: %w", err)
		}
	case ".json":
		data, err = json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config to JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// MergeConfigs merges multiple configurations with precedence (later configs override earlier ones)
func MergeConfigs(configs ...*AdvancedConfig) *AdvancedConfig {
	if len(configs) == 0 {
		return GetDefaultAdvancedConfig()
	}

	result := *configs[0]

	for i := 1; i < len(configs); i++ {
		config := configs[i]
		
		// Merge basic config
		if config.Config.BasePath != "" {
			result.Config.BasePath = config.Config.BasePath
		}
		result.Config.CacheEnabled = config.Config.CacheEnabled
		
		// Merge custom functions
		if config.Config.CustomFuncs != nil {
			if result.Config.CustomFuncs == nil {
				result.Config.CustomFuncs = make(template.FuncMap)
			}
			for name, fn := range config.Config.CustomFuncs {
				result.Config.CustomFuncs[name] = fn
			}
		}

		// Merge cache config
		if config.CacheConfig.Type != "" {
			result.CacheConfig.Type = config.CacheConfig.Type
		}
		if config.CacheConfig.TTL != 0 {
			result.CacheConfig.TTL = config.CacheConfig.TTL
		}
		
		// Merge other sections...
		// (Add more merge logic as needed)
	}

	return &result
}

// ValidateAndNormalize validates and normalizes configuration values
func ValidateAndNormalize(config *AdvancedConfig) error {
	// Normalize paths
	if config.Config.BasePath != "" {
		abs, err := filepath.Abs(config.Config.BasePath)
		if err == nil {
			config.Config.BasePath = abs
		}
	}

	// Normalize cache settings
	if config.CacheConfig.Type == "" {
		config.CacheConfig.Type = "memory"
	}

	// Set default TTL if not specified
	if config.CacheConfig.TTL == 0 {
		config.CacheConfig.TTL = 1 * time.Hour
	}

	// Normalize eviction policy
	policy := strings.ToUpper(config.CacheConfig.Memory.EvictionPolicy)
	if policy != "LRU" && policy != "LFU" && policy != "FIFO" {
		config.CacheConfig.Memory.EvictionPolicy = "LRU"
	}

	return nil
}