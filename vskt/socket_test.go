package vskt_test

import (
	"net"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vskt"
)

func TestFacadeSocketConfig(t *testing.T) {
	cfg := vskt.NewSocketConfig()
	if cfg == nil {
		t.Fatal("expected non-nil socket config")
	}
}

func TestFacadeSocketConfigWithOptions(t *testing.T) {
	cfg := vskt.NewSocketConfigWithOptions(
		vskt.WithThreadPoolSize(2),
		vskt.WithReadTimeout(100),
		vskt.WithWriteTimeout(200),
		vskt.WithReadBufferSize(64),
		vskt.WithWriteBufferSize(128),
	)
	if cfg.ThreadPoolSize != 2 || cfg.ReadTimeout != 100 || cfg.WriteTimeout != 200 ||
		cfg.ReadBufferSize != 64 || cfg.WriteBufferSize != 128 {
		t.Fatalf("NewSocketConfigWithOptions not applied: %+v", cfg)
	}
}

func TestFacadeSocketConnectWithOptions(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()

	go func() {
		conn, _ := ln.Accept()
		if conn != nil {
			_ = conn.Close()
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	conn, err := vskt.SocketConnectWithOptions("127.0.0.1", addr.Port, vskt.WithConnectTimeout(time.Second), vskt.WithConnectNetwork("tcp"))
	if err != nil {
		t.Fatalf("SocketConnectWithOptions failed: %v", err)
	}
	defer func() { _ = conn.Close() }()
	if !vskt.SocketIsConnected(conn) {
		t.Fatal("SocketConnectWithOptions should return a connected socket")
	}
}

func TestFacadeSocketIsConnected(t *testing.T) {
	// nil conn should not be connected
	if vskt.SocketIsConnected(nil) {
		t.Fatal("expected nil conn to be disconnected")
	}
}

func TestFacadeSocketError(t *testing.T) {
	err := vskt.NewSocketErrorMsg("test error")
	if err == nil {
		t.Fatal("expected non-nil socket error")
	}
	if err.Error() != "test error" {
		t.Fatalf("expected 'test error', got %q", err.Error())
	}
}

func TestFacadeOperations(t *testing.T) {
	// verify operation constants are accessible
	_ = vskt.OpRead
	_ = vskt.OpWrite
	_ = vskt.OpConnect
	_ = vskt.OpAccept
}
