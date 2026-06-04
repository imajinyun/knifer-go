package crypto

import (
	stdcrypto "crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

// GenerateRSAKey generates an RSA private key.
func GenerateRSAKey(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}

// RSAEncryptOAEP encrypts data using RSA-OAEP with SHA-256.
func RSAEncryptOAEP(plain []byte, pub *rsa.PublicKey, label []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, plain, label)
}

// RSADecryptOAEP decrypts data using RSA-OAEP with SHA-256.
func RSADecryptOAEP(cipherText []byte, priv *rsa.PrivateKey, label []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, cipherText, label)
}

// RSAEncryptPKCS1v15 encrypts data using RSA PKCS#1 v1.5 padding.
func RSAEncryptPKCS1v15(plain []byte, pub *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, pub, plain)
}

// RSADecryptPKCS1v15 decrypts data using RSA PKCS#1 v1.5 padding.
func RSADecryptPKCS1v15(cipherText []byte, priv *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, priv, cipherText)
}

// RSASignPKCS1v15 signs digest using RSA PKCS#1 v1.5.
func RSASignPKCS1v15(priv *rsa.PrivateKey, hash stdcrypto.Hash, digest []byte) ([]byte, error) {
	return rsa.SignPKCS1v15(rand.Reader, priv, hash, digest)
}

// RSAVerifyPKCS1v15 verifies an RSA PKCS#1 v1.5 signature.
func RSAVerifyPKCS1v15(pub *rsa.PublicKey, hash stdcrypto.Hash, digest, sig []byte) error {
	return rsa.VerifyPKCS1v15(pub, hash, digest, sig)
}

// RSASignPSS signs digest using RSA-PSS.
func RSASignPSS(priv *rsa.PrivateKey, hash stdcrypto.Hash, digest []byte) ([]byte, error) {
	return rsa.SignPSS(rand.Reader, priv, hash, digest, nil)
}

// RSAVerifyPSS verifies an RSA-PSS signature.
func RSAVerifyPSS(pub *rsa.PublicKey, hash stdcrypto.Hash, digest, sig []byte) error {
	return rsa.VerifyPSS(pub, hash, digest, sig, nil)
}

// SignSHA256WithRSA signs data using SHA256withRSA.
func SignSHA256WithRSA(data []byte, priv *rsa.PrivateKey) ([]byte, error) {
	digest := sha256.Sum256(data)
	return RSASignPKCS1v15(priv, stdcrypto.SHA256, digest[:])
}

// VerifySHA256WithRSA verifies SHA256withRSA signature.
func VerifySHA256WithRSA(data, sig []byte, pub *rsa.PublicKey) error {
	digest := sha256.Sum256(data)
	return RSAVerifyPKCS1v15(pub, stdcrypto.SHA256, digest[:], sig)
}
