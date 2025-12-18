package cache

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

const defaultTTL = 30 * time.Second

func CanCacheRequest(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}

	if r.Header.Get("Authorization") != "" {
		return false
	}

	cc := r.Header.Get("Cache-Control")
	if strings.Contains(cc, "no-store") {
		return false
	}

	return true
}

func CanCacheResponse(resp *http.Response) (time.Duration, bool) {
	cc := resp.Header.Get("Cache-Control")
	if strings.Contains(cc, "no-store") {
		return 0, false
	}

	if strings.Contains(cc, "private") {
		return 0, false
	}

	if strings.Contains(cc, "max-age") {
		parts := strings.Split(cc, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(p, "max-age=") {
				secs, err := strconv.Atoi(strings.TrimPrefix(p, "max-age="))
				if err == nil && secs > 0 {
					return time.Duration(secs) * time.Second, true
				}
			}
		}
	}

	return defaultTTL, true
}
