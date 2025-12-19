package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Nelwhix/cachefik/internal/cache"
)

func main() {
	handler := &Proxy{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Cache: cache.NewMemoryCache(),
	}

	server := &http.Server{
		Addr:         ":8000",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Starting server on port 8080")
	log.Fatal(server.ListenAndServe())
}
