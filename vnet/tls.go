package vnet

import (
	"crypto/tls"
	"crypto/x509"

	netimpl "github.com/imajinyun/go-knifer/internal/net"
)

func NewTLSConfigBuilder() *TLSConfigBuilder { return netimpl.NewTLSConfigBuilder() }

func CreateTLSConfig(insecureSkipVerify bool) *tls.Config {
	return netimpl.CreateTLSConfig(insecureSkipVerify)
}

func InsecureTLSConfig() *tls.Config { return netimpl.InsecureTLSConfig() }

func TLSVersion(protocol string) uint16 { return netimpl.TLSVersion(protocol) }

func NewCertPool() *x509.CertPool { return x509.NewCertPool() }
