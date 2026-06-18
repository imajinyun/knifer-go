package vcrypto_test

import (
	"bytes"
	"crypto/cipher"
	"errors"
	"io"
	"testing"

	"github.com/imajinyun/go-knifer/vcrypto"
)

func TestAESRoundTripAndErrors(t *testing.T) {
	key, err := vcrypto.GenAESKey(16)
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 16 {
		t.Fatalf("GenAESKey len = %d", len(key))
	}
	if _, err := vcrypto.GenAESKey(15); !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("GenAESKey invalid error = %v", err)
	}
	optionKey, err := vcrypto.GenAESKeyWithOptions(16, vcrypto.WithRandomReader(bytes.NewReader(bytes.Repeat([]byte{0x42}, 16))))
	if err != nil {
		t.Fatalf("GenAESKeyWithOptions error = %v", err)
	}
	if !bytes.Equal(optionKey, bytes.Repeat([]byte{0x42}, 16)) {
		t.Fatalf("GenAESKeyWithOptions = %x", optionKey)
	}
	randomBytes, err := vcrypto.RandomBytesWithOptions(3, vcrypto.WithRandomReader(bytes.NewReader([]byte{1, 2, 3})))
	if err != nil || !bytes.Equal(randomBytes, []byte{1, 2, 3}) {
		t.Fatalf("RandomBytesWithOptions = %v, %v", randomBytes, err)
	}
	plain := []byte("crypto facade")
	nonce := []byte("123456789012")
	cipherText, err := vcrypto.AESEncryptGCM(plain, key, nonce, []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := vcrypto.AESDecryptGCM(cipherText, key, nonce, []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("AESDecryptGCM() = %q", out)
	}
	customNonce := []byte("1234567890123456")
	cipherText, err = vcrypto.AESEncryptGCMWithOptions(plain, key, customNonce, []byte("aad"), vcrypto.WithGCMNonceSize(len(customNonce)))
	if err != nil {
		t.Fatalf("AESEncryptGCMWithOptions() error = %v", err)
	}
	out, err = vcrypto.AESDecryptGCMWithOptions(cipherText, key, customNonce, []byte("aad"), vcrypto.WithGCMNonceSize(len(customNonce)))
	if err != nil {
		t.Fatalf("AESDecryptGCMWithOptions() error = %v", err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("AESDecryptGCMWithOptions() = %q", out)
	}
}

func TestAdditionalAESGCMAndRandomErrors(t *testing.T) {
	key := bytes.Repeat([]byte{0x11}, 16)
	plain := []byte("authenticated payload")
	nonce, cipherText, err := vcrypto.AESSealGCMWithOptions(plain, key, []byte("aad"), vcrypto.WithGCMRandomOptions(vcrypto.WithRandomReader(bytes.NewReader(bytes.Repeat([]byte{0x22}, 12)))))
	if err != nil {
		t.Fatalf("AESSealGCMWithOptions: %v", err)
	}
	if !bytes.Equal(nonce, bytes.Repeat([]byte{0x22}, 12)) {
		t.Fatalf("AESSealGCMWithOptions nonce = %x", nonce)
	}
	out, err := vcrypto.AESOpenGCM(cipherText, key, nonce, []byte("aad"))
	if err != nil || !bytes.Equal(out, plain) {
		t.Fatalf("AESOpenGCM = %q, %v", out, err)
	}
	out, err = vcrypto.AESOpenGCMWithOptions(cipherText, key, nonce, []byte("aad"))
	if err != nil || !bytes.Equal(out, plain) {
		t.Fatalf("AESOpenGCMWithOptions = %q, %v", out, err)
	}
	if _, _, err := vcrypto.AESSealGCMWithOptions(plain, key, nil, vcrypto.WithGCMNonceSize(12), vcrypto.WithGCMTagSize(16)); err == nil {
		t.Fatal("AESSealGCMWithOptions with nonce and tag size error = nil")
	}
	if _, err := vcrypto.AESEncryptGCMWithOptions(plain, key, nonce, nil, vcrypto.WithGCMBlockFactory(func([]byte) (cipher.Block, error) {
		return nil, errors.New("block factory failed")
	})); err == nil {
		t.Fatal("AESEncryptGCMWithOptions block factory error = nil")
	}
	if _, _, err := vcrypto.AESSealGCMWithOptions(plain, key, nil, vcrypto.WithGCMRandomOptions(vcrypto.WithRandomReader(io.LimitReader(bytes.NewReader(nil), 0)))); err == nil {
		t.Fatal("AESSealGCMWithOptions random reader error = nil")
	}
	if _, err := vcrypto.AESDecryptGCM(cipherText, key, []byte("bad"), nil); !errors.Is(err, vcrypto.ErrInvalidIV) {
		t.Fatalf("AESDecryptGCM invalid nonce error = %v", err)
	}
	if _, err := vcrypto.AESDecryptGCM(cipherText, key, nonce, []byte("wrong aad")); err == nil {
		t.Fatal("AESDecryptGCM wrong aad error = nil")
	}
	if _, err := vcrypto.RandomBytes(-1); !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("RandomBytes negative error = %v", err)
	}
	if _, err := vcrypto.RandomBytesWithOptions(2, vcrypto.WithRandomReader(bytes.NewReader([]byte{1}))); err == nil {
		t.Fatal("RandomBytesWithOptions short reader error = nil")
	}
}

func TestFacadeAESSealGCM(t *testing.T) {
	key := bytes.Repeat([]byte{0x11}, 16)
	plain := []byte("seal plain")
	nonce, ct, err := vcrypto.AESSealGCM(plain, key, []byte("aad"))
	if err != nil {
		t.Fatalf("AESSealGCM: %v", err)
	}
	if len(nonce) == 0 || len(ct) == 0 {
		t.Fatal("AESSealGCM returned empty nonce or ciphertext")
	}
	out, err := vcrypto.AESOpenGCM(ct, key, nonce, []byte("aad"))
	if err != nil || !bytes.Equal(out, plain) {
		t.Fatalf("AESOpenGCM = %q, %v", out, err)
	}
}

func TestFacadeAESGCMValidationErrorClassification(t *testing.T) {
	key := bytes.Repeat([]byte{0x11}, 16)
	nonce := bytes.Repeat([]byte{0x22}, 12)
	plain := []byte("payload")
	cipherText, err := vcrypto.AESEncryptGCM(plain, key, nonce, nil)
	if err != nil {
		t.Fatalf("AESEncryptGCM: %v", err)
	}

	if _, err := vcrypto.AESEncryptGCM(plain, []byte("short"), nonce, nil); !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("AESEncryptGCM invalid key error = %v", err)
	}
	if _, err := vcrypto.AESOpenGCM(cipherText, key, []byte("short"), nil); !errors.Is(err, vcrypto.ErrInvalidIV) {
		t.Fatalf("AESOpenGCM invalid nonce error = %v", err)
	}
	if _, err := vcrypto.AESOpenGCM(cipherText[:len(cipherText)-1], key, nonce, nil); err == nil {
		t.Fatal("AESOpenGCM truncated tag error = nil")
	}
}
