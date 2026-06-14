package socket

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestNioClientReceive(t *testing.T) {
	server, err := NewNioServerAddr(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer closeAndReport(t, server.Close)

	server.SetChannelHandler(ChannelHandlerFunc(func(conn net.Conn) error {
		// The server writes once and then closes the connection.
		_, err := conn.Write([]byte("server-push"))
		_ = conn.Close()
		return err
	}))
	server.ListenAsync()

	addr := server.LocalAddr().(*net.TCPAddr)
	client, err := NewNioClientAddr(addr)
	if err != nil {
		t.Fatal(err)
	}
	defer closeAndReport(t, client.Close)

	var (
		mu      sync.Mutex
		message []byte
		done    = make(chan struct{}, 1)
	)
	client.SetChannelHandler(ChannelHandlerFunc(func(conn net.Conn) error {
		buf := make([]byte, 32)
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			return err
		}
		mu.Lock()
		message = append(message, buf[:n]...)
		mu.Unlock()
		select {
		case done <- struct{}{}:
		default:
		}
		return nil
	}))

	client.Listen()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("未在超时时间内收到数据")
	}

	mu.Lock()
	got := string(message)
	mu.Unlock()
	if got != "server-push" {
		t.Errorf("数据不正确: got=%q", got)
	}
}

func TestNioClientWithOptionsUsesIPParser(t *testing.T) {
	client, server := net.Pipe()
	defer closeAndReport(t, server.Close)

	var parsedHost string
	var dialAddr *net.TCPAddr
	nioClient, err := NewNioClientWithOptions("alias-host", 4321,
		WithSocketIPParser(func(host string) net.IP {
			parsedHost = host
			return net.ParseIP("127.0.0.2")
		}),
		WithConnFactory(func(got *net.TCPAddr) (net.Conn, error) {
			dialAddr = got
			return client, nil
		}),
	)
	if err != nil {
		t.Fatalf("NewNioClientWithOptions: %v", err)
	}
	defer closeAndReport(t, nioClient.Close)
	if parsedHost != "alias-host" {
		t.Fatalf("parsed host = %q", parsedHost)
	}
	if dialAddr == nil || !dialAddr.IP.Equal(net.ParseIP("127.0.0.2")) || dialAddr.Port != 4321 {
		t.Fatalf("dial addr = %#v", dialAddr)
	}
}
