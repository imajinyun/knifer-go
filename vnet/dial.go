package vnet

import (
	stdnet "net"
	"time"

	netimpl "github.com/imajinyun/go-knifer/internal/net"
)

func NetCat(host string, port int, data []byte, timeout time.Duration) error {
	return netimpl.NetCat(host, port, data, timeout)
}

func Ping(ip string, timeout time.Duration) bool { return netimpl.Ping(ip, timeout) }

func IsOpen(address *stdnet.TCPAddr, timeout time.Duration) bool {
	return netimpl.IsOpen(address, timeout)
}

func Connect(hostname string, port int, timeout time.Duration) (stdnet.Conn, error) {
	return netimpl.Connect(hostname, port, timeout)
}
