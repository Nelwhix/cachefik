package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Nelwhix/cachefik/internal/cache"
)

func main() {
	upstream := os.Getenv("UPSTREAM_URL")
	if upstream == "" {
		upstream = "http://backend:8080"
	}

	handler := &Proxy{
		Upstream: upstream,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Cache: cache.NewMemoryCache(),
	}

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Starting server on port 8080")
	log.Fatal(server.ListenAndServe())
}
