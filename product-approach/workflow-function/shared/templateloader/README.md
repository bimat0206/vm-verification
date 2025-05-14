# Template Loader

[![Go Reference](https://pkg.go.dev/badge/github.com/kootoro/template-loader.svg)](https://pkg.go.dev/github.com/kootoro/template-loader)
[![Go Report Card](https://goreportcard.com/badge/github.com/kootoro/template-loader)](https://goreportcard.com/report/github.com/kootoro/template-loader)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A simple, efficient, and configurable Go template loader package with built-in caching, version management, and comprehensive template functions.

## Features

- 🚀 **Simple API** - Easy to use interface for loading and rendering templates
- 📁 **Flexible Storage** - Support for versioned and flat file structures
- ⚡ **Smart Caching** - Configurable in-memory cache with multiple eviction policies (LRU, LFU, FIFO, TTL)
- 🔄 **Version Management** - Automatic discovery and semantic version sorting
- 🛠️ **Rich Functions** - 20+ built-in template functions for common operations
- ⚙️ **Configurable** - YAML/JSON configuration with environment variable overrides
- 🔒 **Thread-Safe** - Concurrent access with proper synchronization
- 📊 **Monitoring** - Comprehensive cache statistics and metrics
- 🧪 **Well-Tested** - Extensive test suite with benchmarks

## Quick Start

### Installation

```bash
go get github.com/kootoro/template-loader
```

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    templateloader "github.com/kootoro/template-loader"
)

func main() {
    // Create a new template loader
    config := templateloader.Config{
        BasePath:     "/path/to/templates",
        CacheEnabled: true,
    }
    
    loader, err := templateloader.New(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Render a template
    data := map[string]interface{}{
        "Name":    "World",
        "Count":   42,
        "Items":   []string{"apple", "banana", "cherry"},
    }
    
    result, err := loader.RenderTemplate("greeting", data)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result)
}
```

## Template Structure

The template loader supports both versioned and flat file structures:

### Versioned Structure (Recommended)
```
templates/
├── greeting/
│   ├── v1.0.0.tmpl
│   ├── v1.1.0.tmpl
│   └── v1.2.0.tmpl
└── notification/
    ├── v1.0.0.tmpl
    └── v2.0.0.tmpl
```

### Flat Structure
```
templates/
├── greeting.tmpl
├── notification.tmpl
└── report.tmpl
```

## Configuration

### Basic Configuration

```go
config := templateloader.Config{
    BasePath:     "/opt/templates",
    CacheEnabled: true,
    CustomFuncs: template.FuncMap{
        "customFunc": func(s string) string {
            return strings.ToUpper(s)
        },
    },
}
```

### Advanced Configuration with YAML

```yaml
# template-config.yaml
discovery:
  recursive: false
  include_patterns: ["*.tmpl", "*.template"]
  exclude_patterns: ["*.backup"]
  extensions: [".tmpl", ".template"]
  watch_enabled: false
  watch_interval: "30s"

cache:
  type: "memory"
  memory:
    max_size: 100
    eviction_policy: "LRU"  # LRU, LFU, FIFO, TTL
    cleanup_interval: "5m"
  ttl: "1h"
  default_ttl: "1h"

performance:
  max_concurrent_loads: 10
  max_memory_usage: "100MB"
  gc_interval: "10m"
  precompile_templates: true

logging:
  level: "info"
  format: "json"
  output: "stdout"

security:
  max_template_size: "1MB"
  sandbox_mode: false
```

Load advanced configuration:

```go
configManager := templateloader.NewConfigManager()
advancedConfig, err := configManager.LoadFromFile("template-config.yaml")
if err != nil {
    log.Fatal(err)
}

loader, err := templateloader.New(advancedConfig.Config)
if err != nil {
    log.Fatal(err)
}
```

## API Reference

### Core Methods

#### Creating a Loader

```go
// Create with basic config
loader, err := templateloader.New(templateloader.Config{
    BasePath:     "/opt/templates",
    CacheEnabled: true,
})

// Create with default config
loader, err := templateloader.New(*templateloader.GetDefaultConfig())
```

#### Loading Templates

```go
// Load latest version
tmpl, err := loader.LoadTemplate("template-name")

// Load specific version
tmpl, err := loader.LoadTemplateWithVersion("template-name", "1.2.0")
```

#### Rendering Templates

```go
// Render latest version
result, err := loader.RenderTemplate("template-name", data)

// Render specific version
result, err := loader.RenderTemplateWithVersion("template-name", "1.2.0", data)
```

#### Version Management

```go
// Get latest version
version := loader.GetLatestVersion("template-name")

// List all versions
versions := loader.ListVersions("template-name")
```

#### Cache Management

```go
// Clear cache
err := loader.ClearCache()

// Refresh version discovery
err := loader.RefreshVersions()
```

## Built-in Template Functions

The template loader includes 20+ built-in functions:

### String Functions
- `split` - Split string into array
- `join` - Join array into string
- `upper` - Convert to uppercase
- `lower` - Convert to lowercase
- `trim` - Trim whitespace

### Math Functions
- `add`, `sub`, `mul`, `div` - Basic arithmetic
- `gt`, `lt`, `eq`, `ne`, `ge`, `le` - Comparisons

### Array Functions
- `index` - Get element by index
- `len` - Get length
- `first` - Get first element
- `last` - Get last element

### Utility Functions
- `ordinal` - Convert number to ordinal (1st, 2nd, 3rd)
- `repeat` - Repeat string N times
- `default` - Provide default value
- `contains` - Check if string contains substring

### Template Example

```html
<!-- greeting.tmpl -->
Hello {{.Name}}!

You have {{.Count}} {{.Count | pluralize "item" "items"}}.

{{if .Items}}
Your items:
{{range $i, $item := .Items}}
  {{add $i 1 | ordinal}} item: {{$item | upper}}
{{end}}
{{end}}

{{if gt .Count 10}}
  You have many items!
{{else}}
  You have {{.Count}} item{{if ne .Count 1}}s{{end}}.
{{end}}
```

## Cache Statistics

Monitor cache performance with built-in statistics:

```go
stats := loader.(*templateloader.Loader).Cache.Stats()
fmt.Printf("Cache hit rate: %.2f%%\n", stats.HitRate*100)
fmt.Printf("Cache size: %d/%d\n", stats.Size, stats.MaxSize)
fmt.Printf("Total hits: %d, misses: %d\n", stats.Hits, stats.Misses)
```

## Environment Variables

Configure the loader using environment variables:

```bash
export TEMPLATE_BASE_PATH="/opt/templates"
export TEMPLATE_CACHE_ENABLED="true"
export TEMPLATE_CACHE_SIZE="100"
export TEMPLATE_CACHE_TTL="1h"
export TEMPLATE_LOG_LEVEL="info"
```

Load from environment:

```go
configManager := templateloader.NewConfigManager()
config, err := configManager.LoadFromEnvironment("TEMPLATE")
if err != nil {
    log.Fatal(err)
}

loader, err := templateloader.New(*config)
```

## Best Practices

### 1. Use Semantic Versioning
```
v1.0.0.tmpl  # Major.Minor.Patch
v1.1.0.tmpl
v2.0.0.tmpl
```

### 2. Enable Caching in Production
```go
config := templateloader.Config{
    BasePath:     "/opt/templates",
    CacheEnabled: true,  // Always enable in production
}
```

### 3. Set Appropriate TTL
```yaml
cache:
  ttl: "1h"  # Adjust based on template update frequency
```

### 4. Monitor Cache Performance
```go
// Log cache stats periodically
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        stats := cache.Stats()
        log.Printf("Cache stats: hit_rate=%.2f%% size=%d",
            stats.HitRate*100, stats.Size)
    }
}()
```

### 5. Handle Errors Gracefully
```go
result, err := loader.RenderTemplate("template", data)
if err != nil {
    // Log error and use fallback template or default content
    log.Printf("Template render error: %v", err)
    result = "Default content"
}
```

### 6. Use Type-Safe Template Data
```go
type TemplateData struct {
    Name    string
    Count   int
    Items   []string
    Active  bool
}

data := TemplateData{
    Name:   "John",
    Count:  5,
    Items:  []string{"a", "b", "c"},
    Active: true,
}

result, err := loader.RenderTemplate("template", data)
```

## Migration from Other Template Systems

### From text/template
```go
// Old way
tmpl, err := template.ParseFiles("template.tmpl")
if err != nil {
    return err
}
result, err := tmpl.Execute(buffer, data)

// New way
loader, _ := templateloader.New(config)
result, err := loader.RenderTemplate("template", data)
```

### From html/template
Template Loader works with text/template. For HTML templates, ensure proper escaping in your template files:

```html
<!-- Use html/template syntax in your .tmpl files -->
<div>{{.Content | html}}</div>
<script>var data = {{.JSData | js}};</script>
```

## Performance

### Benchmarks

```
BenchmarkLoader_LoadTemplate-8     	    5000	    250 ns/op
BenchmarkLoader_RenderTemplate-8   	    2000	    500 ns/op
BenchmarkCache_Get-8               	 1000000	      2 ns/op
BenchmarkCache_Set-8               	  500000	      5 ns/op
```

### Performance Tips

1. **Use Caching**: Always enable caching in production
2. **Preload Templates**: Load frequently used templates at startup
3. **Monitor Memory**: Set appropriate cache size limits
4. **Optimize Templates**: Keep templates simple and avoid complex logic

## Error Handling

The loader provides structured error information:

```go
result, err := loader.RenderTemplate("template", data)
if err != nil {
    if terr, ok := err.(*templateloader.TemplateError); ok {
        log.Printf("Template error: type=%s, template=%s:%s, message=%s",
            terr.Type, terr.TemplateType, terr.Version, terr.Message)
    } else {
        log.Printf("General error: %v", err)
    }
}
```

Error types:
- `ErrorTypeNotFound` - Template not found
- `ErrorTypeParsing` - Template parsing failed
- `ErrorTypeExecution` - Template execution failed
- `ErrorTypeValidation` - Template validation failed
- `ErrorTypeConfiguration` - Configuration error

## Contributing

We welcome contributions! Please follow these guidelines:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Write tests** for your changes
4. **Run tests**: `go test ./...`
5. **Check formatting**: `go fmt ./...`
6. **Submit a pull request**

### Development Setup

```bash
# Clone the repository
git clone https://github.com/kootoro/template-loader.git
cd template-loader

# Install dependencies
go mod download

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./...

# Check coverage
go test -cover ./...
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/kootoro/template-loader)
- **Issues**: [GitHub Issues](https://github.com/kootoro/template-loader/issues)
- **Discussions**: [GitHub Discussions](https://github.com/kootoro/template-loader/discussions)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed list of changes and version history.

---

Made with ❤️ by the Kootoro Team