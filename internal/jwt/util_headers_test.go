package jwt

import "testing"

func TestUtil_CreateTokenWithHeaders(t *testing.T) {
	headers := map[string]any{HeaderKeyID: "kid-1"}
	payload := map[string]any{"a": 1}
	tok, err := CreateTokenWithHeaders(headers, payload, []byte("k"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	j, err := ParseToken(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if j.Header(HeaderKeyID) != "kid-1" {
		t.Fatalf("kid: %v", j.Header(HeaderKeyID))
	}
}
