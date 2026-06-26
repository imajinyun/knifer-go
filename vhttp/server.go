package vhttp

import (
	"context"
	"io/fs"
	"log"
	"net"
	"net/http"
	"time"

	httpx "github.com/imajinyun/knifer-go/internal/httpx/http"
)

// NewSimpleServer creates a simple HTTP server on port.
func NewSimpleServer(port int) *SimpleServer { return NewSimpleServerWithOptions(port) }

// NewSimpleServerWithOptions creates a simple HTTP server on port with options.
func NewSimpleServerWithOptions(port int, opts ...ServerOption) *SimpleServer {
	return httpx.NewSimpleServerWithOptions(port, opts...)
}

// NewSimpleServerAddr delegates to the internal httpx implementation.
func NewSimpleServerAddr(addr string) *SimpleServer {
	return NewSimpleServerAddrWithOptions(addr)
}

// NewSimpleServerAddrWithOptions creates a simple HTTP server on addr with options.
func NewSimpleServerAddrWithOptions(addr string, opts ...ServerOption) *SimpleServer {
	return httpx.NewSimpleServerAddrWithOptions(addr, opts...)
}

// WithReadHeaderTimeout sets the server read-header timeout.
func WithReadHeaderTimeout(timeout time.Duration) ServerOption {
	return httpx.WithReadHeaderTimeout(timeout)
}

// WithReadTimeout sets the server read timeout.
func WithReadTimeout(timeout time.Duration) ServerOption { return httpx.WithReadTimeout(timeout) }

// WithWriteTimeout sets the server write timeout.
func WithWriteTimeout(timeout time.Duration) ServerOption { return httpx.WithWriteTimeout(timeout) }

// WithIdleTimeout sets the server idle timeout.
func WithIdleTimeout(timeout time.Duration) ServerOption { return httpx.WithIdleTimeout(timeout) }

// WithServerErrorLog sets the server error logger.
func WithServerErrorLog(logger *log.Logger) ServerOption { return httpx.WithServerErrorLog(logger) }

// WithBaseContext sets the server base context function.
func WithBaseContext(baseContext func(net.Listener) context.Context) ServerOption {
	return httpx.WithBaseContext(baseContext)
}

// WithConnContext sets the server connection context function.
func WithConnContext(connContext func(context.Context, net.Conn) context.Context) ServerOption {
	return httpx.WithConnContext(connContext)
}

// WithHTTPServer copies supported settings from server into the created SimpleServer.
func WithHTTPServer(server *http.Server) ServerOption { return httpx.WithHTTPServer(server) }

// WithStaticFileSystem sets the file system used by SetRootWithOptions.
func WithStaticFileSystem(fileSystem http.FileSystem) StaticOption {
	return httpx.WithStaticFileSystem(fileSystem)
}

// WithStaticFS sets an fs.FS used by SetRootWithOptions.
func WithStaticFS(fileSystem fs.FS) StaticOption { return httpx.WithStaticFS(fileSystem) }

// WithFileServerFactory sets the handler factory used by SetRootWithOptions.
func WithFileServerFactory(factory func(http.FileSystem) http.Handler) StaticOption {
	return httpx.WithFileServerFactory(factory)
}

// WithStaticHandler sets the static handler directly and takes precedence over file-system options.
func WithStaticHandler(handler http.Handler) StaticOption { return httpx.WithStaticHandler(handler) }

// WithListenAndServeFunc sets the function used to start serving.
func WithListenAndServeFunc(listenAndServe ListenAndServeFunc) ServerOption {
	return httpx.WithListenAndServeFunc(listenAndServe)
}

// WithListener sets a listener used by Start and StartAsync instead of ListenAndServe.
func WithListener(listener net.Listener) ServerOption { return httpx.WithListener(listener) }

// WithAsyncRunner sets the function used by StartAsync to launch the serving task.
func WithAsyncRunner(runner func(func())) ServerOption { return httpx.WithAsyncRunner(runner) }

// ResetServerStarters clears pending starter functions registered while applying server options.
func ResetServerStarters() { httpx.ResetServerStarters() }

// CreateServer delegates to the internal httpx implementation.
func CreateServer(port int) *SimpleServer {
	return CreateServerWithOptions(port)
}

// CreateServerWithOptions creates a simple HTTP server with options.
func CreateServerWithOptions(port int, opts ...ServerOption) *SimpleServer {
	return httpx.CreateServerWithOptions(port, opts...)
}
