package socket

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestAioServerCloseClosesActiveConnections(t *testing.T) {
	client, serverConn := net.Pipe()
	defer closeAndReport(t, client.Close)
	listener := &queuedListener{addr: factoryFakeAddr("aio"), conns: make(chan net.Conn, 1)}
	listener.conns <- serverConn
	accepted := make(chan struct{})
	aio, err := NewAioServerAddrWithOptions(&net.TCPAddr{Port: 1}, nil, WithListenerFactory(func(*net.TCPAddr) (net.Listener, error) {
		return listener, nil
	}))
	if err != nil {
		t.Fatalf("NewAioServerAddrWithOptions: %v", err)
	}
	aio.SetIoAction(&SimpleIoAction{OnAccept: func(*AioSession) { close(accepted) }})
	aio.Start(false)
	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("aio server did not accept connection")
	}
	done := make(chan error, 1)
	go func() { done <- aio.Close() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Close error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("AioServer.Close blocked with active connection")
	}
}

func TestAioServerStartContextWithCanceledContextExits(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	listener := &blockingListener{addr: factoryFakeAddr("aio-canceled"), done: make(chan struct{})}
	aio, err := NewAioServerAddrWithOptions(&net.TCPAddr{Port: 1}, nil, WithListenerFactory(func(*net.TCPAddr) (net.Listener, error) {
		return listener, nil
	}))
	if err != nil {
		t.Fatalf("NewAioServerAddrWithOptions: %v", err)
	}

	done := make(chan struct{})
	go func() {
		aio.StartContext(ctx, true)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("AioServer.StartContext did not exit after canceled context")
	}
	if !listener.closed.Load() || aio.IsOpen() {
		t.Fatalf("AioServer canceled start closed=%v open=%v", listener.closed.Load(), aio.IsOpen())
	}
}
