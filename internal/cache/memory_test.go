package cache

import (
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

		c.mu.RLock()
		_, ok = c.items["expired_key"]
		c.mu.RUnlock()
		assert.False(t, ok)
	})
}
