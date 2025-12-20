package main

import (
	"log"
	"net/http"

	"github.com/Nelwhix/cachefik/internal/cache"
	"github.com/Nelwhix/cachefik/internal/config"
	"github.com/Nelwhix/cachefik/internal/provider/docker"
)

func main() {
	cfg := config.New()
	services, err := docker.DiscoverServices(cfg.DockerHost, cfg.DockerVersion)
	if err != nil {
		log.Fatalf("docker discovery failed: %v", err)
	}

	handler := &Proxy{
		Services: services,
		Client: &http.Client{
			Timeout: cfg.ProxyTimeout,
		},
		Cache: cache.NewMemoryCache(),
	}

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	log.Printf("Starting server on %s", cfg.Addr)
	log.Fatal(server.ListenAndServe())
}
