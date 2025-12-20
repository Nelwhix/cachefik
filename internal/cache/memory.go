package cache

import (
	"container/list"
	"sync"
)

type cacheItem struct {
	key   string
	entry Entry
}

type MemoryCache struct {
	mu       sync.Mutex
	capacity int
	list     *list.List
	items    map[string]*list.Element
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		capacity: 1000,
		list:     list.New(),
		items:    make(map[string]*list.Element),
	}
}

func (c *MemoryCache) Get(key string) (Entry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.items[key]; ok {
		item := element.Value.(*cacheItem)
		if item.entry.Expired() {
			c.list.Remove(element)
			delete(c.items, key)
			return Entry{}, false
		}

		c.list.MoveToFront(element)
		return item.entry, true
	}

	return Entry{}, false
}

func (c *MemoryCache) Set(key string, entry Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.items[key]; ok {
		c.list.MoveToFront(element)
		element.Value.(*cacheItem).entry = entry
		return
	}

	item := &cacheItem{key, entry}
	element := c.list.PushFront(item)
	c.items[key] = element

	if c.list.Len() > c.capacity {
		oldest := c.list.Back()
		if oldest != nil {
			c.list.Remove(oldest)
			delete(c.items, oldest.Value.(*cacheItem).key)
		}
	}
}
