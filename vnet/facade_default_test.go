package vnet_test

import (
	"context"
	"math/big"
	stdnet "net"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vnet"
)

func TestVNetDefaultDelegates(t *testing.T) {
	// Default wrappers for address functions.
	_ = vnet.GetLocalHostName()
	_ = vnet.GetLocalHostNameWithOptions()

	_ = vnet.GetIPByHost("localhost")

	mac := vnet.GetLocalMACAddress()
	if mac == "" {
		t.Log("GetLocalMACAddress returned empty (no network)")
	}

	_ = vnet.GetLocalMACAddressWithOptions(nil)

	// Default wrappers that delegate to WithOptions variants.
	_, _ = vnet.GetNetworkInterface("lo0")
	_, _ = vnet.GetNetworkInterfaces() // may fail, just test call
	_ = vnet.LocalIPv4s()
	_ = vnet.LocalIPv6s()
	_ = vnet.LocalIPs()
	_ = vnet.LocalAddressList(nil)
	_ = vnet.LocalAddressListByInterface(nil, nil)
	_ = vnet.GetLocalhostStr()
	_ = vnet.GetLocalhost()
	_ = vnet.GetLocalMACAddress()
	_ = vnet.GetMACAddress(stdnet.ParseIP("127.0.0.1"))
	_ = vnet.GetHardwareAddress(stdnet.ParseIP("127.0.0.1"))
	_ = vnet.GetLocalHardwareAddress()
}

func TestVNetIPDefaultDelegates(t *testing.T) {
	// Simple option constructors.
	_ = vnet.WithWildcardCompileFunc(nil)
	_ = vnet.WithWildcardIntParser(nil)
	_ = vnet.WithIPIntParser(nil)

	// Default IP wrappers.
	_ = vnet.IPv4ToLongDefault("bad", 42)
	v, err := vnet.IPv6ToBigInt("::1")
	if err != nil || v == nil || v.Sign() == 0 {
		t.Fatalf("IPv6ToBigInt = %v, %v", v, err)
	}
	if !vnet.IsIP("127.0.0.1") || !vnet.IsIPv4("192.168.1.1") || !vnet.IsIPv6("::1") {
		t.Fatal("IsIP/IsIPv4/IsIPv6 default delegates failed")
	}
	if got, err := vnet.FormatIPBlock("192.168.1.0", "255.255.255.0"); err != nil || got != "192.168.1.0/24" {
		t.Fatalf("FormatIPBlock = %q, %v", got, err)
	}
	if got, err := vnet.BeginIP("192.168.1.0", 24); err != nil || got != "192.168.1.0" {
		t.Fatalf("BeginIP = %q, %v", got, err)
	}
	if got, err := vnet.BeginIPLong("192.168.1.0", 24); err != nil || got == 0 {
		t.Fatalf("BeginIPLong = %d, %v", got, err)
	}
	if got, err := vnet.BeginIPLongWithOptions("192.168.1.0", 24); err != nil || got == 0 {
		t.Fatalf("BeginIPLongWithOptions = %d, %v", got, err)
	}
	if got, err := vnet.EndIP("192.168.1.0", 24); err != nil || got != "192.168.1.255" {
		t.Fatalf("EndIP = %q, %v", got, err)
	}
	if got, err := vnet.EndIPLong("192.168.1.0", 24); err != nil || got == 0 {
		t.Fatalf("EndIPLong = %d, %v", got, err)
	}
	if got, err := vnet.EndIPLongWithOptions("192.168.1.0", 24); err != nil || got == 0 {
		t.Fatalf("EndIPLongWithOptions = %d, %v", got, err)
	}
	if got, err := vnet.MaskBitByMask("255.255.255.0"); err != nil || got != 24 {
		t.Fatalf("MaskBitByMask = %d, %v", got, err)
	}
	if _, err := vnet.CountByMaskBit(24, false); err != nil {
		t.Fatalf("CountByMaskBit = %v", err)
	}
	if got, err := vnet.MaskByMaskBit(24); err != nil || got != "255.255.255.0" {
		t.Fatalf("MaskByMaskBit = %q, %v", got, err)
	}
	if got, err := vnet.MaskByIPRange("192.168.1.0", "192.168.1.255"); err != nil || got != "255.255.255.0" {
		t.Fatalf("MaskByIPRange = %q, %v", got, err)
	}
	if got, err := vnet.CountByIPRange("192.168.1.0", "192.168.1.3"); err != nil || got != 4 {
		t.Fatalf("CountByIPRange = %d, %v", got, err)
	}
	if !vnet.IsMaskValid("255.255.255.0") || !vnet.IsMaskBitValid(24) {
		t.Fatal("IsMaskValid/IsMaskBitValid failed")
	}
	if got, err := vnet.ListIPs("192.168.1.0-192.168.1.1", true); err != nil || len(got) != 2 {
		t.Fatalf("ListIPs = %v, %v", got, err)
	}
	if got, err := vnet.ListIPCIDR("192.168.1.0", 30, true); err != nil || len(got) != 4 {
		t.Fatalf("ListIPCIDR = %v, %v", got, err)
	}
	if got, err := vnet.ListIPRange("192.168.1.0", "192.168.1.1"); err != nil || len(got) != 2 {
		t.Fatalf("ListIPRange = %v, %v", got, err)
	}
	if !vnet.MatchesWildcard("192.168.*.1", "192.168.0.1") {
		t.Fatal("MatchesWildcard failed")
	}
	if !vnet.IsInRange("192.168.1.10", "192.168.1.0/24") {
		t.Fatal("IsInRange failed")
	}
	_ = vnet.HideIPPart("192.168.1.2")
}

func TestVNetPortDefaultDelegates(t *testing.T) {
	if !vnet.IsValidPort(80) || vnet.IsValidPort(-1) || vnet.IsValidPort(99999) {
		t.Fatal("IsValidPort failed")
	}
	if vnet.IsUsableLocalPort(99999) {
		t.Log("IsUsableLocalPort(99999) unexpectedly returned true")
	}
	if _, err := vnet.GetUsableLocalPort(); err != nil {
		t.Logf("GetUsableLocalPort: %v (expected with no free port mock)", err)
	}
	if _, err := vnet.GetUsableLocalPortFrom(10000); err != nil {
		t.Logf("GetUsableLocalPortFrom: %v", err)
	}
	if _, err := vnet.GetUsableLocalPortInRange(10000, 10010); err != nil {
		t.Logf("GetUsableLocalPortInRange: %v", err)
	}
	if _, err := vnet.GetUsableLocalPorts(1, 10000, 10010); err != nil {
		t.Logf("GetUsableLocalPorts: %v", err)
	}
	_ = vnet.NewLocalPortGenerator(10000)
}

func TestVNetDialDefaultDelegates(t *testing.T) {
	// Default delegates for dial wrappers — they attempt real connections
	// so we just verify they don't panic and return expected error type.
	if err := vnet.NetCat("localhost", 1, []byte("test"), time.Millisecond); err == nil {
		t.Fatal("NetCat should fail on closed port")
	}
	if vnet.Ping("localhost", time.Millisecond) {
		t.Log("Ping unexpectedly succeeded")
	}
	addr := &stdnet.TCPAddr{IP: stdnet.ParseIP("127.0.0.1"), Port: 1}
	if vnet.IsOpen(addr, time.Millisecond) {
		t.Log("IsOpen unexpectedly succeeded")
	}
	if _, err := vnet.Connect("localhost", 1, time.Millisecond); err == nil {
		t.Fatal("Connect should fail on closed port")
	}
}

func TestVNetGetRemoteAddressAndIsConnected(t *testing.T) {
	client, server := stdnet.Pipe()
	_ = server.Close()

	remote := vnet.GetRemoteAddress(client)
	if remote == "" {
		t.Fatal("GetRemoteAddress should return non-empty")
	}
	if !vnet.IsConnected(client) {
		t.Fatal("IsConnected should return true for pipe conn")
	}
	_ = client.Close()
}

type errConn struct{}

func (errConn) RemoteAddr() stdnet.Addr {
	return &stdnet.TCPAddr{IP: stdnet.ParseIP("0.0.0.0"), Port: 0}
}
func (errConn) LocalAddr() stdnet.Addr             { return nil }
func (errConn) Close() error                       { return nil }
func (errConn) Read([]byte) (int, error)           { return 0, nil }
func (errConn) Write([]byte) (int, error)          { return 0, nil }
func (errConn) SetDeadline(t time.Time) error      { return nil }
func (errConn) SetReadDeadline(t time.Time) error  { return nil }
func (errConn) SetWriteDeadline(t time.Time) error { return nil }

func TestVNetGetRemoteAddressStub(t *testing.T) {
	remote := vnet.GetRemoteAddress(errConn{})
	if remote == "" {
		t.Fatal("GetRemoteAddress should return non-empty for stub")
	}
}

func TestVNetDNSAndTLSDefaultDelegates(t *testing.T) {
	// DNS default delegate.
	if _, err := vnet.GetDNSInfo("localhost"); err != nil {
		t.Logf("GetDNSInfo: %v", err)
	}

	// TLS default delegates.
	b := vnet.NewTLSConfigBuilder()
	if err := vnet.AddRootCAReader(b, strings.NewReader("invalid")); err != nil {
		t.Logf("AddRootCAReader with invalid PEM: %v", err)
	}

	// Option constructors not yet tested.
	_ = vnet.WithTLSReadFile(nil)
	_ = vnet.WithTLSReadAll(nil)
	_ = vnet.WithConnectContext(context.TODO())
	_ = vnet.WithConnectNetwork("")
	_ = vnet.WithConnectDialer(nil)
	_ = vnet.WithPingContext(context.TODO())
	_ = vnet.WithPingNetwork("")
	_ = vnet.WithPingDialer(nil)
	_ = vnet.WithResolveContext(context.TODO())
	_ = vnet.WithResolveNetwork("")
	_ = vnet.WithResolver(nil)
	_ = vnet.WithDNSTypes()
}

func TestVNetMathDelegates(t *testing.T) {
	// BigInteger helpers.
	_, _ = vnet.BigIntToIPv6(big.NewInt(1))
	if !vnet.IsMaskBitValid(24) {
		t.Fatal("IsMaskBitValid(24) = false")
	}
}
