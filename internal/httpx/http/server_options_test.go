package http

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestCreateServerHelper(t *testing.T) {
	srv := CreateServer(0)
	if srv == nil {
		t.Fatal("nil")
	}
}

func TestCreateServerWithOptions(t *testing.T) {
	called := false
	srv := CreateServerWithOptions(0, WithListenAndServeFunc(func(*http.Server) error {
		called = true
		return http.ErrServerClosed
	}))
	if err := srv.Start(); err != http.ErrServerClosed {
		t.Fatalf("Start() = %v, want ErrServerClosed", err)
	}
	if !called {
		t.Fatal("CreateServerWithOptions did not apply options")
	}
}

func TestSimpleServerOptionsAndStopWithContext(t *testing.T) {
	ctxValue := "server-base"
	srv := NewSimpleServerAddrWithOptions("127.0.0.1:0",
		WithReadHeaderTimeout(3*time.Second),
		WithReadTimeout(4*time.Second),
		WithWriteTimeout(5*time.Second),
		WithIdleTimeout(6*time.Second),
		WithBaseContext(func(net.Listener) context.Context {
			return context.WithValue(context.Background(), testServerContextKey{}, ctxValue)
		}),
		WithConnContext(func(ctx context.Context, _ net.Conn) context.Context {
			return context.WithValue(ctx, testConnContextKey{}, "conn")
		}),
	)
	if srv.server.ReadHeaderTimeout != 3*time.Second || srv.server.ReadTimeout != 4*time.Second ||
		srv.server.WriteTimeout != 5*time.Second || srv.server.IdleTimeout != 6*time.Second {
		t.Fatalf("server options not applied: %#v", srv.server)
	}
	srv.AddAction("/ctx", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Context().Value(testServerContextKey{}).(string) + ":" + r.Context().Value(testConnContextKey{}).(string)))
	})

	l, err := net.Listen("tcp", srv.server.Addr)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	errCh := make(chan error, 1)
	go func() {
		err := srv.server.Serve(l)
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	resp, err := http.Get("http://" + l.Addr().String() + "/ctx") //nolint:gosec // test server URL.
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil || string(body) != "server-base:conn" {
		t.Fatalf("body = %q, err=%v", body, err)
	}
	if err := srv.StopWithContext(context.Background()); err != nil {
		t.Fatalf("StopWithContext: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("serve err: %v", err)
	}
}
