package jwt

import (
	"strings"
	"testing"
	"time"
)

// 对应 the utility toolkit-jwt JWTTest。

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
	// 解析回来后能验证通过即可
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
	// 来自 the utility toolkit 的固定测试 token
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

func TestCreateNone(t *testing.T) {
	j := New().
		SetPayload("sub", "1234567890").
		SetPayload("name", "looly").
		SetPayload("admin", true).
		SetSigner(NoneSigner())

	tok, err := j.Sign()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	parts := strings.Split(tok, ".")
	if len(parts) != 3 || parts[2] != "" {
		t.Fatalf("none signature should be empty: %q", tok)
	}
	parsed, err := Of(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !parsed.SetSigner(NoneSigner()).Verify() {
		t.Fatalf("verify failed for none")
	}
}

func TestVerifyRejectsNoneWithoutExplicitSigner(t *testing.T) {
	tok, err := New().SetSigner(NoneSigner()).SetPayload("sub", "public").Sign()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	parsed, err := Of(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.Verify() {
		t.Fatal("Verify should reject alg=none without explicit NoneSigner")
	}
	if parsed.Validate(0) {
		t.Fatal("Validate should reject alg=none without explicit NoneSigner")
	}
	if !parsed.VerifyWith(NoneSigner()) {
		t.Fatal("VerifyWith(NoneSigner) should still support explicit none tokens")
	}
}

func TestNeedSigner(t *testing.T) {
	j := New().SetPayload("sub", "x")
	if _, err := j.Sign(); err == nil {
		t.Fatalf("expected error when no signer set")
	}
}

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

	noneToken, err := CreateTokenWithOptions(WithTokenPayload(map[string]any{"scope": "public"}), WithTokenSigner(NoneSigner()))
	if err != nil {
		t.Fatalf("CreateTokenWithOptions with signer: %v", err)
	}
	if !strings.HasSuffix(noneToken, ".") {
		t.Fatalf("none token should have empty signature: %q", noneToken)
	}
}

func TestAlgMismatch(t *testing.T) {
	// alg=none 时使用非 None signer 应失败
	tok, _ := New().SetSigner(NoneSigner()).SetPayload("a", 1).Sign()
	j, _ := Of(tok)
	hs, _ := NewHMACSigner(AlgHS256, []byte("k"))
	if j.VerifyWith(hs) {
		t.Fatalf("none alg with HS256 signer should fail")
	}
	// alg=HS256 时使用 None signer 应失败
	tok2, _ := New().SetKey([]byte("k")).SetPayload("a", 1).Sign()
	j2, _ := Of(tok2)
	if j2.VerifyWith(NoneSigner()) {
		t.Fatalf("HS256 token with None signer should fail")
	}
}
