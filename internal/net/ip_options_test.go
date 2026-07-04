package net

import (
	"errors"
	"math/big"
	stdnet "net"
	"reflect"
	"regexp"
	"strconv"
	"testing"
)

func TestIPOptionsUseCustomParsers(t *testing.T) {
	parseIPCalls := 0
	parseIP := func(s string) stdnet.IP {
		parseIPCalls++
		switch s {
		case "aliasv4":
			return stdnet.ParseIP("10.1.2.3")
		case "aliasv6":
			return stdnet.ParseIP("::2")
		case "aliasmask":
			return stdnet.ParseIP("255.255.255.0")
		default:
			return stdnet.ParseIP(s)
		}
	}
	if got, err := IPv4ToLongWithOptions("aliasv4", WithIPParser(parseIP)); err != nil || got != 167838211 {
		t.Fatalf("IPv4ToLongWithOptions = %d %v", got, err)
	}
	if got := IPv4ToLongDefaultWithOptions("bad", 42, WithIPParser(func(string) stdnet.IP { return nil })); got != 42 {
		t.Fatalf("IPv4ToLongDefaultWithOptions = %d", got)
	}
	if got, err := IPv6ToBigIntWithOptions("aliasv6", WithIPParser(parseIP)); err != nil || got.Cmp(big.NewInt(2)) != 0 {
		t.Fatalf("IPv6ToBigIntWithOptions = %v %v", got, err)
	}
	if !IsIPWithOptions("aliasv4", WithIPParser(parseIP)) || !IsIPv4WithOptions("aliasv4", WithIPParser(parseIP)) || !IsIPv6WithOptions("aliasv6", WithIPParser(parseIP)) {
		t.Fatal("IP validators should use custom parser")
	}
	if bit, err := MaskBitByMaskWithOptions("aliasmask", WithIPParser(parseIP)); err != nil || bit != 24 {
		t.Fatalf("MaskBitByMaskWithOptions = %d %v", bit, err)
	}
	if block, err := FormatIPBlockWithOptions("10.1.2.3", "aliasmask", WithIPParser(parseIP)); err != nil || block != "10.1.2.3/24" {
		t.Fatalf("FormatIPBlockWithOptions = %q %v", block, err)
	}
	if ips, err := ListIPsWithOptions("aliasv4/30", false, WithIPParser(parseIP), WithIPIntParser(strconv.Atoi)); err != nil || !reflect.DeepEqual(ips, []string{"10.1.2.1", "10.1.2.2"}) {
		t.Fatalf("ListIPsWithOptions = %#v %v", ips, err)
	}
	if parseIPCalls == 0 {
		t.Fatal("custom IP parser was not called")
	}
}

func TestIPRangeOptionsUseCustomParsers(t *testing.T) {
	parseCIDRCalls := 0
	_, network, err := stdnet.ParseCIDR("192.0.2.0/24")
	if err != nil {
		t.Fatal(err)
	}
	parseCIDR := func(s string) (stdnet.IP, *stdnet.IPNet, error) {
		parseCIDRCalls++
		if s != "alias-cidr" {
			return nil, nil, errors.New("unexpected cidr")
		}
		return stdnet.ParseIP("192.0.2.1"), network, nil
	}
	if !IsInRangeWithOptions("alias-ip", "alias-cidr", WithIPParser(func(s string) stdnet.IP {
		if s == "alias-ip" {
			return stdnet.ParseIP("192.0.2.5")
		}
		return nil
	}), WithCIDRParser(parseCIDR)) {
		t.Fatal("IsInRangeWithOptions should use custom parsers")
	}
	if parseCIDRCalls != 1 {
		t.Fatalf("parseCIDR calls = %d", parseCIDRCalls)
	}

	wildcardParseIntCalls := 0
	if !MatchesWildcardWithOptions("10.x.*.2", "alias", WithWildcardIPParser(func(s string) stdnet.IP {
		if s == "alias" {
			return stdnet.ParseIP("10.9.1.2")
		}
		return nil
	}), WithWildcardIntParser(func(s string) (int, error) {
		wildcardParseIntCalls++
		if s == "x" {
			return 9, nil
		}
		return strconv.Atoi(s)
	})) {
		t.Fatal("MatchesWildcardWithOptions should use custom parsers")
	}
	if wildcardParseIntCalls == 0 {
		t.Fatal("custom wildcard int parser was not called")
	}
}

func TestNilIPProviderOptionsDoNotOverwriteConfiguredProviders(t *testing.T) {
	_, network, err := stdnet.ParseCIDR("192.0.2.0/24")
	if err != nil {
		t.Fatal(err)
	}
	parseCIDR := func(string) (stdnet.IP, *stdnet.IPNet, error) {
		return stdnet.ParseIP("192.0.2.1"), network, nil
	}

	wildcardCfg := applyWildcardOptions([]WildcardOption{
		WithWildcardCompileFunc(regexp.Compile),
		WithWildcardCompileFunc(nil),
		WithWildcardIPParser(stdnet.ParseIP),
		WithWildcardIPParser(nil),
		WithWildcardIntParser(strconv.Atoi),
		WithWildcardIntParser(nil),
	})
	if wildcardCfg.compile == nil || wildcardCfg.parseIP == nil || wildcardCfg.parseInt == nil {
		t.Fatalf("nil wildcard provider option overwrote configured provider: %#v", wildcardCfg)
	}

	ipCfg := applyIPOptions([]IPOption{
		WithIPParser(stdnet.ParseIP),
		WithIPParser(nil),
		WithCIDRParser(parseCIDR),
		WithCIDRParser(nil),
		WithIPIntParser(strconv.Atoi),
		WithIPIntParser(nil),
	})
	if ipCfg.parseIP == nil || ipCfg.parseCIDR == nil || ipCfg.parseInt == nil {
		t.Fatalf("nil IP provider option overwrote configured provider: %#v", ipCfg)
	}
}
