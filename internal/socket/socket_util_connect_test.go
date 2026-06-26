package socket

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
)

func TestSocketUtilConnect(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer closeAndReport(t, ln.Close)

	go func() {
		c, _ := ln.Accept()
		if c != nil {
			_ = c.Close()
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	conn, err := Connect("127.0.0.1", addr.Port, time.Second)
	if err != nil {
		t.Fatalf("Connect 失败: %v", err)
	}
	defer closeAndReport(t, conn.Close)

	if !IsConnected(conn) {
		t.Errorf("IsConnected 应返回 true")
	}
	if GetRemoteAddress(conn) == nil {
		t.Errorf("GetRemoteAddress 不应返回 nil")
	}
}

func TestSocketConnectWithOptions(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer closeAndReport(t, ln.Close)

	go func() {
		c, _ := ln.Accept()
		if c != nil {
			_ = c.Close()
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	conn, err := ConnectWithOptions("127.0.0.1", addr.Port, WithConnectTimeout(time.Second), WithConnectNetwork("tcp"))
	if err != nil {
		t.Fatalf("ConnectWithOptions failed: %v", err)
	}
	defer closeAndReport(t, conn.Close)
}

func TestSocketConnectOptionsUseDialerAndContext(t *testing.T) {
	dialer := &fakeDialer{}
	conn, err := ConnectWithOptions("example.com", 443, WithConnectNetwork("tcp4"), WithConnectDialer(dialer))
	if err != nil {
		t.Fatalf("ConnectWithOptions with fake dialer failed: %v", err)
	}
	defer closeAndReport(t, conn.Close)
	defer closeAndReport(t, dialer.server.Close)
	if dialer.calls.Load() != 1 || dialer.network != "tcp4" || dialer.address != "example.com:443" {
		t.Fatalf("dialer call = (%d, %q, %q), want (1, tcp4, example.com:443)", dialer.calls.Load(), dialer.network, dialer.address)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ConnectWithOptions("127.0.0.1", 1, WithConnectContext(ctx)); err == nil {
		t.Fatal("ConnectWithOptions with canceled context error = nil")
	}
}

func TestSocketConnectAddrWithOptionsUsesDialer(t *testing.T) {
	dialer := &fakeDialer{}
	addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
	conn, err := ConnectAddrWithOptions(addr, WithConnectDialer(dialer))
	if err != nil {
		t.Fatalf("ConnectAddrWithOptions failed: %v", err)
	}
	closeAndReport(t, conn.Close)
	closeAndReport(t, dialer.server.Close)
}

func TestSocketConnectErrorAndFallbackBoundaries(t *testing.T) {
	if _, err := ConnectAddrWithOptions(nil); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("ConnectAddrWithOptions(nil) err = %v, want internal socket error", err)
	}

	dialer := &fakeDialer{}
	var nilCtx context.Context
	conn, err := ConnectWithOptions("example.com", 443,
		nil,
		WithConnectContext(nilCtx),
		WithConnectNetwork(""),
		WithConnectDialer(dialer),
	)
	if err != nil {
		t.Fatalf("ConnectWithOptions fallback dialer err = %v", err)
	}
	closeAndReport(t, conn.Close)
	closeAndReport(t, dialer.server.Close)
	if dialer.network != "tcp" || dialer.address != "example.com:443" {
		t.Fatalf("fallback dialer got network=%q address=%q", dialer.network, dialer.address)
	}

	if addr := GetRemoteAddress(nil); addr != nil {
		t.Fatalf("GetRemoteAddress(nil) = %#v, want nil", addr)
	}
	if IsConnected(nil) {
		t.Fatal("IsConnected(nil) should be false")
	}
}
