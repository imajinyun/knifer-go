package vnet

import (
	"context"
	stdnet "net"
	"time"

	netimpl "github.com/imajinyun/go-knifer/internal/net"
)

func NetCat(host string, port int, data []byte, timeout time.Duration) error {
	return netimpl.NetCat(host, port, data, timeout)
}

func Ping(ip string, timeout time.Duration) bool { return netimpl.Ping(ip, timeout) }

func WithPingContext(ctx context.Context) PingOption { return netimpl.WithPingContext(ctx) }

func WithPingTimeout(timeout time.Duration) PingOption { return netimpl.WithPingTimeout(timeout) }

func WithPingPorts(ports ...int) PingOption { return netimpl.WithPingPorts(ports...) }

func WithPingNetwork(network string) PingOption { return netimpl.WithPingNetwork(network) }

func WithPingDialer(d Dialer) PingOption { return netimpl.WithPingDialer(d) }

func PingWithOptions(ip string, opts ...PingOption) bool { return netimpl.PingWithOptions(ip, opts...) }

func IsOpen(address *stdnet.TCPAddr, timeout time.Duration) bool {
	return netimpl.IsOpen(address, timeout)
}

func Connect(hostname string, port int, timeout time.Duration) (stdnet.Conn, error) {
	return netimpl.Connect(hostname, port, timeout)
}
