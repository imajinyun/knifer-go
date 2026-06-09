package vcrypto_test

import (
	"bytes"
	stdcrypto "crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vcrypto"
)

func TestDigestAndHMAC(t *testing.T) {
	if got := vcrypto.SHA256Hex("hello"); got != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("SHA256Hex() = %s", got)
	}
	if got := vcrypto.SHA512Hex("hello"); got == "" {
		t.Fatal("SHA512Hex() is empty")
	}
	if got := vcrypto.HMACSHA256Hex([]byte("key"), []byte("hello")); got != "9307b3b915efb5171ff14d8cb55fbcc798c6c0ef1456d66ded1a6aa723a58b7b" {
		t.Fatalf("HMACSHA256Hex() = %s", got)
	}
	if got := vcrypto.SHA224Hex([]byte("hello")); got != "ea09ae9cc6768c50fcee903ed054556e5bfc8347907f12598aa24193" {
		t.Fatalf("SHA224Hex() = %s", got)
	}
	if !vcrypto.HMACEqual(vcrypto.HMACBytes(sha256.New, []byte("key"), []byte("hello")), vcrypto.HMACBytes(sha256.New, []byte("key"), []byte("hello"))) {
		t.Fatal("HMACEqual() returned false for identical MAC values")
	}
	if !vcrypto.ConstantTimeEqual([]byte("same"), []byte("same")) || vcrypto.ConstantTimeEqual([]byte("same"), []byte("diff")) {
		t.Fatal("ConstantTimeEqual() returned unexpected result")
	}
}

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

func TestAESRoundTripAndErrors(t *testing.T) {
	key, err := vcrypto.GenerateAESKey(16)
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 16 {
		t.Fatalf("GenerateAESKey len = %d", len(key))
	}
	if _, err := vcrypto.GenerateAESKey(15); !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("GenerateAESKey invalid error = %v", err)
	}
	optionKey, err := vcrypto.GenerateAESKeyWithOptions(16, vcrypto.WithRandomReader(bytes.NewReader(bytes.Repeat([]byte{0x42}, 16))))
	if err != nil {
		t.Fatalf("GenerateAESKeyWithOptions error = %v", err)
	}
	if !bytes.Equal(optionKey, bytes.Repeat([]byte{0x42}, 16)) {
		t.Fatalf("GenerateAESKeyWithOptions = %x", optionKey)
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

func TestErrorContract(t *testing.T) {
	if err := vcrypto.ValidateAESKey([]byte("too-short")); !errors.Is(err, knifer.ErrCodeInvalidInput) || !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("ValidateAESKey error should match root code and domain sentinel: %v", err)
	}
	if err := vcrypto.ValidateAESIV([]byte("bad")); !errors.Is(err, knifer.ErrCodeInvalidInput) || !errors.Is(err, vcrypto.ErrInvalidIV) {
		t.Fatalf("ValidateAESIV error should match root code and domain sentinel: %v", err)
	}
	if err := vcrypto.ValidateAESGCMNonce([]byte("bad")); !errors.Is(err, knifer.ErrCodeInvalidInput) || !errors.Is(err, vcrypto.ErrInvalidIV) {
		t.Fatalf("ValidateAESGCMNonce error should match root code and domain sentinel: %v", err)
	}
	if code, ok := knifer.CodeOf(vcrypto.ErrInvalidCipherText); !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(ErrInvalidCipherText) = %q, %v", code, ok)
	}
	if err := vcrypto.ValidateAESKey([]byte("1234567890123456")); err != nil {
		t.Fatalf("ValidateAESKey(valid) = %v", err)
	}
	if err := vcrypto.ValidateAESIV([]byte("1234567890123456")); err != nil {
		t.Fatalf("ValidateAESIV(valid) = %v", err)
	}
	if err := vcrypto.ValidateAESGCMNonce([]byte("123456789012")); err != nil {
		t.Fatalf("ValidateAESGCMNonce(valid) = %v", err)
	}
}

func TestRSAAndPEM(t *testing.T) {
	priv, err := vcrypto.GenerateRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM, err := vcrypto.PublicKeyToPEM(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := vcrypto.ParseRSAPublicKeyPEM(pubPEM)
	if err != nil {
		t.Fatal(err)
	}
	parsedPriv, err := vcrypto.ParseRSAPrivateKeyPEM(vcrypto.PrivateKeyToPEM(priv))
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("rsa message")
	cipherText, err := vcrypto.RSAEncryptOAEP(plain, pub, nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err := vcrypto.RSADecryptOAEP(cipherText, parsedPriv, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("RSADecryptOAEP() = %q", out)
	}

	digest := sha256.Sum256(plain)
	pssSig, err := vcrypto.RSASignPSS(parsedPriv, stdcrypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	if err := vcrypto.RSAVerifyPSS(pub, stdcrypto.SHA256, digest[:], pssSig); err != nil {
		t.Fatal(err)
	}
	pssOptions := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: stdcrypto.SHA256}
	pssSig, err = vcrypto.RSASignPSSWithOptions(parsedPriv, stdcrypto.SHA256, digest[:], vcrypto.WithRSAPSSOptions(pssOptions))
	if err != nil {
		t.Fatal(err)
	}
	if err := vcrypto.RSAVerifyPSSWithOptions(pub, stdcrypto.SHA256, digest[:], pssSig, vcrypto.WithRSAPSSOptions(pssOptions)); err != nil {
		t.Fatal(err)
	}
	oaepCipherText, err := vcrypto.RSAEncryptOAEPWithOptions(plain, pub, []byte("label"), vcrypto.WithRSAOAEPHash(sha256.New))
	if err != nil {
		t.Fatal(err)
	}
	oaepOut, err := vcrypto.RSADecryptOAEPWithOptions(oaepCipherText, parsedPriv, []byte("label"), vcrypto.WithRSAOAEPHash(sha256.New))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(oaepOut, plain) {
		t.Fatalf("RSADecryptOAEPWithOptions() = %q", oaepOut)
	}
	quickSig, err := vcrypto.SignSHA256WithRSA(plain, parsedPriv)
	if err != nil {
		t.Fatal(err)
	}
	if err := vcrypto.VerifySHA256WithRSA(plain, quickSig, pub); err != nil {
		t.Fatal(err)
	}
	pkcs8, err := vcrypto.PrivateKeyToPKCS8PEM(parsedPriv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := vcrypto.ParseRSAPrivateKeyPEM(pkcs8); err != nil {
		t.Fatal(err)
	}
	if _, err := vcrypto.ParseRSAPublicKeyPEM(vcrypto.PublicKeyToPKCS1PEM(pub)); err != nil {
		t.Fatal(err)
	}
}
