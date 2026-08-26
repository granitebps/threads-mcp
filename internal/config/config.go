package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultToolTimeout = 45 * time.Second
	minimumToolTimeout = 5 * time.Second
	maximumToolTimeout = 120 * time.Second
)

type Config struct {
	ToolTimeout time.Duration
	HTTPTimeout time.Duration
	Delay       time.Duration
	Retries     int
	CacheDir    string
	CacheTTL    time.Duration
}

func Load(_ string) (Config, error) {
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return Config{}, fmt.Errorf("resolve cache directory: %w", err)
	}

	cfg := Config{
		ToolTimeout: defaultToolTimeout,
		HTTPTimeout: 30 * time.Second,
		Delay:       time.Second,
		Retries:     2,
		CacheDir:    filepath.Join(cacheRoot, "threads-mcp"),
		CacheTTL:    time.Hour,
	}

	if raw := os.Getenv("THREADS_MCP_TIMEOUT"); raw != "" {
		timeout, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("THREADS_MCP_TIMEOUT must be a duration between 5s and 120s")
		}
		if timeout < minimumToolTimeout || timeout > maximumToolTimeout {
			return Config{}, fmt.Errorf("THREADS_MCP_TIMEOUT must be between 5s and 120s")
		}
		cfg.ToolTimeout = timeout
	}

	return cfg, nil
}
