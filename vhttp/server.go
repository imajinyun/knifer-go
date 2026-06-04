package vhttp

import httpx "github.com/imajinyun/go-knifer/internal/httpx/http"

// NewSimpleServer creates a simple HTTP server on port.
func NewSimpleServer(port int) *SimpleServer { return httpx.NewSimpleServer(port) }

// NewSimpleServerAddr delegates to the internal httpx implementation.
func NewSimpleServerAddr(addr string) *SimpleServer {
	return httpx.NewSimpleServerAddr(addr)
}

// CreateServer delegates to the internal httpx implementation.
func CreateServer(port int) *SimpleServer {
	return httpx.CreateServer(port)
}
