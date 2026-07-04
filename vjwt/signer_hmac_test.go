package vjwt_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vjwt"
)

func TestStrictHMACSignerRejectsWeakKey(t *testing.T) {
	if _, err := vjwt.NewHMACSignerStrict(vjwt.JWTAlgHS256, []byte("weak")); err == nil {
		t.Fatal("NewHMACSignerStrict should reject weak key")
	}
	if _, err := vjwt.CreateSignerStrict(vjwt.JWTAlgHS256, []byte("weak")); err == nil {
		t.Fatal("CreateSignerStrict should reject weak key")
	}
	if minBytes, err := vjwt.MinHMACKeyBytes(vjwt.JWTAlgHS256); err != nil || minBytes != vjwt.MinHMACKeyBytesHS256 {
		t.Fatalf("MinHMACKeyBytes = %d, %v", minBytes, err)
	}
}

func TestHMACSignerFactories(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	for _, tt := range []struct {
		name string
		fn   func([]byte) vjwt.JWTSigner
		alg  string
	}{
		{name: "JWTSignerHS256", fn: vjwt.JWTSignerHS256, alg: vjwt.JWTAlgHS256},
		{name: "HS256", fn: vjwt.HS256, alg: vjwt.JWTAlgHS256},
		{name: "HS384", fn: vjwt.HS384, alg: vjwt.JWTAlgHS384},
		{name: "HS512", fn: vjwt.HS512, alg: vjwt.JWTAlgHS512},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(key).Algorithm(); got != tt.alg {
				t.Fatalf("Algorithm = %q, want %q", got, tt.alg)
			}
		})
	}

	signer, err := vjwt.JWTSignerHMAC(vjwt.JWTAlgHS384, key)
	if err != nil || signer.Algorithm() != vjwt.JWTAlgHS384 {
		t.Fatalf("JWTSignerHMAC alg=%q err=%v", signer.Algorithm(), err)
	}
	if got := vjwt.MustHMACSigner(vjwt.JWTAlgHS512, key).Algorithm(); got != vjwt.JWTAlgHS512 {
		t.Fatalf("MustHMACSigner alg = %q", got)
	}
}

func TestMustHMACSignerPanicsOnInvalidAlgorithm(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("MustHMACSigner should panic on invalid algorithm")
		}
	}()
	_ = vjwt.MustHMACSigner("bad", []byte("secret"))
}
