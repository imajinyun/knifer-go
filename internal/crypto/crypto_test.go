package crypto

import (
	"bytes"
	stdcrypto "crypto"
	"crypto/rand"
	"crypto/rsa"
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

func TestRandomBytesWithOptions(t *testing.T) {
	reader := bytes.NewReader([]byte{1, 2, 3, 4, 5, 6})
	b, err := RandomBytesWithOptions(4, WithRandomReader(reader))
	if err != nil {
		t.Fatalf("RandomBytesWithOptions() error = %v", err)
	}
	if !bytes.Equal(b, []byte{1, 2, 3, 4}) {
		t.Fatalf("RandomBytesWithOptions() = %v", b)
	}
	key, err := GenerateAESKeyWithOptions(16, WithRandomReader(bytes.NewReader(bytes.Repeat([]byte{0x7f}, 16))))
	if err != nil {
		t.Fatalf("GenerateAESKeyWithOptions() error = %v", err)
	}
	if !bytes.Equal(key, bytes.Repeat([]byte{0x7f}, 16)) {
		t.Fatalf("GenerateAESKeyWithOptions() = %x", key)
	}
	if _, err := GenerateAESKeyWithOptions(15, WithRandomReader(bytes.NewReader(nil))); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("GenerateAESKeyWithOptions invalid error = %v", err)
	}
}

func TestPBKDF2AndSignParams(t *testing.T) {
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

	params := map[string]any{"b": 2, "a": 1, "skip": nil}
	if got := SignParams(params, SHA256Hex, "&", "=", true, "secret"); got != SHA256Hex([]byte("a=1&b=2&secret")) {
		t.Fatalf("SignParams() = %s", got)
	}
	if got := SignParamsSHA256(map[string]any{"b": 2, "a": 1}, "z"); got != SHA256Hex([]byte("a1b2z")) {
		t.Fatalf("SignParamsSHA256() = %s", got)
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
	digest := sha256.Sum256(plain)
	pssSig, err := RSASignPSS(priv, stdcrypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	if err := RSAVerifyPSS(&priv.PublicKey, stdcrypto.SHA256, digest[:], pssSig); err != nil {
		t.Fatal(err)
	}
	pssOptions := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: stdcrypto.SHA256}
	pssSig, err = RSASignPSSWithOptions(priv, stdcrypto.SHA256, digest[:], WithRSAPSSOptions(pssOptions))
	if err != nil {
		t.Fatal(err)
	}
	if err := RSAVerifyPSSWithOptions(&priv.PublicKey, stdcrypto.SHA256, digest[:], pssSig, WithRSAPSSOptions(pssOptions)); err != nil {
		t.Fatal(err)
	}
	oaepCipherText, err := RSAEncryptOAEPWithOptions(plain, &priv.PublicKey, []byte("label"), WithRSAOAEPHash(sha256.New))
	if err != nil {
		t.Fatal(err)
	}
	oaepOut, err := RSADecryptOAEPWithOptions(oaepCipherText, priv, []byte("label"), WithRSAOAEPHash(sha256.New))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(oaepOut, plain) {
		t.Fatalf("RSADecryptOAEPWithOptions() = %q", oaepOut)
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
