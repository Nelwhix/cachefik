package cache

import (
	"net/http"
	"time"
)

type Entry struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	ExpiresAt  time.Time
}

func (e Entry) Expired() bool {
	return time.Now().After(e.ExpiresAt)
}

type Cache interface {
	Get(key string) (Entry, bool)
	Set(key string, entry Entry)
}
