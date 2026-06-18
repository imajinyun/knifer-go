package vjwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/imajinyun/go-knifer/vjwt"
)

func TestRSAPSSSignerFactories(t *testing.T) {
	if _, err := vjwt.NewRSAPSSSigner(vjwt.JWTAlgPS256, nil, nil); err == nil {
		t.Fatal("NewRSAPSSSigner(nil keys) error = nil")
	}

	rsaKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	reader := zeroReader{}
	psSigner, err := vjwt.NewRSAPSSSignerWithOptions(vjwt.JWTAlgPS256, rsaKey, &rsaKey.PublicKey, vjwt.WithSignerRandomReader(reader), vjwt.WithRSAPSSOptions(nil))
	if err != nil || psSigner.Algorithm() != vjwt.JWTAlgPS256 {
		t.Fatalf("NewRSAPSSSignerWithOptions alg=%q err=%v", psSigner.Algorithm(), err)
	}
	if got := vjwt.PS256WithOptions(rsaKey, &rsaKey.PublicKey, vjwt.WithSignerRandomReader(reader)).Algorithm(); got != vjwt.JWTAlgPS256 {
		t.Fatalf("PS256WithOptions alg = %q", got)
	}
	if got := vjwt.PS384(rsaKey, &rsaKey.PublicKey).Algorithm(); got != vjwt.JWTAlgPS384 {
		t.Fatalf("PS384 alg = %q", got)
	}
	if got := vjwt.PS512WithOptions(rsaKey, &rsaKey.PublicKey).Algorithm(); got != vjwt.JWTAlgPS512 {
		t.Fatalf("PS512WithOptions alg = %q", got)
	}
}

func TestFacadePS256(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	s := vjwt.PS256(rsaKey, &rsaKey.PublicKey)
	if s.Algorithm() != vjwt.JWTAlgPS256 {
		t.Fatalf("PS256 alg = %q", s.Algorithm())
	}
}

func TestFacadePS384WithOptions(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	s := vjwt.PS384WithOptions(rsaKey, &rsaKey.PublicKey)
	if s.Algorithm() != vjwt.JWTAlgPS384 {
		t.Fatalf("PS384WithOptions alg = %q", s.Algorithm())
	}
}

func TestFacadePS512(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	s := vjwt.PS512(rsaKey, &rsaKey.PublicKey)
	if s.Algorithm() != vjwt.JWTAlgPS512 {
		t.Fatalf("PS512 alg = %q", s.Algorithm())
	}
}
