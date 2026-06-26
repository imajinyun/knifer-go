package vhttp_test

import (
	"net/http"
	"testing"

	"github.com/imajinyun/knifer-go/vhttp"
)

func TestFacadeServerStarterLifecycle(t *testing.T) {
	vhttp.ResetServerStarters()
	t.Cleanup(vhttp.ResetServerStarters)

	called := 0
	server := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0", vhttp.WithListenAndServeFunc(func(server *http.Server) error {
		called++
		return http.ErrServerClosed
	}))
	if err := server.Start(); err != http.ErrServerClosed {
		t.Fatalf("Start() = %v, want ErrServerClosed", err)
	}
	if called != 1 {
		t.Fatalf("custom starter called %d times, want 1", called)
	}
}
