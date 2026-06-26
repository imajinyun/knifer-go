package vnet

import (
	stdnet "net"

	netimpl "github.com/imajinyun/knifer-go/internal/net"
)

// BuildInetSocketAddress resolves host into a TCP address using defaultPort when the host omits a port.
func BuildInetSocketAddress(host string, defaultPort int) (*stdnet.TCPAddr, error) {
	return BuildInetSocketAddressWithOptions(host, defaultPort)
}

func BuildInetSocketAddressWithOptions(host string, defaultPort int, opts ...AddressOption) (*stdnet.TCPAddr, error) {
	return netimpl.BuildInetSocketAddressWithOptions(host, defaultPort, opts...)
}

func CreateAddress(host string, port int) *stdnet.TCPAddr { return netimpl.CreateAddress(host, port) }

func CreateAddressWithOptions(host string, port int, opts ...AddressOption) (*stdnet.TCPAddr, error) {
	return netimpl.CreateAddressWithOptions(host, port, opts...)
}

func WithAddressNetwork(network string) AddressOption { return netimpl.WithAddressNetwork(network) }

func WithTCPAddrResolver(resolver func(network, address string) (*stdnet.TCPAddr, error)) AddressOption {
	return netimpl.WithTCPAddrResolver(resolver)
}

func GetIPByHost(hostName string) string { return netimpl.GetIPByHost(hostName) }

func WithInterfaceByNameFunc(fn func(string) (*stdnet.Interface, error)) InterfaceOption {
	return netimpl.WithInterfaceByNameFunc(fn)
}

func WithInterfacesFunc(fn func() ([]stdnet.Interface, error)) InterfaceOption {
	return netimpl.WithInterfacesFunc(fn)
}

func WithInterfaceAddrsFunc(fn func(stdnet.Interface) ([]stdnet.Addr, error)) InterfaceOption {
	return netimpl.WithInterfaceAddrsFunc(fn)
}

func WithReverseLookupFunc(fn func(string) ([]string, error)) InterfaceOption {
	return netimpl.WithReverseLookupFunc(fn)
}

func WithNetHostnameFunc(fn func() (string, error)) InterfaceOption {
	return netimpl.WithNetHostnameFunc(fn)
}

// GetNetworkInterface returns the network interface with the given name.
func GetNetworkInterface(name string) (*stdnet.Interface, error) {
	return GetNetworkInterfaceWithOptions(name)
}

func GetNetworkInterfaceWithOptions(name string, opts ...InterfaceOption) (*stdnet.Interface, error) {
	return netimpl.GetNetworkInterfaceWithOptions(name, opts...)
}

// GetNetworkInterfaces returns all network interfaces visible to the local host.
func GetNetworkInterfaces() ([]stdnet.Interface, error) { return GetNetworkInterfacesWithOptions() }

func GetNetworkInterfacesWithOptions(opts ...InterfaceOption) ([]stdnet.Interface, error) {
	return netimpl.GetNetworkInterfacesWithOptions(opts...)
}

// LocalIPv4s returns local IPv4 addresses as strings.
func LocalIPv4s() []string { return LocalIPv4sWithOptions() }

func LocalIPv4sWithOptions(opts ...InterfaceOption) []string {
	return netimpl.LocalIPv4sWithOptions(opts...)
}

// LocalIPv6s returns local IPv6 addresses as strings.
func LocalIPv6s() []string { return LocalIPv6sWithOptions() }

func LocalIPv6sWithOptions(opts ...InterfaceOption) []string {
	return netimpl.LocalIPv6sWithOptions(opts...)
}

// LocalIPs returns local IP addresses as strings.
func LocalIPs() []string { return LocalIPsWithOptions() }

func LocalIPsWithOptions(opts ...InterfaceOption) []string {
	return netimpl.LocalIPsWithOptions(opts...)
}

func ToIPList(addressList []stdnet.IP) []string { return netimpl.ToIPList(addressList) }

// LocalAddressList returns local IP addresses accepted by addressFilter.
func LocalAddressList(addressFilter func(stdnet.IP) bool) []stdnet.IP {
	return LocalAddressListWithOptions(addressFilter)
}

func LocalAddressListWithOptions(addressFilter func(stdnet.IP) bool, opts ...InterfaceOption) []stdnet.IP {
	return netimpl.LocalAddressListWithOptions(addressFilter, opts...)
}

// LocalAddressListByInterface returns local IP addresses whose interface and address pass the provided filters.
func LocalAddressListByInterface(interfaceFilter func(stdnet.Interface) bool, addressFilter func(stdnet.IP) bool) []stdnet.IP {
	return LocalAddressListByInterfaceWithOptions(interfaceFilter, addressFilter)
}

func LocalAddressListByInterfaceWithOptions(interfaceFilter func(stdnet.Interface) bool, addressFilter func(stdnet.IP) bool, opts ...InterfaceOption) []stdnet.IP {
	return netimpl.LocalAddressListByInterfaceWithOptions(interfaceFilter, addressFilter, opts...)
}

// GetLocalhostStr returns the preferred local host IP address as a string.
func GetLocalhostStr() string { return GetLocalhostStrWithOptions() }

func GetLocalhostStrWithOptions(opts ...InterfaceOption) string {
	return netimpl.GetLocalhostStrWithOptions(opts...)
}

// GetLocalhost returns the preferred local host IP address.
func GetLocalhost() stdnet.IP { return GetLocalhostWithOptions() }

func GetLocalhostWithOptions(opts ...InterfaceOption) stdnet.IP {
	return netimpl.GetLocalhostWithOptions(opts...)
}

// GetLocalHostName returns the local host name reported by the operating system.
func GetLocalHostName() string { return GetLocalHostNameWithOptions() }

func GetLocalHostNameWithOptions(opts ...InterfaceOption) string {
	return netimpl.GetLocalHostNameWithOptions(opts...)
}

// GetLocalMACAddress returns the local hardware address formatted with an optional separator.
func GetLocalMACAddress(separator ...string) string {
	return GetLocalMACAddressWithOptions(nil, separator...)
}

func GetLocalMACAddressWithOptions(opts []InterfaceOption, separator ...string) string {
	return netimpl.GetLocalMACAddressWithOptions(opts, separator...)
}

// GetMACAddress returns the hardware address for inetAddress formatted with an optional separator.
func GetMACAddress(inetAddress stdnet.IP, separator ...string) string {
	return GetMACAddressWithOptions(inetAddress, nil, separator...)
}

func GetMACAddressWithOptions(inetAddress stdnet.IP, opts []InterfaceOption, separator ...string) string {
	return netimpl.GetMACAddressWithOptions(inetAddress, opts, separator...)
}

// GetHardwareAddress returns the hardware address associated with inetAddress.
func GetHardwareAddress(inetAddress stdnet.IP) stdnet.HardwareAddr {
	return GetHardwareAddressWithOptions(inetAddress)
}

func GetHardwareAddressWithOptions(inetAddress stdnet.IP, opts ...InterfaceOption) stdnet.HardwareAddr {
	return netimpl.GetHardwareAddressWithOptions(inetAddress, opts...)
}

// GetLocalHardwareAddress returns the hardware address for the preferred local network interface.
func GetLocalHardwareAddress() stdnet.HardwareAddr { return GetLocalHardwareAddressWithOptions() }

func GetLocalHardwareAddressWithOptions(opts ...InterfaceOption) stdnet.HardwareAddr {
	return netimpl.GetLocalHardwareAddressWithOptions(opts...)
}

func GetRemoteAddress(conn stdnet.Conn) string { return netimpl.GetRemoteAddress(conn) }

func IsConnected(conn stdnet.Conn) bool { return netimpl.IsConnected(conn) }
