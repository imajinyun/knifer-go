package jwt

import (
	"strings"
	"testing"
	"time"
)

// Matches the utility toolkit-jwt JWTTest.

func TestCreateHS256(t *testing.T) {
	key := []byte("1234567890")
	j := New().
		SetPayload("sub", "1234567890").
		SetPayload("name", "looly").
		SetPayload("admin", true).
		SetExpiresAt(time.Unix(1640966400, 0)).
		SetKey(key)

	tok, err := j.Sign()
	if err != nil {
		t.Fatalf("sign err: %v", err)
	}
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatalf("token parts: %d", len(parts))
	}
	// It is enough that the parsed token verifies successfully.
	parsed, err := Of(tok)
	if err != nil {
		t.Fatalf("parse err: %v", err)
	}
	if !parsed.SetKey(key).Verify() {
		t.Fatalf("verify failed")
	}
	if parsed.Payload("name") != "looly" {
		t.Fatalf("payload name: %v", parsed.Payload("name"))
	}
	if parsed.Algorithm() != AlgHS256 {
		t.Fatalf("alg: %s", parsed.Algorithm())
	}
	if parsed.Type() != "JWT" {
		t.Fatalf("typ: %s", parsed.Type())
	}
}

func TestParseAndVerifyKnownToken(t *testing.T) {
	// Fixed test token from the utility toolkit.
	rightToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9." +
		"eyJzdWIiOiIxMjM0NTY3ODkwIiwiYWRtaW4iOnRydWUsIm5hbWUiOiJsb29seSJ9." +
		"U2aQkC2THYV9L0fTN-yBBI7gmo5xhmvMhATtu8v0zEA"

	j, err := Of(rightToken)
	if err != nil {
		t.Fatalf("parse err: %v", err)
	}
	if !j.SetKey([]byte("1234567890")).Verify() {
		t.Fatalf("verify failed")
	}
	if j.Header(HeaderType) != "JWT" {
		t.Fatalf("type: %v", j.Header(HeaderType))
	}
	if j.Header(HeaderAlgorithm) != "HS256" {
		t.Fatalf("alg: %v", j.Header(HeaderAlgorithm))
	}
	if j.Header(HeaderContentType) != nil {
		t.Fatalf("cty should be nil")
	}
	if j.Payload("sub") != "1234567890" {
		t.Fatalf("sub: %v", j.Payload("sub"))
	}
	if j.Payload("name") != "looly" {
		t.Fatalf("name: %v", j.Payload("name"))
	}
	if j.Payload("admin") != true {
		t.Fatalf("admin: %v", j.Payload("admin"))
	}
}

func TestSetKeyRejectsNone(t *testing.T) {
	tok := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJwdWJsaWMifQ."
	parsed, err := Of(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if parsed.SetKey([]byte("ignored")).Verify() {
		t.Fatal("SetKey should reject alg=none")
	}
	if err := parsed.SetKeyStrict([]byte("ignored")); err == nil {
		t.Fatal("SetKeyStrict should reject alg=none")
	}
	if Verify(tok, []byte("ignored")) {
		t.Fatal("Verify should reject alg=none")
	}
}

func TestSetKeyEReturnsSignerCreationError(t *testing.T) {
	j := New().SetHeader(HeaderAlgorithm, AlgPS256)
	if err := j.SetKeyE([]byte("hmac-key")); err == nil {
		t.Fatal("SetKeyE should return signer creation error for non-HMAC header alg")
	}
}

func TestSetSignerNilIsSafe(t *testing.T) {
	j := New()
	j.SetSigner(nil)
	if j.Signer() != nil {
		t.Fatal("SetSigner(nil) should clear signer")
	}
	if got := j.Algorithm(); got != "" {
		t.Fatalf("SetSigner(nil) should not write alg, got %q", got)
	}
	if err := j.SetSignerE(nil); err == nil {
		t.Fatal("SetSignerE(nil) should return an error")
	}
}

func TestNeedSigner(t *testing.T) {
	j := New().SetPayload("sub", "x")
	if _, err := j.Sign(); err == nil {
		t.Fatalf("expected error when no signer set")
	}
}

func TestSignRejectsEmptySignature(t *testing.T) {
	if tok, err := New().SetPayload("sub", "x").SetSigner(emptySigner{alg: AlgPS256}).Sign(); err == nil || tok != "" {
		t.Fatalf("Sign should reject empty signature, token=%q err=%v", tok, err)
	}
}

type emptySigner struct{ alg string }

func (s emptySigner) Algorithm() string { return s.alg }

func (emptySigner) Sign(string, string) string { return "" }

func (emptySigner) Verify(string, string, string) bool { return false }

func TestVerifyMismatchKey(t *testing.T) {
	tok, err := New().SetPayload("a", 1).SetKey([]byte("right")).Sign()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	j, _ := Of(tok)
	if j.SetKey([]byte("wrong")).Verify() {
		t.Fatalf("should fail with wrong key")
	}
}

func TestCreateTokenWithOptions(t *testing.T) {
	key := []byte("secret")
	tok, err := CreateTokenWithOptions(
		WithTokenHeaders(map[string]any{HeaderKeyID: "kid-1"}),
		WithTokenPayload(map[string]any{PayloadSubject: "alice"}),
		WithTokenKey(key),
		WithTokenAlgorithm(AlgHS384),
	)
	if err != nil {
		t.Fatalf("CreateTokenWithOptions: %v", err)
	}
	parsed, err := Of(tok)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if parsed.Header(HeaderKeyID) != "kid-1" || parsed.Payload(PayloadSubject) != "alice" {
		t.Fatalf("claims = headers:%#v payload:%#v", parsed.Headers(), parsed.Payloads())
	}
	if parsed.Algorithm() != AlgHS384 {
		t.Fatalf("alg = %q", parsed.Algorithm())
	}
	if err := parsed.SetKeyStrict(key); err != nil {
		t.Fatalf("SetKeyStrict: %v", err)
	}
	if !parsed.Verify() {
		t.Fatal("strict verification failed")
	}

	customToken, err := CreateTokenWithOptions(WithTokenPayload(map[string]any{"scope": "public"}), WithTokenSigner(HS256(key)))
	if err != nil {
		t.Fatalf("CreateTokenWithOptions with signer: %v", err)
	}
	if customToken == "" || strings.HasSuffix(customToken, ".") {
		t.Fatalf("custom signer token should be signed: %q", customToken)
	}
}

func TestVerifyWithRejectsAlgorithmMismatch(t *testing.T) {
	tok, _ := New().SetKey([]byte("k")).SetPayload("a", 1).Sign()
	j, _ := Of(tok)
	hs, _ := NewHMACSigner(AlgHS512, []byte("k"))
	if j.VerifyWith(hs) {
		t.Fatalf("HS256 token with HS512 signer should fail")
	}
}
