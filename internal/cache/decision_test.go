package cache

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCanCacheRequest(t *testing.T) {
	testCases := []struct {
		name     string
		method   string
		headers  http.Header
		expected bool
	}{
		{
			name:     "GET request",
			method:   http.MethodGet,
			expected: true,
		},
		{
			name:     "POST request",
			method:   http.MethodPost,
			expected: false,
		},
		{
			name:   "Authorization header",
			method: http.MethodGet,
			headers: http.Header{
				"Authorization": []string{"Bearer token"},
			},
			expected: false,
		},
		{
			name:   "Cache-Control no-store",
			method: http.MethodGet,
			headers: http.Header{
				"Cache-Control": []string{"no-store"},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := http.NewRequest(tc.method, "http://example.com", nil)
			for k, vv := range tc.headers {
				for _, v := range vv {
					r.Header.Add(k, v)
				}
			}
			got := CanCacheRequest(r)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestCanCacheResponse(t *testing.T) {
	testCases := []struct {
		name        string
		headers     http.Header
		expectedTTL time.Duration
		expectedOk  bool
	}{
		{
			name:        "Default TTL",
			headers:     http.Header{},
			expectedTTL: defaultTTL,
			expectedOk:  true,
		},
		{
			name: "max-age",
			headers: http.Header{
				"Cache-Control": []string{"max-age=60"},
			},
			expectedTTL: 60 * time.Second,
			expectedOk:  true,
		},
		{
			name: "no-store",
			headers: http.Header{
				"Cache-Control": []string{"no-store"},
			},
			expectedTTL: 0,
			expectedOk:  false,
		},
		{
			name: "private",
			headers: http.Header{
				"Cache-Control": []string{"private"},
			},
			expectedTTL: 0,
			expectedOk:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{
				Header: tc.headers,
			}
			gotTTL, gotOk := CanCacheResponse(resp)
			assert.Equal(t, tc.expectedTTL, gotTTL)
			assert.Equal(t, tc.expectedOk, gotOk)
		})
	}
}
