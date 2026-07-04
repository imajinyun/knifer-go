package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

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
	failing, err = NewRSAPSSSignerWithOptions(AlgPS256, priv, nil, WithSignerRandomReader(errReader{}), WithSignerRandomReader(nil))
	if err != nil {
		t.Fatalf("failing signer with nil overwrite: %v", err)
	}
	if sig := failing.Sign("a", "b"); sig != "" {
		t.Fatalf("signature after nil random overwrite = %q, want empty", sig)
	}
	if token, err := New().SetPayload("u", 1).SetSigner(failing).Sign(); err == nil || token != "" {
		t.Fatalf("Sign should reject RSA-PSS empty signature, token=%q err=%v", token, err)
	}
	publicOnly, err := NewRSAPSSSignerWithOptions(AlgPS256, nil, &priv.PublicKey)
	if err != nil {
		t.Fatalf("public-only signer: %v", err)
	}
	if token, err := New().SetPayload("u", 1).SetSigner(publicOnly).Sign(); err == nil || token != "" {
		t.Fatalf("Sign should reject public-only RSA-PSS empty signature, token=%q err=%v", token, err)
	}
}
