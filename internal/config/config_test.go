package config

import (
	"testing"
	"time"
)

func TestLoadTimeout(t *testing.T) {
	t.Setenv("THREADS_MCP_TIMEOUT", "30s")
	cfg, err := Load("1.2.3")
	if err != nil || cfg.ToolTimeout != 30*time.Second {
		t.Fatalf("cfg=%+v err=%v", cfg, err)
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("THREADS_MCP_TIMEOUT", "")
	cfg, err := Load("dev")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ToolTimeout != 45*time.Second || cfg.HTTPTimeout != 30*time.Second || cfg.Delay != time.Second || cfg.Retries != 2 || cfg.CacheTTL != time.Hour {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.CacheDir == "" {
		t.Fatal("cache directory is empty")
	}
}

func TestLoadRejectsTimeoutOutsideRange(t *testing.T) {
	for _, value := range []string{"4s", "121s", "not-a-duration"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("THREADS_MCP_TIMEOUT", value)
			if _, err := Load("dev"); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
