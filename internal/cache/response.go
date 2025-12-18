package cache

import "net/http"

func WriteCachedResponse(w http.ResponseWriter, entry Entry) {
	for k, vv := range entry.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.Header().Set("X-Cache", "HIT")
	w.WriteHeader(entry.StatusCode)
	w.Write(entry.Body)
}
