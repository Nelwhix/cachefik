package main

import (
	"net"
	"net/http"
	"strings"
)

var hopByHopHeaders = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

func removeHopByHopHeaders(headers http.Header) {
	for _, hopHeader := range hopByHopHeaders {
		headers.Del(hopHeader)
	}

	if connection := headers.Get("Connection"); connection != "" {
		for _, f := range strings.Split(connection, ",") {
			headers.Del(strings.TrimSpace(f))
		}
	}
}

func copyHeaders(destination, source http.Header) {
	for key, value := range source {
		for _, v := range value {
			destination.Add(key, v)
		}
	}
}

func addForwardedHeaders(r *http.Request) {
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		r.Header.Add("X-Forwarded-For", ip)
	}

	r.Header.Set("X-Forwarded-Proto", scheme(r))
	r.Header.Set("X-Forwarded-Host", r.Host)
}

func scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}

	return "http"
}
