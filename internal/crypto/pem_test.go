package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func TestPublicKeyToPEM(t *testing.T) {
	priv, err := GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM, err := PublicKeyToPEM(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(pubPEM) == 0 {
		t.Fatal("PublicKeyToPEM() is empty")
	}
}

func TestPEMKeys(t *testing.T) {
	priv, err := GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseRSAPrivateKeyPEM(PrivateKeyToPEM(priv))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.N.Cmp(priv.N) != 0 {
		t.Fatal("parsed private key mismatch")
	}

	pkcs8, err := PrivateKeyToPKCS8PEM(priv)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err = ParseRSAPrivateKeyPEM(pkcs8)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.N.Cmp(priv.N) != 0 {
		t.Fatal("parsed PKCS#8 private key mismatch")
	}
	pub, err := ParseRSAPublicKeyPEM(PublicKeyToPKCS1PEM(&priv.PublicKey))
	if err != nil {
		t.Fatal(err)
	}
	if pub.N.Cmp(priv.N) != 0 {
		t.Fatal("parsed PKCS#1 public key mismatch")
	}
}

func TestPEMCertificate(t *testing.T) {
	priv, err := GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := newTestCertificatePEM(t, priv)
	cert, err := ParseX509CertificatePEM(certPEM)
	if err != nil {
		t.Fatal(err)
	}
	if cert.Subject.CommonName != "go-knifer-test" {
		t.Fatalf("certificate subject = %s", cert.Subject.CommonName)
	}
	certPub, err := PublicKeyFromCertificatePEM(certPEM)
	if err != nil {
		t.Fatal(err)
	}
	if certPub.N.Cmp(priv.N) != 0 {
		t.Fatal("certificate public key mismatch")
	}
}

func newTestCertificatePEM(t *testing.T, priv any) []byte {
	t.Helper()
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "go-knifer-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	rsaPriv := priv.(*rsa.PrivateKey)
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &rsaPriv.PublicKey, rsaPriv)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}
