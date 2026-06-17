package net

import (
	"errors"
	stdnet "net"
	"reflect"
	"strconv"
	"testing"
)

type stubListener struct{}

func (stubListener) Accept() (stdnet.Conn, error) { return nil, errors.New("stub listener") }
func (stubListener) Close() error                 { return nil }
func (stubListener) Addr() stdnet.Addr {
	return &stdnet.TCPAddr{IP: stdnet.ParseIP("127.0.0.1"), Port: 12345}
}

func TestPortAndMiscHelpers(t *testing.T) {
	if !IsValidPort(65535) || IsValidPort(70000) {
		t.Fatal("IsValidPort failed")
	}
	ln, err := stdnet.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen local port: %v", err)
	}
	defer func() { _ = ln.Close() }()
	_, portStr, err := stdnet.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("split listener address: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse listener port: %v", err)
	}
	if IsUsableLocalPortWithOptions(port, WithPortHost("127.0.0.1")) {
		t.Fatal("IsUsableLocalPortWithOptions should reject an occupied port")
	}
	g := NewLocalPortGeneratorWithOptions(port, WithPortHost("127.0.0.1"))
	generated, err := g.Gen()
	if err != nil {
		t.Fatalf("LocalPortGenerator.Gen with options: %v", err)
	}
	if generated <= port || generated > PortRangeMax {
		t.Fatalf("LocalPortGenerator generated %d, want > %d", generated, port)
	}
	next, err := g.GenWithOptions(WithPortHost("127.0.0.1"))
	if err != nil {
		t.Fatalf("LocalPortGenerator.GenWithOptions: %v", err)
	}
	if next <= generated || next > PortRangeMax {
		t.Fatalf("LocalPortGenerator next generated %d, want > %d", next, generated)
	}
	if HideIPPart("192.168.1.2") != "192.168.1.*" {
		t.Fatal("HideIPPart failed")
	}
	if got := GetMultistageReverseProxyIP("unknown, 10.0.0.1, 8.8.8.8"); got != "10.0.0.1" {
		t.Fatalf("GetMultistageReverseProxyIP = %q", got)
	}
	if IsUnknown("10.0.0.1") || !IsUnknown("unknown") {
		t.Fatal("IsUnknown failed")
	}
	if ascii, err := IDNToASCII("中国.cn"); err != nil || ascii == "" {
		t.Fatalf("IDNToASCII = %q %v", ascii, err)
	}
	if len(ParseCookies("a=1; b=2")) != 2 {
		t.Fatal("ParseCookies failed")
	}
}

func TestPortOptionsUseListenerFactory(t *testing.T) {
	var network, address string
	factory := func(n, a string) (stdnet.Listener, error) {
		network, address = n, a
		return stubListener{}, nil
	}
	if !IsUsableLocalPortWithOptions(12345, WithPortNetwork("tcp4"), WithPortHost("127.0.0.2"), WithPortListenerFactory(factory)) {
		t.Fatal("IsUsableLocalPortWithOptions should use successful listener factory")
	}
	if network != "tcp4" || address != "127.0.0.2:12345" {
		t.Fatalf("listener target = %s %s", network, address)
	}

	factory = func(n, a string) (stdnet.Listener, error) {
		network, address = n, a
		return nil, errors.New("bind failed")
	}
	if IsUsableLocalPortWithOptions(12345, WithPortListenerFactory(factory)) {
		t.Fatal("IsUsableLocalPortWithOptions should reject listener factory errors")
	}
}

func TestPortRangeWrappersWithListenerFactory(t *testing.T) {
	usable := map[int]bool{1002: true, 1004: true, 1005: true, 1024: true}
	seen := make([]int, 0)
	factory := func(_ string, address string) (stdnet.Listener, error) {
		_, portText, err := stdnet.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		port, err := strconv.Atoi(portText)
		if err != nil {
			return nil, err
		}
		seen = append(seen, port)
		if !usable[port] {
			return nil, errors.New("occupied")
		}
		return stubListener{}, nil
	}
	opts := []PortOption{WithPortListenerFactory(factory), WithPortHost(""), WithPortNetwork("")}

	if got, err := GetUsableLocalPortInRangeWithOptions(1000, 1003, opts...); err != nil || got != 1002 {
		t.Fatalf("GetUsableLocalPortInRangeWithOptions = %d, %v", got, err)
	}
	if got, err := GetUsableLocalPortFromWithOptions(1004, opts...); err != nil || got != 1004 {
		t.Fatalf("GetUsableLocalPortFromWithOptions = %d, %v", got, err)
	}
	if got, err := GetUsableLocalPortWithOptions(opts...); err != nil || got != 1024 {
		t.Fatalf("GetUsableLocalPortWithOptions = %d, %v", got, err)
	}
	if got, err := GetUsableLocalPortsWithOptions(2, 1000, 1005, opts...); err != nil || !reflect.DeepEqual(got, []int{1002, 1004}) {
		t.Fatalf("GetUsableLocalPortsWithOptions = %#v, %v", got, err)
	}
	if got, err := GetUsableLocalPortsWithOptions(0, 1000, 1005, opts...); err != nil || got != nil {
		t.Fatalf("GetUsableLocalPortsWithOptions zero = %#v, %v", got, err)
	}
	if _, err := GetUsableLocalPortInRangeWithOptions(1005, 1000, opts...); err == nil {
		t.Fatalf("GetUsableLocalPortInRangeWithOptions should reject invalid range")
	}
	if got, err := GetUsableLocalPortsWithOptions(2, 1000, 1003, opts...); err == nil || !reflect.DeepEqual(got, []int{1002}) {
		t.Fatalf("partial usable ports = %#v, %v", got, err)
	}
	if len(seen) == 0 {
		t.Fatalf("listener factory was not used")
	}
}

func TestLocalPortGeneratorBoundaries(t *testing.T) {
	if _, err := (*LocalPortGenerator)(nil).Gen(); err == nil {
		t.Fatalf("nil generator should return error")
	}
	factory := func(_ string, _ string) (stdnet.Listener, error) { return stubListener{}, nil }
	g := NewLocalPortGenerator(1234)
	port, err := g.GenWithOptions(WithPortListenerFactory(factory))
	if err != nil || port != 1234 {
		t.Fatalf("NewLocalPortGenerator.GenWithOptions = %d, %v", port, err)
	}
	if got := HideIPPart("localhost"); got != "localhost" {
		t.Fatalf("HideIPPart(localhost) = %q", got)
	}
	if got := HideIPPartLong(0x7f000001); got != "127.0.0.*" {
		t.Fatalf("HideIPPartLong = %q", got)
	}
}
