package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// CacheManager manages concurrent access to cached data
type CacheManager struct {
	mu    sync.RWMutex
	cache map[string]*CacheItem
	stats CacheStats
}

// CacheItem represents a single cached item
type CacheItem struct {
	Value      interface{}
	Expiration int64 // Unix timestamp in nanoseconds
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits    int64 `json:"hits"`
	Misses  int64 `json:"misses"`
	Sets    int64 `json:"sets"`
	Updates int64 `json:"updates"`
	Deletes int64 `json:"deletes"`
	Expired int64 `json:"expired"`
}

// NewCacheManager creates a new cache manager
// func NewCacheManager(cleanupInterval time.Duration) *CacheManager {
func NewCacheManager(cleanupInterval time.Duration) *CacheManager {
	cm := &CacheManager{
		cache: make(map[string]*CacheItem),
	}
	return cm
}

// getInternal retrieves a value without counting statistics
func (cm *CacheManager) getInternal(key string) (interface{}, bool) {
	cm.mu.RLock()
	item, found := cm.cache[key]
	cm.mu.RUnlock()
	if !found {
		return nil, false
	}
	// Check expiration
	now := time.Now().UnixNano()
	if item.Expiration > 0 && item.Expiration <= now {
		return nil, false
	}
	return item.Value, true
}

// setInternal helper for setting with expiration
func (cm *CacheManager) setInternal(key string, value interface{}, ttl time.Duration) {
	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}
	cm.cache[key] = &CacheItem{
		Value:      value,
		Expiration: expiration,
	}
	atomic.AddInt64(&cm.stats.Sets, 1)
}

// GetOrSet retrieves a value or sets it if not found
func (cm *CacheManager) GetOrSet(key string, setFunc func() (interface{}, time.Duration)) interface{} {

	// Try to get existing value
	if val, found := cm.getInternal(key); found {
		atomic.AddInt64(&cm.stats.Hits, 1)
		return val
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Double-check after acquiring lock, somebody could refresh while Lock waiting.
	if item, found := cm.cache[key]; found {
		if item.Expiration == 0 || item.Expiration > time.Now().UnixNano() {
			atomic.AddInt64(&cm.stats.Hits, 1)
			return item.Value
		}
		// Item is expired - clean it up
		//delete(cm.cache, key)
		atomic.AddInt64(&cm.stats.Expired, 1)
	}

	// Call set function to get value
	value, ttl := setFunc()
	cm.setInternal(key, value, ttl)
	atomic.AddInt64(&cm.stats.Misses, 1) // This was a cache miss
	return value
}
