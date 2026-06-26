package vcrypto_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/imajinyun/knifer-go/vcrypto"
)

func TestKDFAndParamSigning(t *testing.T) {
	key, err := vcrypto.PBKDF2SHA256([]byte("password"), []byte("salt"), 1, 32)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(key); got != "120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b" {
		t.Fatalf("PBKDF2SHA256() = %s", got)
	}
	params := map[string]any{"b": 2, "a": 1, "skip": nil}
	if got := vcrypto.SignParams(params, vcrypto.SHA256HexBytes, "&", "=", true, "secret"); got != vcrypto.SHA256Hex("a=1&b=2&secret") {
		t.Fatalf("SignParams() = %s", got)
	}
	if got := vcrypto.SignParamsSHA256(map[string]any{"b": 2, "a": 1}, "z"); got != vcrypto.SHA256Hex("a1b2z") {
		t.Fatalf("SignParamsSHA256() = %s", got)
	}
}

func TestPBKDF2Errors(t *testing.T) {
	if _, err := vcrypto.PBKDF2([]byte("password"), []byte("salt"), 0, 32, sha256.New); !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("PBKDF2 invalid iterations error = %v", err)
	}
	if _, err := vcrypto.PBKDF2([]byte("password"), []byte("salt"), 1, 0, sha256.New); !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("PBKDF2 invalid key length error = %v", err)
	}
	if _, err := vcrypto.PBKDF2([]byte("password"), []byte("salt"), 1, 32, nil); !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("PBKDF2 nil hash error = %v", err)
	}
}
