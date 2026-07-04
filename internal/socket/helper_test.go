package socket

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

func closeAndReport(t *testing.T, closeFn func() error) {
	t.Helper()
	if err := closeFn(); err != nil {
		t.Errorf("close failed: %v", err)
	}
}

func waitForInt32(t *testing.T, get func() int32, want int32) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if get() >= want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("value = %d, want %d", get(), want)
}

func waitForBool(t *testing.T, get func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if get() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition was not met before timeout")
}

type fakeDialer struct {
	calls   atomic.Int32
	network string
	address string
	server  net.Conn
}

type factoryFakeAddr string

func (a factoryFakeAddr) Network() string { return "tcp" }
func (a factoryFakeAddr) String() string  { return string(a) }

type fakeListener struct {
	addr   net.Addr
	closed atomic.Bool
}

func (l *fakeListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (l *fakeListener) Close() error {
	l.closed.Store(true)
	return nil
}
func (l *fakeListener) Addr() net.Addr { return l.addr }

type blockingListener struct {
	addr   net.Addr
	done   chan struct{}
	closed atomic.Bool
}

func (l *blockingListener) Accept() (net.Conn, error) {
	<-l.done
	return nil, net.ErrClosed
}

func (l *blockingListener) Close() error {
	if l.closed.Swap(true) {
		return nil
	}
	close(l.done)
	return nil
}

func (l *blockingListener) Addr() net.Addr { return l.addr }

type queuedListener struct {
	addr   net.Addr
	conns  chan net.Conn
	closed atomic.Bool
}

func (l *queuedListener) Accept() (net.Conn, error) {
	conn, ok := <-l.conns
	if !ok {
		return nil, net.ErrClosed
	}
	return conn, nil
}

func (l *queuedListener) Close() error {
	if l.closed.Swap(true) {
		return nil
	}
	close(l.conns)
	return nil
}

func (l *queuedListener) Addr() net.Addr { return l.addr }

func (d *fakeDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d.calls.Add(1)
	d.network = network
	d.address = address
	client, server := net.Pipe()
	d.server = server
	return client, nil
}
