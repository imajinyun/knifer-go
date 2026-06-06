package socket

import (
	"bytes"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// echoIoAction implements IoAction and writes received data back as-is.
type echoIoAction struct {
	accepted atomic.Int32
	failed   atomic.Int32
}

func (a *echoIoAction) Accept(session *AioSession) {
	a.accepted.Add(1)
}

func (a *echoIoAction) DoAction(session *AioSession, data *bytes.Buffer) {
	if data == nil || data.Len() == 0 {
		return
	}
	_, _ = session.Write(data.Bytes())
}

func (a *echoIoAction) Failed(err error, session *AioSession) {
	a.failed.Add(1)
}

type blockingIoAction struct {
	accepted atomic.Int32
	release  chan struct{}
}

type deadlineConn struct {
	readDeadline  time.Time
	writeDeadline time.Time
}

func (c *deadlineConn) Read([]byte) (int, error)          { return 0, io.EOF }
func (c *deadlineConn) Write(p []byte) (int, error)       { return len(p), nil }
func (c *deadlineConn) Close() error                      { return nil }
func (c *deadlineConn) LocalAddr() net.Addr               { return fakeAddr("local") }
func (c *deadlineConn) RemoteAddr() net.Addr              { return fakeAddr("remote") }
func (c *deadlineConn) SetDeadline(time.Time) error       { return nil }
func (c *deadlineConn) SetReadDeadline(t time.Time) error { c.readDeadline = t; return nil }
func (c *deadlineConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline = t
	return nil
}

type fakeAddr string

func (a fakeAddr) Network() string { return string(a) }
func (a fakeAddr) String() string  { return string(a) }

func (a *blockingIoAction) Accept(session *AioSession) {
	a.accepted.Add(1)
	<-a.release
}

func (a *blockingIoAction) DoAction(session *AioSession, data *bytes.Buffer) {}

func (a *blockingIoAction) Failed(err error, session *AioSession) {}

func TestAioSessionClockControlsDeadlines(t *testing.T) {
	fixed := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)
	cfg := NewSocketConfigWithOptions(
		WithReadTimeout(1000),
		WithWriteTimeout(2000),
		WithClock(func() time.Time { return fixed }),
	)

	writeConn := &deadlineConn{}
	writeSession := NewAioSession(writeConn, nil, cfg)
	if _, err := writeSession.Write([]byte("x")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if want := fixed.Add(2 * time.Second); !writeConn.writeDeadline.Equal(want) {
		t.Fatalf("write deadline = %v, want %v", writeConn.writeDeadline, want)
	}

	readConn := &deadlineConn{}
	readSession := NewAioSession(readConn, nil, cfg)
	if readSession.doRead() {
		t.Fatal("doRead should fail when fake connection returns EOF")
	}
	if want := fixed.Add(time.Second); !readConn.readDeadline.Equal(want) {
		t.Fatalf("read deadline = %v, want %v", readConn.readDeadline, want)
	}
}

func TestAioServerEcho(t *testing.T) {
	action := &echoIoAction{}
	server, err := NewAioServerAddr(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}, NewSocketConfig())
	if err != nil {
		t.Fatal(err)
	}
	server.SetIoAction(action)
	defer closeAndReport(t, server.Close)

	server.Start(false)

	addr := server.LocalAddr().(*net.TCPAddr)
	conn, err := net.DialTimeout("tcp", addr.String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer closeAndReport(t, conn.Close)

	want := []byte("hello-aio")
	if _, err := conn.Write(want); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	got := make([]byte, len(want))
	if _, err := conn.Read(got); err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("AioServer 回显数据不一致: got=%q want=%q", got, want)
	}
	if action.accepted.Load() != 1 {
		t.Errorf("Accept 应被回调 1 次，实际 %d", action.accepted.Load())
	}
}

func TestAioServerThreadPoolSizeLimitsHandlers(t *testing.T) {
	action := &blockingIoAction{release: make(chan struct{})}
	server, err := NewAioServerAddr(
		&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
		NewSocketConfigWithOptions(WithThreadPoolSize(1)),
	)
	if err != nil {
		t.Fatal(err)
	}
	server.SetIoAction(action)
	defer closeAndReport(t, server.Close)
	server.Start(false)

	addr := server.LocalAddr().String()
	first, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	waitForInt32(t, func() int32 { return action.accepted.Load() }, 1)

	second, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer closeAndReport(t, second.Close)
	time.Sleep(50 * time.Millisecond)
	if got := action.accepted.Load(); got != 1 {
		t.Fatalf("accepted = %d, want 1 while first handler occupies the only slot", got)
	}

	close(action.release)
	closeAndReport(t, first.Close)
	waitForInt32(t, func() int32 { return action.accepted.Load() }, 2)
}

// clientIoAction notifies done after receiving one message.
type clientIoAction struct {
	mu      sync.Mutex
	message []byte
	done    chan struct{}
}

func (a *clientIoAction) Accept(session *AioSession) {}

func (a *clientIoAction) DoAction(session *AioSession, data *bytes.Buffer) {
	a.mu.Lock()
	a.message = append(a.message, data.Bytes()...)
	a.mu.Unlock()
	select {
	case a.done <- struct{}{}:
	default:
	}
}

func (a *clientIoAction) Failed(err error, session *AioSession) {}

func TestAioClient(t *testing.T) {
	server, err := NewAioServerAddr(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}, NewSocketConfig())
	if err != nil {
		t.Fatal(err)
	}
	server.SetIoAction(&echoIoAction{})
	defer closeAndReport(t, server.Close)
	server.Start(false)

	addr := server.LocalAddr().(*net.TCPAddr)
	clientAction := &clientIoAction{done: make(chan struct{}, 1)}

	client, err := NewAioClient(addr, clientAction)
	if err != nil {
		t.Fatal(err)
	}
	defer closeAndReport(t, client.Close)

	if _, err := client.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	client.Read()

	select {
	case <-clientAction.done:
	case <-time.After(2 * time.Second):
		t.Fatal("AioClient 未在超时内收到回显")
	}

	clientAction.mu.Lock()
	got := string(clientAction.message)
	clientAction.mu.Unlock()
	if got != "ping" {
		t.Errorf("AioClient 收到错误数据: %q", got)
	}
}

func TestSimpleIoAction(t *testing.T) {
	var (
		acceptCalled bool
		failed       error
		received     []byte
	)
	action := &SimpleIoAction{
		OnAccept: func(session *AioSession) { acceptCalled = true },
		OnDoAction: func(session *AioSession, data *bytes.Buffer) {
			received = append(received, data.Bytes()...)
		},
		OnFailed: func(err error, session *AioSession) { failed = err },
	}

	action.Accept(nil)
	action.DoAction(nil, bytes.NewBufferString("hi"))
	action.Failed(NewSocketErrorMsg("oops"), nil)

	if !acceptCalled {
		t.Errorf("OnAccept 未被调用")
	}
	if string(received) != "hi" {
		t.Errorf("OnDoAction 数据错误: %q", received)
	}
	if failed == nil {
		t.Errorf("OnFailed 未传递错误")
	}
}
