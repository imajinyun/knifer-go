package http

import (
	"io"
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

func TestSimpleServerStartAsyncUsesRunner(t *testing.T) {
	started := make(chan struct{})
	runnerCalled := false
	srv := NewSimpleServerAddrWithOptions("127.0.0.1:0",
		WithAsyncRunner(func(fn func()) {
			runnerCalled = true
			go fn()
		}),
		WithListenAndServeFunc(func(server *http.Server) error {
			close(started)
			return http.ErrServerClosed
		}),
	)
	errCh := srv.StartAsync()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("custom starter was not launched")
	}
	if !runnerCalled {
		t.Fatal("custom async runner was not used")
	}
	if err, ok := <-errCh; ok || err != nil {
		t.Fatalf("StartAsync err channel = (%v, %v), want closed", err, ok)
	}
}
