package system

import (
	"errors"
	"net"
	"strings"
	"testing"
)

func TestHostInfo(t *testing.T) {
	h := GetHostInfo()
	if h == nil {
		t.Fatal("HostInfo 不应为 nil")
	}
	if h.GetName() == "" {
		t.Errorf("Host Name 不应为空")
	}
	if !strings.Contains(h.String(), "Host Name:") {
		t.Errorf("HostInfo.String 缺少 caption: %s", h.String())
	}
}

func TestHostInfoWithOptions(t *testing.T) {
	_, ipNet, err := net.ParseCIDR("10.0.0.2/24")
	if err != nil {
		t.Fatal(err)
	}
	ipNet.IP = net.ParseIP("10.0.0.2")
	h := NewHostInfoWithOptions(
		WithHostNameFunc(func() (string, error) { return "option-host", nil }),
		WithHostInterfaceAddrsFunc(func() ([]net.Addr, error) { return []net.Addr{ipNet}, nil }),
	)
	if h.GetName() != "option-host" || h.GetAddress() != "10.0.0.2" {
		t.Fatalf("NewHostInfoWithOptions = %#v", h)
	}

	h = GetHostInfoWithOptions(WithHostAddressFunc(func() string { return "192.0.2.10" }))
	if h.GetAddress() != "192.0.2.10" {
		t.Fatalf("GetHostInfoWithOptions address = %q", h.GetAddress())
	}
}

type systemTestAddr string

func (a systemTestAddr) Network() string { return "system-test" }
func (a systemTestAddr) String() string  { return string(a) }

func TestHostInfoErrorAndAddressBoundaries(t *testing.T) {
	h := NewHostInfoWithOptions(
		WithHostNameFunc(func() (string, error) { return "", errors.New("hostname unavailable") }),
		WithHostInterfaceAddrsFunc(func() ([]net.Addr, error) { return nil, errors.New("interfaces unavailable") }),
	)
	if h.GetName() != "" || h.GetAddress() != "" {
		t.Fatalf("host error fallback = %#v", h)
	}

	_, loopback, err := net.ParseCIDR("127.0.0.1/8")
	if err != nil {
		t.Fatal(err)
	}
	_, ipv6, err := net.ParseCIDR("2001:db8::1/64")
	if err != nil {
		t.Fatal(err)
	}
	_, ipv4, err := net.ParseCIDR("192.0.2.42/24")
	if err != nil {
		t.Fatal(err)
	}
	loopback.IP = net.ParseIP("127.0.0.1")
	ipv6.IP = net.ParseIP("2001:db8::1")
	ipv4.IP = net.ParseIP("192.0.2.42")
	got := firstNonLoopbackIPv4(func() ([]net.Addr, error) {
		return []net.Addr{systemTestAddr("not-ipnet"), loopback, ipv6, ipv4}, nil
	})
	if got != "192.0.2.42" {
		t.Fatalf("firstNonLoopbackIPv4 = %q", got)
	}
}

func TestHostInfoNilOptionsFallBackToDefaults(t *testing.T) {
	h := NewHostInfoWithOptions(nil, WithHostNameFunc(nil), WithHostInterfaceAddrsFunc(nil), WithHostAddressFunc(func() string { return "203.0.113.7" }))
	if h.GetName() == "" || h.GetAddress() != "203.0.113.7" {
		t.Fatalf("nil option host fallback = %#v", h)
	}
}
