package socket

import (
	"bytes"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// AioSession represents one client connection, aligned with the utility AIO AioSession.
// It owns read/write buffers and triggers IoAction.DoAction via asynchronous callbacks.
type AioSession struct {
	conn        net.Conn
	ioAction    IoAction[*bytes.Buffer]
	readBuffer  *bytes.Buffer
	writeBuffer *bytes.Buffer
	scratch     []byte // Temporary space for one Read call.

	readTimeout  time.Duration
	writeTimeout time.Duration
	clock        func() time.Time

	closed atomic.Bool
	mu     sync.Mutex
}

// NewAioSession creates an AioSession.
func NewAioSession(conn net.Conn, ioAction IoAction[*bytes.Buffer], cfg *SocketConfig) *AioSession {
	if cfg == nil {
		cfg = NewSocketConfig()
	}
	readSize := cfg.ReadBufferSize
	if readSize <= 0 {
		readSize = DefaultBufferSize
	}
	writeSize := cfg.WriteBufferSize
	if writeSize <= 0 {
		writeSize = DefaultBufferSize
	}
	return &AioSession{
		conn:         conn,
		ioAction:     ioAction,
		readBuffer:   bytes.NewBuffer(make([]byte, 0, readSize)),
		writeBuffer:  bytes.NewBuffer(make([]byte, 0, writeSize)),
		scratch:      make([]byte, readSize),
		readTimeout:  time.Duration(cfg.ReadTimeout) * time.Millisecond,
		writeTimeout: time.Duration(cfg.WriteTimeout) * time.Millisecond,
		clock:        cfg.Clock,
	}
}

func (s *AioSession) now() time.Time {
	if s.clock != nil {
		return s.clock()
	}
	return time.Now()
}

// Conn returns the underlying connection.
func (s *AioSession) Conn() net.Conn { return s.conn }

// ReadBuffer returns the read buffer.
func (s *AioSession) ReadBuffer() *bytes.Buffer { return s.readBuffer }

// WriteBuffer returns the write buffer.
func (s *AioSession) WriteBuffer() *bytes.Buffer { return s.writeBuffer }

// IoAction returns the IO action.
func (s *AioSession) IoAction() IoAction[*bytes.Buffer] { return s.ioAction }

// RemoteAddress returns the remote address.
func (s *AioSession) RemoteAddress() net.Addr {
	return GetRemoteAddress(s.conn)
}

// Read asynchronously reads once, equivalent to the utility toolkit's read().
// In Go this is implemented with a goroutine and a completion callback to IoAction.
func (s *AioSession) Read() *AioSession {
	if !s.IsOpen() {
		return s
	}
	go s.doRead()
	return s
}

// doRead reads once and invokes callbacks; false means read failed or the connection closed.
func (s *AioSession) doRead() bool {
	if !s.IsOpen() {
		return false
	}
	if s.readTimeout > 0 {
		_ = s.conn.SetReadDeadline(s.now().Add(s.readTimeout))
	} else {
		_ = s.conn.SetReadDeadline(time.Time{})
	}
	n, err := s.conn.Read(s.scratch)
	if err != nil {
		if s.ioAction != nil {
			s.ioAction.Failed(err, s)
		}
		_ = s.Close()
		return false
	}
	s.readBuffer.Reset()
	s.readBuffer.Write(s.scratch[:n])
	s.callbackRead()
	return true
}

// callbackRead invokes the IoAction.DoAction callback.
func (s *AioSession) callbackRead() {
	if s.ioAction != nil {
		s.ioAction.DoAction(s, s.readBuffer)
	}
}

// Write writes data.
func (s *AioSession) Write(data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.IsOpen() {
		return 0, NewSocketErrorMsg("session is closed")
	}
	if s.writeTimeout > 0 {
		_ = s.conn.SetWriteDeadline(s.now().Add(s.writeTimeout))
	} else {
		_ = s.conn.SetWriteDeadline(time.Time{})
	}
	n, err := s.conn.Write(data)
	if err != nil {
		return n, NewSocketError(err)
	}
	return n, nil
}

// WriteAndClose writes data and closes the write side.
func (s *AioSession) WriteAndClose(data []byte) error {
	if _, err := s.Write(data); err != nil {
		return err
	}
	return s.CloseOut()
}

// IsOpen reports whether the session is still open.
func (s *AioSession) IsOpen() bool {
	return s.conn != nil && !s.closed.Load()
}

// CloseIn closes the read side.
func (s *AioSession) CloseIn() error {
	if tc, ok := s.conn.(*net.TCPConn); ok {
		if err := tc.CloseRead(); err != nil {
			return NewSocketError(err)
		}
	}
	return nil
}

// CloseOut closes the write side.
func (s *AioSession) CloseOut() error {
	if tc, ok := s.conn.(*net.TCPConn); ok {
		if err := tc.CloseWrite(); err != nil {
			return NewSocketError(err)
		}
	}
	return nil
}

// Close closes the session.
func (s *AioSession) Close() error {
	if s.closed.Swap(true) {
		return nil
	}
	if s.conn != nil {
		_ = s.conn.Close()
	}
	s.readBuffer = nil
	s.writeBuffer = nil
	return nil
}
