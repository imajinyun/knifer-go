package jwt

import "testing"

func TestUtil_CreateAndVerifyStrictWithAlgorithm(t *testing.T) {
	key := []byte("secret")
	tok, err := CreateTokenWithAlgorithm(map[string]any{"a": 1}, key, AlgHS512)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	parsed, err := ParseToken(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.Algorithm() != AlgHS512 {
		t.Fatalf("alg = %q, want %q", parsed.Algorithm(), AlgHS512)
	}
	if !VerifyStrict(tok, key) {
		t.Fatal("VerifyStrict failed")
	}
	if !Verify(tok, key) {
		t.Fatal("Verify should use header algorithm without fallback")
	}
	if _, err := CreateTokenWithAlgorithm(map[string]any{"a": 1}, key, "bad"); err == nil {
		t.Fatal("CreateTokenWithAlgorithm bad alg error = nil")
	}
}

func TestUtil_VerifyRejectsUnsupportedHeaderAlgorithm(t *testing.T) {
	key := []byte("secret")
	tok, err := New().SetHeader(HeaderAlgorithm, "BAD").SetPayload("a", 1).SetSigner(HS256(key)).Sign()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if Verify(tok, key) {
		t.Fatal("Verify should reject unsupported header alg instead of falling back")
	}
}
