package http

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const defaultReadHeaderTimeout = 10 * time.Second

// SimpleServer is a simple HTTP server, aligned with the utility toolkit-http SimpleServer.
type SimpleServer struct {
	addr           string
	mux            *http.ServeMux
	server         *http.Server
	listenAndServe ListenAndServeFunc
	asyncRunner    func(func())
}

type staticConfig struct {
	fileSystem http.FileSystem
	fileServer func(http.FileSystem) http.Handler
	handler    http.Handler
}

// ServerOption customizes SimpleServer construction.
type ServerOption func(*http.Server)

// StaticOption customizes SimpleServer static file registration.
type StaticOption func(*staticConfig)

// WithStaticFileSystem sets the file system used by SetRootWithOptions.
func WithStaticFileSystem(fileSystem http.FileSystem) StaticOption {
	return func(c *staticConfig) { c.fileSystem = fileSystem }
}

// WithStaticFS sets an fs.FS used by SetRootWithOptions.
func WithStaticFS(fileSystem fs.FS) StaticOption {
	return func(c *staticConfig) {
		if fileSystem != nil {
			c.fileSystem = http.FS(fileSystem)
		}
	}
}

// WithFileServerFactory sets the handler factory used by SetRootWithOptions.
func WithFileServerFactory(factory func(http.FileSystem) http.Handler) StaticOption {
	return func(c *staticConfig) {
		if factory != nil {
			c.fileServer = factory
		}
	}
}

// WithStaticHandler sets the static handler directly and takes precedence over file-system options.
func WithStaticHandler(handler http.Handler) StaticOption {
	return func(c *staticConfig) { c.handler = handler }
}

func applyStaticOptions(dir string, opts []StaticOption) staticConfig {
	cfg := staticConfig{fileSystem: http.Dir(dir), fileServer: http.FileServer}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.fileSystem == nil {
		cfg.fileSystem = http.Dir(dir)
	}
	if cfg.fileServer == nil {
		cfg.fileServer = http.FileServer
	}
	return cfg
}

// ListenAndServeFunc starts serving with the provided HTTP server.
type ListenAndServeFunc func(*http.Server) error

var (
	serverStarters     sync.Map
	serverAsyncRunners sync.Map
)

func storeServerStarter(server *http.Server, listenAndServe ListenAndServeFunc) {
	serverStarters.Store(server, listenAndServe)
}

func takeServerStarter(server *http.Server) (ListenAndServeFunc, bool) {
	starter, ok := serverStarters.LoadAndDelete(server)
	if !ok {
		return nil, false
	}
	return starter.(ListenAndServeFunc), true
}

func storeServerAsyncRunner(server *http.Server, runner func(func())) {
	serverAsyncRunners.Store(server, runner)
}

func takeServerAsyncRunner(server *http.Server) (func(func()), bool) {
	runner, ok := serverAsyncRunners.LoadAndDelete(server)
	if !ok {
		return nil, false
	}
	return runner.(func(func())), true
}

// ResetServerStarters clears pending starter functions registered while applying server options.
func ResetServerStarters() {
	serverStarters.Clear()
	serverAsyncRunners.Clear()
}

// NewSimpleServer creates a simple server on the specified port.
func NewSimpleServer(port int) *SimpleServer {
	return NewSimpleServerWithOptions(port)
}

// NewSimpleServerWithOptions creates a simple server on the specified port with options.
func NewSimpleServerWithOptions(port int, opts ...ServerOption) *SimpleServer {
	return NewSimpleServerAddrWithOptions(fmt.Sprintf(":%d", port), opts...)
}

// NewSimpleServerAddr creates a simple server with the specified listen address.
func NewSimpleServerAddr(addr string) *SimpleServer {
	return NewSimpleServerAddrWithOptions(addr)
}

// NewSimpleServerAddrWithOptions creates a simple server with the specified listen address and options.
func NewSimpleServerAddrWithOptions(addr string, opts ...ServerOption) *SimpleServer {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
	}
	defer serverStarters.Delete(server)
	for _, opt := range opts {
		if opt != nil {
			opt(server)
		}
	}
	if server.Addr == "" {
		server.Addr = addr
	}
	if server.Handler == nil {
		server.Handler = mux
	}
	listenAndServe := defaultListenAndServe
	if starter, ok := takeServerStarter(server); ok {
		listenAndServe = starter
	}
	asyncRunner := defaultAsyncRunner
	if runner, ok := takeServerAsyncRunner(server); ok {
		asyncRunner = runner
	}
	return &SimpleServer{
		addr:           server.Addr,
		mux:            mux,
		server:         server,
		listenAndServe: listenAndServe,
		asyncRunner:    asyncRunner,
	}
}

func defaultListenAndServe(server *http.Server) error { return server.ListenAndServe() }

func defaultAsyncRunner(fn func()) { go fn() }

// WithListenAndServeFunc sets the function used to start serving.
func WithListenAndServeFunc(listenAndServe ListenAndServeFunc) ServerOption {
	return func(s *http.Server) {
		if listenAndServe != nil {
			storeServerStarter(s, listenAndServe)
		}
	}
}

// WithListener sets a listener used by Start and StartAsync instead of ListenAndServe.
func WithListener(listener net.Listener) ServerOption {
	return WithListenAndServeFunc(func(server *http.Server) error {
		if listener == nil {
			return defaultListenAndServe(server)
		}
		return server.Serve(listener)
	})
}

// WithAsyncRunner sets the function used by StartAsync to launch the serving task.
func WithAsyncRunner(runner func(func())) ServerOption {
	return func(s *http.Server) {
		if runner != nil {
			storeServerAsyncRunner(s, runner)
		}
	}
}

// WithReadHeaderTimeout sets the server read-header timeout.
func WithReadHeaderTimeout(timeout time.Duration) ServerOption {
	return func(s *http.Server) { s.ReadHeaderTimeout = timeout }
}

// WithReadTimeout sets the server read timeout.
func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(s *http.Server) { s.ReadTimeout = timeout }
}

// WithWriteTimeout sets the server write timeout.
func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(s *http.Server) { s.WriteTimeout = timeout }
}

// WithIdleTimeout sets the server idle timeout.
func WithIdleTimeout(timeout time.Duration) ServerOption {
	return func(s *http.Server) { s.IdleTimeout = timeout }
}

// WithServerErrorLog sets the server error logger.
func WithServerErrorLog(logger *log.Logger) ServerOption {
	return func(s *http.Server) { s.ErrorLog = logger }
}

// WithBaseContext sets the server base context function.
func WithBaseContext(baseContext func(net.Listener) context.Context) ServerOption {
	return func(s *http.Server) { s.BaseContext = baseContext }
}

// WithConnContext sets the server connection context function.
func WithConnContext(connContext func(context.Context, net.Conn) context.Context) ServerOption {
	return func(s *http.Server) { s.ConnContext = connContext }
}

// WithHTTPServer copies supported settings from server into the created SimpleServer.
func WithHTTPServer(server *http.Server) ServerOption {
	return func(s *http.Server) {
		if server == nil {
			return
		}
		s.Addr = server.Addr
		s.Handler = server.Handler
		s.DisableGeneralOptionsHandler = server.DisableGeneralOptionsHandler
		s.TLSConfig = server.TLSConfig
		s.ReadTimeout = server.ReadTimeout
		s.ReadHeaderTimeout = server.ReadHeaderTimeout
		s.WriteTimeout = server.WriteTimeout
		s.IdleTimeout = server.IdleTimeout
		s.MaxHeaderBytes = server.MaxHeaderBytes
		s.TLSNextProto = server.TLSNextProto
		s.ConnState = server.ConnState
		s.ErrorLog = server.ErrorLog
		s.BaseContext = server.BaseContext
		s.ConnContext = server.ConnContext
		s.HTTP2 = server.HTTP2
		s.Protocols = server.Protocols
	}
}

// AddAction registers a path handler.
func (s *SimpleServer) AddAction(path string, handler http.HandlerFunc) *SimpleServer {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	s.mux.HandleFunc(path, handler)
	return s
}

// AddHandler registers an http.Handler.
func (s *SimpleServer) AddHandler(path string, handler http.Handler) *SimpleServer {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	s.mux.Handle(path, handler)
	return s
}

// SetRoot sets the static file root directory.
func (s *SimpleServer) SetRoot(dir string) *SimpleServer {
	return s.SetRootWithOptions(dir)
}

// SetRootWithOptions sets the static file root directory with options.
func (s *SimpleServer) SetRootWithOptions(dir string, opts ...StaticOption) *SimpleServer {
	cfg := applyStaticOptions(dir, opts)
	handler := cfg.handler
	if handler == nil {
		handler = cfg.fileServer(cfg.fileSystem)
	}
	s.mux.Handle("/", handler)
	return s
}

// Start starts the server synchronously and blocks.
func (s *SimpleServer) Start() error {
	listenAndServe := s.listenAndServe
	if listenAndServe == nil {
		listenAndServe = defaultListenAndServe
	}
	return listenAndServe(s.server)
}

// StartAsync starts the server asynchronously and returns an error channel.
func (s *SimpleServer) StartAsync() <-chan error {
	ch := make(chan error, 1)
	runner := s.asyncRunner
	if runner == nil {
		runner = defaultAsyncRunner
	}
	runner(func() {
		listenAndServe := s.listenAndServe
		if listenAndServe == nil {
			listenAndServe = defaultListenAndServe
		}
		err := listenAndServe(s.server)
		if err != nil && err != http.ErrServerClosed {
			ch <- err
		}
		close(ch)
	})
	return ch
}

// Stop shuts down the server gracefully.
func (s *SimpleServer) Stop(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.StopWithContext(ctx)
}

// StopWithContext shuts down the server gracefully using ctx.
func (s *SimpleServer) StopWithContext(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.server.Shutdown(ctx)
}

// CreateServer creates a simple HTTP server, aligned with HttpUtil.createServer.
func CreateServer(port int) *SimpleServer { return CreateServerWithOptions(port) }

// CreateServerWithOptions creates a simple HTTP server with options.
func CreateServerWithOptions(port int, opts ...ServerOption) *SimpleServer {
	return NewSimpleServerWithOptions(port, opts...)
}
