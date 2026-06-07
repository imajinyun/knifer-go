package socket

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

// NioServer is an event-driven TCP server aligned with the utility NIO NioServer.
// In Go, goroutines plus blocking Accept/Read calls provide equivalent semantics.
type NioServer struct {
	listener net.Listener
	handler  ChannelHandler
	addr     *net.TCPAddr
	config   *SocketConfig
	limiter  chan struct{}
	done     chan struct{}

	closed atomic.Bool
	wg     sync.WaitGroup
	mu     sync.Mutex
}

// NewNioServer creates and initializes a server on the given port.
func NewNioServer(port int) (*NioServer, error) {
	return NewNioServerWithOptions(port)
}

// NewNioServerWithOptions creates and initializes a server on the given port with custom config options.
func NewNioServerWithOptions(port int, opts ...ConfigOption) (*NioServer, error) {
	return NewNioServerAddrWithConfig(&net.TCPAddr{Port: port}, NewSocketConfigWithOptions(opts...))
}

// NewNioServerWithConfig creates and initializes a server on the given port with config.
func NewNioServerWithConfig(port int, cfg *SocketConfig) (*NioServer, error) {
	return NewNioServerAddrWithConfig(&net.TCPAddr{Port: port}, cfg)
}

// NewNioServerAddr creates a server from the specified address.
func NewNioServerAddr(addr *net.TCPAddr) (*NioServer, error) {
	return NewNioServerAddrWithConfig(addr, NewSocketConfig())
}

// NewNioServerAddrWithConfig creates a server from the specified address and configuration.
func NewNioServerAddrWithConfig(addr *net.TCPAddr, cfg *SocketConfig) (*NioServer, error) {
	return NewNioServerAddrWithOptions(addr, cfg)
}

// NewNioServerAddrWithOptions creates a server from the specified address, base config, and custom config options.
func NewNioServerAddrWithOptions(addr *net.TCPAddr, cfg *SocketConfig, opts ...ConfigOption) (*NioServer, error) {
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
	s := &NioServer{addr: addr, config: cfg, limiter: newConcurrencyLimiter(cfg), done: make(chan struct{})}
	if err := s.init(addr); err != nil {
		return nil, err
	}
	return s, nil
}

// init initializes the listener.
func (s *NioServer) init(addr *net.TCPAddr) error {
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

// SetChannelHandler sets the data handler.
func (s *NioServer) SetChannelHandler(handler ChannelHandler) *NioServer {
	s.handler = handler
	return s
}

// Listener returns the underlying net.Listener.
func (s *NioServer) Listener() net.Listener {
	return s.listener
}

// LocalAddr returns the local listen address, useful for dynamic port tests.
func (s *NioServer) LocalAddr() net.Addr {
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// Config returns the configuration.
func (s *NioServer) Config() *SocketConfig { return s.config }

// Start begins listening and blocks the current goroutine.
func (s *NioServer) Start() {
	s.Listen()
}

// Listen starts synchronous blocking listening.
func (s *NioServer) Listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
			continue
		}
		s.handleAccept(conn)
	}
}

// ListenAsync starts listening asynchronously and closes the returned channel when done.
func (s *NioServer) ListenAsync() <-chan struct{} {
	done := make(chan struct{})
	runWithConfig(s.config, func() {
		defer close(done)
		s.Listen()
	})
	return done
}

// handleAccept handles read events from a connection in a new goroutine.
func (s *NioServer) handleAccept(conn net.Conn) {
	if s.handler == nil {
		_ = conn.Close()
		return
	}
	if !acquireConcurrencySlot(s.limiter, s.done) {
		_ = conn.Close()
		return
	}
	s.wg.Add(1)
	runWithConfig(s.config, func() {
		defer s.wg.Done()
		defer releaseConcurrencySlot(s.limiter)
		defer func() { _ = conn.Close() }()
		for {
			if s.closed.Load() {
				return
			}
			// Simulate NIO read events by invoking the handler when the connection is readable.
			// The handler usually calls conn.Read once to consume data.
			if err := s.handler.Handle(conn); err != nil {
				return
			}
		}
	})
}

// Close closes the server.
func (s *NioServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed.Swap(true) {
		return nil
	}
	close(s.done)
	if s.listener != nil {
		_ = s.listener.Close()
	}
	s.wg.Wait()
	return nil
}

// IsOpen reports whether the server is still running.
func (s *NioServer) IsOpen() bool {
	return !s.closed.Load()
}
