package config

import (
	"os"
	"time"
)

type Config struct {
	Addr          string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	ProxyTimeout  time.Duration
	DockerHost    string
	DockerVersion string
	LogLevel      string
}

func New() *Config {
	return &Config{
		Addr:          getEnv("CACHEFIK_ADDR", ":8000"),
		ReadTimeout:   getDurationEnv("CACHEFIK_READ_TIMEOUT", 5*time.Second),
		WriteTimeout:  getDurationEnv("CACHEFIK_WRITE_TIMEOUT", 10*time.Second),
		ProxyTimeout:  getDurationEnv("CACHEFIK_PROXY_TIMEOUT", 10*time.Second),
		DockerHost:    getEnv("CACHEFIK_DOCKER_HOST", ""),
		DockerVersion: getEnv("CACHEFIK_DOCKER_VERSION", ""),
		LogLevel:      getEnv("CACHEFIK_LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return d
}
