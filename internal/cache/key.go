package cache

import (
	"fmt"
	"net/http"
)

func Key(r *http.Request) string {
	queryString := r.URL.Query().Encode()

	return fmt.Sprintf(
		"%s:%s://%s%s?%s",
		r.Method,
		scheme(r),
		r.Host,
		r.URL.Path,
		queryString,
	)
}

func scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}
