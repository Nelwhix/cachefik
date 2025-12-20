package cache

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKey(t *testing.T) {
	testCases := []struct {
		name     string
		request  *http.Request
		expected string
	}{
		{
			name: "HTTP GET",
			request: &http.Request{
				Method: http.MethodGet,
				Host:   "example.com",
				URL: &url.URL{
					Path:     "/test",
					RawQuery: "foo=bar",
				},
			},
			expected: "GET:http://example.com/test?foo=bar",
		},
		{
			name: "HTTPS GET",
			request: &http.Request{
				Method: http.MethodGet,
				Host:   "example.com",
				TLS:    &tls.ConnectionState{},
				URL: &url.URL{
					Path:     "/test",
					RawQuery: "foo=bar",
				},
			},
			expected: "GET:https://example.com/test?foo=bar",
		},
		{
			name: "HTTP POST",
			request: &http.Request{
				Method: http.MethodPost,
				Host:   "example.com",
				URL: &url.URL{
					Path: "/post",
				},
			},
			expected: "POST:http://example.com/post?",
		},
		{
			name: "Query Parameter Sorting",
			request: &http.Request{
				Method: http.MethodGet,
				Host:   "example.com",
				URL: &url.URL{
					Path:     "/search",
					RawQuery: "q=go&page=1&sort=desc",
				},
			},
			expected: "GET:http://example.com/search?page=1&q=go&sort=desc",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Key(tc.request)
			assert.Equal(t, tc.expected, got)
		})
	}
}
