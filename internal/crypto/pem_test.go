package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
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

func TestPEMErrorContractsAndNonRSABoundaries(t *testing.T) {
	if _, err := ParseRSAPrivateKeyPEM([]byte("not pem")); !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ParseRSAPrivateKeyPEM invalid PEM err = %v", err)
	}
	if _, err := ParseRSAPublicKeyPEM([]byte("not pem")); !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ParseRSAPublicKeyPEM invalid PEM err = %v", err)
	}
	if _, err := ParseX509CertificatePEM([]byte("not pem")); !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ParseX509CertificatePEM invalid PEM err = %v", err)
	}

	badBlock := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: []byte("bad der")})
	if _, err := ParseRSAPublicKeyPEM(badBlock); err == nil {
		t.Fatal("ParseRSAPublicKeyPEM malformed DER should fail")
	}
	if _, err := ParseRSAPrivateKeyPEM(badBlock); err == nil {
		t.Fatal("ParseRSAPrivateKeyPEM malformed DER should fail")
	}
	if _, err := ParseX509CertificatePEM(badBlock); err == nil {
		t.Fatal("ParseX509CertificatePEM malformed DER should fail")
	}

	ecdsaPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ecdsaPKCS8, err := x509.MarshalPKCS8PrivateKey(ecdsaPriv)
	if err != nil {
		t.Fatal(err)
	}
	nonRSAPrivatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: ecdsaPKCS8})
	if _, err := ParseRSAPrivateKeyPEM(nonRSAPrivatePEM); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("ParseRSAPrivateKeyPEM non-RSA err = %v, want invalid key", err)
	}
	ecdsaPublicDER, err := x509.MarshalPKIXPublicKey(&ecdsaPriv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	nonRSAPublicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: ecdsaPublicDER})
	if _, err := ParseRSAPublicKeyPEM(nonRSAPublicPEM); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("ParseRSAPublicKeyPEM non-RSA err = %v, want invalid key", err)
	}

	ecdsaCertPEM := newTestCertificatePEM(t, ecdsaPriv)
	if _, err := PublicKeyFromCertificatePEM(ecdsaCertPEM); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("PublicKeyFromCertificatePEM non-RSA cert err = %v, want invalid key", err)
	}
}

func TestPEMPublicKeyRoundTripPKIX(t *testing.T) {
	priv, err := GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM, err := PublicKeyToPEM(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseRSAPublicKeyPEM(pubPEM)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.N.Cmp(priv.N) != 0 || parsed.E != priv.E {
		t.Fatal("parsed PKIX public key mismatch")
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
	var pub any
	switch p := priv.(type) {
	case *rsa.PrivateKey:
		pub = &p.PublicKey
	case *ecdsa.PrivateKey:
		pub = &p.PublicKey
	default:
		t.Fatalf("unsupported private key type %T", priv)
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, priv)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}
