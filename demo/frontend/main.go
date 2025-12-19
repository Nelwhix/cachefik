package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=10")
		fmt.Fprintf(w, "frontend response at %s\n", time.Now())
	})

	http.ListenAndServe(":8080", nil)
}
