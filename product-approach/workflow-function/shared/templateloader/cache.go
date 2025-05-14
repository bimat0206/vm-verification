package templateloader

import (
	"container/list"
	"fmt"
	"sync"
	"text/template"
	"time"
	"strings"
)

// TemplateCache defines the interface for template caching
type TemplateCache interface {
	Get(key string) (*template.Template, *TemplateMetadata, bool)
	Set(key string, tmpl *template.Template, metadata *TemplateMetadata, ttl time.Duration) error
	Delete(key string) error
	Clear() error
	Size() int
	Stats() CacheStats
	Keys() []string
	HasExpired(key string) bool
	Cleanup() int // Returns number of items removed
}

// CacheStats provides statistics about cache usage
type CacheStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	Sets        int64     `json:"sets"`
	Deletes     int64     `json:"deletes"`
	Evictions   int64     `json:"evictions"`
	Size        int       `json:"size"`
	MaxSize     int       `json:"max_size"`
	HitRate     float64   `json:"hit_rate"`
	LastAccess  time.Time `json:"last_access"`
	LastSet     time.Time `json:"last_set"`
	LastCleanup time.Time `json:"last_cleanup"`
}

// CacheItem represents a cached template with metadata
type CacheItem struct {
	Key         string             `json:"key"`
	Template    *template.Template `json:"-"`
	Metadata    *TemplateMetadata  `json:"metadata"`
	CreatedAt   time.Time          `json:"created_at"`
	LastAccess  time.Time          `json:"last_access"`
	AccessCount int64              `json:"access_count"`
	TTL         time.Duration      `json:"ttl"`
	ExpiresAt   time.Time          `json:"expires_at"`
	Size        int64              `json:"size"`
}

// EvictionPolicy defines cache eviction policies
type EvictionPolicy int

const (
	EvictionPolicyLRU EvictionPolicy = iota
	EvictionPolicyLFU
	EvictionPolicyFIFO
	EvictionPolicyTTL
)

// String returns string representation of EvictionPolicy
func (ep EvictionPolicy) String() string {
	switch ep {
	case EvictionPolicyLRU:
		return "LRU"
	case EvictionPolicyLFU:
		return "LFU"
	case EvictionPolicyFIFO:
		return "FIFO"
	case EvictionPolicyTTL:
		return "TTL"
	default:
		return "UNKNOWN"
	}
}

// MemoryCache implements an in-memory template cache with configurable eviction policies
type MemoryCache struct {
	items          map[string]*CacheItem
	accessList     *list.List          // For LRU
	itemElements   map[string]*list.Element // Maps keys to list elements
	maxSize        int
	evictionPolicy EvictionPolicy
	defaultTTL     time.Duration
	stats          CacheStats
	mu             sync.RWMutex
	cleanupTicker  *time.Ticker
	stopCleanup    chan struct{}
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(maxSize int, evictionPolicy string, defaultTTL time.Duration) (*MemoryCache, error) {
	if maxSize <= 0 {
		return nil, fmt.Errorf("max size must be positive, got %d", maxSize)
	}

	policy := parseEvictionPolicy(evictionPolicy)

	cache := &MemoryCache{
		items:          make(map[string]*CacheItem),
		accessList:     list.New(),
		itemElements:   make(map[string]*list.Element),
		maxSize:        maxSize,
		evictionPolicy: policy,
		defaultTTL:     defaultTTL,
		stats: CacheStats{
			MaxSize: maxSize,
		},
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine if TTL is used
	if defaultTTL > 0 {
		cache.startCleanupRoutine(defaultTTL / 4) // Cleanup every quarter of TTL
	}

	return cache, nil
}

// parseEvictionPolicy converts string to EvictionPolicy
func parseEvictionPolicy(policy string) EvictionPolicy {
	switch policy {
	case "LRU":
		return EvictionPolicyLRU
	case "LFU":
		return EvictionPolicyLFU
	case "FIFO":
		return EvictionPolicyFIFO
	case "TTL":
		return EvictionPolicyTTL
	default:
		return EvictionPolicyLRU
	}
}

// Get retrieves a template from the cache
func (mc *MemoryCache) Get(key string) (*template.Template, *TemplateMetadata, bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	item, exists := mc.items[key]
	if !exists {
		mc.stats.Misses++
		return nil, nil, false
	}

	// Check if item has expired
	if mc.isExpired(item) {
		delete(mc.items, key)
		if element, exists := mc.itemElements[key]; exists {
			mc.accessList.Remove(element)
			delete(mc.itemElements, key)
		}
		mc.stats.Misses++
		mc.stats.Size--
		return nil, nil, false
	}

	// Update access information
	item.LastAccess = time.Now()
	item.AccessCount++
	mc.stats.Hits++
	mc.stats.LastAccess = time.Now()

	// Update LRU list
	if mc.evictionPolicy == EvictionPolicyLRU {
		if element, exists := mc.itemElements[key]; exists {
			mc.accessList.MoveToFront(element)
		}
	}

	return item.Template, item.Metadata, true
}

// Set stores a template in the cache
func (mc *MemoryCache) Set(key string, tmpl *template.Template, metadata *TemplateMetadata, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Use default TTL if not specified
	if ttl == 0 {
		ttl = mc.defaultTTL
	}

	// Create cache item
	item := &CacheItem{
		Key:         key,
		Template:    tmpl,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
		LastAccess:  time.Now(),
		AccessCount: 0,
		TTL:         ttl,
		ExpiresAt:   time.Now().Add(ttl),
		Size:        mc.estimateSize(tmpl, metadata),
	}

	// Check if item already exists
	if existingItem, exists := mc.items[key]; exists {
		// Update existing item
		item.AccessCount = existingItem.AccessCount
		mc.items[key] = item
		
		// Update LRU list element
		if element, exists := mc.itemElements[key]; exists {
			element.Value = item
			if mc.evictionPolicy == EvictionPolicyLRU {
				mc.accessList.MoveToFront(element)
			}
		}
	} else {
		// Add new item
		mc.items[key] = item
		mc.stats.Size++

		// Add to LRU list
		if mc.evictionPolicy == EvictionPolicyLRU || mc.evictionPolicy == EvictionPolicyFIFO {
			element := mc.accessList.PushFront(item)
			mc.itemElements[key] = element
		}

		// Evict if necessary
		if mc.stats.Size > mc.maxSize {
			mc.evict()
		}
	}

	mc.stats.Sets++
	mc.stats.LastSet = time.Now()

	return nil
}

// Delete removes a template from the cache
func (mc *MemoryCache) Delete(key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.items[key]; exists {
		delete(mc.items, key)
		mc.stats.Size--
		mc.stats.Deletes++

		// Remove from LRU list
		if element, exists := mc.itemElements[key]; exists {
			mc.accessList.Remove(element)
			delete(mc.itemElements, key)
		}
	}

	return nil
}

// Clear removes all templates from the cache
func (mc *MemoryCache) Clear() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items = make(map[string]*CacheItem)
	mc.accessList = list.New()
	mc.itemElements = make(map[string]*list.Element)
	mc.stats.Size = 0

	return nil
}

// Size returns the current number of items in the cache
func (mc *MemoryCache) Size() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.stats.Size
}

// Stats returns cache statistics
func (mc *MemoryCache) Stats() CacheStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	stats := mc.stats
	
	// Calculate hit rate
	totalRequests := stats.Hits + stats.Misses
	if totalRequests > 0 {
		stats.HitRate = float64(stats.Hits) / float64(totalRequests)
	}

	return stats
}

// Keys returns all cache keys
func (mc *MemoryCache) Keys() []string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	keys := make([]string, 0, len(mc.items))
	for key := range mc.items {
		keys = append(keys, key)
	}
	return keys
}

// HasExpired checks if a cache item has expired
func (mc *MemoryCache) HasExpired(key string) bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	item, exists := mc.items[key]
	if !exists {
		return true
	}
	
	return mc.isExpired(item)
}

// Cleanup removes expired items from the cache
func (mc *MemoryCache) Cleanup() int {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	removed := 0
	now := time.Now()

	for key, item := range mc.items {
		if mc.isExpired(item) {
			delete(mc.items, key)
			removed++
			mc.stats.Size--

			// Remove from LRU list
			if element, exists := mc.itemElements[key]; exists {
				mc.accessList.Remove(element)
				delete(mc.itemElements, key)
			}
		}
	}

	mc.stats.LastCleanup = now
	return removed
}

// Close stops the cache and cleanup routines
func (mc *MemoryCache) Close() error {
	if mc.cleanupTicker != nil {
		mc.cleanupTicker.Stop()
	}
	
	close(mc.stopCleanup)
	return nil
}

// Private methods

// isExpired checks if an item has expired (must be called with lock held)
func (mc *MemoryCache) isExpired(item *CacheItem) bool {
	if item.TTL == 0 {
		return false
	}
	return time.Now().After(item.ExpiresAt)
}

// evict removes items based on the eviction policy
func (mc *MemoryCache) evict() {
	if mc.stats.Size <= mc.maxSize {
		return
	}

	var keyToEvict string

	switch mc.evictionPolicy {
	case EvictionPolicyLRU:
		// Remove least recently used (back of list)
		if element := mc.accessList.Back(); element != nil {
			item := element.Value.(*CacheItem)
			keyToEvict = item.Key
		}

	case EvictionPolicyLFU:
		// Find least frequently used
		var minAccess int64 = -1
		for key, item := range mc.items {
			if minAccess == -1 || item.AccessCount < minAccess {
				minAccess = item.AccessCount
				keyToEvict = key
			}
		}

	case EvictionPolicyFIFO:
		// Remove first in (back of list)
		if element := mc.accessList.Back(); element != nil {
			item := element.Value.(*CacheItem)
			keyToEvict = item.Key
		}

	case EvictionPolicyTTL:
		// Remove item with closest expiration
		var earliestExpiry time.Time
		for key, item := range mc.items {
			if item.TTL > 0 && (earliestExpiry.IsZero() || item.ExpiresAt.Before(earliestExpiry)) {
				earliestExpiry = item.ExpiresAt
				keyToEvict = key
			}
		}
	}

	// Remove the selected item
	if keyToEvict != "" {
		delete(mc.items, keyToEvict)
		mc.stats.Size--
		mc.stats.Evictions++

		// Remove from LRU list
		if element, exists := mc.itemElements[keyToEvict]; exists {
			mc.accessList.Remove(element)
			delete(mc.itemElements, keyToEvict)
		}
	}
}

// estimateSize estimates the memory size of a template and metadata
func (mc *MemoryCache) estimateSize(tmpl *template.Template, metadata *TemplateMetadata) int64 {
	// This is a rough estimate
	// In practice, you might want a more accurate calculation
	size := int64(200) // Base size for template struct

	if metadata != nil {
		size += int64(len(metadata.Type))
		size += int64(len(metadata.Version))
		size += int64(len(metadata.Path))
		// Add more fields as needed
	}

	return size
}

// startCleanupRoutine starts a background routine to clean expired items
func (mc *MemoryCache) startCleanupRoutine(interval time.Duration) {
	mc.cleanupTicker = time.NewTicker(interval)
	
	go func() {
		for {
			select {
			case <-mc.cleanupTicker.C:
				mc.Cleanup()
			case <-mc.stopCleanup:
				return
			}
		}
	}()
}

// NoCache is a cache implementation that doesn't cache anything
type NoCache struct{}

// NewNoCache creates a new no-op cache
func NewNoCache() *NoCache {
	return &NoCache{}
}

// Get always returns cache miss
func (nc *NoCache) Get(key string) (*template.Template, *TemplateMetadata, bool) {
	return nil, nil, false
}

// Set does nothing
func (nc *NoCache) Set(key string, tmpl *template.Template, metadata *TemplateMetadata, ttl time.Duration) error {
	return nil
}

// Delete does nothing
func (nc *NoCache) Delete(key string) error {
	return nil
}

// Clear does nothing
func (nc *NoCache) Clear() error {
	return nil
}

// Size always returns 0
func (nc *NoCache) Size() int {
	return 0
}

// Stats returns empty stats
func (nc *NoCache) Stats() CacheStats {
	return CacheStats{}
}

// Keys returns empty slice
func (nc *NoCache) Keys() []string {
	return []string{}
}

// HasExpired always returns true
func (nc *NoCache) HasExpired(key string) bool {
	return true
}

// Cleanup does nothing
func (nc *NoCache) Cleanup() int {
	return 0
}

// CacheFactory creates cache instances based on configuration
type CacheFactory struct{}

// NewCacheFactory creates a new cache factory
func NewCacheFactory() *CacheFactory {
	return &CacheFactory{}
}

// CreateCache creates a cache based on the configuration
func (cf *CacheFactory) CreateCache(config CacheConfig) (TemplateCache, error) {
	switch strings.ToLower(config.Type) {
	case "memory":
		return NewMemoryCache(
			config.Memory.MaxSize,
			config.Memory.EvictionPolicy,
			config.DefaultTTL,
		)
	case "none", "disabled":
		return NewNoCache(), nil
	default:
		return nil, fmt.Errorf("unsupported cache type: %s", config.Type)
	}
}

