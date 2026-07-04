package net

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"os"
)

type tlsFileConfig struct {
	readFile func(string) ([]byte, error)
	readAll  func(io.Reader) ([]byte, error)
}

// TLSFileOption customizes TLS file loading helpers.
type TLSFileOption func(*tlsFileConfig)

// WithTLSReadFile sets the file reader used by TLS file helpers.
func WithTLSReadFile(readFile func(string) ([]byte, error)) TLSFileOption {
	return func(c *tlsFileConfig) {
		if readFile != nil {
			c.readFile = readFile
		}
	}
}

// WithTLSReadAll sets the reader drain function used by TLS reader helpers.
func WithTLSReadAll(readAll func(io.Reader) ([]byte, error)) TLSFileOption {
	return func(c *tlsFileConfig) {
		if readAll != nil {
			c.readAll = readAll
		}
	}
}

func applyTLSFileOptions(opts []TLSFileOption) tlsFileConfig {
	cfg := tlsFileConfig{readFile: os.ReadFile, readAll: io.ReadAll}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.readFile == nil {
		cfg.readFile = os.ReadFile
	}
	if cfg.readAll == nil {
		cfg.readAll = io.ReadAll
	}
	return cfg
}

const (
	// SSL is a legacy SSL protocol label.
	SSL = "SSL"
	// SSLv2 is a legacy SSLv2 protocol label.
	SSLv2 = "SSLv2"
	// SSLv3 is a legacy SSLv3 protocol label.
	SSLv3 = "SSLv3"
	// TLS is the TLS protocol label.
	TLS = "TLS"
	// TLSv1 is TLS 1.0.
	TLSv1 = "TLSv1"
	// TLSv11 is TLS 1.1.
	TLSv11 = "TLSv1.1"
	// TLSv12 is TLS 1.2.
	TLSv12 = "TLSv1.2"
	// TLSv13 is TLS 1.3.
	TLSv13 = "TLSv1.3"
)

// TLSConfigBuilder builds tls.Config values.
type TLSConfigBuilder struct {
	config tls.Config
}

// NewTLSConfigBuilder creates a TLS config builder.
func NewTLSConfigBuilder() *TLSConfigBuilder { return &TLSConfigBuilder{} }

// SetMinVersion sets the minimum TLS version.
func (b *TLSConfigBuilder) SetMinVersion(version uint16) *TLSConfigBuilder {
	b.config.MinVersion = version
	return b
}

// SetMaxVersion sets the maximum TLS version.
func (b *TLSConfigBuilder) SetMaxVersion(version uint16) *TLSConfigBuilder {
	b.config.MaxVersion = version
	return b
}

// SetServerName sets the TLS server name.
func (b *TLSConfigBuilder) SetServerName(name string) *TLSConfigBuilder {
	b.config.ServerName = name
	return b
}

// SetRootCAs sets root CAs.
func (b *TLSConfigBuilder) SetRootCAs(pool *x509.CertPool) *TLSConfigBuilder {
	b.config.RootCAs = pool
	return b
}

// AddRootCAFile appends PEM certificates from path to RootCAs.
func (b *TLSConfigBuilder) AddRootCAFile(path string) error {
	return b.AddRootCAFileWithOptions(path)
}

// AddRootCAFileWithOptions appends PEM certificates from path to RootCAs using options.
func (b *TLSConfigBuilder) AddRootCAFileWithOptions(path string, opts ...TLSFileOption) error {
	cfg := applyTLSFileOptions(opts)
	pem, err := cfg.readFile(path) // #nosec G304 -- caller controls certificate path.
	if err != nil {
		return err
	}
	return b.AddRootCABytes(pem)
}

// AddRootCAReader appends PEM certificates from r to RootCAs.
func (b *TLSConfigBuilder) AddRootCAReader(r io.Reader) error {
	return b.AddRootCAReaderWithOptions(r)
}

// AddRootCAReaderWithOptions appends PEM certificates from r to RootCAs using options.
func (b *TLSConfigBuilder) AddRootCAReaderWithOptions(r io.Reader, opts ...TLSFileOption) error {
	cfg := applyTLSFileOptions(opts)
	pem, err := cfg.readAll(r)
	if err != nil {
		return err
	}
	return b.AddRootCABytes(pem)
}

// AddRootCABytes appends PEM certificates from bytes to RootCAs.
func (b *TLSConfigBuilder) AddRootCABytes(pem []byte) error {
	pool := b.config.RootCAs
	if pool == nil {
		pool = x509.NewCertPool()
	}
	pool.AppendCertsFromPEM(pem)
	b.config.RootCAs = pool
	return nil
}

// SetCertificates sets client certificates.
func (b *TLSConfigBuilder) SetCertificates(certs []tls.Certificate) *TLSConfigBuilder {
	b.config.Certificates = certs
	return b
}

// Build returns a cloned tls.Config.
func (b *TLSConfigBuilder) Build() *tls.Config { return b.config.Clone() }

// CreateTLSConfig creates a TLS config using TLS 1.2 as the minimum version.
func CreateTLSConfig() *tls.Config {
	return (&tls.Config{MinVersion: tls.VersionTLS12}).Clone()
}

// TLSVersion maps a protocol label to crypto/tls version constants.
func TLSVersion(protocol string) uint16 {
	switch protocol {
	case TLSv1:
		return tls.VersionTLS10
	case TLSv11:
		return tls.VersionTLS11
	case TLSv12:
		return tls.VersionTLS12
	case TLSv13:
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12
	}
}
