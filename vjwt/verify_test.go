package vjwt_test

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vjwt"
)

func TestVerifyRejectsNoneToken(t *testing.T) {
	token := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJwdWJsaWMifQ."

	if vjwt.Verify(token, []byte("ignored")) {
		t.Fatal("Verify should reject alg=none tokens")
	}
	if vjwt.VerifyStrict(token, []byte("ignored")) {
		t.Fatal("VerifyStrict should reject alg=none tokens")
	}
	if _, err := vjwt.CreateSigner("none", []byte("ignored")); !errors.Is(err, knifer.ErrCodeUnsupported) {
		t.Fatalf("CreateSigner(none) error = %v, want unsupported", err)
	}
}

func TestVerifyWithSignerRejectsAlgorithmMismatch(t *testing.T) {
	key := []byte("secret")
	token, err := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenAlgorithm(vjwt.JWTAlgHS256),
		vjwt.WithTokenKey(key),
		vjwt.WithTokenPayload(map[string]any{vjwt.JWTPayloadSubject: "alice"}),
	)
	if err != nil {
		t.Fatalf("CreateTokenWithOptions: %v", err)
	}

	wrongAlgSigner, err := vjwt.NewHMACSigner(vjwt.JWTAlgHS512, key)
	if err != nil {
		t.Fatalf("NewHMACSigner: %v", err)
	}
	if vjwt.VerifyWithSigner(token, wrongAlgSigner) {
		t.Fatal("VerifyWithSigner should reject signer/token algorithm mismatch")
	}
	if err := vjwt.ValidateAlgorithm(token, wrongAlgSigner); err == nil {
		t.Fatal("ValidateAlgorithm should reject signer/token algorithm mismatch")
	}
}
