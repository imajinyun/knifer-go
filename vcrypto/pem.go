package vcrypto

import (
	"crypto/rsa"
	"crypto/x509"

	cryptoimpl "github.com/imajinyun/go-knifer/internal/crypto"
)

// PrivateKeyToPEM encodes an RSA private key as PKCS#1 PEM.
func PrivateKeyToPEM(priv *rsa.PrivateKey) []byte { return cryptoimpl.PrivateKeyToPEM(priv) }

// PublicKeyToPEM encodes an RSA public key as PKIX PEM.
func PublicKeyToPEM(pub *rsa.PublicKey) ([]byte, error) { return cryptoimpl.PublicKeyToPEM(pub) }

// ParseRSAPrivateKeyPEM parses a PKCS#1 or PKCS#8 RSA private key PEM.
func ParseRSAPrivateKeyPEM(data []byte) (*rsa.PrivateKey, error) {
	return cryptoimpl.ParseRSAPrivateKeyPEM(data)
}

// ParseRSAPublicKeyPEM parses a PKIX or PKCS#1 RSA public key PEM.
func ParseRSAPublicKeyPEM(data []byte) (*rsa.PublicKey, error) {
	return cryptoimpl.ParseRSAPublicKeyPEM(data)
}

// PrivateKeyToPKCS8PEM encodes an RSA private key as PKCS#8 PEM.
func PrivateKeyToPKCS8PEM(priv *rsa.PrivateKey) ([]byte, error) {
	return cryptoimpl.PrivateKeyToPKCS8PEM(priv)
}

// PublicKeyToPKCS1PEM encodes an RSA public key as PKCS#1 PEM.
func PublicKeyToPKCS1PEM(pub *rsa.PublicKey) []byte { return cryptoimpl.PublicKeyToPKCS1PEM(pub) }

// ParseX509CertificatePEM parses an X.509 certificate from PEM data.
func ParseX509CertificatePEM(data []byte) (*x509.Certificate, error) {
	return cryptoimpl.ParseX509CertificatePEM(data)
}

// PublicKeyFromCertificatePEM extracts an RSA public key from an X.509 certificate PEM.
func PublicKeyFromCertificatePEM(data []byte) (*rsa.PublicKey, error) {
	return cryptoimpl.PublicKeyFromCertificatePEM(data)
}
