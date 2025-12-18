package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteCachedResponse(t *testing.T) {
	entry := Entry{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"text/plain"},
		},
		Body: []byte("cached content"),
	}

	w := httptest.NewRecorder()
	WriteCachedResponse(w, entry)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Equal(t, "HIT", w.Header().Get("X-Cache"))
	assert.Equal(t, "cached content", w.Body.String())
}
