package vskt_test

import (
	"net"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vskt"
)

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

func TestFacadeSocketConnectOptionVariants(t *testing.T) {
	addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
	dialer := &facadeFakeDialer{}
	conn, err := vskt.SocketConnectAddrWithOptions(addr, vskt.WithConnectDialer(dialer), vskt.WithConnectNetwork("tcp4"))
	if err != nil {
		t.Fatalf("SocketConnectAddrWithOptions failed: %v", err)
	}
	_ = conn.Close()
	_ = dialer.server.Close()
	if dialer.calls.Load() != 1 || dialer.network != "tcp4" || dialer.address != "127.0.0.1:1234" {
		t.Fatalf("dialer call = (%d, %q, %q)", dialer.calls.Load(), dialer.network, dialer.address)
	}

	dialer = &facadeFakeDialer{}
	conn, err = vskt.ChannelDialWithOptions(addr, vskt.WithConnectDialer(dialer))
	if err != nil {
		t.Fatalf("ChannelDialWithOptions failed: %v", err)
	}
	_ = conn.Close()
	_ = dialer.server.Close()

	dialer = &facadeFakeDialer{}
	client, err := vskt.NewAioClientWithOptions(addr, &vskt.SimpleIoAction{}, vskt.WithConnectDialer(dialer))
	if err != nil {
		t.Fatalf("NewAioClientWithOptions failed: %v", err)
	}
	_ = client.Close()
	_ = dialer.server.Close()
}
