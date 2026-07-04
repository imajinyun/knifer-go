package socket

import (
	"bytes"
	"context"
	"errors"
	"net"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
)

func TestFuncDecoderAndEncoder(t *testing.T) {
	var got bytes.Buffer
	decoder := FuncDecoder[*bytes.Buffer](func(session *AioSession, buf *bytes.Buffer) (*bytes.Buffer, bool) {
		return buf, buf.Len() > 0
	})
	encoder := FuncEncoder[*bytes.Buffer](func(session *AioSession, buf *bytes.Buffer, data *bytes.Buffer) {
		got.Write(data.Bytes())
	})

	buf := bytes.NewBufferString("hello")
	result, ok := decoder.Decode(nil, buf)
	if !ok || result.String() != "hello" {
		t.Fatalf("Decode = (%q, %t), want (\"hello\", true)", result.String(), ok)
	}

	encoder.Encode(nil, nil, bytes.NewBufferString("world"))
	if got.String() != "world" {
		t.Fatalf("Encode result = %q, want \"world\"", got.String())
	}
}

func TestAioClientNilSession(t *testing.T) {
	c := &AioClient{}
	if s := c.Session(); s != nil {
		t.Fatalf("Session = %v, want nil", s)
	}
	if act := c.IoAction(); act != nil {
		t.Fatalf("IoAction = %v, want nil", act)
	}
	if _, err := c.Write([]byte("x")); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("Write with nil session err = %v, want internal socket error", err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close with nil session err = %v", err)
	}
}

func TestNilConnectDialerOptionDoesNotOverwriteConfiguredDialer(t *testing.T) {
	dialer := socketRecordingDialer{conn: pipeConn(t), called: make(chan string, 1)}
	conn, err := ConnectWithOptions("example.com", 80, WithConnectDialer(dialer), WithConnectDialer(nil))
	if err != nil {
		t.Fatalf("ConnectWithOptions nil overwrite error = %v", err)
	}
	_ = conn.Close()
	if dialer.called == nil {
		t.Fatal("test dialer did not record calls")
	}
	if got := <-dialer.called; got != "tcp/example.com:80" {
		t.Fatalf("dial target = %q", got)
	}
}

func TestAioClientWithConnNilActionAndClosedWrite(t *testing.T) {
	client, server := net.Pipe()
	defer closeAndReport(t, client.Close)
	defer closeAndReport(t, server.Close)

	c := NewAioClientWithConn(client, nil, nil)
	if c.Session() == nil {
		t.Fatal("NewAioClientWithConn should create a session even when action is nil")
	}
	if c.IoAction() != nil {
		t.Fatal("IoAction should be nil when action provider is nil")
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close err = %v", err)
	}
	if _, err := c.Write([]byte("after-close")); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("Write after Close err = %v, want internal socket error", err)
	}
}

func TestAioSessionGetters(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	action := &SimpleIoAction{}
	cfg := NewSocketConfig()
	session := NewAioSession(server, action, cfg)

	if c := session.Conn(); c != server {
		t.Fatal("Conn mismatch")
	}
	if act := session.IoAction(); act != action {
		t.Fatal("IoAction mismatch")
	}
	if addr := session.RemoteAddress(); addr == nil {
		t.Fatal("RemoteAddress should not be nil")
	}
}

func TestAioSessionWriteAndClose(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	action := &SimpleIoAction{}
	session := NewAioSession(server, action, NewSocketConfig())

	go func() {
		buf := make([]byte, 1024)
		n, _ := client.Read(buf)
		if string(buf[:n]) != "test-data" {
			t.Errorf("received = %q, want %q", string(buf[:n]), "test-data")
		}
	}()

	if err := session.WriteAndClose([]byte("test-data")); err != nil {
		t.Fatalf("WriteAndClose error = %v", err)
	}
}

func TestAioSessionCloseInAndCloseOut(t *testing.T) {
	// net.Pipe connections are not *net.TCPConn,
	// so CloseIn/CloseOut should be no-ops (nil error).
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	session := NewAioSession(server, &SimpleIoAction{}, NewSocketConfig())

	if err := session.CloseIn(); err != nil {
		t.Fatalf("CloseIn error = %v", err)
	}
	if err := session.CloseOut(); err != nil {
		t.Fatalf("CloseOut error = %v", err)
	}
}

func TestAioSessionNilConnectionBoundaries(t *testing.T) {
	session := NewAioSession(nil, &SimpleIoAction{}, NewSocketConfigWithOptions(WithReadBufferSize(0), WithWriteBufferSize(0)))
	if session.IsOpen() {
		t.Fatal("nil connection session should not be open")
	}
	if session.Read() != session {
		t.Fatal("Read should return the session itself")
	}
	if err := session.CloseIn(); err != nil {
		t.Fatalf("CloseIn nil connection err = %v", err)
	}
	if err := session.CloseOut(); err != nil {
		t.Fatalf("CloseOut nil connection err = %v", err)
	}
	if _, err := session.Write([]byte("x")); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("Write nil connection err = %v, want internal socket error", err)
	}
	if err := session.Close(); err != nil {
		t.Fatalf("Close nil connection err = %v", err)
	}
}

func TestAioServerGetters(t *testing.T) {
	cfg := NewSocketConfig()
	cfg.ListenerFactory = func(addr *net.TCPAddr) (net.Listener, error) {
		return &fakeListener{addr: factoryFakeAddr("0.0.0.0:9999")}, nil
	}
	server, err := NewAioServerAddrWithOptions(&net.TCPAddr{Port: 9999}, cfg)
	if err != nil {
		t.Fatalf("NewAioServerAddrWithOptions error = %v", err)
	}

	action := &SimpleIoAction{}
	server.SetIoAction(action)

	if act := server.IoAction(); act != action {
		t.Fatal("IoAction mismatch")
	}
	if c := server.Config(); c == nil {
		t.Fatal("Config should not be nil")
	}
	if !server.IsOpen() {
		t.Fatal("IsOpen should be true")
	}
	server.Close()
	if server.IsOpen() {
		t.Fatal("IsOpen should be false after Close")
	}
}

func TestAioServerConstructorsAndDeterministicStart(t *testing.T) {
	runs := 0
	server, err := NewAioServerWithOptions(0,
		WithRunner(func(fn func()) { runs++; fn() }),
		WithListenerFactory(func(*net.TCPAddr) (net.Listener, error) {
			return &fakeListener{addr: factoryFakeAddr("aio-start")}, nil
		}),
	)
	if err != nil {
		t.Fatalf("NewAioServerWithOptions = %v", err)
	}
	server.Start(false)
	if runs != 1 {
		t.Fatalf("Start(false) runner calls = %d, want 1", runs)
	}
	closeAndReport(t, server.Close)

	syncServer, err := NewAioServer(0)
	if err != nil {
		t.Fatalf("NewAioServer = %v", err)
	}
	closeAndReport(t, syncServer.Close)
}

func TestNioServerGetters(t *testing.T) {
	cfg := NewSocketConfig()
	cfg.ListenerFactory = func(addr *net.TCPAddr) (net.Listener, error) {
		return &fakeListener{addr: factoryFakeAddr("0.0.0.0:9998")}, nil
	}
	server, err := NewNioServerWithConfig(9998, cfg)
	if err != nil {
		t.Fatalf("NewNioServerWithConfig error = %v", err)
	}

	if c := server.Config(); c == nil {
		t.Fatal("Config should not be nil")
	}
	if !server.IsOpen() {
		t.Fatal("IsOpen should be true")
	}
	server.Close()
	if server.IsOpen() {
		t.Fatal("IsOpen should be false after Close")
	}
}

func TestNioClientWriteBoundaries(t *testing.T) {
	client, server := net.Pipe()
	defer closeAndReport(t, server.Close)
	nio := &NioClient{conn: client, config: NewSocketConfig()}
	if got, err := nio.Write(nil, []byte("")); err != nil || got != nio {
		t.Fatalf("Write empty fragments = (%v, %v), want self nil", got, err)
	}
	readDone := make(chan string, 1)
	go func() {
		buf := make([]byte, 8)
		n, _ := server.Read(buf)
		readDone <- string(buf[:n])
	}()
	if got, err := nio.Write([]byte("hello")); err != nil || got != nio {
		t.Fatalf("Write data = (%v, %v), want self nil", got, err)
	}
	select {
	case got := <-readDone:
		if got != "hello" {
			t.Fatalf("server read = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not receive Write data")
	}
	closeAndReport(t, nio.Close)
	if _, err := nio.Write([]byte("closed")); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("Write after Close err = %v, want internal socket error", err)
	}
	if _, err := (&NioClient{}).Write([]byte("nil")); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("Write nil connection err = %v, want internal socket error", err)
	}
}

func TestNioServerConstructorsAndStart(t *testing.T) {
	runs := 0
	server, err := NewNioServerWithOptions(0,
		WithRunner(func(fn func()) { runs++; fn() }),
		WithListenerFactory(func(*net.TCPAddr) (net.Listener, error) {
			return &fakeListener{addr: factoryFakeAddr("nio-start")}, nil
		}),
	)
	if err != nil {
		t.Fatalf("NewNioServerWithOptions = %v", err)
	}
	done := server.ListenAsync()
	<-done
	if runs != 1 {
		t.Fatalf("ListenAsync runner calls = %d, want 1", runs)
	}
	server.Start()
	closeAndReport(t, server.Close)

	realServer, err := NewNioServer(0)
	if err != nil {
		t.Fatalf("NewNioServer = %v", err)
	}
	closeAndReport(t, realServer.Close)
}

func TestConnectAddr(t *testing.T) {
	conn, err := ConnectAddr(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 19999}, time.Millisecond)
	if err == nil {
		conn.Close()
		t.Fatal("expected connection error to localhost:19999")
	}
}

func TestConfigSettersNilGuard(t *testing.T) {
	cfg := NewSocketConfig()

	result := cfg.SetClock(nil)
	if result != cfg {
		t.Fatal("SetClock(nil) should return self")
	}
	result = cfg.SetRunner(nil)
	if result != cfg {
		t.Fatal("SetRunner(nil) should return self")
	}
	result = cfg.SetListenerFactory(nil)
	if result != cfg {
		t.Fatal("SetListenerFactory(nil) should return self")
	}
	result = cfg.SetConnFactory(nil)
	if result != cfg {
		t.Fatal("SetConnFactory(nil) should return self")
	}
	result = cfg.SetSocketIPParser(nil)
	if result != cfg {
		t.Fatal("SetSocketIPParser(nil) should return self")
	}

	// Verify non-nil values are actually set
	clock := time.Now
	cfg.SetClock(clock)
	if cfg.Clock == nil {
		t.Fatal("Clock should be set after SetClock with non-nil")
	}

	runner := func(fn func()) { fn() }
	cfg.SetRunner(runner)
	if cfg.Runner == nil {
		t.Fatal("Runner should be set after SetRunner with non-nil")
	}

	factory := func(addr *net.TCPAddr) (net.Listener, error) {
		return &fakeListener{addr: factoryFakeAddr("0.0.0.0:0")}, nil
	}
	cfg.SetListenerFactory(factory)
	if cfg.ListenerFactory == nil {
		t.Fatal("ListenerFactory should be set after SetListenerFactory with non-nil")
	}

	connFactory := func(addr *net.TCPAddr) (net.Conn, error) {
		client, _ := net.Pipe()
		return client, nil
	}
	cfg.SetConnFactory(connFactory)
	if cfg.ConnFactory == nil {
		t.Fatal("ConnFactory should be set after SetConnFactory with non-nil")
	}

	parser := net.ParseIP
	cfg.SetSocketIPParser(parser)
	if cfg.IPParser == nil {
		t.Fatal("IPParser should be set after SetSocketIPParser with non-nil")
	}
}

type socketRecordingDialer struct {
	conn   net.Conn
	called chan string
}

func (d socketRecordingDialer) DialContext(_ context.Context, network, address string) (net.Conn, error) {
	if d.called != nil {
		d.called <- network + "/" + address
	}
	return d.conn, nil
}

func pipeConn(t *testing.T) net.Conn {
	t.Helper()
	client, server := net.Pipe()
	t.Cleanup(func() { _ = server.Close() })
	return client
}
