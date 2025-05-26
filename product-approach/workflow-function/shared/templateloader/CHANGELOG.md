# Changelog

All notable changes to the Template Loader package will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Redis cache implementation
- Template hot-reloading support
- GraphQL-like template queries
- Template composition and inheritance
 - Performance profiling tools

## [1.0.1] - 2025-05-31

### Fixed
- Consolidated duplicate cache entry types to use `CacheItem` only.
- Removed unused `TemplateDiscovery` struct.
- Refactored `Loader` to utilize the `TemplateCache` interface with `MemoryCache` or `NoCache`.
- Logged errors from cache `Set` without failing template loads.

### Added
- Conceptual hook for file watching and hot-reloading.

## [1.0.0] - 2025-01-20

### Added
- Initial release of Template Loader package
- Core template loading and rendering functionality
- Support for versioned and flat file structures
- Semantic version management with automatic discovery
- In-memory cache with multiple eviction policies (LRU, LFU, FIFO, TTL)
- 20+ built-in template functions (math, string, array, utility)
- Comprehensive configuration system (YAML/JSON/Environment)
- Thread-safe concurrent access with RWMutex
- Detailed cache statistics and monitoring
- Custom template function registration
- Structured error handling with context
- Extensive test suite with benchmarks
- Complete API documentation

### Features
- **Template Loading**: Load templates by name or specific version
- **Smart Caching**: Configurable cache with automatic cleanup
- **Version Discovery**: Automatic semantic version sorting
- **Rich Functions**: Built-in functions for common template operations
- **Configuration**: Multiple configuration sources with validation
- **Error Handling**: Structured errors with detailed context
- **Monitoring**: Cache statistics and performance metrics
- **Thread Safety**: Safe for concurrent use

### Template Functions
- String: `split`, `join`, `upper`, `lower`, `trim`, `contains`
- Math: `add`, `sub`, `mul`, `div`, `gt`, `lt`, `eq`, `ne`, `ge`, `le`
- Array: `index`, `len`, `first`, `last`
- Utility: `ordinal`, `repeat`, `default`, `formatArray`

### Cache Policies
- **LRU** (Least Recently Used) - Default
- **LFU** (Least Frequently Used)
- **FIFO** (First In, First Out)
- **TTL** (Time To Live)

### Configuration Options
- Basic configuration via `Config` struct
- Advanced configuration via `AdvancedConfig`
- Environment variable overrides
- YAML and JSON configuration files
- Validation with custom rules

### Performance
- Efficient in-memory caching
- Concurrent template loading
- Background cleanup of expired items
- Optimized version comparison
- Minimal memory allocation

### Security
- Template size limits
- Sandbox mode support
- Restricted path access
- Input validation

### Documentation
- Comprehensive README with examples
- Complete API documentation
- Best practices guide
- Migration guide
- Performance benchmarks

## [0.9.0] - 2025-01-15

### Added
- Beta release for testing
- Core template loading functionality
- Basic caching implementation
- Version management system
- Initial documentation

### Fixed
- Template parsing edge cases
- Memory leaks in cache cleanup
- Race conditions in concurrent access

## [0.8.0] - 2025-01-10

### Added
- Alpha release
- Proof of concept implementation
- Basic template loading
- Simple in-memory cache

### Known Issues
- Limited error handling
- No version management
- Basic caching only
- Minimal documentation

---

### Legend

- **Added** - New features
- **Changed** - Changes in existing functionality
- **Deprecated** - Soon-to-be removed features
- **Removed** - Removed features
- **Fixed** - Bug fixes
- **Security** - Security improvements

### Version Numbering

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

### Contributing

When contributing, please:

1. Update this changelog with your changes
2. Follow the format established above
3. Place new entries under the "Unreleased" section
4. Move entries to a version section when releasing

### Migration Notes

#### From 0.x to 1.0.0

Breaking changes in 1.0.0:

1. **Configuration Structure**: 
   ```go
   // Old (0.x)
   config := Config{Path: "/templates"}
   
   // New (1.0.0)
   config := Config{BasePath: "/templates", CacheEnabled: true}
   ```

2. **Error Handling**:
   ```go
   // Old (0.x)
   template, err := loader.Load("name")
   
   // New (1.0.0)
   template, err := loader.LoadTemplate("name")
   ```

3. **Cache Interface**:
   ```go
   // Old (0.x)
   cache.Get(key) (Template, bool)
   
   // New (1.0.0)
   cache.Get(key) (*template.Template, *TemplateMetadata, bool)
   ```

### Support Policy

- **1.x.x**: Full support with bug fixes and security updates
- **0.x.x**: End of life - upgrade to 1.x.x recommended

### Dependencies

- Go 1.19+ required
- No external dependencies for core functionality
- Optional: `gopkg.in/yaml.v3` for YAML configuration

For more details on any version, see the corresponding [GitHub Release](https://github.com/kootoro/template-loader/releases).