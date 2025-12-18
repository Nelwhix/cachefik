package main

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Nelwhix/cachefik/internal/cache"
)

type Proxy struct {
	Upstream string
	Client   *http.Client
	Cache    cache.Cache
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.Cache != nil && cache.CanCacheRequest(r) {
		key := cache.Key(r)

		if entry, ok := p.Cache.Get(key); ok {
			cache.WriteCachedResponse(w, entry)
			return
		}
	}

	upstreamURL, err := url.Parse(p.Upstream)
	if err != nil {
		http.Error(w, "bad upstream", http.StatusInternalServerError)
		return
	}

	outRequest := p.cloneRequest(r, upstreamURL)
	resp, err := p.Client.Do(outRequest)
	if err != nil {
		http.Error(w, "upstream error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadGateway)
		return
	}

	copyHeaders(w.Header(), resp.Header)
	removeHopByHopHeaders(w.Header())

	if p.Cache != nil {
		if ttl, ok := cache.CanCacheResponse(resp); ok && cache.CanCacheRequest(r) {
			p.Cache.Set(cache.Key(r), cache.Entry{
				StatusCode: resp.StatusCode,
				Header:     resp.Header,
				Body:       body,
				ExpiresAt:  time.Now().Add(ttl),
			})
			w.Header().Set("X-Cache", "MISS")
		} else {
			w.Header().Set("X-Cache", "BYPASS")
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
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
