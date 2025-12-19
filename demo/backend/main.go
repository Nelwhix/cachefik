package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprintf(w, "backend response at %s\n", time.Now())
	})

	http.ListenAndServe(":8080", nil)
}
