package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestPBKDF2(t *testing.T) {
	key, err := PBKDF2SHA256([]byte("password"), []byte("salt"), 1, 32)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(key); got != "120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b" {
		t.Fatalf("PBKDF2SHA256() = %s", got)
	}
	if _, err := PBKDF2([]byte("password"), []byte("salt"), 0, 32, sha256.New); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("PBKDF2 invalid iterations error = %v", err)
	}
	if _, err := PBKDF2([]byte("password"), []byte("salt"), 0, 32, sha256.New); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("PBKDF2 invalid iterations error = %v, want ErrCodeInvalidInput", err)
	}
}

func TestPBKDF2ErrorContracts(t *testing.T) {
	password := []byte("password")
	salt := []byte("salt")
	for _, tt := range []struct {
		name       string
		iterations int
		keyLen     int
		fn         func() hash.Hash
	}{
		{name: "zero iterations", iterations: 0, keyLen: 32, fn: sha256.New},
		{name: "negative iterations", iterations: -1, keyLen: 32, fn: sha256.New},
		{name: "zero key length", iterations: 1, keyLen: 0, fn: sha256.New},
		{name: "nil hash", iterations: 1, keyLen: 32, fn: nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PBKDF2(password, salt, tt.iterations, tt.keyLen, tt.fn)
			if !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("PBKDF2 error = %v, want invalid key/input", err)
			}
		})
	}
}
