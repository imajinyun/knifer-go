package vnet

import (
	stdnet "net"

	netimpl "github.com/imajinyun/go-knifer/internal/net"
)

func BuildInetSocketAddress(host string, defaultPort int) (*stdnet.TCPAddr, error) {
	return netimpl.BuildInetSocketAddress(host, defaultPort)
}

func CreateAddress(host string, port int) *stdnet.TCPAddr { return netimpl.CreateAddress(host, port) }

func GetIPByHost(hostName string) string { return netimpl.GetIPByHost(hostName) }

func GetNetworkInterface(name string) (*stdnet.Interface, error) {
	return netimpl.GetNetworkInterface(name)
}

func GetNetworkInterfaces() ([]stdnet.Interface, error) { return netimpl.GetNetworkInterfaces() }

func LocalIPv4s() []string { return netimpl.LocalIPv4s() }

func LocalIPv6s() []string { return netimpl.LocalIPv6s() }

func LocalIPs() []string { return netimpl.LocalIPs() }

func ToIPList(addressList []stdnet.IP) []string { return netimpl.ToIPList(addressList) }

func LocalAddressList(addressFilter func(stdnet.IP) bool) []stdnet.IP {
	return netimpl.LocalAddressList(addressFilter)
}

func LocalAddressListByInterface(interfaceFilter func(stdnet.Interface) bool, addressFilter func(stdnet.IP) bool) []stdnet.IP {
	return netimpl.LocalAddressListByInterface(interfaceFilter, addressFilter)
}

func GetLocalhostStr() string { return netimpl.GetLocalhostStr() }

func GetLocalhost() stdnet.IP { return netimpl.GetLocalhost() }

func GetLocalHostName() string { return netimpl.GetLocalHostName() }

func GetLocalMACAddress(separator ...string) string { return netimpl.GetLocalMACAddress(separator...) }

func GetMACAddress(inetAddress stdnet.IP, separator ...string) string {
	return netimpl.GetMACAddress(inetAddress, separator...)
}

func GetHardwareAddress(inetAddress stdnet.IP) stdnet.HardwareAddr {
	return netimpl.GetHardwareAddress(inetAddress)
}

func GetLocalHardwareAddress() stdnet.HardwareAddr { return netimpl.GetLocalHardwareAddress() }

func GetRemoteAddress(conn stdnet.Conn) string { return netimpl.GetRemoteAddress(conn) }

func IsConnected(conn stdnet.Conn) bool { return netimpl.IsConnected(conn) }
