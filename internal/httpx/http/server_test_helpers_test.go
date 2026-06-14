package http

import (
	"net"
	"strconv"
	"testing"
	"time"
)

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
