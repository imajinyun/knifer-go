package socket

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestNioServerCloseClosesActiveConnections(t *testing.T) {
	client, serverConn := net.Pipe()
	defer closeAndReport(t, client.Close)
	listener := &queuedListener{addr: factoryFakeAddr("nio"), conns: make(chan net.Conn, 1)}
	listener.conns <- serverConn
	entered := make(chan struct{})
	nio, err := NewNioServerAddrWithOptions(&net.TCPAddr{Port: 1}, nil, WithListenerFactory(func(*net.TCPAddr) (net.Listener, error) {
		return listener, nil
	}))
	if err != nil {
		t.Fatalf("NewNioServerAddrWithOptions: %v", err)
	}
	nio.SetChannelHandler(ChannelHandlerFunc(func(conn net.Conn) error {
		select {
		case <-entered:
		default:
			close(entered)
		}
		buf := make([]byte, 1)
		_, err := conn.Read(buf)
		return err
	}))
	nio.ListenAsync()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("nio server did not handle connection")
	}
	done := make(chan error, 1)
	go func() { done <- nio.Close() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Close error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("NioServer.Close blocked with active connection")
	}
}

func TestNioServerListenAsyncContextWithCanceledContextExits(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	listener := &blockingListener{addr: factoryFakeAddr("nio-canceled"), done: make(chan struct{})}
	nio, err := NewNioServerAddrWithOptions(&net.TCPAddr{Port: 1}, nil, WithListenerFactory(func(*net.TCPAddr) (net.Listener, error) {
		return listener, nil
	}))
	if err != nil {
		t.Fatalf("NewNioServerAddrWithOptions: %v", err)
	}

	done := nio.ListenAsyncContext(ctx)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("NioServer.ListenAsyncContext did not exit after canceled context")
	}
	if !listener.closed.Load() || nio.IsOpen() {
		t.Fatalf("NioServer canceled start closed=%v open=%v", listener.closed.Load(), nio.IsOpen())
	}
}
