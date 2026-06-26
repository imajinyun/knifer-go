package vcron_test

import (
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vcron"
)

func TestFacadeConfig(t *testing.T) {
	cfg := vcron.NewConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	loc := time.FixedZone("facade-config", 8*3600)
	cfg = vcron.NewConfigWithOptions(vcron.WithConfigLocation(loc), vcron.WithConfigMatchSecond(true))
	if cfg.Location != loc || !cfg.MatchSecond {
		t.Fatalf("NewConfigWithOptions = %#v", cfg)
	}
}
