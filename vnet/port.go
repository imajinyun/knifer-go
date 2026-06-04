package vnet

import netimpl "github.com/imajinyun/go-knifer/internal/net"

func IsValidPort(port int) bool { return netimpl.IsValidPort(port) }

func IsUsableLocalPort(port int) bool { return netimpl.IsUsableLocalPort(port) }

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
