package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"
)

func TestRSASigner_RoundTrip(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	for _, alg := range []string{AlgRS256, AlgRS384, AlgRS512} {
		signer, err := NewRSASigner(alg, priv, nil)
		if err != nil {
			t.Fatalf("%s: %v", alg, err)
		}
		token, err := New().AddPayloads(map[string]any{"u": 1}).SetSigner(signer).Sign()
		if err != nil {
			t.Fatalf("%s sign: %v", alg, err)
		}
		// 验签
		j, err := Of(token)
		if err != nil {
			t.Fatalf("%s parse: %v", alg, err)
		}
		if !j.VerifyWith(signer) {
			t.Fatalf("%s verify failed", alg)
		}
		// 用单独 pub 验签
		verifier, err := NewRSASigner(alg, nil, &priv.PublicKey)
		if err != nil {
			t.Fatal(err)
		}
		if !j.VerifyWith(verifier) {
			t.Fatalf("%s pub verify failed", alg)
		}
	}
}

func TestRSAPSSSigner_RoundTrip(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	for _, alg := range []string{AlgPS256, AlgPS384, AlgPS512} {
		signer, err := NewRSAPSSSigner(alg, priv, nil)
		if err != nil {
			t.Fatalf("%s: %v", alg, err)
		}
		token, err := New().AddPayloads(map[string]any{"u": 1}).SetSigner(signer).Sign()
		if err != nil {
			t.Fatalf("%s sign: %v", alg, err)
		}
		j, err := Of(token)
		if err != nil {
			t.Fatal(err)
		}
		if !j.VerifyWith(signer) {
			t.Fatalf("%s verify failed", alg)
		}
	}
}

func TestRSAPSSSignerWithOptions(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := NewRSAPSSSignerWithOptions(AlgPS256, priv, nil, WithRSAPSSOptions(&rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto}))
	if err != nil {
		t.Fatalf("NewRSAPSSSignerWithOptions: %v", err)
	}
	token, err := New().SetPayload("u", 1).SetSigner(signer).Sign()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	j, err := Of(token)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := NewRSAPSSSignerWithOptions(AlgPS256, nil, &priv.PublicKey, WithRSAPSSOptions(&rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto}))
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	if !j.VerifyWith(verifier) {
		t.Fatal("verify with custom PSS options failed")
	}
	failing, err := NewRSAPSSSignerWithOptions(AlgPS256, priv, nil, WithSignerRandomReader(errReader{}))
	if err != nil {
		t.Fatalf("failing signer: %v", err)
	}
	if sig := failing.Sign("a", "b"); sig != "" {
		t.Fatalf("signature with failing random reader = %q, want empty", sig)
	}
}

func TestECDSASigner_RoundTrip(t *testing.T) {
	cases := []struct {
		alg   string
		curve elliptic.Curve
	}{
		{AlgES256, elliptic.P256()},
		{AlgES384, elliptic.P384()},
		{AlgES512, elliptic.P521()},
	}
	for _, c := range cases {
		priv, err := ecdsa.GenerateKey(c.curve, rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		signer, err := NewECDSASigner(c.alg, priv, nil)
		if err != nil {
			t.Fatalf("%s: %v", c.alg, err)
		}
		token, err := New().AddPayloads(map[string]any{"u": 1}).SetSigner(signer).Sign()
		if err != nil {
			t.Fatalf("%s sign: %v", c.alg, err)
		}
		j, err := Of(token)
		if err != nil {
			t.Fatal(err)
		}
		if !j.VerifyWith(signer) {
			t.Fatalf("%s verify failed", c.alg)
		}
	}
}

func TestECDSASignerWithOptionsRandomReader(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := NewECDSASignerWithOptions(AlgES256, priv, nil, WithSignerRandomReader(rand.Reader))
	if err != nil {
		t.Fatalf("NewECDSASignerWithOptions: %v", err)
	}
	token, err := New().SetPayload("u", 1).SetSigner(signer).Sign()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	j, err := Of(token)
	if err != nil {
		t.Fatal(err)
	}
	if !j.VerifyWith(signer) {
		t.Fatal("verify with custom ECDSA random reader failed")
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

func TestECDSASigner_CurveMismatch(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewECDSASigner(AlgES256, priv, nil); err == nil {
		t.Fatal("expected curve mismatch error")
	}
}

func TestSignerUtilFactories(t *testing.T) {
	// HS*
	if HS256([]byte("k")).Algorithm() != AlgHS256 {
		t.Fatal()
	}
	if HS384([]byte("k")).Algorithm() != AlgHS384 {
		t.Fatal()
	}
	if HS512([]byte("k")).Algorithm() != AlgHS512 {
		t.Fatal()
	}
	if None().Algorithm() != AlgNone {
		t.Fatal()
	}

	// RS*
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if RS256(priv, nil).Algorithm() != AlgRS256 {
		t.Fatal()
	}
	if PS256(priv, nil).Algorithm() != AlgPS256 {
		t.Fatal()
	}

	// ES*
	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if ES256(ec, nil).Algorithm() != AlgES256 {
		t.Fatal()
	}
}

func TestAlgorithmName(t *testing.T) {
	pairs := map[string]string{
		AlgHS256: "HmacSHA256",
		AlgHS384: "HmacSHA384",
		AlgHS512: "HmacSHA512",
		AlgRS256: "SHA256withRSA",
		AlgPS256: "SHA256withRSA_PSS",
		AlgES256: "SHA256withECDSA",
		AlgNone:  "None",
	}
	for id, name := range pairs {
		if got := AlgorithmName(id); got != name {
			t.Fatalf("%s -> %s, want %s", id, got, name)
		}
	}
	if AlgorithmName("UNKNOWN") != "UNKNOWN" {
		t.Fatal("unknown should be returned as-is")
	}
}

func TestJWTValidator_Chain(t *testing.T) {
	signer := HS256([]byte("secret"))
	now := time.Now()
	token, err := New().
		AddPayloads(map[string]any{
			PayloadIssuer:    "alice",
			PayloadIssuedAt:  now.Unix(),
			PayloadExpiresAt: now.Add(time.Hour).Unix(),
		}).
		SetSigner(signer).
		Sign()
	if err != nil {
		t.Fatal(err)
	}

	if err := OfValidator(token).
		ValidateAlgorithm(signer).
		ValidateDate(now, 0).
		Err(); err != nil {
		t.Fatalf("validator should pass: %v", err)
	}

	// 算法不匹配
	if err := OfValidator(token).ValidateAlgorithm(HS384([]byte("secret"))).Err(); err == nil {
		t.Fatal("expected algorithm mismatch error")
	}

	// 过期场景
	expired, _ := New().
		AddPayloads(map[string]any{PayloadExpiresAt: now.Add(-time.Hour).Unix()}).
		SetSigner(signer).
		Sign()
	if err := OfValidator(expired).
		ValidateAlgorithm(signer).
		ValidateDate(now, 0).
		Err(); err == nil {
		t.Fatal("expected expired error")
	}
}
