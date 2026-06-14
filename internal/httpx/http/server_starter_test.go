package http

import (
	"net/http"
	"testing"
)

func TestSimpleServerStarterLifecycle(t *testing.T) {
	ResetServerStarters()
	t.Cleanup(ResetServerStarters)

	called := 0
	srv := NewSimpleServerAddrWithOptions("127.0.0.1:0", WithListenAndServeFunc(func(server *http.Server) error {
		called++
		if server.Addr != "127.0.0.1:0" {
			t.Fatalf("server addr = %q, want 127.0.0.1:0", server.Addr)
		}
		return http.ErrServerClosed
	}))
	if err := srv.Start(); err != http.ErrServerClosed {
		t.Fatalf("Start() = %v, want ErrServerClosed", err)
	}
	if called != 1 {
		t.Fatalf("custom starter called %d times, want 1", called)
	}

	called = 0
	ResetServerStarters()
	srv = NewSimpleServerAddrWithOptions("127.0.0.1:0", WithListenAndServeFunc(func(server *http.Server) error {
		called++
		return nil
	}), func(server *http.Server) {
		ResetServerStarters()
	})
	if called != 0 {
		t.Fatalf("custom starter leaked after reset: called %d times", called)
	}
	if srv.listenAndServe == nil {
		t.Fatal("server starter should fall back to default function")
	}
}
