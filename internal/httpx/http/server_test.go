package http

import (
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"
)

// Covers the action routing example from the utility toolkit-http server/SimpleServerTest.

func TestSimpleServerStartAndStop(t *testing.T) {
	port := pickFreePort(t)
	srv := NewSimpleServer(port)

	srv.AddAction("/get", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Path))
	})
	srv.AddAction("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	})
	srv.AddAction("/zero", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("0"))
	})

	errCh := srv.StartAsync()
	defer func() {
		_ = srv.Stop(2 * time.Second)
		select {
		case err := <-errCh:
			if err != nil {
				t.Logf("server err: %v", err)
			}
		case <-time.After(2 * time.Second):
		}
	}()

	waitServerReady(t, port)
	base := "http://127.0.0.1:" + strconv.Itoa(port)

	if got := Get(base + "/get").Execute().Body(); got != "/get" {
		t.Fatalf("get: %q", got)
	}

	body := Post(base + "/echo").BodyString(`{"a":1}`).Execute().Body()
	if body != `{"a":1}` {
		t.Fatalf("echo: %q", body)
	}

	if got := Get(base + "/zero").Execute().Body(); got != "0" {
		t.Fatalf("zero: %q", got)
	}
}

func TestCreateServerHelper(t *testing.T) {
	srv := CreateServer(0)
	if srv == nil {
		t.Fatal("nil")
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

type testServerContextKey struct{}

type testConnContextKey struct{}

// pickFreePort reserves a free port, releases it immediately, and returns the port number.
func pickFreePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}

// waitServerReady polls until the port is connectable.
func waitServerReady(t *testing.T, port int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(port), 100*time.Millisecond)
		if err == nil {
			_ = c.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("server on %d not ready", port)
}
