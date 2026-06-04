package vcrypto_test

import (
	"bytes"
	stdcrypto "crypto"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vcrypto"
)

func TestDigestAndHMAC(t *testing.T) {
	if got := vcrypto.MD5Hex("hello"); got != "5d41402abc4b2a76b9719d911017c592" {
		t.Fatalf("MD5Hex() = %s", got)
	}
	if got := vcrypto.MD5HexBytes([]byte("hello")); got != "5d41402abc4b2a76b9719d911017c592" {
		t.Fatalf("MD5HexBytes() = %s", got)
	}
	if got := vcrypto.SHA1Hex("hello"); got != "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d" {
		t.Fatalf("SHA1Hex() = %s", got)
	}
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
	if _, err := vcrypto.PBKDF2SHA1([]byte("password"), []byte("salt"), 0, 20); !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("PBKDF2SHA1 invalid iterations error = %v", err)
	}

	params := map[string]any{"b": 2, "a": 1, "skip": nil}
	if got := vcrypto.SignParams(params, vcrypto.MD5HexBytes, "&", "=", true, "secret"); got != vcrypto.MD5Hex("a=1&b=2&secret") {
		t.Fatalf("SignParams() = %s", got)
	}
	if got := vcrypto.SignParamsSHA1(map[string]any{"b": 2, "a": 1}, "z"); got != vcrypto.SHA1Hex("a1b2z") {
		t.Fatalf("SignParamsSHA1() = %s", got)
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
	iv := []byte("1234567890123456")
	plain := []byte("crypto facade")
	cipherText, err := vcrypto.AESEncryptCBC(plain, key, iv)
	if err != nil {
		t.Fatal(err)
	}
	out, err := vcrypto.AESDecryptCBC(cipherText, key, iv)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("AESDecryptCBC() = %q", out)
	}
	if _, err := vcrypto.AESEncryptCBC(plain, key, []byte("bad")); !errors.Is(err, vcrypto.ErrInvalidIV) {
		t.Fatalf("AESEncryptCBC invalid iv error = %v", err)
	}

	nonce := []byte("123456789012")
	cipherText, err = vcrypto.AESEncryptGCM(plain, key, nonce, []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	out, err = vcrypto.AESDecryptGCM(cipherText, key, nonce, []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("AESDecryptGCM() = %q", out)
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

func TestSymmetricHelpers(t *testing.T) {
	key := []byte("1234567890123456")
	iv := []byte("abcdefghijklmnop")
	plain := []byte("block and stream facade")

	tests := []struct {
		name    string
		encrypt func([]byte, []byte, []byte) ([]byte, error)
		decrypt func([]byte, []byte, []byte) ([]byte, error)
	}{
		{"CTR", vcrypto.AESEncryptCTR, vcrypto.AESDecryptCTR},
		{"CFB", vcrypto.AESEncryptCFB, vcrypto.AESDecryptCFB},
		{"OFB", vcrypto.AESEncryptOFB, vcrypto.AESDecryptOFB},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipherText, err := tt.encrypt(plain, key, iv)
			if err != nil {
				t.Fatal(err)
			}
			out, err := tt.decrypt(cipherText, key, iv)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out, plain) {
				t.Fatalf("decrypt() = %q", out)
			}
		})
	}

	cipherText, err := vcrypto.AESEncryptECB(plain, key)
	if err != nil {
		t.Fatal(err)
	}
	out, err := vcrypto.AESDecryptECB(cipherText, key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("AESDecryptECB() = %q", out)
	}

	desCipherText, err := vcrypto.DESEncryptCBC(plain, []byte("12345678"), []byte("abcdefgh"))
	if err != nil {
		t.Fatal(err)
	}
	out, err = vcrypto.DESDecryptCBC(desCipherText, []byte("12345678"), []byte("abcdefgh"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("DESDecryptCBC() = %q", out)
	}

	tripleKey := []byte("123456789012345678901234")
	tripleCipherText, err := vcrypto.TripleDESEncryptCBC(plain, tripleKey, []byte("abcdefgh"))
	if err != nil {
		t.Fatal(err)
	}
	out, err = vcrypto.TripleDESDecryptCBC(tripleCipherText, tripleKey, []byte("abcdefgh"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("TripleDESDecryptCBC() = %q", out)
	}
}

func TestClassicCiphers(t *testing.T) {
	plain := []byte("stream payload")
	cipherText, err := vcrypto.RC4Crypt(plain, []byte("stream-key"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := vcrypto.RC4Crypt(cipherText, []byte("stream-key"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("RC4Crypt() = %q", out)
	}

	vigenereCipher, err := vcrypto.VigenereEncrypt("printable text", "key")
	if err != nil {
		t.Fatal(err)
	}
	vigenereOut, err := vcrypto.VigenereDecrypt(vigenereCipher, "key")
	if err != nil {
		t.Fatal(err)
	}
	if vigenereOut != "printable text" {
		t.Fatalf("VigenereDecrypt() = %q", vigenereOut)
	}

	xxteaCipher := vcrypto.XXTEAEncrypt([]byte("payload"), []byte("secret"))
	xxteaOut, err := vcrypto.XXTEADecrypt(xxteaCipher, []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if string(xxteaOut) != "payload" {
		t.Fatalf("XXTEADecrypt() = %q", xxteaOut)
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

	pkcs1CipherText, err := vcrypto.RSAEncryptPKCS1v15(plain, pub)
	if err != nil {
		t.Fatal(err)
	}
	out, err = vcrypto.RSADecryptPKCS1v15(pkcs1CipherText, parsedPriv)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("RSADecryptPKCS1v15() = %q", out)
	}

	digest := sha256.Sum256(plain)
	sig, err := vcrypto.RSASignPKCS1v15(parsedPriv, stdcrypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	if err := vcrypto.RSAVerifyPKCS1v15(pub, stdcrypto.SHA256, digest[:], sig); err != nil {
		t.Fatal(err)
	}
	pssSig, err := vcrypto.RSASignPSS(parsedPriv, stdcrypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	if err := vcrypto.RSAVerifyPSS(pub, stdcrypto.SHA256, digest[:], pssSig); err != nil {
		t.Fatal(err)
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
