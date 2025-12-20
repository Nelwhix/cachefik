package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/Nelwhix/cachefik/internal/cache"
	"github.com/Nelwhix/cachefik/internal/config"
	"github.com/Nelwhix/cachefik/internal/provider/docker"
)

func main() {
	cfg := config.New()

	var level slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)
	slog.Info("Starting Cachefik", "addr", cfg.Addr, "log_level", cfg.LogLevel)

	services, err := docker.DiscoverServices(cfg.DockerHost, cfg.DockerVersion)
	if err != nil {
		slog.Error("Docker discovery failed", "error", err)
		os.Exit(1)
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

	log.Fatal(server.ListenAndServe())
}
