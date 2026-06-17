package net

import (
	"errors"
	stdnet "net"
	"reflect"
	"testing"
)

func TestAddressOptionsUseResolver(t *testing.T) {
	var network, address string
	resolver := func(n, a string) (*stdnet.TCPAddr, error) {
		network, address = n, a
		return &stdnet.TCPAddr{IP: stdnet.ParseIP("10.0.0.1"), Port: 4321}, nil
	}
	addr, err := BuildInetSocketAddressWithOptions("example.com", 8080, WithAddressNetwork("tcp4"), WithTCPAddrResolver(resolver))
	if err != nil || addr.Port != 4321 {
		t.Fatalf("BuildInetSocketAddressWithOptions = %#v %v", addr, err)
	}
	if network != "tcp4" || address != "example.com:8080" {
		t.Fatalf("resolver target = %s %s", network, address)
	}

	addr, err = CreateAddressWithOptions("example.org", 9090, WithTCPAddrResolver(resolver))
	if err != nil || addr.Port != 4321 {
		t.Fatalf("CreateAddressWithOptions = %#v %v", addr, err)
	}
	if network != "tcp" || address != "example.org:9090" {
		t.Fatalf("resolver target = %s %s", network, address)
	}

	addr, err = BuildInetSocketAddress("127.0.0.1:8081", 9090)
	if err != nil || addr.Port != 8081 {
		t.Fatalf("BuildInetSocketAddress host:port = %#v, %v", addr, err)
	}
	addr = CreateAddress("127.0.0.1", 8082)
	if addr == nil || addr.Port != 8082 {
		t.Fatalf("CreateAddress = %#v", addr)
	}
}

func TestInterfaceOptions(t *testing.T) {
	iface := stdnet.Interface{Name: "eth-test", HardwareAddr: stdnet.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}}
	_, ipNet, err := stdnet.ParseCIDR("10.2.3.4/24")
	if err != nil {
		t.Fatal(err)
	}
	ipNet.IP = stdnet.ParseIP("10.2.3.4")
	opts := []InterfaceOption{
		WithInterfaceByNameFunc(func(name string) (*stdnet.Interface, error) {
			if name != "eth-test" {
				return nil, errors.New("unexpected interface")
			}
			return &iface, nil
		}),
		WithInterfacesFunc(func() ([]stdnet.Interface, error) { return []stdnet.Interface{iface}, nil }),
		WithInterfaceAddrsFunc(func(got stdnet.Interface) ([]stdnet.Addr, error) {
			if got.Name != iface.Name {
				return nil, errors.New("unexpected addrs interface")
			}
			return []stdnet.Addr{ipNet}, nil
		}),
		WithReverseLookupFunc(func(addr string) ([]string, error) {
			if addr != "10.2.3.4" {
				return nil, errors.New("unexpected reverse lookup")
			}
			return []string{"host.example."}, nil
		}),
		WithNetHostnameFunc(func() (string, error) { return "fallback-host", nil }),
	}

	gotIface, err := GetNetworkInterfaceWithOptions("eth-test", opts...)
	if err != nil || gotIface.Name != "eth-test" {
		t.Fatalf("GetNetworkInterfaceWithOptions = %#v %v", gotIface, err)
	}
	if ifaces, err := GetNetworkInterfacesWithOptions(opts...); err != nil || len(ifaces) != 1 || ifaces[0].Name != "eth-test" {
		t.Fatalf("GetNetworkInterfacesWithOptions = %#v %v", ifaces, err)
	}
	if got := LocalIPv4sWithOptions(opts...); !reflect.DeepEqual(got, []string{"10.2.3.4"}) {
		t.Fatalf("LocalIPv4sWithOptions = %#v", got)
	}
	if got := GetLocalhostStrWithOptions(opts...); got != "10.2.3.4" {
		t.Fatalf("GetLocalhostStrWithOptions = %q", got)
	}
	if got := GetLocalHostNameWithOptions(opts...); got != "host.example" {
		t.Fatalf("GetLocalHostNameWithOptions = %q", got)
	}
	if got := GetHardwareAddressWithOptions(stdnet.ParseIP("10.2.3.4"), opts...); !reflect.DeepEqual(got, iface.HardwareAddr) {
		t.Fatalf("GetHardwareAddressWithOptions = %v", got)
	}
	if got := GetLocalHardwareAddressWithOptions(opts...); !reflect.DeepEqual(got, iface.HardwareAddr) {
		t.Fatalf("GetLocalHardwareAddressWithOptions = %v", got)
	}
	if got := GetMACAddressWithOptions(stdnet.ParseIP("10.2.3.4"), opts, "-"); got != "aa-bb-cc-dd-ee-ff" {
		t.Fatalf("GetMACAddressWithOptions = %q", got)
	}
	if got := GetLocalMACAddressWithOptions(opts, "-"); got != "aa-bb-cc-dd-ee-ff" {
		t.Fatalf("GetLocalMACAddressWithOptions = %q", got)
	}
	if got, err := GetNetworkInterface("eth-test"); err == nil || got != nil {
		t.Fatalf("GetNetworkInterface real lookup unexpectedly returned %#v, %v", got, err)
	}
}

func TestInterfaceWrapperAndFallbackHelpers(t *testing.T) {
	iface := stdnet.Interface{Name: "lo-test", Flags: stdnet.FlagLoopback}
	_, ipNet, err := stdnet.ParseCIDR("2001:db8::1/64")
	if err != nil {
		t.Fatal(err)
	}
	ipNet.IP = stdnet.ParseIP("2001:db8::1")
	opts := []InterfaceOption{
		WithInterfacesFunc(func() ([]stdnet.Interface, error) { return []stdnet.Interface{iface}, nil }),
		WithInterfaceAddrsFunc(func(stdnet.Interface) ([]stdnet.Addr, error) {
			return []stdnet.Addr{ipNet, &stdnet.IPAddr{IP: stdnet.ParseIP("127.0.0.1")}, fakeAddr("bad")}, nil
		}),
		WithReverseLookupFunc(func(string) ([]string, error) { return nil, errors.New("no ptr") }),
		WithNetHostnameFunc(func() (string, error) { return "fallback-host", nil }),
	}
	if got := LocalIPv6sWithOptions(opts...); !reflect.DeepEqual(got, []string{"2001:db8::1"}) {
		t.Fatalf("LocalIPv6sWithOptions = %#v", got)
	}
	if got := LocalIPsWithOptions(opts...); !reflect.DeepEqual(got, []string{"2001:db8::1", "127.0.0.1"}) {
		t.Fatalf("LocalIPsWithOptions = %#v", got)
	}
	if got := LocalAddressListByInterfaceWithOptions(func(stdnet.Interface) bool { return false }, nil, opts...); len(got) != 0 {
		t.Fatalf("filtered LocalAddressListByInterfaceWithOptions = %#v", got)
	}
	if got := GetLocalhostStrWithOptions(opts...); got != LocalIP {
		t.Fatalf("GetLocalhostStrWithOptions fallback = %q", got)
	}
	if got := GetLocalhostWithOptions(opts...); !got.Equal(stdnet.ParseIP(LocalIP)) {
		t.Fatalf("GetLocalhostWithOptions = %v", got)
	}
	if got := GetLocalHostNameWithOptions(opts...); got != "fallback-host" {
		t.Fatalf("GetLocalHostNameWithOptions fallback = %q", got)
	}
	if got := GetLocalMACAddressWithOptions(opts); got != "" {
		t.Fatalf("GetLocalMACAddressWithOptions loopback = %q", got)
	}
	if got := GetMACAddressWithOptions(stdnet.ParseIP("192.0.2.1"), opts); got != "" {
		t.Fatalf("GetMACAddressWithOptions missing = %q", got)
	}
	if got := GetHardwareAddress(stdnet.ParseIP("192.0.2.1")); got != nil {
		t.Fatalf("GetHardwareAddress real lookup unexpectedly returned %v", got)
	}
}

func TestInterfaceProviderErrors(t *testing.T) {
	opts := []InterfaceOption{WithInterfacesFunc(func() ([]stdnet.Interface, error) { return nil, errors.New("interfaces") })}
	if got := LocalAddressListWithOptions(nil, opts...); got != nil {
		t.Fatalf("LocalAddressListWithOptions error = %#v", got)
	}
	if got := GetHardwareAddressWithOptions(stdnet.ParseIP("127.0.0.1"), opts...); got != nil {
		t.Fatalf("GetHardwareAddressWithOptions interfaces error = %v", got)
	}
	if got := GetLocalHardwareAddressWithOptions(opts...); got != nil {
		t.Fatalf("GetLocalHardwareAddressWithOptions interfaces error = %v", got)
	}
	iface := stdnet.Interface{Name: "broken"}
	opts = []InterfaceOption{
		WithInterfacesFunc(func() ([]stdnet.Interface, error) { return []stdnet.Interface{iface}, nil }),
		WithInterfaceAddrsFunc(func(stdnet.Interface) ([]stdnet.Addr, error) { return nil, errors.New("addrs") }),
	}
	if got := LocalAddressListWithOptions(nil, opts...); len(got) != 0 {
		t.Fatalf("LocalAddressListWithOptions addrs error = %#v", got)
	}
}

type fakeAddr string

func (a fakeAddr) Network() string { return "fake" }
func (a fakeAddr) String() string  { return string(a) }
