package net

import (
	"errors"
	stdnet "net"
	"testing"
	"time"
)

func TestConnectHelpersWithOptionsUseDialerNetworkAndTimeout(t *testing.T) {
	dialer := &recordingDialer{data: make(chan []byte, 1)}
	conn, err := ConnectWithOptions(
		"example.com", 8080,
		WithConnectNetwork("tcp4"),
		WithConnectTimeout(time.Second),
		WithConnectDialer(dialer),
	)
	if err != nil {
		t.Fatalf("ConnectWithOptions: %v", err)
	}
	_ = conn.Close()
	if dialer.network != "tcp4" || dialer.address != "example.com:8080" {
		t.Fatalf("dial target = %s %s", dialer.network, dialer.address)
	}

	dialer = &recordingDialer{data: make(chan []byte, 1)}
	if err := NetCatWithOptions("127.0.0.1", 1234, []byte("hello"), WithConnectDialer(dialer)); err != nil {
		t.Fatalf("NetCatWithOptions: %v", err)
	}
	if got := string(<-dialer.data); got != "hello" {
		t.Fatalf("NetCatWithOptions wrote %q", got)
	}
	if dialer.network != "tcp" || dialer.address != "127.0.0.1:1234" {
		t.Fatalf("netcat dial target = %s %s", dialer.network, dialer.address)
	}

	addr := &stdnet.TCPAddr{IP: stdnet.ParseIP("127.0.0.1"), Port: 4321}
	dialer = &recordingDialer{data: make(chan []byte, 1)}
	if !IsOpenWithOptions(addr, WithConnectDialer(dialer), WithConnectNetwork("tcp4")) {
		t.Fatal("IsOpenWithOptions should report true for successful dialer")
	}
	if dialer.network != "tcp4" || dialer.address != addr.String() {
		t.Fatalf("is-open dial target = %s %s", dialer.network, dialer.address)
	}
}

func TestConnectWrapperBoundaries(t *testing.T) {
	dialer := &recordingDialer{data: make(chan []byte, 1)}
	addr := &stdnet.TCPAddr{IP: stdnet.ParseIP("127.0.0.1"), Port: 4321}
	if !IsOpenWithOptions(addr, WithConnectDialer(dialer)) || dialer.address == "" {
		t.Fatalf("IsOpenWithOptions did not use injected dialer")
	}
	if IsOpenWithOptions(nil, WithConnectDialer(dialer)) {
		t.Fatalf("IsOpenWithOptions(nil) = true")
	}

	wantErr := errors.New("dial failed")
	dialer = &recordingDialer{err: wantErr, data: make(chan []byte, 1)}
	if err := NetCatWithOptions("127.0.0.1", 1234, []byte("hello"), WithConnectDialer(dialer)); !errors.Is(err, wantErr) {
		t.Fatalf("NetCatWithOptions dial error = %v", err)
	}
	if GetRemoteAddress(nil) != "" || IsConnected(nil) {
		t.Fatalf("nil connection helpers should be empty/false")
	}
	conn, err := ConnectWithOptions("127.0.0.1", 1234, WithConnectDialer(&recordingDialer{data: make(chan []byte, 1)}))
	if err != nil {
		t.Fatalf("ConnectWithOptions pipe: %v", err)
	}
	defer func() { _ = conn.Close() }()
	if GetRemoteAddress(conn) == "" || !IsConnected(conn) {
		t.Fatalf("connection helpers remote=%q connected=%v", GetRemoteAddress(conn), IsConnected(conn))
	}
}

func TestNilDialerOptionsDoNotOverwriteConfiguredDialers(t *testing.T) {
	connectDialer := &recordingDialer{data: make(chan []byte, 1)}
	conn, err := ConnectWithOptions("example.com", 80, WithConnectDialer(connectDialer), WithConnectDialer(nil))
	if err != nil {
		t.Fatalf("ConnectWithOptions nil overwrite error = %v", err)
	}
	_ = conn.Close()
	if connectDialer.address == "" {
		t.Fatal("nil connect dialer option overwrote configured dialer")
	}

	pingDialer := &recordingDialer{data: make(chan []byte, 1)}
	if !PingWithOptions("example.com", WithPingPorts(80), WithPingDialer(pingDialer), WithPingDialer(nil)) {
		t.Fatal("PingWithOptions should use configured dialer after nil option")
	}
	if pingDialer.address == "" {
		t.Fatal("nil ping dialer option overwrote configured dialer")
	}
}
