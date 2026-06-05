package vnet_test

import (
	stdnet "net"
	"strconv"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vnet"
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
	if vnet.CreateTLSConfig(false) == nil || vnet.NewUploadSetting().MemoryThreshold == 0 {
		t.Fatal("TLS/upload helpers failed")
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
