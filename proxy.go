package main

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Nelwhix/cachefik/internal/cache"
	"github.com/Nelwhix/cachefik/internal/provider/docker"
)

type Proxy struct {
	Services     []docker.Service
	Client       *http.Client
	Cache        cache.Cache
	MaxCacheSize int64
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)

	if p.Cache != nil && cache.CanCacheRequest(r) {
		key := cache.Key(r)

		if entry, ok := p.Cache.Get(key); ok {
			cache.WriteCachedResponse(w, entry)
			return
		}
	}

	target := p.pickUpstream(r)
	if target == "" {
		logger.Warn("no upstream found")
		sendJSONError(w, "no upstream found", http.StatusNotFound)
		return
	}

	logger = logger.With("upstream", target)

	upstreamURL, _ := url.Parse(target)
	outRequest := p.cloneRequest(r, upstreamURL)
	resp, err := p.Client.Do(outRequest)
	if err != nil {
		logger.Error("upstream request failed", "error", err)
		sendJSONError(w, "upstream error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	ttl, ok := cache.CanCacheResponse(resp)
	canCache := ok && p.Cache != nil && cache.CanCacheRequest(r)

	var buf bytes.Buffer
	var bodyWriter = io.Discard
	var lw *limitedWriter
	if canCache {
		lw = &limitedWriter{
			W:     &buf,
			Limit: p.MaxCacheSize,
		}
		bodyWriter = lw
	}

	copyHeaders(w.Header(), resp.Header)
	removeHopByHopHeaders(w.Header())

	if p.Cache != nil {
		if canCache {
			w.Header().Set("X-Cache", "MISS")
		} else {
			w.Header().Set("X-Cache", "BYPASS")
		}
	}

	w.WriteHeader(resp.StatusCode)

	tee := io.TeeReader(resp.Body, bodyWriter)
	_, err = io.Copy(w, tee)
	if err != nil {
		// Too late to send an error to the client as headers/status are already sent
		logger.Error("streaming failed", "error", err)
		return
	}

	if canCache && !lw.Exceeded {
		p.Cache.Set(cache.Key(r), cache.Entry{
			StatusCode: resp.StatusCode,
			Header:     resp.Header,
			Body:       buf.Bytes(),
			ExpiresAt:  time.Now().Add(ttl),
		})
	}
}

func (p *Proxy) cloneRequest(r *http.Request, upstream *url.URL) *http.Request {
	outRequest := r.Clone(context.Background())
	outRequest.URL.Scheme = upstream.Scheme
	outRequest.URL.Host = upstream.Host
	outRequest.URL.Path = singleJoiningSlash(upstream.Path, r.URL.Path)
	outRequest.RequestURI = ""

	outRequest.Host = upstream.Host

	copyHeaders(outRequest.Header, r.Header)
	removeHopByHopHeaders(outRequest.Header)
	addForwardedHeaders(outRequest)

	return outRequest
}

func (p *Proxy) pickUpstream(r *http.Request) string {
	for _, svc := range p.Services {
		if strings.HasPrefix(r.URL.Path, svc.PathPrefix()) {
			return svc.Upstream
		}
	}

	return ""
}

type limitedWriter struct {
	W        io.Writer
	Limit    int64
	Written  int64
	Exceeded bool
}

func (lw *limitedWriter) Write(p []byte) (n int, err error) {
	if lw.Exceeded {
		return len(p), nil
	}

	remaining := lw.Limit - lw.Written
	if int64(len(p)) > remaining {
		lw.Exceeded = true
		n, _ = lw.W.Write(p[:remaining])
		lw.Written += int64(n)
		return len(p), nil
	}

	n, err = lw.W.Write(p)
	lw.Written += int64(n)
	return n, err
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")

	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	default:
		return a + b
	}
}
