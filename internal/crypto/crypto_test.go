package crypto

import (
	"bytes"
	stdcrypto "crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"math/big"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
)

func TestDigestAndHMAC(t *testing.T) {
	if got := MD5Hex([]byte("hello")); got != "5d41402abc4b2a76b9719d911017c592" {
		t.Fatalf("MD5Hex() = %s", got)
	}
	if got := hex.EncodeToString(MD5([]byte("hello"))); got != "5d41402abc4b2a76b9719d911017c592" {
		t.Fatalf("MD5() = %s", got)
	}
	if got := hex.EncodeToString(SHA1([]byte("hello"))); got != "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d" {
		t.Fatalf("SHA1() = %s", got)
	}
	if got := SHA224Hex([]byte("hello")); got != "ea09ae9cc6768c50fcee903ed054556e5bfc8347907f12598aa24193" {
		t.Fatalf("SHA224Hex() = %s", got)
	}
	if got := SHA256Hex([]byte("hello")); got != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("SHA256Hex() = %s", got)
	}
	if got := SHA384Hex([]byte("hello")); got != "59e1748777448c69de6b800d7a33bbfb9ff1b463e44354c3553bcdb9c666fa90125a3c79f90397bdf5f6a13de828684f" {
		t.Fatalf("SHA384Hex() = %s", got)
	}
	if got := HMACSHA256Hex([]byte("key"), []byte("hello")); got == "" {
		t.Fatal("HMACSHA256Hex() is empty")
	}
	mac := HMACBytes(sha256.New, []byte("key"), []byte("hello"))
	if !HMACEqual(mac, HMACBytes(sha256.New, []byte("key"), []byte("hello"))) {
		t.Fatal("HMACEqual() returned false for identical MAC values")
	}
	if !ConstantTimeEqual([]byte("same"), []byte("same")) || ConstantTimeEqual([]byte("same"), []byte("diff")) {
		t.Fatal("ConstantTimeEqual() returned unexpected result")
	}
}

func TestPBKDF2AndSignParams(t *testing.T) {
	key, err := PBKDF2SHA1([]byte("password"), []byte("salt"), 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(key); got != "0c60c80f961f0e71f3a9b524af6012062fe037a6" {
		t.Fatalf("PBKDF2SHA1() = %s", got)
	}

	key, err = PBKDF2SHA256([]byte("password"), []byte("salt"), 1, 32)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(key); got != "120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b" {
		t.Fatalf("PBKDF2SHA256() = %s", got)
	}
	if _, err := PBKDF2([]byte("password"), []byte("salt"), 0, 32, sha1.New); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("PBKDF2 invalid iterations error = %v", err)
	}
	if _, err := PBKDF2([]byte("password"), []byte("salt"), 0, 32, sha1.New); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("PBKDF2 invalid iterations error = %v, want ErrCodeInvalidInput", err)
	}

	params := map[string]any{"b": 2, "a": 1, "skip": nil}
	if got := SignParams(params, MD5Hex, "&", "=", true, "secret"); got != MD5Hex([]byte("a=1&b=2&secret")) {
		t.Fatalf("SignParams() = %s", got)
	}
	if got := SignParamsMD5(map[string]any{"b": 2, "a": 1}, "z"); got != MD5Hex([]byte("a1b2z")) {
		t.Fatalf("SignParamsMD5() = %s", got)
	}
}

func TestAESCBCAndGCM(t *testing.T) {
	key := []byte("1234567890123456")
	iv := []byte("abcdefghijklmnop")
	plain := []byte("hello crypto")
	cipherText, err := AESEncryptCBC(plain, key, iv)
	if err != nil {
		t.Fatal(err)
	}
	out, err := AESDecryptCBC(cipherText, key, iv)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("AESDecryptCBC() = %q", out)
	}

	nonce := []byte("123456789012")
	cipherText, err = AESEncryptGCM(plain, key, nonce, nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err = AESDecryptGCM(cipherText, key, nonce, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("AESDecryptGCM() = %q", out)
	}
}

func TestSymmetricModesRoundTrip(t *testing.T) {
	key := []byte("1234567890123456")
	iv := []byte("abcdefghijklmnop")
	plain := []byte("block and stream modes")

	modeCases := []struct {
		name    string
		encrypt func([]byte, []byte, []byte) ([]byte, error)
		decrypt func([]byte, []byte, []byte) ([]byte, error)
	}{
		{"CTR", AESEncryptCTR, AESDecryptCTR},
		{"CFB", AESEncryptCFB, AESDecryptCFB},
		{"OFB", AESEncryptOFB, AESDecryptOFB},
	}
	for _, tc := range modeCases {
		t.Run(tc.name, func(t *testing.T) {
			cipherText, err := tc.encrypt(plain, key, iv)
			if err != nil {
				t.Fatal(err)
			}
			out, err := tc.decrypt(cipherText, key, iv)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out, plain) {
				t.Fatalf("decrypt() = %q", out)
			}
		})
	}

	ecbCipherText, err := AESEncryptECB(plain, key)
	if err != nil {
		t.Fatal(err)
	}
	out, err := AESDecryptECB(ecbCipherText, key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("AESDecryptECB() = %q", out)
	}
	if _, err := AESDecryptECB([]byte("short"), key); !errors.Is(err, ErrInvalidCipherText) {
		t.Fatalf("AESDecryptECB invalid data error = %v", err)
	}

	desCipherText, err := DESEncryptCBC(plain, []byte("12345678"), []byte("abcdefgh"))
	if err != nil {
		t.Fatal(err)
	}
	out, err = DESDecryptCBC(desCipherText, []byte("12345678"), []byte("abcdefgh"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("DESDecryptCBC() = %q", out)
	}

	tripleKey := []byte("123456789012345678901234")
	tripleCipherText, err := TripleDESEncryptCBC(plain, tripleKey, []byte("abcdefgh"))
	if err != nil {
		t.Fatal(err)
	}
	out, err = TripleDESDecryptCBC(tripleCipherText, tripleKey, []byte("abcdefgh"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("TripleDESDecryptCBC() = %q", out)
	}

	rc4CipherText, err := RC4Crypt(plain, []byte("stream-key"))
	if err != nil {
		t.Fatal(err)
	}
	out, err = RC4Crypt(rc4CipherText, []byte("stream-key"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("RC4Crypt() = %q", out)
	}
}

func TestVigenereAndXXTEA(t *testing.T) {
	plain := "printable text 123"
	cipherText, err := VigenereEncrypt(plain, "key")
	if err != nil {
		t.Fatal(err)
	}
	out, err := VigenereDecrypt(cipherText, "key")
	if err != nil {
		t.Fatal(err)
	}
	if out != plain {
		t.Fatalf("VigenereDecrypt() = %q", out)
	}
	if _, err := VigenereEncrypt(plain, ""); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("VigenereEncrypt empty key error = %v", err)
	}

	data := []byte("lightweight block cipher payload")
	xxteaCipherText := XXTEAEncrypt(data, []byte("secret"))
	xxteaOut, err := XXTEADecrypt(xxteaCipherText, []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(xxteaOut, data) {
		t.Fatalf("XXTEADecrypt() = %q", xxteaOut)
	}
	if _, err := XXTEADecrypt([]byte{1, 2, 3, 4}, []byte("secret")); !errors.Is(err, ErrInvalidCipherText) {
		t.Fatalf("XXTEADecrypt invalid data error = %v", err)
	}
}

func TestRSAOAEPAndPEM(t *testing.T) {
	priv, err := GenerateRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("hello")
	cipherText, err := RSAEncryptOAEP(plain, &priv.PublicKey, nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err := RSADecryptOAEP(cipherText, priv, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("RSADecryptOAEP() = %q", out)
	}
	parsed, err := ParseRSAPrivateKeyPEM(PrivateKeyToPEM(priv))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.N.Cmp(priv.N) != 0 {
		t.Fatal("parsed private key mismatch")
	}

	pkcs8, err := PrivateKeyToPKCS8PEM(priv)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err = ParseRSAPrivateKeyPEM(pkcs8)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.N.Cmp(priv.N) != 0 {
		t.Fatal("parsed PKCS#8 private key mismatch")
	}
	pub, err := ParseRSAPublicKeyPEM(PublicKeyToPKCS1PEM(&priv.PublicKey))
	if err != nil {
		t.Fatal(err)
	}
	if pub.N.Cmp(priv.N) != 0 {
		t.Fatal("parsed PKCS#1 public key mismatch")
	}
}

func TestRSAPKCS1PSSAndCertificate(t *testing.T) {
	priv, err := GenerateRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("signed payload")
	cipherText, err := RSAEncryptPKCS1v15(plain, &priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	out, err := RSADecryptPKCS1v15(cipherText, priv)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("RSADecryptPKCS1v15() = %q", out)
	}

	digest := sha256.Sum256(plain)
	sig, err := RSASignPKCS1v15(priv, stdcrypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	if err := RSAVerifyPKCS1v15(&priv.PublicKey, stdcrypto.SHA256, digest[:], sig); err != nil {
		t.Fatal(err)
	}
	pssSig, err := RSASignPSS(priv, stdcrypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	if err := RSAVerifyPSS(&priv.PublicKey, stdcrypto.SHA256, digest[:], pssSig); err != nil {
		t.Fatal(err)
	}
	quickSig, err := SignSHA256WithRSA(plain, priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifySHA256WithRSA(plain, quickSig, &priv.PublicKey); err != nil {
		t.Fatal(err)
	}

	certPEM := newTestCertificatePEM(t, priv)
	cert, err := ParseX509CertificatePEM(certPEM)
	if err != nil {
		t.Fatal(err)
	}
	if cert.Subject.CommonName != "go-knifer-test" {
		t.Fatalf("certificate subject = %s", cert.Subject.CommonName)
	}
	certPub, err := PublicKeyFromCertificatePEM(certPEM)
	if err != nil {
		t.Fatal(err)
	}
	if certPub.N.Cmp(priv.N) != 0 {
		t.Fatal("certificate public key mismatch")
	}
}

func newTestCertificatePEM(t *testing.T, priv any) []byte {
	t.Helper()
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "go-knifer-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	rsaPriv := priv.(*rsa.PrivateKey)
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &rsaPriv.PublicKey, rsaPriv)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}
