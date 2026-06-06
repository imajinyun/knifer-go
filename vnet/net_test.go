package vnet_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	stdnet "net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vnet"
)

type recordingDialer struct {
	network string
	address string
	data    chan []byte
}

func (d *recordingDialer) DialContext(_ context.Context, network, address string) (stdnet.Conn, error) {
	d.network = network
	d.address = address
	client, server := stdnet.Pipe()
	go func() {
		defer func() { _ = server.Close() }()
		payload, _ := io.ReadAll(server)
		d.data <- payload
	}()
	return client, nil
}

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
	g := vnet.NewLocalPortGeneratorWithOptions(port, vnet.WithPortHost("127.0.0.1"))
	generated, err := g.Generate()
	if err != nil {
		t.Fatalf("LocalPortGenerator.Generate with options: %v", err)
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

func TestVNetConnectOptionsFacade(t *testing.T) {
	dialer := &recordingDialer{data: make(chan []byte, 1)}
	conn, err := vnet.ConnectWithOptions(
		"example.com", 8080,
		vnet.WithConnectNetwork("tcp4"),
		vnet.WithConnectTimeout(time.Second),
		vnet.WithConnectDialer(dialer),
	)
	if err != nil {
		t.Fatalf("ConnectWithOptions: %v", err)
	}
	_ = conn.Close()
	if dialer.network != "tcp4" || dialer.address != "example.com:8080" {
		t.Fatalf("dial target = %s %s", dialer.network, dialer.address)
	}

	dialer = &recordingDialer{data: make(chan []byte, 1)}
	if err := vnet.NetCatWithOptions("127.0.0.1", 1234, []byte("hello"), vnet.WithConnectDialer(dialer)); err != nil {
		t.Fatalf("NetCatWithOptions: %v", err)
	}
	if got := string(<-dialer.data); got != "hello" {
		t.Fatalf("NetCatWithOptions wrote %q", got)
	}

	addr := &stdnet.TCPAddr{IP: stdnet.ParseIP("127.0.0.1"), Port: 4321}
	dialer = &recordingDialer{data: make(chan []byte, 1)}
	if !vnet.IsOpenWithOptions(addr, vnet.WithConnectDialer(dialer)) {
		t.Fatal("IsOpenWithOptions should report true")
	}
}

func TestVNetUploadSaveOptionsFacade(t *testing.T) {
	req := multipartRequest(t, "avatar", "a.txt", "hello")
	form, err := vnet.ParseMultipartForm(req, vnet.NewUploadSetting())
	if err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}
	file := form.GetFile("avatar")
	if file == nil {
		t.Fatal("uploaded file is nil")
	}
	if vnet.UploadFileName(file) != "a.txt" || vnet.UploadFileSize(file) != int64(len("hello")) || vnet.UploadFileContentType(file) == "" {
		t.Fatalf("upload metadata = name:%q size:%d type:%q", vnet.UploadFileName(file), vnet.UploadFileSize(file), vnet.UploadFileContentType(file))
	}

	dir := t.TempDir()
	dest := filepath.Join(dir, "nested", "a.txt")
	if err := vnet.SaveUploadedFile(file, dest, vnet.WithUploadFilePerm(0o600), vnet.WithUploadDirPerm(0o700)); err != nil {
		t.Fatalf("SaveUploadedFile: %v", err)
	}
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("stat saved file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("saved file perm = %v", info.Mode().Perm())
	}
	if err := vnet.SaveUploadedFile(file, dest, vnet.WithUploadOverwrite(false)); err == nil {
		t.Fatal("SaveUploadedFile should reject overwrite when disabled")
	}
	missingParent := filepath.Join(dir, "missing", "b.txt")
	if err := vnet.SaveUploadedFile(file, missingParent, vnet.WithUploadCreateParents(false)); err == nil {
		t.Fatal("SaveUploadedFile should reject missing parent when parent creation is disabled")
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
