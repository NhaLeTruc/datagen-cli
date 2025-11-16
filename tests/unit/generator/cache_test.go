package generator_test

import (
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFKCache(t *testing.T) {
	tests := []struct {
		name string
		size int
		want int
	}{
		{
			name: "create cache with valid size",
			size: 100,
			want: 100,
		},
		{
			name: "create cache with zero size uses default",
			size: 0,
			want: 10000, // Default size
		},
		{
			name: "create cache with negative size uses default",
			size: -10,
			want: 10000, // Default size
		},
		{
			name: "create cache with large size",
			size: 1000000,
			want: 1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := generator.NewFKCache(tt.size)
			require.NotNil(t, cache)
			assert.Equal(t, tt.want, cache.Capacity())
		})
	}
}

func TestFKCachePutAndGet(t *testing.T) {
	cache := generator.NewFKCache(10)

	t.Run("get from empty cache returns false", func(t *testing.T) {
		val, ok := cache.Get("users", 1)
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("put and get returns same value", func(t *testing.T) {
		cache.Put("users", 1, int64(100))
		val, ok := cache.Get("users", 1)
		assert.True(t, ok)
		assert.Equal(t, int64(100), val)
	})

	t.Run("put multiple values and get them", func(t *testing.T) {
		cache.Put("posts", 1, "post-1-id")
		cache.Put("posts", 2, "post-2-id")
		cache.Put("comments", 1, map[string]interface{}{"id": 1, "text": "comment"})

		val1, ok1 := cache.Get("posts", 1)
		assert.True(t, ok1)
		assert.Equal(t, "post-1-id", val1)

		val2, ok2 := cache.Get("posts", 2)
		assert.True(t, ok2)
		assert.Equal(t, "post-2-id", val2)

		val3, ok3 := cache.Get("comments", 1)
		assert.True(t, ok3)
		assert.Equal(t, map[string]interface{}{"id": 1, "text": "comment"}, val3)
	})

	t.Run("overwrite existing value", func(t *testing.T) {
		cache.Put("users", 1, int64(100))
		val1, ok1 := cache.Get("users", 1)
		assert.True(t, ok1)
		assert.Equal(t, int64(100), val1)

		cache.Put("users", 1, int64(200))
		val2, ok2 := cache.Get("users", 1)
		assert.True(t, ok2)
		assert.Equal(t, int64(200), val2)
	})
}

func TestFKCacheEviction(t *testing.T) {
	t.Run("LRU eviction when cache is full", func(t *testing.T) {
		cache := generator.NewFKCache(3)

		// Fill cache to capacity
		cache.Put("users", 1, 100)
		cache.Put("users", 2, 200)
		cache.Put("users", 3, 300)

		// All values should be present
		_, ok1 := cache.Get("users", 1)
		assert.True(t, ok1)
		_, ok2 := cache.Get("users", 2)
		assert.True(t, ok2)
		_, ok3 := cache.Get("users", 3)
		assert.True(t, ok3)

		// Add one more item, should evict LRU (users:1)
		cache.Put("users", 4, 400)

		// users:1 should be evicted
		_, ok := cache.Get("users", 1)
		assert.False(t, ok)

		// Others should still be present
		_, ok2 = cache.Get("users", 2)
		assert.True(t, ok2)
		_, ok3 = cache.Get("users", 3)
		assert.True(t, ok3)
		_, ok4 := cache.Get("users", 4)
		assert.True(t, ok4)
	})

	t.Run("Get refreshes item to prevent eviction", func(t *testing.T) {
		cache := generator.NewFKCache(3)

		cache.Put("users", 1, 100)
		cache.Put("users", 2, 200)
		cache.Put("users", 3, 300)

		// Access users:1 to make it recently used
		cache.Get("users", 1)

		// Add new item, should evict users:2 (oldest unreferenced)
		cache.Put("users", 4, 400)

		// users:1 should still be present (was accessed)
		_, ok1 := cache.Get("users", 1)
		assert.True(t, ok1)

		// users:2 should be evicted
		_, ok2 := cache.Get("users", 2)
		assert.False(t, ok2)
	})
}

func TestFKCacheMultipleTables(t *testing.T) {
	cache := generator.NewFKCache(100)

	t.Run("cache handles multiple tables independently", func(t *testing.T) {
		// Add data for different tables with same row IDs
		cache.Put("users", 1, "user-1")
		cache.Put("posts", 1, "post-1")
		cache.Put("comments", 1, "comment-1")

		// Each table should have its own entry
		val1, ok1 := cache.Get("users", 1)
		assert.True(t, ok1)
		assert.Equal(t, "user-1", val1)

		val2, ok2 := cache.Get("posts", 1)
		assert.True(t, ok2)
		assert.Equal(t, "post-1", val2)

		val3, ok3 := cache.Get("comments", 1)
		assert.True(t, ok3)
		assert.Equal(t, "comment-1", val3)
	})
}

func TestFKCacheClear(t *testing.T) {
	cache := generator.NewFKCache(10)

	t.Run("clear removes all entries", func(t *testing.T) {
		cache.Put("users", 1, 100)
		cache.Put("posts", 1, 200)
		cache.Put("comments", 1, 300)

		// Verify entries exist
		_, ok := cache.Get("users", 1)
		assert.True(t, ok)

		// Clear cache
		cache.Clear()

		// All entries should be gone
		_, ok1 := cache.Get("users", 1)
		assert.False(t, ok1)
		_, ok2 := cache.Get("posts", 1)
		assert.False(t, ok2)
		_, ok3 := cache.Get("comments", 1)
		assert.False(t, ok3)
	})

	t.Run("can add entries after clear", func(t *testing.T) {
		cache.Clear()
		cache.Put("users", 1, 100)

		val, ok := cache.Get("users", 1)
		assert.True(t, ok)
		assert.Equal(t, 100, val)
	})
}

func TestFKCacheStats(t *testing.T) {
	cache := generator.NewFKCache(10)

	t.Run("reports accurate statistics", func(t *testing.T) {
		// Add some entries
		cache.Put("users", 1, 100)
		cache.Put("users", 2, 200)
		cache.Put("posts", 1, 300)

		// Perform some gets (hits and misses)
		cache.Get("users", 1)    // hit
		cache.Get("users", 2)    // hit
		cache.Get("users", 999)  // miss
		cache.Get("posts", 1)    // hit
		cache.Get("posts", 999)  // miss

		stats := cache.Stats()
		assert.Equal(t, 3, stats.Hits)
		assert.Equal(t, 2, stats.Misses)
		assert.Equal(t, 3, stats.Size)
		assert.Equal(t, 0, stats.Evictions)
	})

	t.Run("tracks evictions", func(t *testing.T) {
		smallCache := generator.NewFKCache(2)

		smallCache.Put("users", 1, 100)
		smallCache.Put("users", 2, 200)
		smallCache.Put("users", 3, 300) // causes eviction

		stats := smallCache.Stats()
		assert.Equal(t, 2, stats.Size)
		assert.Equal(t, 1, stats.Evictions)
	})
}

func TestFKCacheConcurrency(t *testing.T) {
	t.Run("cache is safe for concurrent access", func(t *testing.T) {
		cache := generator.NewFKCache(100)
		done := make(chan bool)

		// Multiple goroutines writing
		for i := 0; i < 10; i++ {
			go func(id int) {
				for j := 0; j < 100; j++ {
					cache.Put("users", j, id*1000+j)
				}
				done <- true
			}(i)
		}

		// Multiple goroutines reading
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					cache.Get("users", j)
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 20; i++ {
			<-done
		}

		// Verify cache is still functional
		stats := cache.Stats()
		assert.Greater(t, stats.Hits+stats.Misses, 0)
	})
}

func TestFKCacheKeyFormat(t *testing.T) {
	cache := generator.NewFKCache(10)

	t.Run("different key formats are treated as different entries", func(t *testing.T) {
		// Test that table name and row ID are both part of the key
		cache.Put("users", 1, "value-1")
		cache.Put("users_backup", 1, "value-2")

		val1, ok1 := cache.Get("users", 1)
		assert.True(t, ok1)
		assert.Equal(t, "value-1", val1)

		val2, ok2 := cache.Get("users_backup", 1)
		assert.True(t, ok2)
		assert.Equal(t, "value-2", val2)
	})
}
