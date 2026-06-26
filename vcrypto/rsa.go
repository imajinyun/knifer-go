package vcrypto

import (
	stdcrypto "crypto"
	"crypto/rsa"
	"hash"
	"io"

	cryptoimpl "github.com/imajinyun/knifer-go/internal/crypto"
)

// RSAOption customizes RSA helper behavior.
type RSAOption = cryptoimpl.RSAOption

// RSADigestOption customizes RSA data-signing helpers per call.
type RSADigestOption = cryptoimpl.RSADigestOption

// WithRSARandomReader sets the entropy source used by RSA helpers.
func WithRSARandomReader(reader io.Reader) RSAOption { return cryptoimpl.WithRSARandomReader(reader) }

// WithRSAOAEPHash sets the hash function used by RSA-OAEP helpers.
func WithRSAOAEPHash(newHash func() hash.Hash) RSAOption { return cryptoimpl.WithRSAOAEPHash(newHash) }

// WithRSAPSSOptions sets the PSS options used by RSA-PSS helpers.
func WithRSAPSSOptions(opts *rsa.PSSOptions) RSAOption { return cryptoimpl.WithRSAPSSOptions(opts) }

// WithRSADigestHash sets the hash used by RSA data-signing helpers.
func WithRSADigestHash(hashID stdcrypto.Hash, newHash func() hash.Hash) RSADigestOption {
	return cryptoimpl.WithRSADigestHash(hashID, newHash)
}

// WithRSADigestRandomReader sets the entropy source used by RSA data-signing helpers.
func WithRSADigestRandomReader(reader io.Reader) RSADigestOption {
	return cryptoimpl.WithRSADigestRandomReader(reader)
}

// WithRSADigestPSS signs and verifies using RSA-PSS instead of PKCS#1 v1.5.
func WithRSADigestPSS(opts *rsa.PSSOptions) RSADigestOption {
	return cryptoimpl.WithRSADigestPSS(opts)
}

// GenRSAKey generates an RSA private key.
func GenRSAKey(bits int) (*rsa.PrivateKey, error) { return cryptoimpl.GenRSAKey(bits) }

// GenRSAKeyWithOptions generates an RSA private key with options.
func GenRSAKeyWithOptions(bits int, opts ...RSAOption) (*rsa.PrivateKey, error) {
	return cryptoimpl.GenRSAKeyWithOptions(bits, opts...)
}

// RSAEncryptOAEP encrypts data using RSA-OAEP with SHA-256.
func RSAEncryptOAEP(plain []byte, pub *rsa.PublicKey, label []byte) ([]byte, error) {
	return cryptoimpl.RSAEncryptOAEP(plain, pub, label)
}

// RSAEncryptOAEPWithOptions encrypts data using RSA-OAEP with options.
func RSAEncryptOAEPWithOptions(plain []byte, pub *rsa.PublicKey, label []byte, opts ...RSAOption) ([]byte, error) {
	return cryptoimpl.RSAEncryptOAEPWithOptions(plain, pub, label, opts...)
}

// RSADecryptOAEP decrypts data using RSA-OAEP with SHA-256.
func RSADecryptOAEP(cipherText []byte, priv *rsa.PrivateKey, label []byte) ([]byte, error) {
	return cryptoimpl.RSADecryptOAEP(cipherText, priv, label)
}

// RSADecryptOAEPWithOptions decrypts data using RSA-OAEP with options.
func RSADecryptOAEPWithOptions(cipherText []byte, priv *rsa.PrivateKey, label []byte, opts ...RSAOption) ([]byte, error) {
	return cryptoimpl.RSADecryptOAEPWithOptions(cipherText, priv, label, opts...)
}

// RSASignPSS signs digest using RSA-PSS.
func RSASignPSS(priv *rsa.PrivateKey, hash stdcrypto.Hash, digest []byte) ([]byte, error) {
	return cryptoimpl.RSASignPSS(priv, hash, digest)
}

// RSASignPSSWithOptions signs digest using RSA-PSS with options.
func RSASignPSSWithOptions(priv *rsa.PrivateKey, hash stdcrypto.Hash, digest []byte, opts ...RSAOption) ([]byte, error) {
	return cryptoimpl.RSASignPSSWithOptions(priv, hash, digest, opts...)
}

// RSAVerifyPSS verifies an RSA-PSS signature.
func RSAVerifyPSS(pub *rsa.PublicKey, hash stdcrypto.Hash, digest, sig []byte) error {
	return cryptoimpl.RSAVerifyPSS(pub, hash, digest, sig)
}

// RSAVerifyPSSWithOptions verifies an RSA-PSS signature with options.
func RSAVerifyPSSWithOptions(pub *rsa.PublicKey, hash stdcrypto.Hash, digest, sig []byte, opts ...RSAOption) error {
	return cryptoimpl.RSAVerifyPSSWithOptions(pub, hash, digest, sig, opts...)
}

// SignSHA256WithRSA signs data using SHA256withRSA.
func SignSHA256WithRSA(data []byte, priv *rsa.PrivateKey) ([]byte, error) {
	return cryptoimpl.SignSHA256WithRSA(data, priv)
}

// VerifySHA256WithRSA verifies SHA256withRSA signature.
func VerifySHA256WithRSA(data, sig []byte, pub *rsa.PublicKey) error {
	return cryptoimpl.VerifySHA256WithRSA(data, sig, pub)
}

// SignWithRSAOptions hashes data and signs it with configurable RSA options.
func SignWithRSAOptions(data []byte, priv *rsa.PrivateKey, opts ...RSADigestOption) ([]byte, error) {
	return cryptoimpl.SignWithRSAOptions(data, priv, opts...)
}

// VerifyWithRSAOptions hashes data and verifies an RSA signature with configurable options.
func VerifyWithRSAOptions(data, sig []byte, pub *rsa.PublicKey, opts ...RSADigestOption) error {
	return cryptoimpl.VerifyWithRSAOptions(data, sig, pub, opts...)
}
