package crypto

import (
	"bytes"
	stdcrypto "crypto"
	"crypto/rsa"
	"crypto/sha256"
	"testing"
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
