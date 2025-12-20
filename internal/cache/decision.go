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

	if strings.Contains(cc, "no-store") || strings.Contains(cc, "private") {
		return 0, false
	}

	for _, part := range strings.Split(cc, ",") {
		part = strings.TrimSpace(part)
		value, ok := strings.CutPrefix(part, "max-age=")
		if !ok {
			continue
		}

		secs, err := strconv.Atoi(value)
		if err != nil || secs <= 0 {
			return 0, false
		}

		return time.Duration(secs) * time.Second, true
	}

	return defaultTTL, true
}
