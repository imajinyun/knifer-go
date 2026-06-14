package socket

import (
	"net"
	"sync/atomic"
	"testing"
	"time"
)

// echoChannelHandler echoes data read from the connection.
type echoChannelHandler struct{}

func (h *echoChannelHandler) Handle(conn net.Conn) error {
	buf := make([]byte, 1024)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	_, err = conn.Write(buf[:n])
	return err
}

type blockingChannelHandler struct {
	started atomic.Int32
	release chan struct{}
}

func (h *blockingChannelHandler) Handle(conn net.Conn) error {
	h.started.Add(1)
	<-h.release
	return nil
}

func TestNioServerEcho(t *testing.T) {
	server, err := NewNioServerAddr(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	server.SetChannelHandler(&echoChannelHandler{})
	defer closeAndReport(t, server.Close)

	server.ListenAsync()

	addr := server.LocalAddr().(*net.TCPAddr)
	conn, err := net.DialTimeout("tcp", addr.String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer closeAndReport(t, conn.Close)

	want := []byte("hello-nio")
	if _, err := conn.Write(want); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	got := make([]byte, len(want))
	if _, err := conn.Read(got); err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("回显数据不一致: got=%q want=%q", got, want)
	}
}

func TestNioServerThreadPoolSizeLimitsHandlers(t *testing.T) {
	handler := &blockingChannelHandler{release: make(chan struct{})}
	server, err := NewNioServerAddrWithConfig(
		&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
		NewSocketConfigWithOptions(WithThreadPoolSize(1)),
	)
	if err != nil {
		t.Fatal(err)
	}
	server.SetChannelHandler(handler)
	defer closeAndReport(t, server.Close)
	server.ListenAsync()

	addr := server.LocalAddr().String()
	first, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	waitForInt32(t, func() int32 { return handler.started.Load() }, 1)

	second, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer closeAndReport(t, second.Close)
	time.Sleep(50 * time.Millisecond)
	if got := handler.started.Load(); got != 1 {
		t.Fatalf("started handlers = %d, want 1 while first handler occupies the only slot", got)
	}

	close(handler.release)
	closeAndReport(t, first.Close)
	waitForInt32(t, func() int32 { return handler.started.Load() }, 2)
}

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
