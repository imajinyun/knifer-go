package socket

import (
	"bytes"
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

// AioServer is an AIO-style socket server aligned with the utility AIO AioServer.
// In Go, goroutines plus blocking reads are used to simulate AIO callback semantics.
type AioServer struct {
	listener net.Listener
	ioAction IoAction[*bytes.Buffer]
	config   *SocketConfig
	limiter  chan struct{}
	done     chan struct{}

	closed atomic.Bool
	wg     sync.WaitGroup
	mu     sync.Mutex
	conns  map[net.Conn]struct{}
}

// NewAioServer creates a server on the given port.
func NewAioServer(port int) (*AioServer, error) {
	return NewAioServerWithOptions(port)
}

// NewAioServerWithOptions creates a server on the given port with custom config options.
func NewAioServerWithOptions(port int, opts ...ConfigOption) (*AioServer, error) {
	return NewAioServerAddr(&net.TCPAddr{Port: port}, NewSocketConfigWithOptions(opts...))
}

// NewAioServerAddr creates a server from an address and configuration.
func NewAioServerAddr(addr *net.TCPAddr, cfg *SocketConfig) (*AioServer, error) {
	return NewAioServerAddrWithOptions(addr, cfg)
}

// NewAioServerAddrWithOptions creates a server from an address, base config, and custom config options.
func NewAioServerAddrWithOptions(addr *net.TCPAddr, cfg *SocketConfig, opts ...ConfigOption) (*AioServer, error) {
	if addr == nil {
		return nil, NewSocketErrorMsg("address must not be nil")
	}
	if cfg == nil {
		cfg = NewSocketConfig()
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
	s := &AioServer{config: cfg, limiter: newConcurrencyLimiter(cfg), done: make(chan struct{}), conns: make(map[net.Conn]struct{})}
	if err := s.init(addr); err != nil {
		return nil, err
	}
	return s, nil
}

// init initializes the listener.
func (s *AioServer) init(addr *net.TCPAddr) error {
	listenerFactory := func(addr *net.TCPAddr) (net.Listener, error) { return net.ListenTCP("tcp", addr) }
	if s.config != nil && s.config.ListenerFactory != nil {
		listenerFactory = s.config.ListenerFactory
	}
	ln, err := listenerFactory(addr)
	if err != nil {
		return NewSocketError(err)
	}
	s.listener = ln
	return nil
}

// SetIoAction sets the IO action.
func (s *AioServer) SetIoAction(action IoAction[*bytes.Buffer]) *AioServer {
	s.ioAction = action
	return s
}

// IoAction returns the IO action.
func (s *AioServer) IoAction() IoAction[*bytes.Buffer] {
	return s.ioAction
}

// Listener returns the underlying listener.
func (s *AioServer) Listener() net.Listener {
	return s.listener
}

// LocalAddr returns the local listen address.
func (s *AioServer) LocalAddr() net.Addr {
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// Config returns the configuration.
func (s *AioServer) Config() *SocketConfig { return s.config }

// IsOpen reports whether the server is still running.
func (s *AioServer) IsOpen() bool { return !s.closed.Load() }

// Start starts the server; sync controls whether it blocks the current goroutine.
func (s *AioServer) Start(sync bool) {
	if sync {
		s.acceptLoop()
		return
	}
	runWithConfig(s.config, s.acceptLoop)
}

// acceptLoop keeps accepting new connections.
func (s *AioServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() || errors.Is(err, net.ErrClosed) {
				return
			}
			if s.ioAction != nil {
				s.ioAction.Failed(NewSocketError(err), nil)
			}
			continue
		}
		s.handleAccept(conn)
	}
}

// handleAccept creates an AioSession for each connection and triggers callbacks.
func (s *AioServer) handleAccept(conn net.Conn) {
	if s.ioAction == nil {
		_ = conn.Close()
		return
	}
	if !acquireConcurrencySlot(s.limiter, s.done) {
		_ = conn.Close()
		return
	}
	if !s.registerConn(conn) {
		releaseConcurrencySlot(s.limiter)
		_ = conn.Close()
		return
	}
	s.wg.Add(1)
	runWithConfig(s.config, func() {
		defer s.wg.Done()
		defer s.unregisterConn(conn)
		defer releaseConcurrencySlot(s.limiter)

		session := NewAioSession(conn, s.ioAction, s.config)
		defer func() { _ = session.Close() }()
		// Trigger Accept after the connection obtains a handler slot so ThreadPoolSize
		// limits both accept callbacks and read callbacks.
		s.ioAction.Accept(session)

		// Keep reading to simulate chained AIO callbacks.
		for session.IsOpen() && !s.closed.Load() {
			if !session.doRead() {
				return
			}
		}
	})
}

// Close closes the server.
func (s *AioServer) Close() error {
	if s.closed.Swap(true) {
		return nil
	}
	close(s.done)
	if s.listener != nil {
		_ = s.listener.Close()
	}
	s.closeActiveConns()
	s.wg.Wait()
	return nil
}

func (s *AioServer) registerConn(conn net.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed.Load() {
		return false
	}
	if s.conns == nil {
		s.conns = make(map[net.Conn]struct{})
	}
	s.conns[conn] = struct{}{}
	return true
}

func (s *AioServer) unregisterConn(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.conns, conn)
}

func (s *AioServer) closeActiveConns() {
	s.mu.Lock()
	conns := make([]net.Conn, 0, len(s.conns))
	for conn := range s.conns {
		conns = append(conns, conn)
	}
	s.mu.Unlock()
	for _, conn := range conns {
		_ = conn.Close()
	}
}
