package vnet_test

import (
	"context"
	"io"
	stdnet "net"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vnet"
)

type recordingDialer struct {
	network string
	address string
	ctxErr  error
	data    chan []byte
}

func (d *recordingDialer) DialContext(ctx context.Context, network, address string) (stdnet.Conn, error) {
	d.network = network
	d.address = address
	d.ctxErr = ctx.Err()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	client, server := stdnet.Pipe()
	go func() {
		defer func() { _ = server.Close() }()
		payload, _ := io.ReadAll(server)
		if d.data != nil {
			d.data <- payload
		}
	}()
	return client, nil
}

func TestVNetConnectOptionsFacade(t *testing.T) {
	dialer := &recordingDialer{data: make(chan []byte, 1)}
	conn, err := vnet.ConnectWithOptions(
		"example.com", 8080,
		vnet.WithConnectNetwork("tcp4"),
		vnet.WithConnectTimeout(time.Second),
		vnet.WithConnectDialer(dialer),
	)
	if err != nil {
		t.Fatalf("ConnectWithOptions: %v", err)
	}
	_ = conn.Close()
	if dialer.network != "tcp4" || dialer.address != "example.com:8080" {
		t.Fatalf("dial target = %s %s", dialer.network, dialer.address)
	}

	dialer = &recordingDialer{data: make(chan []byte, 1)}
	if err := vnet.NetCatWithOptions("127.0.0.1", 1234, []byte("hello"), vnet.WithConnectDialer(dialer)); err != nil {
		t.Fatalf("NetCatWithOptions: %v", err)
	}
	if got := string(<-dialer.data); got != "hello" {
		t.Fatalf("NetCatWithOptions wrote %q", got)
	}

	addr := &stdnet.TCPAddr{IP: stdnet.ParseIP("127.0.0.1"), Port: 4321}
	dialer = &recordingDialer{data: make(chan []byte, 1)}
	if !vnet.IsOpenWithOptions(addr, vnet.WithConnectDialer(dialer)) {
		t.Fatal("IsOpenWithOptions should report true")
	}
}

func TestVNetConnectAndPingContextFacade(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	dialer := &recordingDialer{}
	if _, err := vnet.ConnectWithOptions("example.com", 80, vnet.WithConnectContext(canceled), vnet.WithConnectDialer(dialer)); err == nil {
		t.Fatal("ConnectWithOptions should return canceled context error")
	}
	if dialer.ctxErr == nil {
		t.Fatal("ConnectWithOptions did not pass the configured context to dialer")
	}

	dialer = &recordingDialer{}
	if vnet.PingWithOptions("example.com", vnet.WithPingContext(canceled), vnet.WithPingDialer(dialer), vnet.WithPingPorts(80)) {
		t.Fatal("PingWithOptions should report false for canceled context")
	}
	if dialer.ctxErr == nil {
		t.Fatal("PingWithOptions did not pass the configured context to dialer")
	}
}
