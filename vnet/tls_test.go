package vnet_test

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"strings"
	"testing"

	"github.com/imajinyun/go-knifer/vnet"
)

func TestVNetTLSFileOptionsFacade(t *testing.T) {
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
	b := vnet.NewTLSConfigBuilder()
	if err := b.AddRootCAFileWithOptions("ca.pem", vnet.WithTLSReadFile(func(path string) ([]byte, error) {
		readPath = path
		return []byte(certPEM), nil
	})); err != nil {
		t.Fatalf("AddRootCAFileWithOptions: %v", err)
	}
	if readPath != "ca.pem" || b.Build().RootCAs == nil {
		t.Fatalf("TLS read provider not applied path=%q cfg=%#v", readPath, b.Build())
	}

	b = vnet.NewTLSConfigBuilder()
	if err := b.AddRootCAReader(strings.NewReader(certPEM)); err != nil || b.Build().RootCAs == nil {
		t.Fatalf("AddRootCAReader rootCAs=%#v err=%v", b.Build().RootCAs, err)
	}
}

func TestVNetTLSFacadeWrappers(t *testing.T) {
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

	b := vnet.NewTLSConfigBuilder()
	if err := vnet.AddRootCAFileWithOptions(b, "ca.pem", vnet.WithTLSReadFile(func(path string) ([]byte, error) {
		return []byte(certPEM), nil
	})); err != nil {
		t.Fatalf("AddRootCAFileWithOptions wrapper: %v", err)
	}
	if b.Build().RootCAs == nil {
		t.Fatal("AddRootCAFileWithOptions wrapper did not set RootCAs")
	}

	b = vnet.NewTLSConfigBuilder()
	if err := vnet.AddRootCAReaderWithOptions(b, strings.NewReader("ignored"), vnet.WithTLSReadAll(func(io.Reader) ([]byte, error) {
		return []byte(certPEM), nil
	})); err != nil {
		t.Fatalf("AddRootCAReaderWithOptions wrapper: %v", err)
	}
	if b.Build().RootCAs == nil {
		t.Fatal("AddRootCAReaderWithOptions wrapper did not set RootCAs")
	}

	b = vnet.NewTLSConfigBuilder()
	if err := vnet.AddRootCABytes(b, []byte(certPEM)); err != nil || b.Build().RootCAs == nil {
		t.Fatalf("AddRootCABytes wrapper rootCAs=%#v err=%v", b.Build().RootCAs, err)
	}
	if pool := vnet.NewCertPool(); pool == nil || pool.Equal(x509.NewCertPool()) != true {
		t.Fatalf("NewCertPool = %#v", pool)
	}
	if vnet.TLSVersion(vnet.TLSv13) != tls.VersionTLS13 || vnet.TLSVersion("unknown") != tls.VersionTLS12 {
		t.Fatal("TLSVersion returned unexpected version")
	}
}
