package net

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestTLSReaderWithOptionsUsesReadAll(t *testing.T) {
	b := NewTLSConfigBuilder()
	called := false
	err := b.AddRootCAReaderWithOptions(strings.NewReader("ignored"), WithTLSReadAll(func(r io.Reader) ([]byte, error) {
		called = true
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		if string(data) != "ignored" {
			t.Fatalf("reader data = %q", data)
		}
		return []byte("not a certificate"), nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("custom TLS readAll provider was not called")
	}
}

func TestTLSHelpers(t *testing.T) {
	pool := x509.NewCertPool()
	certs := []tls.Certificate{{Certificate: [][]byte{[]byte("cert")}}}
	cfg := NewTLSConfigBuilder().
		SetMinVersion(tls.VersionTLS12).
		SetMaxVersion(tls.VersionTLS13).
		SetServerName("example.com").
		SetRootCAs(pool).
		SetCertificates(certs).
		Build()
	if cfg.MinVersion != tls.VersionTLS12 || cfg.MaxVersion != tls.VersionTLS13 || cfg.ServerName != "example.com" || cfg.RootCAs != pool || len(cfg.Certificates) != 1 {
		t.Fatalf("TLS builder failed: %#v", cfg)
	}
	created := CreateTLSConfig()
	if created.MinVersion != tls.VersionTLS12 {
		t.Fatalf("CreateTLSConfig MinVersion = %x", created.MinVersion)
	}
	tests := map[string]uint16{
		TLSv1:  tls.VersionTLS10,
		TLSv11: tls.VersionTLS11,
		TLSv12: tls.VersionTLS12,
		TLSv13: tls.VersionTLS13,
		SSL:    tls.VersionTLS12,
	}
	for protocol, want := range tests {
		if got := TLSVersion(protocol); got != want {
			t.Fatalf("TLSVersion(%q) = %x, want %x", protocol, got, want)
		}
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

	wantErr := errors.New("read failed")
	if err := NewTLSConfigBuilder().AddRootCAFileWithOptions("missing.pem", WithTLSReadFile(func(string) ([]byte, error) { return nil, wantErr })); !errors.Is(err, wantErr) {
		t.Fatalf("AddRootCAFileWithOptions error = %v", err)
	}
	if err := NewTLSConfigBuilder().AddRootCAReaderWithOptions(strings.NewReader("ignored"), WithTLSReadAll(func(io.Reader) ([]byte, error) { return nil, wantErr })); !errors.Is(err, wantErr) {
		t.Fatalf("AddRootCAReaderWithOptions error = %v", err)
	}
	if err := NewTLSConfigBuilder().AddRootCAFileWithOptions("ca.pem", WithTLSReadFile(nil)); err == nil {
		t.Fatalf("AddRootCAFileWithOptions nil reader should use default and fail missing file")
	}
}
