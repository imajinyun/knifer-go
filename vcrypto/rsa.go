package vcrypto

import (
	stdcrypto "crypto"
	"crypto/rsa"

	cryptoimpl "github.com/imajinyun/go-knifer/internal/crypto"
)

// GenerateRSAKey generates an RSA private key.
func GenerateRSAKey(bits int) (*rsa.PrivateKey, error) { return cryptoimpl.GenerateRSAKey(bits) }

// RSAEncryptOAEP encrypts data using RSA-OAEP with SHA-256.
func RSAEncryptOAEP(plain []byte, pub *rsa.PublicKey, label []byte) ([]byte, error) {
	return cryptoimpl.RSAEncryptOAEP(plain, pub, label)
}

// RSADecryptOAEP decrypts data using RSA-OAEP with SHA-256.
func RSADecryptOAEP(cipherText []byte, priv *rsa.PrivateKey, label []byte) ([]byte, error) {
	return cryptoimpl.RSADecryptOAEP(cipherText, priv, label)
}

// RSAEncryptPKCS1v15 encrypts data using RSA PKCS#1 v1.5 padding.
func RSAEncryptPKCS1v15(plain []byte, pub *rsa.PublicKey) ([]byte, error) {
	return cryptoimpl.RSAEncryptPKCS1v15(plain, pub)
}

// RSADecryptPKCS1v15 decrypts data using RSA PKCS#1 v1.5 padding.
func RSADecryptPKCS1v15(cipherText []byte, priv *rsa.PrivateKey) ([]byte, error) {
	return cryptoimpl.RSADecryptPKCS1v15(cipherText, priv)
}

// RSASignPKCS1v15 signs digest using RSA PKCS#1 v1.5.
func RSASignPKCS1v15(priv *rsa.PrivateKey, hash stdcrypto.Hash, digest []byte) ([]byte, error) {
	return cryptoimpl.RSASignPKCS1v15(priv, hash, digest)
}

// RSAVerifyPKCS1v15 verifies an RSA PKCS#1 v1.5 signature.
func RSAVerifyPKCS1v15(pub *rsa.PublicKey, hash stdcrypto.Hash, digest, sig []byte) error {
	return cryptoimpl.RSAVerifyPKCS1v15(pub, hash, digest, sig)
}

// RSASignPSS signs digest using RSA-PSS.
func RSASignPSS(priv *rsa.PrivateKey, hash stdcrypto.Hash, digest []byte) ([]byte, error) {
	return cryptoimpl.RSASignPSS(priv, hash, digest)
}

// RSAVerifyPSS verifies an RSA-PSS signature.
func RSAVerifyPSS(pub *rsa.PublicKey, hash stdcrypto.Hash, digest, sig []byte) error {
	return cryptoimpl.RSAVerifyPSS(pub, hash, digest, sig)
}

// SignSHA256WithRSA signs data using SHA256withRSA.
func SignSHA256WithRSA(data []byte, priv *rsa.PrivateKey) ([]byte, error) {
	return cryptoimpl.SignSHA256WithRSA(data, priv)
}

// VerifySHA256WithRSA verifies SHA256withRSA signature.
func VerifySHA256WithRSA(data, sig []byte, pub *rsa.PublicKey) error {
	return cryptoimpl.VerifySHA256WithRSA(data, sig, pub)
}
