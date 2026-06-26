package vskt_test

import (
	"net"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vskt"
)

func TestFacadeSocketConfig(t *testing.T) {
	cfg := vskt.NewSocketConfig()
	if cfg == nil {
		t.Fatal("expected non-nil socket config")
	}
}

func TestFacadeSocketConfigWithOptions(t *testing.T) {
	listener := &facadeFakeListener{addr: facadeFakeAddr("listener")}
	client, server := net.Pipe()
	defer func() { _ = server.Close() }()
	cfg := vskt.NewSocketConfigWithOptions(
		vskt.WithThreadPoolSize(2),
		vskt.WithReadTimeout(100),
		vskt.WithWriteTimeout(200),
		vskt.WithReadBufferSize(64),
		vskt.WithWriteBufferSize(128),
		vskt.WithListenerFactory(func(*net.TCPAddr) (net.Listener, error) { return listener, nil }),
		vskt.WithConnFactory(func(*net.TCPAddr) (net.Conn, error) { return client, nil }),
	)
	if cfg.ThreadPoolSize != 2 || cfg.ReadTimeout != 100 || cfg.WriteTimeout != 200 ||
		cfg.ReadBufferSize != 64 || cfg.WriteBufferSize != 128 {
		t.Fatalf("NewSocketConfigWithOptions not applied: %+v", cfg)
	}
	if cfg.ListenerFactory == nil || cfg.ConnFactory == nil {
		t.Fatal("expected listener and connection factories")
	}
}

func TestFacadeSocketConfigThreadPoolSizeFunc(t *testing.T) {
	calls := 0
	cfg := vskt.NewSocketConfigWithOptions(vskt.WithThreadPoolSizeFunc(func() int {
		calls++
		return 9
	}))
	if calls != 1 || cfg.ThreadPoolSize != 9 {
		t.Fatalf("WithThreadPoolSizeFunc calls=%d size=%d, want 1/9", calls, cfg.ThreadPoolSize)
	}
}

func TestFacadeSocketConfigProviderOptions(t *testing.T) {
	now := time.Unix(123, 0)
	runnerCalled := false
	parsedIP := net.IPv4(10, 0, 0, 8)
	cfg := vskt.NewSocketConfigWithOptions(
		vskt.WithClock(func() time.Time { return now }),
		vskt.WithRunner(func(fn func()) {
			runnerCalled = true
			fn()
		}),
		vskt.WithSocketIPParser(func(host string) net.IP {
			if host != "example.test" {
				t.Fatalf("host = %q, want example.test", host)
			}
			return parsedIP
		}),
	)
	if got := cfg.Clock(); !got.Equal(now) {
		t.Fatalf("Clock() = %v, want %v", got, now)
	}
	cfg.Runner(func() {})
	if !runnerCalled {
		t.Fatal("WithRunner should install the runner provider")
	}
	if got := cfg.IPParser("example.test"); !got.Equal(parsedIP) {
		t.Fatalf("IPParser() = %v, want %v", got, parsedIP)
	}
}
