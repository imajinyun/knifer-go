package vskt_test

import (
	"errors"
	"net"
	"testing"

	"github.com/imajinyun/go-knifer/vskt"
)

func TestFacadeSocketFactories(t *testing.T) {
	addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9999}
	listener := &facadeFakeListener{addr: facadeFakeAddr("facade-aio")}
	aio, err := vskt.NewAioServerAddrWithOptions(addr, nil, vskt.WithListenerFactory(func(got *net.TCPAddr) (net.Listener, error) {
		if got != addr {
			return nil, errors.New("unexpected aio addr")
		}
		return listener, nil
	}))
	if err != nil {
		t.Fatalf("NewAioServerAddrWithOptions: %v", err)
	}
	if aio.Listener() != listener || aio.LocalAddr().String() != "facade-aio" {
		t.Fatalf("aio listener = %#v addr=%v", aio.Listener(), aio.LocalAddr())
	}
	_ = aio.Close()

	listener = &facadeFakeListener{addr: facadeFakeAddr("facade-nio")}
	nio, err := vskt.NewNioServerAddrWithOptions(addr, nil, vskt.WithListenerFactory(func(got *net.TCPAddr) (net.Listener, error) {
		if got != addr {
			return nil, errors.New("unexpected nio addr")
		}
		return listener, nil
	}))
	if err != nil {
		t.Fatalf("NewNioServerAddrWithOptions: %v", err)
	}
	if nio.Listener() != listener || nio.LocalAddr().String() != "facade-nio" {
		t.Fatalf("nio listener = %#v addr=%v", nio.Listener(), nio.LocalAddr())
	}
	_ = nio.Close()

	client, server := net.Pipe()
	defer func() { _ = server.Close() }()
	nioClient, err := vskt.NewNioClientAddrWithOptions(addr, vskt.WithConnFactory(func(got *net.TCPAddr) (net.Conn, error) {
		if got != addr {
			return nil, errors.New("unexpected client addr")
		}
		return client, nil
	}))
	if err != nil {
		t.Fatalf("NewNioClientAddrWithOptions: %v", err)
	}
	if nioClient.Channel() != client {
		t.Fatalf("nio client channel = %#v", nioClient.Channel())
	}
	_ = nioClient.Close()
}
