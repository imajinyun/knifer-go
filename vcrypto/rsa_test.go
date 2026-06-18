package vcrypto_test

import (
	"bytes"
	stdcrypto "crypto"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"testing"

	"github.com/imajinyun/go-knifer/vcrypto"
)

func TestRSAEncryptDecryptAndSignVerify(t *testing.T) {
	priv, err := vcrypto.GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	pub := &priv.PublicKey
	plain := []byte("rsa message")

	cipherText, err := vcrypto.RSAEncryptOAEP(plain, pub, nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err := vcrypto.RSADecryptOAEP(cipherText, priv, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("RSADecryptOAEP() = %q", out)
	}

	digest := sha256.Sum256(plain)
	pssSig, err := vcrypto.RSASignPSS(priv, stdcrypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	if err := vcrypto.RSAVerifyPSS(pub, stdcrypto.SHA256, digest[:], pssSig); err != nil {
		t.Fatal(err)
	}
	quickSig, err := vcrypto.SignSHA256WithRSA(plain, priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := vcrypto.VerifySHA256WithRSA(plain, quickSig, pub); err != nil {
		t.Fatal(err)
	}
}

func TestRSAFacadeNilKeyErrors(t *testing.T) {
	plain := []byte("rsa message")
	digest := sha256.Sum256(plain)

	if _, err := vcrypto.RSAEncryptOAEP(plain, nil, nil); err == nil {
		t.Fatal("RSAEncryptOAEP nil public key error = nil")
	}
	if _, err := vcrypto.RSADecryptOAEP([]byte("cipher"), nil, nil); err == nil {
		t.Fatal("RSADecryptOAEP nil private key error = nil")
	}
	if _, err := vcrypto.RSASignPSS(nil, stdcrypto.SHA256, digest[:]); err == nil {
		t.Fatal("RSASignPSS nil private key error = nil")
	}
	if err := vcrypto.RSAVerifyPSS(nil, stdcrypto.SHA256, digest[:], []byte("sig")); err == nil {
		t.Fatal("RSAVerifyPSS nil public key error = nil")
	}
	if _, err := vcrypto.PBKDF2SHA256([]byte("password"), []byte("salt"), 0, 32); !errors.Is(err, vcrypto.ErrInvalidKey) {
		t.Fatalf("PBKDF2SHA256 invalid iterations error = %v", err)
	}
}

func TestRSAOptionsAndErrorPaths(t *testing.T) {
	priv, err := vcrypto.GenRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	pub := &priv.PublicKey
	plain := []byte("rsa message")
	digest := sha256.Sum256(plain)

	pssOptions := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: stdcrypto.SHA256}
	pssSig, err := vcrypto.RSASignPSSWithOptions(priv, stdcrypto.SHA256, digest[:], vcrypto.WithRSAPSSOptions(pssOptions))
	if err != nil {
		t.Fatal(err)
	}
	if err := vcrypto.RSAVerifyPSSWithOptions(pub, stdcrypto.SHA256, digest[:], pssSig, vcrypto.WithRSAPSSOptions(pssOptions)); err != nil {
		t.Fatal(err)
	}
	if err := vcrypto.RSAVerifyPSS(pub, stdcrypto.SHA256, digest[:], append([]byte(nil), pssSig[:len(pssSig)-1]...)); err == nil {
		t.Fatal("RSAVerifyPSS tampered signature error = nil")
	}

	oaepCipherText, err := vcrypto.RSAEncryptOAEPWithOptions(plain, pub, []byte("label"), vcrypto.WithRSAOAEPHash(sha256.New))
	if err != nil {
		t.Fatal(err)
	}
	oaepOut, err := vcrypto.RSADecryptOAEPWithOptions(oaepCipherText, priv, []byte("label"), vcrypto.WithRSAOAEPHash(sha256.New))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(oaepOut, plain) {
		t.Fatalf("RSADecryptOAEPWithOptions() = %q", oaepOut)
	}

	digestSig, err := vcrypto.SignWithRSAOptions(
		[]byte("rsa digest payload"),
		priv,
		vcrypto.WithRSADigestHash(stdcrypto.SHA256, sha256.New),
		vcrypto.WithRSADigestRandomReader(bytes.NewReader(bytes.Repeat([]byte{0x66}, 512))),
	)
	if err != nil {
		t.Fatalf("SignWithRSAOptions: %v", err)
	}
	if err := vcrypto.VerifyWithRSAOptions(
		[]byte("rsa digest payload"),
		digestSig,
		pub,
		vcrypto.WithRSADigestHash(stdcrypto.SHA256, sha256.New),
	); err != nil {
		t.Fatalf("VerifyWithRSAOptions: %v", err)
	}
	if err := vcrypto.VerifyWithRSAOptions([]byte("different"), digestSig, pub); err == nil {
		t.Fatal("VerifyWithRSAOptions tampered data error = nil")
	}
}

func TestFacadeRSAOptionSetters(t *testing.T) {
	if vcrypto.WithRSARandomReader(nil) == nil {
		t.Fatal("WithRSARandomReader returned nil")
	}
	if vcrypto.WithRSADigestPSS(nil) == nil {
		t.Fatal("WithRSADigestPSS returned nil")
	}
}

func TestFacadeGenRSAKeyWithOptions(t *testing.T) {
	priv, err := vcrypto.GenRSAKeyWithOptions(1024)
	if err != nil {
		t.Fatalf("GenRSAKeyWithOptions: %v", err)
	}
	if priv == nil {
		t.Fatal("GenRSAKeyWithOptions returned nil key")
	}
}
