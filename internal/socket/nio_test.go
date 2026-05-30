package socket

import (
	"net"
	"sync"
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
