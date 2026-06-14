package http

import (
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestSimpleServerWithListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := NewSimpleServerAddrWithOptions("127.0.0.1:0", WithListener(listener))
	srv.AddAction("/listener", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	errCh := srv.StartAsync()
	defer func() {
		_ = srv.Stop(time.Second)
		<-errCh
	}()
	resp, err := http.Get("http://" + listener.Addr().String() + "/listener") //nolint:gosec // test server URL.
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil || string(body) != "ok" {
		t.Fatalf("body = %q, err=%v", body, err)
	}
}
