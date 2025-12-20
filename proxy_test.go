package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Nelwhix/cachefik/internal/cache"
	"github.com/Nelwhix/cachefik/internal/provider/docker"
	"github.com/stretchr/testify/assert"
)

func TestProxyIntegration(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("X-Backend", "true")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("backend response"))
	}))
	defer backend.Close()

	proxy := &Proxy{
		Services: []docker.Service{
			{Rule: "PathPrefix(`/`)", Upstream: backend.URL},
		},
		Client:       &http.Client{},
		Cache:        cache.NewMemoryCache(),
		MaxCacheSize: 1024 * 1024,
	}

	t.Run("Cache MISS then HIT", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		// First request - MISS
		proxy.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))
		assert.Equal(t, "backend response", w.Body.String())

		// Second request - HIT
		w = httptest.NewRecorder()
		proxy.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "HIT", w.Header().Get("X-Cache"))
		assert.Equal(t, "backend response", w.Body.String())
	})

	t.Run("Cache BYPASS with Authorization", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/bypass", nil)
		req.Header.Set("Authorization", "Bearer token")
		w := httptest.NewRecorder()

		proxy.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "BYPASS", w.Header().Get("X-Cache"))
	})

	t.Run("Upstream Error", func(t *testing.T) {
		// Using a closed server to trigger a connection error
		p := &Proxy{
			Services: []docker.Service{
				{Rule: "PathPrefix(`/`)", Upstream: "http://localhost:1"},
			},
			Client: &http.Client{
				Timeout: 100 * time.Millisecond,
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		w := httptest.NewRecorder()

		p.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), `"error":"upstream error"`)
	})

	t.Run("No Upstream Found", func(t *testing.T) {
		p := &Proxy{
			Services: []docker.Service{},
			Client:   &http.Client{},
		}
		req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
		w := httptest.NewRecorder()

		p.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), `"error":"no upstream found"`)
		assert.Contains(t, w.Body.String(), `"code":404`)
	})

	t.Run("TTL Expiration", func(t *testing.T) {
		// Mock backend with short max-age
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "max-age=1")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("expiring content"))
		}))
		defer backend.Close()

		p := &Proxy{
			Services: []docker.Service{
				{Rule: "PathPrefix(`/`)", Upstream: backend.URL},
			},
			Client:       &http.Client{},
			Cache:        cache.NewMemoryCache(),
			MaxCacheSize: 1024 * 1024,
		}

		req := httptest.NewRequest(http.MethodGet, "/expire", nil)

		// First request - MISS
		w := httptest.NewRecorder()
		p.ServeHTTP(w, req)
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))

		// Second request - HIT
		w = httptest.NewRecorder()
		p.ServeHTTP(w, req)
		assert.Equal(t, "HIT", w.Header().Get("X-Cache"))

		// Wait for expiration
		time.Sleep(1500 * time.Millisecond)

		// Third request - MISS
		w = httptest.NewRecorder()
		p.ServeHTTP(w, req)
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))
	})

	t.Run("Header Forwarding", func(t *testing.T) {
		var capturedHeader http.Header
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedHeader = r.Header
			w.WriteHeader(http.StatusOK)
		}))
		defer backend.Close()

		p := &Proxy{
			Services: []docker.Service{
				{Rule: "PathPrefix(`/`)", Upstream: backend.URL},
			},
			Client:       &http.Client{},
			Cache:        cache.NewMemoryCache(),
			MaxCacheSize: 1024 * 1024,
		}

		req := httptest.NewRequest(http.MethodGet, "/headers", nil)
		req.Header.Set("X-Custom", "value")
		req.RemoteAddr = "1.2.3.4:1234"

		p.ServeHTTP(httptest.NewRecorder(), req)

		assert.Equal(t, "value", capturedHeader.Get("X-Custom"))
		assert.Equal(t, "1.2.3.4", capturedHeader.Get("X-Forwarded-For"))
		assert.Equal(t, "http", capturedHeader.Get("X-Forwarded-Proto"))
	})

	t.Run("Large Response (No Cache)", func(t *testing.T) {
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(make([]byte, 2000)) // 2000 bytes
		}))
		defer backend.Close()

		p := &Proxy{
			Services: []docker.Service{
				{Rule: "PathPrefix(`/`)", Upstream: backend.URL},
			},
			Client:       &http.Client{},
			Cache:        cache.NewMemoryCache(),
			MaxCacheSize: 1000, // Limit 1000 bytes
		}

		req := httptest.NewRequest(http.MethodGet, "/large", nil)
		w := httptest.NewRecorder()

		// First request - MISS (too large to cache)
		p.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))
		assert.Equal(t, 2000, len(w.Body.Bytes()))

		// Second request - MISS (should not be in cache)
		w = httptest.NewRecorder()
		p.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "MISS", w.Header().Get("X-Cache"))
	})
}
