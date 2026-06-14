package cron

import (
	"testing"
	"time"
)

func TestConfigOptions(t *testing.T) {
	loc := time.FixedZone("config", 9*3600)
	cfg := NewConfigWithOptions(WithConfigLocation(loc), WithConfigMatchSecond(true))
	if cfg.Location != loc || !cfg.MatchSecond {
		t.Fatalf("config options not applied: %#v", cfg)
	}
	cfg = NewConfigWithOptions(WithConfigLocation(nil))
	if cfg.Location == nil {
		t.Fatal("nil config location should fall back to local")
	}
}
