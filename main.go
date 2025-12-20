package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Nelwhix/cachefik/internal/cache"
	"github.com/Nelwhix/cachefik/internal/provider/docker"
)

func main() {
	services, err := docker.DiscoverServices()
	if err != nil {
		log.Fatalf("docker discovery failed: %v", err)
	}

	handler := &Proxy{
		Services: services,
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
