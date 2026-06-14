package vhttp_test

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vhttp"
)

func TestFacadeSimpleServerOptions(t *testing.T) {
	server := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0",
		vhttp.WithReadHeaderTimeout(time.Second),
		vhttp.WithReadTimeout(time.Second),
		vhttp.WithWriteTimeout(time.Second),
		vhttp.WithIdleTimeout(time.Second),
		vhttp.WithHTTPServer(&http.Server{Addr: "127.0.0.1:0"}),
	)
	if server == nil {
		t.Fatal("NewSimpleServerAddrWithOptions returned nil")
	}
	if err := server.StopWithContext(context.Background()); err != nil {
		t.Fatalf("StopWithContext on idle server = %v", err)
	}
}

func TestFacadeAdditionalServerOptions(t *testing.T) {
	baseContextCalled := false
	connContextCalled := false
	runnerCalled := false
	logger := log.New(io.Discard, "", 0)
	simple := vhttp.NewSimpleServerWithOptions(0,
		vhttp.WithServerErrorLog(logger),
		vhttp.WithBaseContext(func(net.Listener) context.Context {
			baseContextCalled = true
			return context.Background()
		}),
		vhttp.WithConnContext(func(ctx context.Context, conn net.Conn) context.Context {
			connContextCalled = conn != nil
			return ctx
		}),
		vhttp.WithAsyncRunner(func(run func()) {
			runnerCalled = true
			run()
		}),
		vhttp.WithListenAndServeFunc(func(server *http.Server) error {
			if server.BaseContext != nil {
				_ = server.BaseContext(nil)
			}
			if server.ConnContext != nil {
				_ = server.ConnContext(context.Background(), nil)
			}
			return http.ErrServerClosed
		}),
	)
	errCh := simple.StartAsync()
	if err, ok := <-errCh; ok || err != nil {
		t.Fatalf("StartAsync channel = (%v, %v), want closed", err, ok)
	}
	if !runnerCalled || !baseContextCalled {
		t.Fatalf("server option calls runner=%v base=%v conn=%v", runnerCalled, baseContextCalled, connContextCalled)
	}
	if created := vhttp.CreateServer(0); created == nil {
		t.Fatal("CreateServer returned nil")
	}
	if created := vhttp.CreateServerWithOptions(0, vhttp.WithHTTPServer(&http.Server{Addr: ":0"})); created == nil {
		t.Fatal("CreateServerWithOptions returned nil")
	}
}
