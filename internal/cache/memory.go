package cache

import "sync"

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]Entry
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]Entry),
	}
}

func (c *MemoryCache) Get(key string) (Entry, bool) {
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()

	if !ok || entry.Expired() {
		if ok {
			c.mu.Lock()
			delete(c.items, key)
			c.mu.Unlock()
		}

		return Entry{}, false
	}

	return entry, true
}

func (c *MemoryCache) Set(key string, entry Entry) {
	c.mu.Lock()
	c.items[key] = entry
	c.mu.Unlock()
}
