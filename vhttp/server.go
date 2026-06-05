package vhttp

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	httpx "github.com/imajinyun/go-knifer/internal/httpx/http"
)

// NewSimpleServer creates a simple HTTP server on port.
func NewSimpleServer(port int) *SimpleServer { return httpx.NewSimpleServer(port) }

// NewSimpleServerWithOptions creates a simple HTTP server on port with options.
func NewSimpleServerWithOptions(port int, opts ...ServerOption) *SimpleServer {
	return httpx.NewSimpleServerWithOptions(port, opts...)
}

// NewSimpleServerAddr delegates to the internal httpx implementation.
func NewSimpleServerAddr(addr string) *SimpleServer {
	return httpx.NewSimpleServerAddr(addr)
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

// CreateServer delegates to the internal httpx implementation.
func CreateServer(port int) *SimpleServer {
	return httpx.CreateServer(port)
}
