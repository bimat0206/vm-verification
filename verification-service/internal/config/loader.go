package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigLoader manages loading and refreshing configuration
type ConfigLoader struct {
	currentConfig *Config
	configPath    string
	listeners     []ConfigChangeListener
}

// ConfigChangeListener is a function that gets called when config changes
type ConfigChangeListener func(oldConfig, newConfig *Config)

// NewConfigLoader creates a new config loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		listeners: make([]ConfigChangeListener, 0),
	}
}

// LoadConfig loads configuration from environment variables
func (l *ConfigLoader) LoadConfig() (*Config, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	l.currentConfig = cfg
	return cfg, nil
}

// LoadConfigFromFile loads configuration from environment variables and optionally overrides with values from a JSON file
func (l *ConfigLoader) LoadConfigFromFile(filePath string) (*Config, error) {
	cfg, err := LoadConfigFromFile(filePath)
	if err != nil {
		return nil, err
	}

	// Set the config path for future refreshes
	absPath, err := filepath.Abs(filePath)
	if err == nil {
		l.configPath = absPath
	} else {
		l.configPath = filePath
	}

	oldConfig := l.currentConfig
	l.currentConfig = cfg

	// Notify listeners if this is a refresh
	if oldConfig != nil {
		l.notifyListeners(oldConfig, cfg)
	}

	return cfg, nil
}

// GetCurrentConfig returns the current configuration
func (l *ConfigLoader) GetCurrentConfig() *Config {
	return l.currentConfig
}

// RefreshConfig reloads configuration from environment variables and the config file if set
func (l *ConfigLoader) RefreshConfig() (*Config, error) {
	if l.configPath == "" {
		return l.LoadConfig()
	}
	return l.LoadConfigFromFile(l.configPath)
}

// AddChangeListener adds a listener that gets called when config changes
func (l *ConfigLoader) AddChangeListener(listener ConfigChangeListener) {
	l.listeners = append(l.listeners, listener)
}

// notifyListeners notifies all registered listeners of a config change
func (l *ConfigLoader) notifyListeners(oldConfig, newConfig *Config) {
	for _, listener := range l.listeners {
		listener(oldConfig, newConfig)
	}
}

// InitDefaultConfig initializes a default configuration file if it doesn't exist
func (l *ConfigLoader) InitDefaultConfig(filePath string) error {
	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		// File already exists
		return fmt.Errorf("config file already exists: %s", filePath)
	}

	// Load default config from environment
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Save to file
	if err := SaveConfigToFile(cfg, filePath); err != nil {
		return err
	}

	// Set the config path
	absPath, err := filepath.Abs(filePath)
	if err == nil {
		l.configPath = absPath
	} else {
		l.configPath = filePath
	}

	l.currentConfig = cfg
	return nil
}