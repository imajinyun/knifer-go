package jwt

import (
	"strings"
	"testing"
)

// Simplified utility toolkit-jwt JWTUtilTest.

func TestUtil_CreateAndVerify(t *testing.T) {
	key := []byte("1234567890")
	payload := map[string]any{
		"sub":  "1234567890",
		"name": "looly",
	}
	tok, err := CreateToken(payload, key)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !Verify(tok, key) {
		t.Fatalf("verify failed")
	}
	if Verify(tok, []byte("wrong")) {
		t.Fatalf("verify should fail with wrong key")
	}
}

func TestCreateTokenWithOptionsStrictKey(t *testing.T) {
	if token, err := CreateTokenWithOptions(
		WithTokenPayload(map[string]any{PayloadSubject: "alice"}),
		WithTokenKey([]byte("weak")),
		WithTokenStrictKey(),
	); err == nil || token != "" {
		t.Fatalf("CreateTokenWithOptions strict weak key token=%q err=%v, want error", token, err)
	}

	strong := []byte(strings.Repeat("k", MinHMACKeyBytesHS256))
	token, err := CreateTokenWithOptions(
		WithTokenPayload(map[string]any{PayloadSubject: "alice"}),
		WithTokenKey(strong),
		WithTokenStrictKey(),
	)
	if err != nil {
		t.Fatalf("CreateTokenWithOptions strict strong key: %v", err)
	}
	if token == "" {
		t.Fatal("CreateTokenWithOptions strict strong key returned empty token")
	}
}
