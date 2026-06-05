package net

import (
	"bytes"
	"crypto/tls"
	"math/big"
	"mime/multipart"
	stdnet "net"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestIPv4Helpers(t *testing.T) {
	v, err := IPv4ToLong("127.0.0.1")
	if err != nil || v != 2130706433 {
		t.Fatalf("IPv4ToLong = %d %v", v, err)
	}
	if got := LongToIPv4(v); got != "127.0.0.1" {
		t.Fatalf("LongToIPv4 = %q", got)
	}
	if !IsIPv4("192.168.1.1") || IsIPv4("999.1.1.1") || !IsIPv6("::1") || !IsIP("::1") {
		t.Fatal("IP validators failed")
	}
	if !IsInnerIP("192.168.1.1") || IsInnerIP("8.8.8.8") {
		t.Fatal("IsInnerIP failed")
	}
	if got, _ := BeginIP("192.168.1.9", 24); got != "192.168.1.0" {
		t.Fatalf("BeginIP = %q", got)
	}
	if got, _ := EndIP("192.168.1.9", 24); got != "192.168.1.255" {
		t.Fatalf("EndIP = %q", got)
	}
	if bit, _ := MaskBitByMask("255.255.255.0"); bit != 24 {
		t.Fatalf("MaskBitByMask = %d", bit)
	}
	if mask, _ := MaskByMaskBit(24); mask != "255.255.255.0" {
		t.Fatalf("MaskByMaskBit = %q", mask)
	}
	if count, _ := CountByMaskBit(30, false); count != 2 {
		t.Fatalf("CountByMaskBit = %d", count)
	}
	if ips, _ := ListIPCIDR("192.168.1.0", 30, false); !reflect.DeepEqual(ips, []string{"192.168.1.1", "192.168.1.2"}) {
		t.Fatalf("ListIPCIDR = %#v", ips)
	}
	if !MatchesWildcard("192.168.*.*", "192.168.1.2") || !IsInRange("192.168.1.2", "192.168.1.0/24") {
		t.Fatal("range matching failed")
	}
}

func TestIPv6BigInt(t *testing.T) {
	v, err := IPv6ToBigInt("::1")
	if err != nil || v.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("IPv6ToBigInt = %v %v", v, err)
	}
	if got, err := BigIntToIPv6(big.NewInt(1)); err != nil || got != "::1" {
		t.Fatalf("BigIntToIPv6 = %q %v", got, err)
	}
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

func TestPingWithOptions(t *testing.T) {
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
	if !PingWithOptions("127.0.0.1", WithPingPorts(port), WithPingTimeout(time.Second), WithPingNetwork("tcp")) {
		t.Fatal("PingWithOptions failed to reach local listener")
	}
	<-done
}

func TestResolveWithOptions(t *testing.T) {
	ips, err := GetIPByHostWithOptions("localhost", WithResolveNetwork("ip4"), WithResolveTimeout(time.Second))
	if err != nil {
		t.Fatalf("GetIPByHostWithOptions: %v", err)
	}
	if len(ips) == 0 {
		t.Fatal("GetIPByHostWithOptions returned no IPs")
	}
	dns, err := GetDNSInfoWithOptions("localhost", WithDNSTypes("A"), WithResolveTimeout(time.Second))
	if err != nil {
		t.Fatalf("GetDNSInfoWithOptions: %v", err)
	}
	if len(dns) == 0 {
		t.Fatal("GetDNSInfoWithOptions returned no A records")
	}
}

func TestMultipartFileExts(t *testing.T) {
	req := multipartRequest(t, "avatar", "a.txt", "hello")
	setting := NewUploadSetting()
	setting.FileExts = []string{".jpg"}
	setting.AllowFileExts = true
	if _, err := ParseMultipartForm(req, setting); err == nil {
		t.Fatal("ParseMultipartForm should reject extension outside allow list")
	}

	req = multipartRequest(t, "avatar", "a.txt", "hello")
	setting.FileExts = []string{"txt"}
	setting.AllowFileExts = true
	if _, err := ParseMultipartForm(req, setting); err != nil {
		t.Fatalf("ParseMultipartForm should accept allowed extension: %v", err)
	}

	req = multipartRequest(t, "avatar", "a.exe", "hello")
	setting.FileExts = []string{".exe"}
	setting.AllowFileExts = false
	if _, err := ParseMultipartForm(req, setting); err == nil {
		t.Fatal("ParseMultipartForm should reject extension in deny list")
	}
}

func multipartRequest(t *testing.T, field, filename, content string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, "/upload", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestTLSHelpers(t *testing.T) {
	cfg := NewTLSConfigBuilder().SetMinVersion(tls.VersionTLS12).SetServerName("example.com").Build()
	if cfg.MinVersion != tls.VersionTLS12 || cfg.ServerName != "example.com" {
		t.Fatalf("TLS builder failed: %#v", cfg)
	}
	if TLSVersion(TLSv13) != tls.VersionTLS13 {
		t.Fatal("TLSVersion failed")
	}
}
