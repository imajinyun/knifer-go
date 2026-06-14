package jwt

import "testing"

func TestUtil_CreateWithSigner(t *testing.T) {
	signer := MustHMACSigner(AlgHS512, []byte("secret"))
	tok, err := CreateTokenWithSigner(map[string]any{"a": 1}, signer)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !VerifyWithSigner(tok, signer) {
		t.Fatalf("verify failed")
	}
}
