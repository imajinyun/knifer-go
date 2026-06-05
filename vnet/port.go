package vnet

import netimpl "github.com/imajinyun/go-knifer/internal/net"

func IsValidPort(port int) bool { return netimpl.IsValidPort(port) }

func IsUsableLocalPort(port int) bool { return netimpl.IsUsableLocalPort(port) }

func WithPortNetwork(network string) PortOption { return netimpl.WithPortNetwork(network) }

func WithPortHost(host string) PortOption { return netimpl.WithPortHost(host) }

func IsUsableLocalPortWithOptions(port int, opts ...PortOption) bool {
	return netimpl.IsUsableLocalPortWithOptions(port, opts...)
}

func GetUsableLocalPort() (int, error) { return netimpl.GetUsableLocalPort() }

func GetUsableLocalPortFrom(minPort int) (int, error) { return netimpl.GetUsableLocalPortFrom(minPort) }

func GetUsableLocalPortInRange(minPort, maxPort int) (int, error) {
	return netimpl.GetUsableLocalPortInRange(minPort, maxPort)
}

func GetUsableLocalPorts(numRequested, minPort, maxPort int) ([]int, error) {
	return netimpl.GetUsableLocalPorts(numRequested, minPort, maxPort)
}

func NewLocalPortGenerator(beginPort int) *LocalPortGenerator {
	return netimpl.NewLocalPortGenerator(beginPort)
}
