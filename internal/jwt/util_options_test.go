package jwt

import (
	"strings"
	"testing"
)

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
