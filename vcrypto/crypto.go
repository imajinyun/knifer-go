package vcrypto

import (
	"crypto/rsa"

	cryptoimpl "github.com/imajinyun/go-knifer/internal/crypto"
)

// ErrInvalidKey indicates an invalid cryptographic key.
var ErrInvalidKey = cryptoimpl.ErrInvalidKey

// ErrInvalidIV indicates an invalid initialization vector.
var ErrInvalidIV = cryptoimpl.ErrInvalidIV

// ErrInvalidCipherText indicates invalid encrypted data.
var ErrInvalidCipherText = cryptoimpl.ErrInvalidCipherText

// MD5Hex returns the MD5 digest of s in lower-case hex form.
// For lightweight checksum shortcuts with no security requirement, vhash.MD5Hex is also available.
func MD5Hex(s string) string { return cryptoimpl.MD5Hex([]byte(s)) }

// MD5HexBytes returns the MD5 digest of data in lower-case hex form.
func MD5HexBytes(data []byte) string { return cryptoimpl.MD5Hex(data) }

// SHA1Hex returns the SHA1 digest of s in lower-case hex form.
// For lightweight checksum shortcuts with no security requirement, vhash.SHA1Hex is also available.
func SHA1Hex(s string) string { return cryptoimpl.SHA1Hex([]byte(s)) }

// SHA256Hex returns the SHA256 digest of s in lower-case hex form.
// For lightweight checksum shortcuts with no security requirement, vhash.SHA256Hex is also available.
func SHA256Hex(s string) string { return cryptoimpl.SHA256Hex([]byte(s)) }

// SHA512Hex returns the SHA512 digest of s in lower-case hex form.
func SHA512Hex(s string) string { return cryptoimpl.SHA512Hex([]byte(s)) }

// HMACMD5Hex returns HMAC-MD5 in lower-case hex form.
func HMACMD5Hex(key, data []byte) string { return cryptoimpl.HMACMD5Hex(key, data) }

// HMACSHA1Hex returns HMAC-SHA1 in lower-case hex form.
func HMACSHA1Hex(key, data []byte) string { return cryptoimpl.HMACSHA1Hex(key, data) }

// HMACSHA256Hex returns HMAC-SHA256 in lower-case hex form.
func HMACSHA256Hex(key, data []byte) string { return cryptoimpl.HMACSHA256Hex(key, data) }

// HMACSHA512Hex returns HMAC-SHA512 in lower-case hex form.
func HMACSHA512Hex(key, data []byte) string { return cryptoimpl.HMACSHA512Hex(key, data) }

// RandomBytes returns n cryptographically secure random bytes.
func RandomBytes(n int) ([]byte, error) { return cryptoimpl.RandomBytes(n) }

// GenerateAESKey returns a random AES key. Valid sizes are 16, 24, or 32 bytes.
func GenerateAESKey(size int) ([]byte, error) { return cryptoimpl.GenerateAESKey(size) }

// AESEncryptCBC encrypts plain data using AES-CBC with PKCS#7 padding.
func AESEncryptCBC(plain, key, iv []byte) ([]byte, error) {
	return cryptoimpl.AESEncryptCBC(plain, key, iv)
}

// AESDecryptCBC decrypts AES-CBC data using PKCS#7 padding.
func AESDecryptCBC(cipherText, key, iv []byte) ([]byte, error) {
	return cryptoimpl.AESDecryptCBC(cipherText, key, iv)
}

// AESEncryptGCM encrypts plain data using AES-GCM.
func AESEncryptGCM(plain, key, nonce, additionalData []byte) ([]byte, error) {
	return cryptoimpl.AESEncryptGCM(plain, key, nonce, additionalData)
}

// AESDecryptGCM decrypts AES-GCM data.
func AESDecryptGCM(cipherText, key, nonce, additionalData []byte) ([]byte, error) {
	return cryptoimpl.AESDecryptGCM(cipherText, key, nonce, additionalData)
}

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

// PrivateKeyToPEM encodes an RSA private key as PKCS#1 PEM.
func PrivateKeyToPEM(priv *rsa.PrivateKey) []byte { return cryptoimpl.PrivateKeyToPEM(priv) }

// PublicKeyToPEM encodes an RSA public key as PKIX PEM.
func PublicKeyToPEM(pub *rsa.PublicKey) ([]byte, error) { return cryptoimpl.PublicKeyToPEM(pub) }

// ParseRSAPrivateKeyPEM parses a PKCS#1 or PKCS#8 RSA private key PEM.
func ParseRSAPrivateKeyPEM(data []byte) (*rsa.PrivateKey, error) {
	return cryptoimpl.ParseRSAPrivateKeyPEM(data)
}

// ParseRSAPublicKeyPEM parses a PKIX or PKCS#1 RSA public key PEM.
func ParseRSAPublicKeyPEM(data []byte) (*rsa.PublicKey, error) {
	return cryptoimpl.ParseRSAPublicKeyPEM(data)
}
