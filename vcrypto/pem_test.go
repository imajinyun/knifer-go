package vcrypto_test

import (
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vcrypto"
)

func TestPEMEncodeParseRSAKeys(t *testing.T) {
	priv, err := vcrypto.GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM, err := vcrypto.PublicKeyToPEM(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := vcrypto.ParseRSAPublicKeyPEM(pubPEM)
	if err != nil {
		t.Fatal(err)
	}
	parsedPriv, err := vcrypto.ParseRSAPrivateKeyPEM(vcrypto.PrivateKeyToPEM(priv))
	if err != nil {
		t.Fatal(err)
	}
	if parsedPriv.N.Cmp(priv.N) != 0 || pub.N.Cmp(priv.N) != 0 {
		t.Fatal("parsed PEM keys do not match generated key")
	}

	pkcs8, err := vcrypto.PrivateKeyToPKCS8PEM(parsedPriv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := vcrypto.ParseRSAPrivateKeyPEM(pkcs8); err != nil {
		t.Fatal(err)
	}
	if _, err := vcrypto.ParseRSAPublicKeyPEM(vcrypto.PublicKeyToPKCS1PEM(pub)); err != nil {
		t.Fatal(err)
	}
}

func TestPEMParseInvalidKeys(t *testing.T) {
	if _, err := vcrypto.ParseRSAPrivateKeyPEM([]byte("not pem")); !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("ParseRSAPrivateKeyPEM invalid = %v", err)
	}
	if _, err := vcrypto.ParseRSAPublicKeyPEM([]byte("not pem")); !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("ParseRSAPublicKeyPEM invalid = %v", err)
	}
	if _, err := vcrypto.ParseX509CertificatePEM([]byte("not pem")); !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("ParseX509CertificatePEM invalid = %v", err)
	}
}

func TestPEMCertificateParseAndPublicKey(t *testing.T) {
	priv, err := vcrypto.GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	pub := &priv.PublicKey
	certDER, err := x509.CreateCertificate(bytes.NewReader(bytes.Repeat([]byte{0x42}, 1024)), &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "knifer-go.test"},
		NotBefore:    time.Unix(0, 0),
		NotAfter:     time.Unix(3600, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}, &x509.Certificate{}, pub, priv)
	if err != nil {
		t.Fatalf("CreateCertificate: %v", err)
	}
	certPEM := pemEncode("CERTIFICATE", certDER)
	cert, err := vcrypto.ParseX509CertificatePEM(certPEM)
	if err != nil {
		t.Fatalf("ParseX509CertificatePEM: %v", err)
	}
	if cert.Subject.CommonName != "knifer-go.test" {
		t.Fatalf("certificate CN = %q", cert.Subject.CommonName)
	}
	certPub, err := vcrypto.PublicKeyFromCertificatePEM(certPEM)
	if err != nil {
		t.Fatalf("PublicKeyFromCertificatePEM: %v", err)
	}
	if certPub.N.Cmp(pub.N) != 0 {
		t.Fatal("PublicKeyFromCertificatePEM returned different key")
	}
}

func pemEncode(typ string, der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der})
}
