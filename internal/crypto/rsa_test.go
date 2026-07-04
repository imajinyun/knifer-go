package crypto

import (
	"bytes"
	stdcrypto "crypto"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestRSADigestOptions(t *testing.T) {
	if WithRSARandomReader(nil) == nil {
		t.Fatal("WithRSARandomReader() = nil")
	}
	if WithRSADigestHash(0, nil) == nil {
		t.Fatal("WithRSADigestHash() = nil")
	}
	if WithRSADigestRandomReader(nil) == nil {
		t.Fatal("WithRSADigestRandomReader() = nil")
	}
	if WithRSADigestPSS(nil) == nil {
		t.Fatal("WithRSADigestPSS() = nil")
	}
}

func TestRSAOAEPWithOptions(t *testing.T) {
	priv, err := GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("hello")
	cipherText, err := RSAEncryptOAEPWithOptions(plain, &priv.PublicKey, []byte("label"), WithRSARandomReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	out, err := RSADecryptOAEPWithOptions(cipherText, priv, []byte("label"), WithRSARandomReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("RSADecryptOAEPWithOptions() = %q", out)
	}
}

func TestRSASignWithDigestOptions(t *testing.T) {
	priv, err := GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("payload")
	sig, err := SignWithRSAOptions(plain, priv, WithRSADigestHash(stdcrypto.SHA256, sha256.New))
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyWithRSAOptions(plain, sig, &priv.PublicKey, WithRSADigestHash(stdcrypto.SHA256, sha256.New)); err != nil {
		t.Fatal(err)
	}
}

func TestRSASignWithPSSOptions(t *testing.T) {
	priv, err := GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	pssOptions := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: stdcrypto.SHA256}
	plain := []byte("pss payload")
	sig, err := SignWithRSAOptions(plain, priv, WithRSADigestPSS(pssOptions))
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyWithRSAOptions(plain, sig, &priv.PublicKey, WithRSADigestPSS(pssOptions)); err != nil {
		t.Fatal(err)
	}
}

func TestSignWithDigestRandomReader(t *testing.T) {
	priv, err := GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("custom reader")
	sig, err := SignWithRSAOptions(plain, priv, WithRSADigestRandomReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyWithRSAOptions(plain, sig, &priv.PublicKey, WithRSADigestRandomReader(nil)); err != nil {
		t.Fatal(err)
	}
}

func TestRSAOAEP(t *testing.T) {
	priv, err := GenRSAKey(1024)
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
}

func TestRSAPKCS1PSS(t *testing.T) {
	priv, err := GenRSAKey(1024)
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
}

func TestRSANilKeyAndProviderErrorContracts(t *testing.T) {
	data := []byte("payload")
	digest := sha256.Sum256(data)
	entropyErr := errors.New("entropy exhausted")

	if _, err := RSAEncryptOAEP(data, nil, nil); !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("RSAEncryptOAEP nil key err = %v", err)
	}
	if _, err := RSADecryptOAEP(data, nil, nil); !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("RSADecryptOAEP nil key err = %v", err)
	}
	if _, err := RSASignPSS(nil, stdcrypto.SHA256, digest[:]); !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("RSASignPSS nil key err = %v", err)
	}
	if err := RSAVerifyPSS(nil, stdcrypto.SHA256, digest[:], []byte("sig")); !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("RSAVerifyPSS nil key err = %v", err)
	}
	if _, err := SignWithRSAOptions(data, nil); !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("SignWithRSAOptions nil key err = %v", err)
	}
	if err := VerifyWithRSAOptions(data, []byte("sig"), nil); !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("VerifyWithRSAOptions nil key err = %v", err)
	}

	priv, err := GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RSAEncryptOAEPWithOptions(data, &priv.PublicKey, nil, WithRSARandomReader(errorReader{err: entropyErr})); !errors.Is(err, entropyErr) {
		t.Fatalf("RSAEncryptOAEPWithOptions provider err = %v, want entropyErr", err)
	}
	if _, err := RSAEncryptOAEPWithOptions(data, &priv.PublicKey, nil, WithRSARandomReader(errorReader{err: entropyErr}), WithRSARandomReader(nil)); !errors.Is(err, entropyErr) {
		t.Fatalf("RSAEncryptOAEPWithOptions nil random overwrite err = %v, want entropyErr", err)
	}
	if _, err := RSASignPSSWithOptions(priv, stdcrypto.SHA256, digest[:], WithRSARandomReader(errorReader{err: entropyErr})); !errors.Is(err, entropyErr) {
		t.Fatalf("RSASignPSSWithOptions provider err = %v, want entropyErr", err)
	}
	if _, err := RSASignPSSWithOptions(priv, stdcrypto.SHA256, digest[:], WithRSARandomReader(errorReader{err: entropyErr}), WithRSARandomReader(nil)); !errors.Is(err, entropyErr) {
		t.Fatalf("RSASignPSSWithOptions nil random overwrite err = %v, want entropyErr", err)
	}
	if _, err := SignWithRSAOptions(data, priv, WithRSADigestRandomReader(errorReader{err: entropyErr})); !errors.Is(err, entropyErr) {
		t.Fatalf("SignWithRSAOptions provider err = %v, want entropyErr", err)
	}
	if _, err := SignWithRSAOptions(data, priv, WithRSADigestRandomReader(errorReader{err: entropyErr}), WithRSADigestRandomReader(nil)); !errors.Is(err, entropyErr) {
		t.Fatalf("SignWithRSAOptions nil random overwrite err = %v, want entropyErr", err)
	}
}

func TestRSARealUserEncryptSignVerifyScenario(t *testing.T) {
	priv, err := GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	message := []byte("order_id=42&amount=19.99")
	label := []byte("knifer-go:vcrypto")
	cipherText, err := RSAEncryptOAEP(message, &priv.PublicKey, label)
	if err != nil {
		t.Fatalf("RSAEncryptOAEP = %v", err)
	}
	opened, err := RSADecryptOAEP(cipherText, priv, label)
	if err != nil {
		t.Fatalf("RSADecryptOAEP = %v", err)
	}
	if !bytes.Equal(opened, message) {
		t.Fatalf("opened message = %q", opened)
	}
	if _, err := RSADecryptOAEP(cipherText, priv, []byte("wrong label")); err == nil {
		t.Fatal("RSADecryptOAEP should reject wrong label")
	}

	pssOptions := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: stdcrypto.SHA256}
	sig, err := SignWithRSAOptions(message, priv, WithRSADigestPSS(pssOptions))
	if err != nil {
		t.Fatalf("SignWithRSAOptions PSS = %v", err)
	}
	if err := VerifyWithRSAOptions(message, sig, &priv.PublicKey, WithRSADigestPSS(pssOptions)); err != nil {
		t.Fatalf("VerifyWithRSAOptions PSS = %v", err)
	}
	if err := VerifyWithRSAOptions([]byte("tampered"), sig, &priv.PublicKey, WithRSADigestPSS(pssOptions)); err == nil {
		t.Fatal("VerifyWithRSAOptions should reject tampered message")
	}
}
