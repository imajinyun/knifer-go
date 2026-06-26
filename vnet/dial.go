package vnet

import (
	"context"
	stdnet "net"
	"time"

	netimpl "github.com/imajinyun/knifer-go/internal/net"
)

// NetCat sends data to host:port over TCP.
func NetCat(host string, port int, data []byte, timeout time.Duration) error {
	return NetCatWithOptions(host, port, data, WithConnectTimeout(timeout))
}

// NetCatWithOptions sends data to host:port using custom connection options.
func NetCatWithOptions(host string, port int, data []byte, opts ...ConnectOption) error {
	return netimpl.NetCatWithOptions(host, port, data, opts...)
}

// WithConnectContext sets the context used by connection helpers.
func WithConnectContext(ctx context.Context) ConnectOption { return netimpl.WithConnectContext(ctx) }

// WithConnectTimeout bounds connection attempts made by connection helpers.
func WithConnectTimeout(timeout time.Duration) ConnectOption {
	return netimpl.WithConnectTimeout(timeout)
}

// WithConnectNetwork sets the network used by connection helpers, such as tcp, tcp4, or tcp6.
func WithConnectNetwork(network string) ConnectOption { return netimpl.WithConnectNetwork(network) }

// WithConnectDialer sets the dialer used by connection helpers.
func WithConnectDialer(d Dialer) ConnectOption { return netimpl.WithConnectDialer(d) }

// Ping checks whether an IP or host is reachable by opening a TCP connection to common ports.
func Ping(ip string, timeout time.Duration) bool {
	return PingWithOptions(ip, WithPingTimeout(timeout))
}

// WithPingContext sets the context used by PingWithOptions.
func WithPingContext(ctx context.Context) PingOption { return netimpl.WithPingContext(ctx) }

// WithPingTimeout sets the timeout for each connection attempt made by PingWithOptions.
func WithPingTimeout(timeout time.Duration) PingOption { return netimpl.WithPingTimeout(timeout) }

// WithPingPorts sets the destination ports PingWithOptions probes.
func WithPingPorts(ports ...int) PingOption { return netimpl.WithPingPorts(ports...) }

// WithPingNetwork sets the network used by PingWithOptions, such as tcp, tcp4, or tcp6.
func WithPingNetwork(network string) PingOption { return netimpl.WithPingNetwork(network) }

// WithPingDialer sets the dialer used by PingWithOptions.
func WithPingDialer(d Dialer) PingOption { return netimpl.WithPingDialer(d) }

// PingWithOptions checks whether an IP or host is reachable with custom probe options.
func PingWithOptions(ip string, opts ...PingOption) bool { return netimpl.PingWithOptions(ip, opts...) }

// IsOpen reports whether address can be opened within timeout.
func IsOpen(address *stdnet.TCPAddr, timeout time.Duration) bool {
	return IsOpenWithOptions(address, WithConnectTimeout(timeout))
}

// IsOpenWithOptions reports whether address can be opened with custom connection options.
func IsOpenWithOptions(address *stdnet.TCPAddr, opts ...ConnectOption) bool {
	return netimpl.IsOpenWithOptions(address, opts...)
}

// Connect opens a TCP connection to host:port.
func Connect(hostname string, port int, timeout time.Duration) (stdnet.Conn, error) {
	return ConnectWithOptions(hostname, port, WithConnectTimeout(timeout))
}

// ConnectWithOptions opens a connection to host:port using custom connection options.
func ConnectWithOptions(hostname string, port int, opts ...ConnectOption) (stdnet.Conn, error) {
	return netimpl.ConnectWithOptions(hostname, port, opts...)
}
