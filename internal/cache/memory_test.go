package cache

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCache(t *testing.T) {
	c := NewMemoryCache()

	t.Run("Set and Get", func(t *testing.T) {
		entry := Entry{
			StatusCode: http.StatusOK,
			Body:       []byte("test"),
			ExpiresAt:  time.Now().Add(1 * time.Hour),
		}
		c.Set("key1", entry)

		got, ok := c.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, entry, got)
	})

	t.Run("Get missing", func(t *testing.T) {
		_, ok := c.Get("missing")
		assert.False(t, ok)
	})

	t.Run("Expiration", func(t *testing.T) {
		entry := Entry{
			StatusCode: http.StatusOK,
			Body:       []byte("expired"),
			ExpiresAt:  time.Now().Add(-1 * time.Hour),
		}
		c.Set("expired_key", entry)

		_, ok := c.Get("expired_key")
		assert.False(t, ok)

		c.mu.Lock()
		_, ok = c.items["expired_key"]
		c.mu.Unlock()
		assert.False(t, ok)
	})

	t.Run("LRU Eviction", func(t *testing.T) {
		c := NewMemoryCache()
		// Fill it up to capacity (1000)
		for i := 0; i < 1000; i++ {
			c.Set(fmt.Sprintf("key%d", i), Entry{ExpiresAt: time.Now().Add(1 * time.Hour)})
		}

		// Access key0, so it's now the most recently used
		c.Get("key0")

		// Add one more, which should trigger eviction of key1 (the next oldest)
		c.Set("new", Entry{ExpiresAt: time.Now().Add(1 * time.Hour)})

		// key0 should still be there
		_, ok := c.Get("key0")
		assert.True(t, ok)

		// key1 should be gone
		_, ok = c.Get("key1")
		assert.False(t, ok)

		// newest one should be there
		_, ok = c.Get("new")
		assert.True(t, ok)
	})
}
