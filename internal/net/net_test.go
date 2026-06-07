package net

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"io/fs"
	"math/big"
	"mime/multipart"
	stdnet "net"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

type recordingDialer struct {
	network string
	address string
	err     error
	data    chan []byte
}

type stubListener struct{}

func (stubListener) Accept() (stdnet.Conn, error) { return nil, errors.New("stub listener") }
func (stubListener) Close() error                 { return nil }
func (stubListener) Addr() stdnet.Addr {
	return &stdnet.TCPAddr{IP: stdnet.ParseIP("127.0.0.1"), Port: 12345}
}

func (d *recordingDialer) DialContext(_ context.Context, network, address string) (stdnet.Conn, error) {
	d.network = network
	d.address = address
	if d.err != nil {
		return nil, d.err
	}
	client, server := stdnet.Pipe()
	go func() {
		defer func() { _ = server.Close() }()
		payload, _ := io.ReadAll(server)
		d.data <- payload
	}()
	return client, nil
}

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
	var compiled string
	if !MatchesWildcardWithOptions("10.0.*.2", "10.0.1.2", WithWildcardCompileFunc(func(pattern string) (*regexp.Regexp, error) {
		compiled = pattern
		return regexp.Compile(pattern)
	})) {
		t.Fatal("MatchesWildcardWithOptions failed")
	}
	if compiled != `^10\.0\.\d{1,3}\.2$` {
		t.Fatalf("compiled wildcard pattern = %q", compiled)
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
	g := NewLocalPortGeneratorWithOptions(port, WithPortHost("127.0.0.1"))
	generated, err := g.Generate()
	if err != nil {
		t.Fatalf("LocalPortGenerator.Generate with options: %v", err)
	}
	if generated <= port || generated > PortRangeMax {
		t.Fatalf("LocalPortGenerator generated %d, want > %d", generated, port)
	}
	next, err := g.GenerateWithOptions(WithPortHost("127.0.0.1"))
	if err != nil {
		t.Fatalf("LocalPortGenerator.GenerateWithOptions: %v", err)
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

func TestConnectHelpersWithOptionsUseDialerNetworkAndTimeout(t *testing.T) {
	dialer := &recordingDialer{data: make(chan []byte, 1)}
	conn, err := ConnectWithOptions(
		"example.com", 8080,
		WithConnectNetwork("tcp4"),
		WithConnectTimeout(time.Second),
		WithConnectDialer(dialer),
	)
	if err != nil {
		t.Fatalf("ConnectWithOptions: %v", err)
	}
	_ = conn.Close()
	if dialer.network != "tcp4" || dialer.address != "example.com:8080" {
		t.Fatalf("dial target = %s %s", dialer.network, dialer.address)
	}

	dialer = &recordingDialer{data: make(chan []byte, 1)}
	if err := NetCatWithOptions("127.0.0.1", 1234, []byte("hello"), WithConnectDialer(dialer)); err != nil {
		t.Fatalf("NetCatWithOptions: %v", err)
	}
	if got := string(<-dialer.data); got != "hello" {
		t.Fatalf("NetCatWithOptions wrote %q", got)
	}
	if dialer.network != "tcp" || dialer.address != "127.0.0.1:1234" {
		t.Fatalf("netcat dial target = %s %s", dialer.network, dialer.address)
	}

	addr := &stdnet.TCPAddr{IP: stdnet.ParseIP("127.0.0.1"), Port: 4321}
	dialer = &recordingDialer{data: make(chan []byte, 1)}
	if !IsOpenWithOptions(addr, WithConnectDialer(dialer), WithConnectNetwork("tcp4")) {
		t.Fatal("IsOpenWithOptions should report true for successful dialer")
	}
	if dialer.network != "tcp4" || dialer.address != addr.String() {
		t.Fatalf("is-open dial target = %s %s", dialer.network, dialer.address)
	}
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
	req := multipartAvatarRequest(t, "a.txt")
	setting := NewUploadSetting()
	setting.FileExts = []string{".jpg"}
	setting.AllowFileExts = true
	if _, err := ParseMultipartForm(req, setting); err == nil {
		t.Fatal("ParseMultipartForm should reject extension outside allow list")
	}

	req = multipartAvatarRequest(t, "a.txt")
	setting.FileExts = []string{"txt"}
	setting.AllowFileExts = true
	if _, err := ParseMultipartForm(req, setting); err != nil {
		t.Fatalf("ParseMultipartForm should accept allowed extension: %v", err)
	}

	req = multipartAvatarRequest(t, "a.exe")
	setting.FileExts = []string{".exe"}
	setting.AllowFileExts = false
	if _, err := ParseMultipartForm(req, setting); err == nil {
		t.Fatal("ParseMultipartForm should reject extension in deny list")
	}
}

func TestSaveUploadedFileProviderOptions(t *testing.T) {
	req := multipartAvatarRequest(t, "a.txt")
	form, err := ParseMultipartForm(req, NewUploadSetting())
	if err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}
	file := form.GetFile("avatar")
	if file == nil {
		t.Fatal("uploaded file is nil")
	}

	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	var written bytes.Buffer
	err = SaveUploadedFile(file, "/virtual/upload/a.txt",
		WithUploadMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return nil
		}),
		WithUploadOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openPath, openFlag, openPerm = path, flag, perm
			return nopWriteCloser{Writer: &written}, nil
		}),
		WithUploadDirPerm(0o700), WithUploadFilePerm(0o600),
	)
	if err != nil {
		t.Fatalf("SaveUploadedFile provider: %v", err)
	}
	if mkdirPath != "/virtual/upload" || mkdirPerm != 0o700 || openPath != "/virtual/upload/a.txt" || openPerm != 0o600 || openFlag&os.O_CREATE == 0 || written.String() != "hello" {
		t.Fatalf("providers mkdir=%q/%v open=%q flag=%#x perm=%v content=%q", mkdirPath, mkdirPerm, openPath, openFlag, openPerm, written.String())
	}
}

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

func multipartAvatarRequest(t *testing.T, filename string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile("avatar", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("hello")); err != nil {
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

func TestTLSRootCAProviderOptions(t *testing.T) {
	const certPEM = `-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIRAPWQSq0Qr7yZD5twH61BxFIwCgYIKoZIzj0EAwIwEjEQ
MA4GA1UEChMHZ28tdGVzdDAeFw0yNjA2MDYwMDAwMDBaFw0yNzA2MDYwMDAwMDBa
MBIxEDAOBgNVBAoTB2dvLXRlc3QwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASm
1YPqMC7UTw4R7ovbHYgk4+LALoU6hr61VnsBiKCdsMCMScpLob8ldIl+6o4f/ntM
5kmXvEFd9Mp6FfaHkgnbo0IwQDAOBgNVHQ8BAf8EBAMCAqQwDwYDVR0TAQH/BAUw
AwEB/zAdBgNVHQ4EFgQUX90U1OkOXbGUzD2JNoWlqQtk3/0wCgYIKoZIzj0EAwID
SQAwRgIhANw7UzN0vtxOfygWqANg00uGOo7y98q1/Ac3N1wQxVBkAiEA7QjQRHtH
LA6wKo8yoCnW36b+nvxlhHvzrIxwWCgwCWM=
-----END CERTIFICATE-----`
	readPath := ""
	b := NewTLSConfigBuilder()
	if err := b.AddRootCAFileWithOptions("ca.pem", WithTLSReadFile(func(path string) ([]byte, error) {
		readPath = path
		return []byte(certPEM), nil
	})); err != nil {
		t.Fatalf("AddRootCAFileWithOptions: %v", err)
	}
	if readPath != "ca.pem" || b.Build().RootCAs == nil {
		t.Fatalf("TLS read provider not applied path=%q cfg=%#v", readPath, b.Build())
	}

	b = NewTLSConfigBuilder()
	if err := b.AddRootCAReader(strings.NewReader(certPEM)); err != nil {
		t.Fatalf("AddRootCAReader: %v", err)
	}
	if b.Build().RootCAs == nil {
		t.Fatal("AddRootCAReader should initialize RootCAs")
	}
}
