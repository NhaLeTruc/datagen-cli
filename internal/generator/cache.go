package generator

import (
	"fmt"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

// FKCache is a thread-safe LRU cache for foreign key lookups
// It stores generated primary key values to be used as foreign keys
type FKCache struct {
	cache      *lru.Cache[string, interface{}]
	mu         sync.RWMutex
	stats      CacheStats
	capacity   int
}

// CacheStats contains statistics about cache operations
type CacheStats struct {
	Hits      int
	Misses    int
	Size      int
	Evictions int
}

const (
	// DefaultCacheSize is the default maximum number of entries in the cache
	DefaultCacheSize = 10000
)

// NewFKCache creates a new foreign key cache with the specified capacity
// If size <= 0, DefaultCacheSize is used
func NewFKCache(size int) *FKCache {
	if size <= 0 {
		size = DefaultCacheSize
	}

	// Create LRU cache with eviction callback
	cache, err := lru.NewWithEvict[string, interface{}](size, func(key string, value interface{}) {
		// This callback is called when an item is evicted
		// We'll track this in stats
	})
	if err != nil {
		// This should never happen with valid size, but handle it gracefully
		panic(fmt.Sprintf("failed to create LRU cache: %v", err))
	}

	fkCache := &FKCache{
		cache:    cache,
		capacity: size,
		stats:    CacheStats{},
	}

	// Wrap the cache with eviction tracking
	cache, _ = lru.NewWithEvict[string, interface{}](size, fkCache.onEvict)
	fkCache.cache = cache

	return fkCache
}

// onEvict is called when an item is evicted from the cache
func (c *FKCache) onEvict(key string, value interface{}) {
	c.stats.Evictions++
}

// makeKey creates a cache key from table name and row ID
func (c *FKCache) makeKey(tableName string, rowID int) string {
	return fmt.Sprintf("%s:%d", tableName, rowID)
}

// Put adds or updates a value in the cache
// tableName is the source table, rowID is the row number, value is the generated PK value
func (c *FKCache) Put(tableName string, rowID int, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.makeKey(tableName, rowID)
	c.cache.Add(key, value)
}

// Get retrieves a value from the cache
// Returns (value, true) if found, (nil, false) if not found
func (c *FKCache) Get(tableName string, rowID int) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.makeKey(tableName, rowID)
	value, ok := c.cache.Get(key)

	if ok {
		c.stats.Hits++
	} else {
		c.stats.Misses++
	}

	return value, ok
}

// Clear removes all entries from the cache
func (c *FKCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Purge()
	c.stats = CacheStats{}
}

// Stats returns the current cache statistics
func (c *FKCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = c.cache.Len()
	return stats
}

// Capacity returns the maximum number of entries the cache can hold
func (c *FKCache) Capacity() int {
	return c.capacity
}

// Contains checks if a key exists in the cache without updating LRU
func (c *FKCache) Contains(tableName string, rowID int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.makeKey(tableName, rowID)
	return c.cache.Contains(key)
}

// Remove removes a specific entry from the cache
func (c *FKCache) Remove(tableName string, rowID int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.makeKey(tableName, rowID)
	return c.cache.Remove(key)
}
