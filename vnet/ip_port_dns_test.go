package vnet_test

import (
	"context"
	"math/big"
	stdnet "net"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vnet"
)

func TestVNetFacade(t *testing.T) {
	v, err := vnet.IPv4ToLong("127.0.0.1")
	if err != nil || vnet.LongToIPv4(v) != "127.0.0.1" {
		t.Fatalf("IPv4 facade failed: %d %v", v, err)
	}
	if !vnet.IsIPv4("192.168.1.1") || !vnet.IsIPv6("::1") || !vnet.IsInnerIP("10.0.0.1") {
		t.Fatal("IP validators failed")
	}
	if !vnet.IsValidPort(80) || vnet.HideIPPart("192.168.1.2") != "192.168.1.*" {
		t.Fatal("port or hide helper failed")
	}
	if vnet.CreateTLSConfig() == nil || vnet.NewUploadSetting().MemoryThreshold == 0 {
		t.Fatal("TLS/upload helpers failed")
	}
}

func TestVNetIPFacadeWrappers(t *testing.T) {
	parseIP := func(s string) stdnet.IP {
		if s == "alias" {
			return stdnet.ParseIP("192.0.2.1")
		}
		return stdnet.ParseIP(s)
	}
	parseCIDR := func(s string) (stdnet.IP, *stdnet.IPNet, error) {
		if s == "alias/24" {
			return stdnet.ParseCIDR("192.0.2.0/24")
		}
		return stdnet.ParseCIDR(s)
	}

	if got, err := vnet.IPv4ToLongWithOptions("alias", vnet.WithIPParser(parseIP)); err != nil || got != 3221225985 {
		t.Fatalf("IPv4ToLongWithOptions = %d, %v", got, err)
	}
	if got := vnet.IPv4ToLongDefaultWithOptions("bad", 42, vnet.WithIPParser(func(string) stdnet.IP { return nil })); got != 42 {
		t.Fatalf("IPv4ToLongDefaultWithOptions = %d", got)
	}
	if got, err := vnet.IPv6ToBigIntWithOptions("::1", vnet.WithIPParser(parseIP)); err != nil || got.Sign() == 0 {
		t.Fatalf("IPv6ToBigIntWithOptions = %v, %v", got, err)
	}
	if got, err := vnet.BigIntToIPv6(big.NewInt(1)); err != nil || got != "::1" {
		t.Fatalf("BigIntToIPv6 = %q, %v", got, err)
	}
	if !vnet.IsIPWithOptions("alias", vnet.WithIPParser(parseIP)) || !vnet.IsIPv4WithOptions("alias", vnet.WithIPParser(parseIP)) {
		t.Fatal("IP validators did not use parser option")
	}
	if vnet.IsIPv6WithOptions("alias", vnet.WithIPParser(parseIP)) {
		t.Fatal("IsIPv6WithOptions should reject IPv4 alias")
	}
	if !vnet.IsInnerIPWithOptions("10.0.0.1", vnet.WithIPParser(parseIP)) {
		t.Fatal("IsInnerIPWithOptions should accept private IPv4")
	}
	if got, err := vnet.FormatIPBlockWithOptions("alias", "255.255.255.0", vnet.WithIPParser(parseIP)); err != nil || got != "alias/24" {
		t.Fatalf("FormatIPBlockWithOptions = %q, %v", got, err)
	}
	if got, err := vnet.BeginIPWithOptions("alias", 24, vnet.WithIPParser(parseIP)); err != nil || got != "192.0.2.0" {
		t.Fatalf("BeginIPWithOptions = %q, %v", got, err)
	}
	if got, err := vnet.EndIPWithOptions("alias", 24, vnet.WithIPParser(parseIP)); err != nil || got != "192.0.2.255" {
		t.Fatalf("EndIPWithOptions = %q, %v", got, err)
	}
	if got, err := vnet.MaskBitByMaskWithOptions("255.255.255.0", vnet.WithIPParser(parseIP)); err != nil || got != 24 {
		t.Fatalf("MaskBitByMaskWithOptions = %d, %v", got, err)
	}
	if got, err := vnet.MaskByIPRangeWithOptions("192.0.2.0", "192.0.2.255", vnet.WithIPParser(parseIP)); err != nil || got != "255.255.255.0" {
		t.Fatalf("MaskByIPRangeWithOptions = %q, %v", got, err)
	}
	if got, err := vnet.CountByIPRangeWithOptions("192.0.2.0", "192.0.2.3", vnet.WithIPParser(parseIP)); err != nil || got != 4 {
		t.Fatalf("CountByIPRangeWithOptions = %d, %v", got, err)
	}
	if !vnet.IsMaskValidWithOptions("255.255.255.0", vnet.WithIPParser(parseIP)) || !vnet.IsMaskBitValid(24) {
		t.Fatal("mask validators failed")
	}
	if got, err := vnet.ListIPsWithOptions("192.0.2.1-192.0.2.2", true, vnet.WithIPParser(parseIP)); err != nil || !reflect.DeepEqual(got, []string{"192.0.2.1", "192.0.2.2"}) {
		t.Fatalf("ListIPsWithOptions = %#v, %v", got, err)
	}
	if got, err := vnet.ListIPCIDRWithOptions("alias", 30, true, vnet.WithIPParser(parseIP)); err != nil || len(got) != 4 {
		t.Fatalf("ListIPCIDRWithOptions = %#v, %v", got, err)
	}
	if got, err := vnet.ListIPRangeWithOptions("192.0.2.1", "192.0.2.2", vnet.WithIPParser(parseIP)); err != nil || !reflect.DeepEqual(got, []string{"192.0.2.1", "192.0.2.2"}) {
		t.Fatalf("ListIPRangeWithOptions = %#v, %v", got, err)
	}
	if !vnet.IsInRangeWithOptions("alias", "alias/24", vnet.WithIPParser(parseIP), vnet.WithCIDRParser(parseCIDR)) {
		t.Fatal("IsInRangeWithOptions should use custom parsers")
	}
	if got := vnet.HideIPPartLong(0xC0000201); got != "192.0.2.*" {
		t.Fatalf("HideIPPartLong = %q", got)
	}
}

func TestVNetWildcardFacadeWrappers(t *testing.T) {
	if !vnet.MatchesWildcardWithOptions("10.0.*.1", "10.0.2.1") {
		t.Fatal("MatchesWildcardWithOptions should match wildcard segment")
	}
	calledParser := false
	matched := vnet.MatchesWildcardWithOptions("10.0.*.1", "alias",
		vnet.WithWildcardIPParser(func(s string) stdnet.IP {
			calledParser = true
			if s == "alias" {
				return stdnet.ParseIP("10.0.2.1")
			}
			return stdnet.ParseIP(s)
		}),
	)
	if !matched || !calledParser {
		t.Fatalf("MatchesWildcardWithOptions matched=%v calledParser=%v", matched, calledParser)
	}
}

func TestVNetDNSFacadeWrappers(t *testing.T) {
	ascii, err := vnet.IDNToASCII("bücher.example")
	if err != nil || ascii != "xn--bcher-kva.example" {
		t.Fatalf("IDNToASCII = %q, %v", ascii, err)
	}
	if got := vnet.GetMultistageReverseProxyIP("unknown, , 203.0.113.10"); got != "203.0.113.10" {
		t.Fatalf("GetMultistageReverseProxyIP = %q", got)
	}
	if !vnet.IsUnknown(" UNKNOWN ") || vnet.IsUnknown("203.0.113.10") {
		t.Fatal("IsUnknown returned unexpected result")
	}
	cookies := vnet.ParseCookies("sid=abc; theme=dark")
	if len(cookies) != 2 || cookies[0].Name != "sid" || cookies[0].Value != "abc" {
		t.Fatalf("ParseCookies = %#v", cookies)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := vnet.GetIPByHostWithOptions("example.com", vnet.WithResolveContext(ctx), vnet.WithResolver(stdnet.DefaultResolver)); err == nil {
		t.Fatal("GetIPByHostWithOptions should return canceled context error")
	}
}

func TestVNetFacadeOptions(t *testing.T) {
	ln, err := stdnet.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen local port: %v", err)
	}
	defer func() { _ = ln.Close() }()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err == nil {
			_ = conn.Close()
		}
	}()

	_, portStr, err := stdnet.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("split listener address: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse listener port: %v", err)
	}
	if !vnet.PingWithOptions("127.0.0.1", vnet.WithPingPorts(port), vnet.WithPingTimeout(time.Second), vnet.WithPingNetwork("tcp")) {
		t.Fatal("PingWithOptions should reach local listener")
	}
	<-done

	if vnet.IsUsableLocalPortWithOptions(port, vnet.WithPortHost("127.0.0.1")) {
		t.Fatal("IsUsableLocalPortWithOptions should reject occupied port")
	}
	g := vnet.NewLocalPortGeneratorWithOptions(port, vnet.WithPortHost("127.0.0.1"))
	generated, err := g.Gen()
	if err != nil {
		t.Fatalf("LocalPortGenerator.Gen with options: %v", err)
	}
	if generated <= port || generated > vnet.PortRangeMax {
		t.Fatalf("LocalPortGenerator generated %d, want > %d", generated, port)
	}
	freePort, err := vnet.GetUsableLocalPortInRangeWithOptions(port+1, port+20, vnet.WithPortHost("127.0.0.1"))
	if err != nil || freePort < port+1 || freePort > port+20 {
		t.Fatalf("GetUsableLocalPortInRangeWithOptions = %d, %v", freePort, err)
	}
	ports, err := vnet.GetUsableLocalPortsWithOptions(1, port+1, port+20, vnet.WithPortHost("127.0.0.1"))
	if err != nil || len(ports) != 1 {
		t.Fatalf("GetUsableLocalPortsWithOptions = %v, %v; want one port", ports, err)
	}
	ips, err := vnet.GetIPByHostWithOptions("localhost", vnet.WithResolveNetwork("ip4"), vnet.WithResolveTimeout(time.Second))
	if err != nil || len(ips) == 0 {
		t.Fatalf("GetIPByHostWithOptions = %v, %v; want at least one IPv4", ips, err)
	}
	dns, err := vnet.GetDNSInfoWithOptions("localhost", vnet.WithDNSTypes("A"), vnet.WithResolveTimeout(time.Second))
	if err != nil || len(dns) == 0 {
		t.Fatalf("GetDNSInfoWithOptions = %v, %v; want at least one A record", dns, err)
	}
}
