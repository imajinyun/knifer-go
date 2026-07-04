package crypto

import (
	stdcrypto "crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"hash"
	"io"
)

type rsaConfig struct {
	random     io.Reader
	oaepHash   func() hash.Hash
	pssOptions *rsa.PSSOptions
}

type rsaDigestConfig struct {
	rsaConfig
	hashID  stdcrypto.Hash
	newHash func() hash.Hash
	pss     bool
}

// RSAOption customizes RSA helper behavior.
type RSAOption func(*rsaConfig)

// RSADigestOption customizes RSA data-signing helpers per call.
type RSADigestOption func(*rsaDigestConfig)

// WithRSARandomReader sets the entropy source used by RSA helpers.
func WithRSARandomReader(reader io.Reader) RSAOption {
	return func(c *rsaConfig) {
		if reader != nil {
			c.random = reader
		}
	}
}

// WithRSAOAEPHash sets the hash function used by RSA-OAEP helpers.
func WithRSAOAEPHash(newHash func() hash.Hash) RSAOption {
	return func(c *rsaConfig) {
		if newHash != nil {
			c.oaepHash = newHash
		}
	}
}

// WithRSAPSSOptions sets the PSS options used by RSA-PSS sign and verify helpers.
func WithRSAPSSOptions(opts *rsa.PSSOptions) RSAOption {
	return func(c *rsaConfig) { c.pssOptions = opts }
}

func applyRSAOptions(opts ...RSAOption) rsaConfig {
	cfg := rsaConfig{random: rand.Reader, oaepHash: sha256.New}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.random == nil {
		cfg.random = rand.Reader
	}
	if cfg.oaepHash == nil {
		cfg.oaepHash = sha256.New
	}
	return cfg
}

func applyRSADigestOptions(opts ...RSADigestOption) rsaDigestConfig {
	base := applyRSAOptions()
	cfg := rsaDigestConfig{rsaConfig: base, hashID: stdcrypto.SHA256, newHash: sha256.New}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.random == nil {
		cfg.random = rand.Reader
	}
	if cfg.newHash == nil {
		cfg.newHash = sha256.New
	}
	return cfg
}

// WithRSADigestHash sets the hash used by RSA data-signing helpers.
func WithRSADigestHash(hashID stdcrypto.Hash, newHash func() hash.Hash) RSADigestOption {
	return func(c *rsaDigestConfig) {
		c.hashID = hashID
		if newHash != nil {
			c.newHash = newHash
		}
	}
}

// WithRSADigestRandomReader sets the entropy source used by RSA data-signing helpers.
func WithRSADigestRandomReader(reader io.Reader) RSADigestOption {
	return func(c *rsaDigestConfig) {
		if reader != nil {
			c.random = reader
		}
	}
}

// WithRSADigestPSS signs and verifies using RSA-PSS instead of PKCS#1 v1.5.
func WithRSADigestPSS(opts *rsa.PSSOptions) RSADigestOption {
	return func(c *rsaDigestConfig) {
		c.pss = true
		c.pssOptions = opts
	}
}

// GenRSAKey generates an RSA private key.
func GenRSAKey(bits int) (*rsa.PrivateKey, error) {
	return GenRSAKeyWithOptions(bits)
}

// GenRSAKeyWithOptions generates an RSA private key with options.
func GenRSAKeyWithOptions(bits int, opts ...RSAOption) (*rsa.PrivateKey, error) {
	cfg := applyRSAOptions(opts...)
	return rsa.GenerateKey(cfg.random, bits)
}

// RSAEncryptOAEP encrypts data using RSA-OAEP with SHA-256.
func RSAEncryptOAEP(plain []byte, pub *rsa.PublicKey, label []byte) ([]byte, error) {
	return RSAEncryptOAEPWithOptions(plain, pub, label)
}

// RSAEncryptOAEPWithOptions encrypts data using RSA-OAEP with options.
func RSAEncryptOAEPWithOptions(plain []byte, pub *rsa.PublicKey, label []byte, opts ...RSAOption) ([]byte, error) {
	if pub == nil {
		return nil, ErrInvalidKey
	}
	cfg := applyRSAOptions(opts...)
	return rsa.EncryptOAEP(cfg.oaepHash(), cfg.random, pub, plain, label)
}

// RSADecryptOAEP decrypts data using RSA-OAEP with SHA-256.
func RSADecryptOAEP(cipherText []byte, priv *rsa.PrivateKey, label []byte) ([]byte, error) {
	return RSADecryptOAEPWithOptions(cipherText, priv, label)
}

// RSADecryptOAEPWithOptions decrypts data using RSA-OAEP with options.
func RSADecryptOAEPWithOptions(cipherText []byte, priv *rsa.PrivateKey, label []byte, opts ...RSAOption) ([]byte, error) {
	if priv == nil {
		return nil, ErrInvalidKey
	}
	cfg := applyRSAOptions(opts...)
	return rsa.DecryptOAEP(cfg.oaepHash(), cfg.random, priv, cipherText, label)
}

// RSASignPSS signs digest using RSA-PSS.
func RSASignPSS(priv *rsa.PrivateKey, hash stdcrypto.Hash, digest []byte) ([]byte, error) {
	return RSASignPSSWithOptions(priv, hash, digest)
}

// RSASignPSSWithOptions signs digest using RSA-PSS with options.
func RSASignPSSWithOptions(priv *rsa.PrivateKey, hash stdcrypto.Hash, digest []byte, opts ...RSAOption) ([]byte, error) {
	if priv == nil {
		return nil, ErrInvalidKey
	}
	cfg := applyRSAOptions(opts...)
	return rsa.SignPSS(cfg.random, priv, hash, digest, cfg.pssOptions)
}

// RSAVerifyPSS verifies an RSA-PSS signature.
func RSAVerifyPSS(pub *rsa.PublicKey, hash stdcrypto.Hash, digest, sig []byte) error {
	return RSAVerifyPSSWithOptions(pub, hash, digest, sig)
}

// RSAVerifyPSSWithOptions verifies an RSA-PSS signature with options.
func RSAVerifyPSSWithOptions(pub *rsa.PublicKey, hash stdcrypto.Hash, digest, sig []byte, opts ...RSAOption) error {
	if pub == nil {
		return ErrInvalidKey
	}
	cfg := applyRSAOptions(opts...)
	return rsa.VerifyPSS(pub, hash, digest, sig, cfg.pssOptions)
}

// SignSHA256WithRSA signs data using SHA256withRSA.
func SignSHA256WithRSA(data []byte, priv *rsa.PrivateKey) ([]byte, error) {
	return SignWithRSAOptions(data, priv)
}

// VerifySHA256WithRSA verifies SHA256withRSA signature.
func VerifySHA256WithRSA(data, sig []byte, pub *rsa.PublicKey) error {
	return VerifyWithRSAOptions(data, sig, pub)
}

// SignWithRSAOptions hashes data and signs it with configurable RSA options.
func SignWithRSAOptions(data []byte, priv *rsa.PrivateKey, opts ...RSADigestOption) ([]byte, error) {
	if priv == nil {
		return nil, ErrInvalidKey
	}
	cfg := applyRSADigestOptions(opts...)
	digest := Digest(data, cfg.newHash)
	return rsa.SignPSS(cfg.random, priv, cfg.hashID, digest, cfg.pssOptions)
}

// VerifyWithRSAOptions hashes data and verifies an RSA signature with configurable options.
func VerifyWithRSAOptions(data, sig []byte, pub *rsa.PublicKey, opts ...RSADigestOption) error {
	if pub == nil {
		return ErrInvalidKey
	}
	cfg := applyRSADigestOptions(opts...)
	digest := Digest(data, cfg.newHash)
	return rsa.VerifyPSS(pub, cfg.hashID, digest, sig, cfg.pssOptions)
}
