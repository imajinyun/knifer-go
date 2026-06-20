package crypto

import (
	"bytes"
	"crypto/cipher"
	"errors"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

type stubBlock struct{}

func (stubBlock) BlockSize() int          { return 8 }
func (stubBlock) Encrypt(dst, src []byte) { copy(dst, src) }
func (stubBlock) Decrypt(dst, src []byte) { copy(dst, src) }

func TestAESGCMOptions(t *testing.T) {
	if WithGCMTagSize(16) == nil {
		t.Fatal("WithGCMTagSize() = nil")
	}
	if WithGCMBlockFactory(nil) == nil {
		t.Fatal("WithGCMBlockFactory() = nil")
	}
}

func TestAESSealGCM(t *testing.T) {
	key := []byte("1234567890123456")
	plain := []byte("hello")
	nonce, sealed, err := AESSealGCM(plain, key, []byte("aad"))
	if err != nil {
		t.Fatalf("AESSealGCM() error = %v", err)
	}
	if len(nonce) == 0 || len(sealed) == 0 {
		t.Fatal("AESSealGCM() returned empty data")
	}
	opened, err := AESOpenGCMWithOptions(sealed, key, nonce, []byte("aad"))
	if err != nil {
		t.Fatalf("AESOpenGCMWithOptions() error = %v", err)
	}
	if string(opened) != "hello" {
		t.Fatalf("AESOpenGCMWithOptions() = %q", opened)
	}
}

func TestAESGCM(t *testing.T) {
	key := []byte("1234567890123456")
	plain := []byte("hello crypto")
	nonce := []byte("123456789012")
	generatedNonce, sealed, err := AESSealGCMWithOptions(
		plain,
		key,
		[]byte("aad"),
		WithGCMRandomOptions(WithRandomReader(bytes.NewReader([]byte("abcdefghijkl")))),
	)
	if err != nil {
		t.Fatalf("AESSealGCMWithOptions() error = %v", err)
	}
	if !bytes.Equal(generatedNonce, []byte("abcdefghijkl")) {
		t.Fatalf("AESSealGCMWithOptions() nonce = %q", generatedNonce)
	}
	opened, err := AESOpenGCM(sealed, key, generatedNonce, []byte("aad"))
	if err != nil {
		t.Fatalf("AESOpenGCM() error = %v", err)
	}
	if !bytes.Equal(opened, plain) {
		t.Fatalf("AESOpenGCM() = %q", opened)
	}

	cipherText, err := AESEncryptGCM(plain, key, nonce, nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err := AESDecryptGCM(cipherText, key, nonce, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("AESDecryptGCM() = %q", out)
	}

	customNonce := []byte("1234567890123456")
	cipherText, err = AESEncryptGCMWithOptions(plain, key, customNonce, []byte("aad"), WithGCMNonceSize(len(customNonce)))
	if err != nil {
		t.Fatalf("AESEncryptGCMWithOptions() error = %v", err)
	}
	out, err = AESDecryptGCMWithOptions(cipherText, key, customNonce, []byte("aad"), WithGCMNonceSize(len(customNonce)))
	if err != nil {
		t.Fatalf("AESDecryptGCMWithOptions() error = %v", err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("AESDecryptGCMWithOptions() = %q", out)
	}
}

func TestAESGCMProviderAndSecurityBoundaries(t *testing.T) {
	key := []byte("1234567890123456")
	plain := []byte("sensitive payload")
	nonce := []byte("123456789012")
	blockErr := errors.New("block factory failed")

	if _, _, err := AESSealGCMWithOptions(plain, []byte("short"), nil); !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("AESSealGCMWithOptions invalid key err = %v", err)
	}
	if _, _, err := AESSealGCMWithOptions(plain, key, nil, WithGCMBlockFactory(func([]byte) (cipher.Block, error) {
		return nil, blockErr
	})); !errors.Is(err, blockErr) {
		t.Fatalf("AESSealGCMWithOptions block factory err = %v, want blockErr", err)
	}
	if _, _, err := AESSealGCMWithOptions(plain, key, nil, WithGCMNonceSize(12), WithGCMTagSize(16)); err == nil {
		t.Fatal("AESSealGCMWithOptions should reject simultaneous nonce and tag sizes")
	}
	if _, _, err := AESSealGCMWithOptions(plain, key, nil, WithGCMBlockFactory(func([]byte) (cipher.Block, error) {
		return stubBlock{}, nil
	})); err == nil {
		t.Fatal("AESSealGCMWithOptions should reject non-AES block")
	}

	cipherText, err := AESEncryptGCMWithOptions(plain, key, nonce, []byte("aad"), nil, WithGCMBlockFactory(nil))
	if err != nil {
		t.Fatalf("AESEncryptGCMWithOptions nil options/factory = %v", err)
	}
	if _, err := AESDecryptGCMWithOptions(cipherText, key, []byte("bad"), []byte("aad")); !errors.Is(err, ErrInvalidIV) {
		t.Fatalf("AESDecryptGCMWithOptions bad nonce err = %v, want invalid iv", err)
	}
	if _, err := AESDecryptGCMWithOptions(cipherText, key, nonce, []byte("wrong aad")); !errors.Is(err, ErrInvalidCipherText) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("AESDecryptGCMWithOptions auth err = %v, want invalid cipher text/input", err)
	}
	if _, err := AESEncryptGCMWithOptions(plain, key, []byte("bad"), nil); !errors.Is(err, ErrInvalidIV) {
		t.Fatalf("AESEncryptGCMWithOptions bad nonce err = %v, want invalid iv", err)
	}
}

func TestAESSealGCMDeterministicNonceReader(t *testing.T) {
	key := []byte("1234567890123456")
	reader := bytes.NewReader([]byte("abcdefghijklmnopqrstuvwx"))
	firstNonce, firstCipher, err := AESSealGCMWithOptions([]byte("one"), key, nil, WithGCMRandomOptions(WithRandomReader(reader)))
	if err != nil {
		t.Fatalf("first AESSealGCMWithOptions = %v", err)
	}
	secondNonce, secondCipher, err := AESSealGCMWithOptions([]byte("two"), key, nil, WithGCMRandomOptions(WithRandomReader(reader)))
	if err != nil {
		t.Fatalf("second AESSealGCMWithOptions = %v", err)
	}
	if !bytes.Equal(firstNonce, []byte("abcdefghijkl")) || !bytes.Equal(secondNonce, []byte("mnopqrstuvwx")) {
		t.Fatalf("generated nonces = %q/%q", firstNonce, secondNonce)
	}
	if bytes.Equal(firstCipher, secondCipher) {
		t.Fatal("different plaintext/nonces should not produce identical ciphertext")
	}
}
