package vnet_test

import (
	"errors"
	stdnet "net"
	"reflect"
	"testing"

	"github.com/imajinyun/knifer-go/vnet"
)

type stubListener struct{}

func (stubListener) Accept() (stdnet.Conn, error) { return nil, errors.New("stub listener") }
func (stubListener) Close() error                 { return nil }
func (stubListener) Addr() stdnet.Addr {
	return &stdnet.TCPAddr{IP: stdnet.ParseIP("127.0.0.1"), Port: 12345}
}

func TestVNetProviderOptionsFacade(t *testing.T) {
	var network, address string
	addr, err := vnet.BuildInetSocketAddressWithOptions("example.com", 8080, vnet.WithAddressNetwork("tcp4"), vnet.WithTCPAddrResolver(func(n, a string) (*stdnet.TCPAddr, error) {
		network, address = n, a
		return &stdnet.TCPAddr{IP: stdnet.ParseIP("10.0.0.2"), Port: 8080}, nil
	}))
	if err != nil || addr.Port != 8080 {
		t.Fatalf("BuildInetSocketAddressWithOptions = %#v %v", addr, err)
	}
	if network != "tcp4" || address != "example.com:8080" {
		t.Fatalf("address resolver target = %s %s", network, address)
	}

	if !vnet.IsUsableLocalPortWithOptions(23456, vnet.WithPortNetwork("tcp4"), vnet.WithPortHost("127.0.0.2"), vnet.WithPortListenerFactory(func(n, a string) (stdnet.Listener, error) {
		network, address = n, a
		return stubListener{}, nil
	})) {
		t.Fatal("IsUsableLocalPortWithOptions should use listener factory")
	}
	if network != "tcp4" || address != "127.0.0.2:23456" {
		t.Fatalf("listener target = %s %s", network, address)
	}

	iface := stdnet.Interface{Name: "vnet0", HardwareAddr: stdnet.HardwareAddr{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}}
	_, ipNet, err := stdnet.ParseCIDR("10.9.8.7/24")
	if err != nil {
		t.Fatal(err)
	}
	ipNet.IP = stdnet.ParseIP("10.9.8.7")
	opts := []vnet.InterfaceOption{
		vnet.WithInterfaceByNameFunc(func(name string) (*stdnet.Interface, error) { return &iface, nil }),
		vnet.WithInterfacesFunc(func() ([]stdnet.Interface, error) { return []stdnet.Interface{iface}, nil }),
		vnet.WithInterfaceAddrsFunc(func(stdnet.Interface) ([]stdnet.Addr, error) { return []stdnet.Addr{ipNet}, nil }),
		vnet.WithReverseLookupFunc(func(string) ([]string, error) { return []string{"vnet.local."}, nil }),
		vnet.WithNetHostnameFunc(func() (string, error) { return "fallback", nil }),
	}
	gotIface, err := vnet.GetNetworkInterfaceWithOptions("vnet0", opts...)
	if err != nil || gotIface.Name != "vnet0" {
		t.Fatalf("GetNetworkInterfaceWithOptions = %#v %v", gotIface, err)
	}
	if got := vnet.LocalIPv4sWithOptions(opts...); len(got) != 1 || got[0] != "10.9.8.7" {
		t.Fatalf("LocalIPv4sWithOptions = %#v", got)
	}
	if got := vnet.GetLocalHostNameWithOptions(opts...); got != "vnet.local" {
		t.Fatalf("GetLocalHostNameWithOptions = %q", got)
	}
	if got := vnet.GetLocalMACAddressWithOptions(opts, "-"); got != "01-02-03-04-05-06" {
		t.Fatalf("GetLocalMACAddressWithOptions = %q", got)
	}
}

func TestVNetAddressDefaultFacadeWrappers(t *testing.T) {
	var network, address string
	resolver := func(n, a string) (*stdnet.TCPAddr, error) {
		network, address = n, a
		return &stdnet.TCPAddr{IP: stdnet.ParseIP("10.0.0.3"), Port: 9090}, nil
	}

	addr, err := vnet.BuildInetSocketAddress("127.0.0.1", 8080)
	if err != nil || addr.Port != 8080 {
		t.Fatalf("BuildInetSocketAddress = %#v, %v", addr, err)
	}
	addr, err = vnet.CreateAddressWithOptions("example.com", 9090, vnet.WithAddressNetwork("tcp4"), vnet.WithTCPAddrResolver(resolver))
	if err != nil || addr.Port != 9090 || network != "tcp4" || address != "example.com:9090" {
		t.Fatalf("CreateAddressWithOptions addr=%#v err=%v target=%s %s", addr, err, network, address)
	}
	addr = vnet.CreateAddress("127.0.0.1", 8081)
	if addr == nil || addr.Port != 8081 {
		t.Fatalf("CreateAddress = %#v", addr)
	}
}

func TestVNetInterfaceFacadeWrappers(t *testing.T) {
	iface := stdnet.Interface{Name: "vnet0", HardwareAddr: stdnet.HardwareAddr{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}}
	_, ipNet, err := stdnet.ParseCIDR("10.9.8.7/24")
	if err != nil {
		t.Fatal(err)
	}
	ipNet.IP = stdnet.ParseIP("10.9.8.7")
	opts := []vnet.InterfaceOption{
		vnet.WithInterfaceByNameFunc(func(name string) (*stdnet.Interface, error) { return &iface, nil }),
		vnet.WithInterfacesFunc(func() ([]stdnet.Interface, error) { return []stdnet.Interface{iface}, nil }),
		vnet.WithInterfaceAddrsFunc(func(stdnet.Interface) ([]stdnet.Addr, error) { return []stdnet.Addr{ipNet}, nil }),
		vnet.WithReverseLookupFunc(func(string) ([]string, error) { return nil, errors.New("no reverse") }),
		vnet.WithNetHostnameFunc(func() (string, error) { return "fallback-host", nil }),
	}

	ifaces, err := vnet.GetNetworkInterfacesWithOptions(opts...)
	if err != nil || len(ifaces) != 1 || ifaces[0].Name != "vnet0" {
		t.Fatalf("GetNetworkInterfacesWithOptions = %#v, %v", ifaces, err)
	}
	if got := vnet.LocalAddressListWithOptions(nil, opts...); len(got) != 1 || !got[0].Equal(stdnet.ParseIP("10.9.8.7")) {
		t.Fatalf("LocalAddressListWithOptions = %#v", got)
	}
	if got := vnet.LocalAddressListByInterfaceWithOptions(func(i stdnet.Interface) bool { return i.Name == "vnet0" }, nil, opts...); len(got) != 1 {
		t.Fatalf("LocalAddressListByInterfaceWithOptions = %#v", got)
	}
	if got := vnet.LocalIPsWithOptions(opts...); !reflect.DeepEqual(got, []string{"10.9.8.7"}) {
		t.Fatalf("LocalIPsWithOptions = %#v", got)
	}
	if got := vnet.LocalIPv6sWithOptions(opts...); len(got) != 0 {
		t.Fatalf("LocalIPv6sWithOptions = %#v", got)
	}
	if got := vnet.GetLocalhostStrWithOptions(opts...); got != "10.9.8.7" {
		t.Fatalf("GetLocalhostStrWithOptions = %q", got)
	}
	if got := vnet.GetLocalhostWithOptions(opts...); !got.Equal(stdnet.ParseIP("10.9.8.7")) {
		t.Fatalf("GetLocalhostWithOptions = %v", got)
	}
	if got := vnet.GetLocalHostNameWithOptions(opts...); got != "fallback-host" {
		t.Fatalf("GetLocalHostNameWithOptions fallback = %q", got)
	}
	if got := vnet.GetHardwareAddressWithOptions(stdnet.ParseIP("10.9.8.7"), opts...); !reflect.DeepEqual(got, iface.HardwareAddr) {
		t.Fatalf("GetHardwareAddressWithOptions = %v", got)
	}
	if got := vnet.GetMACAddressWithOptions(stdnet.ParseIP("10.9.8.7"), opts, "-"); got != "01-02-03-04-05-06" {
		t.Fatalf("GetMACAddressWithOptions = %q", got)
	}
	if got := vnet.GetLocalHardwareAddressWithOptions(opts...); !reflect.DeepEqual(got, iface.HardwareAddr) {
		t.Fatalf("GetLocalHardwareAddressWithOptions = %v", got)
	}
	if got := vnet.ToIPList([]stdnet.IP{stdnet.ParseIP("10.0.0.1"), stdnet.ParseIP("10.0.0.1"), nil}); !reflect.DeepEqual(got, []string{"10.0.0.1"}) {
		t.Fatalf("ToIPList = %#v", got)
	}
}
