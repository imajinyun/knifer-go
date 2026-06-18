package vjwt_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/imajinyun/go-knifer/vjwt"
)

func TestECDSASignerFactories(t *testing.T) {
	if _, err := vjwt.NewECDSASigner(vjwt.JWTAlgES256, nil, nil); err == nil {
		t.Fatal("NewECDSASigner(nil keys) error = nil")
	}

	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa.GenerateKey: %v", err)
	}
	ecSigner, err := vjwt.JWTSignerECDSA(vjwt.JWTAlgES256, ecdsaKey, &ecdsaKey.PublicKey)
	if err != nil || ecSigner.Algorithm() != vjwt.JWTAlgES256 {
		t.Fatalf("JWTSignerECDSA alg=%q err=%v", ecSigner.Algorithm(), err)
	}
	if got := vjwt.JWTSignerES256(ecdsaKey, &ecdsaKey.PublicKey).Algorithm(); got != vjwt.JWTAlgES256 {
		t.Fatalf("JWTSignerES256 alg = %q", got)
	}
	if got := vjwt.ES256(ecdsaKey, &ecdsaKey.PublicKey).Algorithm(); got != vjwt.JWTAlgES256 {
		t.Fatalf("ES256 alg = %q", got)
	}
	reader := zeroReader{}
	if got := vjwt.ES256WithOptions(ecdsaKey, &ecdsaKey.PublicKey, vjwt.WithSignerRandomReader(reader)).Algorithm(); got != vjwt.JWTAlgES256 {
		t.Fatalf("ES256WithOptions alg = %q", got)
	}
	ecdsa384Key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa.GenerateKey P384: %v", err)
	}
	if got := vjwt.ES384(ecdsa384Key, &ecdsa384Key.PublicKey).Algorithm(); got != vjwt.JWTAlgES384 {
		t.Fatalf("ES384 alg = %q", got)
	}
	if got := vjwt.ES384WithOptions(ecdsa384Key, &ecdsa384Key.PublicKey).Algorithm(); got != vjwt.JWTAlgES384 {
		t.Fatalf("ES384WithOptions alg = %q", got)
	}
	ecdsa521Key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa.GenerateKey P521: %v", err)
	}
	if got := vjwt.ES512(ecdsa521Key, &ecdsa521Key.PublicKey).Algorithm(); got != vjwt.JWTAlgES512 {
		t.Fatalf("ES512 alg = %q", got)
	}
	if got := vjwt.ES512WithOptions(ecdsa521Key, &ecdsa521Key.PublicKey).Algorithm(); got != vjwt.JWTAlgES512 {
		t.Fatalf("ES512WithOptions alg = %q", got)
	}
}

func TestFacadeNewECDSASignerWithOptions(t *testing.T) {
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa.GenerateKey: %v", err)
	}
	s, err := vjwt.NewECDSASignerWithOptions(vjwt.JWTAlgES256, ecdsaKey, &ecdsaKey.PublicKey)
	if err != nil || s.Algorithm() != vjwt.JWTAlgES256 {
		t.Fatalf("NewECDSASignerWithOptions alg=%q err=%v", s.Algorithm(), err)
	}
}
