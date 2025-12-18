package cache

import (
	"fmt"
	"net/http"
)

func Key(r *http.Request) string {
	return fmt.Sprintf(
		"%s:%s://%s%s?%s",
		r.Method,
		scheme(r),
		r.Host,
		r.URL.Path,
		r.URL.RawQuery,
	)
}

func scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}
