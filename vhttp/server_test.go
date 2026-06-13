package vhttp_test

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
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

func TestFacadeClientAndAdditionalServerOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method + ":" + r.Header.Get("X-Client-Default")))
	}))
	defer server.Close()

	client := vhttp.NewClient(vhttp.WithClientRequestOptions(vhttp.WithHeader("X-Client-Default", "shared")))
	if got := client.Get(server.URL).Execute().Body(); got != "GET:shared" {
		t.Fatalf("client.Get body = %q", got)
	}
	if got := client.Post(server.URL).Execute().Body(); got != "POST:shared" {
		t.Fatalf("client.Post body = %q", got)
	}
	if got := client.NewRequest(vhttp.MethodPut, server.URL).Execute().Body(); got != "PUT:shared" {
		t.Fatalf("client.NewRequest body = %q", got)
	}

	cfg := vhttp.SnapshotGlobalConfig()
	cfg.Headers.Set("X-Client-Default", "configured")
	if got := vhttp.NewClientWithConfig(cfg).Get(server.URL).Execute().Body(); got != "GET:configured" {
		t.Fatalf("NewClientWithConfig body = %q", got)
	}
	isolated := vhttp.NewIsolatedClient(vhttp.WithClientGlobalConfig(cfg))
	if got := isolated.Get(server.URL).Execute().Body(); got != "GET:configured" {
		t.Fatalf("NewIsolatedClient body = %q", got)
	}
	if resp := client.GetSafe(server.URL).Execute(); resp.Err() == nil {
		t.Fatal("client.GetSafe(localhost default policy) error = nil")
	}
	if resp := client.PostSafe(server.URL, vhttp.WithURLPolicy(vhttp.URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})).Execute(); resp.Err() != nil {
		t.Fatalf("client.PostSafe allowed error = %v", resp.Err())
	}

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

func TestFacadeErrorCodesAndSimpleServerWrappers(t *testing.T) {
	cause := errors.New("bad request")
	err := vhttp.NewErrorWithCode(knifer.ErrCodeInvalidInput, "invalid request", cause)
	if !errors.Is(err, cause) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("NewErrorWithCode does not unwrap cause or code: %v", err)
	}
	if got := vhttp.ErrorfWithCode(knifer.ErrCodeInvalidInput, "status %d", http.StatusBadRequest).Error(); got != "status 400" {
		t.Fatalf("ErrorfWithCode = %q", got)
	}
	if vhttp.NewSimpleServer(0) == nil {
		t.Fatal("NewSimpleServer returned nil")
	}
	if vhttp.NewSimpleServerAddr("127.0.0.1:0") == nil {
		t.Fatal("NewSimpleServerAddr returned nil")
	}

	static := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0")
	static.SetRootWithOptions(".",
		vhttp.WithStaticFileSystem(http.Dir(".")),
		vhttp.WithStaticFS(os.DirFS(".")),
		vhttp.WithFileServerFactory(func(http.FileSystem) http.Handler {
			return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			})
		}),
		vhttp.WithStaticHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		})),
	)

	listener, listenErr := net.Listen("tcp", "127.0.0.1:0")
	if listenErr != nil {
		t.Fatal(listenErr)
	}
	defer listener.Close()
	if server := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0",
		vhttp.WithListener(listener),
		vhttp.WithListenAndServeFunc(func(*http.Server) error {
			return http.ErrServerClosed
		}),
	); server == nil {
		t.Fatal("WithListener server returned nil")
	}
}
