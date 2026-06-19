package vnet

import (
	stdnet "net"

	netimpl "github.com/imajinyun/go-knifer/internal/net"
)

func IsValidPort(port int) bool { return netimpl.IsValidPort(port) }

// IsUsableLocalPort reports whether port can be listened on locally.
func IsUsableLocalPort(port int) bool { return IsUsableLocalPortWithOptions(port) }

func WithPortNetwork(network string) PortOption { return netimpl.WithPortNetwork(network) }

func WithPortHost(host string) PortOption { return netimpl.WithPortHost(host) }

func WithPortListenerFactory(factory func(network, address string) (stdnet.Listener, error)) PortOption {
	return netimpl.WithPortListenerFactory(factory)
}

func IsUsableLocalPortWithOptions(port int, opts ...PortOption) bool {
	return netimpl.IsUsableLocalPortWithOptions(port, opts...)
}

// GetUsableLocalPort returns an available local port.
func GetUsableLocalPort() (int, error) { return GetUsableLocalPortWithOptions() }

func GetUsableLocalPortWithOptions(opts ...PortOption) (int, error) {
	return netimpl.GetUsableLocalPortWithOptions(opts...)
}

// GetUsableLocalPortFrom returns an available local port greater than or equal to minPort.
func GetUsableLocalPortFrom(minPort int) (int, error) {
	return GetUsableLocalPortFromWithOptions(minPort)
}

func GetUsableLocalPortFromWithOptions(minPort int, opts ...PortOption) (int, error) {
	return netimpl.GetUsableLocalPortFromWithOptions(minPort, opts...)
}

// GetUsableLocalPortInRange returns an available local port within the inclusive range.
func GetUsableLocalPortInRange(minPort, maxPort int) (int, error) {
	return GetUsableLocalPortInRangeWithOptions(minPort, maxPort)
}

func GetUsableLocalPortInRangeWithOptions(minPort, maxPort int, opts ...PortOption) (int, error) {
	return netimpl.GetUsableLocalPortInRangeWithOptions(minPort, maxPort, opts...)
}

// GetUsableLocalPorts returns up to numRequested available local ports within the inclusive range.
func GetUsableLocalPorts(numRequested, minPort, maxPort int) ([]int, error) {
	return GetUsableLocalPortsWithOptions(numRequested, minPort, maxPort)
}

func GetUsableLocalPortsWithOptions(numRequested, minPort, maxPort int, opts ...PortOption) ([]int, error) {
	return netimpl.GetUsableLocalPortsWithOptions(numRequested, minPort, maxPort, opts...)
}

// NewLocalPortGenerator returns a generator that scans local ports starting at beginPort.
func NewLocalPortGenerator(beginPort int) *LocalPortGenerator {
	return NewLocalPortGeneratorWithOptions(beginPort)
}

func NewLocalPortGeneratorWithOptions(beginPort int, opts ...PortOption) *LocalPortGenerator {
	return netimpl.NewLocalPortGeneratorWithOptions(beginPort, opts...)
}
