package main

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type Proxy struct {
	Upstream string
	Client   *http.Client
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	copyHeaders(w.Header(), resp.Header)
	removeHopByHopHeaders(w.Header())

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
