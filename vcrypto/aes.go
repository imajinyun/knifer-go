package vcrypto

import (
	"crypto/cipher"

	cryptoimpl "github.com/imajinyun/go-knifer/internal/crypto"
)

// AESGCMOption customizes AES-GCM helper behavior.
type AESGCMOption = cryptoimpl.AESGCMOption

// WithGCMNonceSize sets a custom nonce size for AES-GCM helpers.
func WithGCMNonceSize(size int) AESGCMOption { return cryptoimpl.WithGCMNonceSize(size) }

// WithGCMTagSize sets a custom tag size for AES-GCM helpers.
func WithGCMTagSize(size int) AESGCMOption { return cryptoimpl.WithGCMTagSize(size) }

// WithGCMBlockFactory sets the cipher block factory used by AES-GCM helpers.
func WithGCMBlockFactory(factory func([]byte) (cipher.Block, error)) AESGCMOption {
	return cryptoimpl.WithGCMBlockFactory(factory)
}

// WithGCMRandomOptions sets the entropy source options used when AESSealGCM generates a nonce.
func WithGCMRandomOptions(opts ...RandomOption) AESGCMOption {
	return cryptoimpl.WithGCMRandomOptions(opts...)
}

// AESSealGCM encrypts plain data using AES-GCM and a freshly generated nonce.
// Prefer this helper for new encryption because AES-GCM authenticates both the
// ciphertext and additionalData. The returned nonce is not secret, but it must be
// stored or transmitted with the ciphertext and must never be reused with the
// same key.
func AESSealGCM(plain, key, additionalData []byte) (nonce, cipherText []byte, err error) {
	return cryptoimpl.AESSealGCM(plain, key, additionalData)
}

// AESSealGCMWithOptions encrypts plain data using AES-GCM and a freshly generated nonce.
func AESSealGCMWithOptions(plain, key, additionalData []byte, opts ...AESGCMOption) (nonce, cipherText []byte, err error) {
	return cryptoimpl.AESSealGCMWithOptions(plain, key, additionalData, opts...)
}

// AESOpenGCM decrypts data produced by AESSealGCM or AESEncryptGCM.
func AESOpenGCM(cipherText, key, nonce, additionalData []byte) ([]byte, error) {
	return cryptoimpl.AESOpenGCM(cipherText, key, nonce, additionalData)
}

// AESOpenGCMWithOptions decrypts AES-GCM data with options.
func AESOpenGCMWithOptions(cipherText, key, nonce, additionalData []byte, opts ...AESGCMOption) ([]byte, error) {
	return cryptoimpl.AESOpenGCMWithOptions(cipherText, key, nonce, additionalData, opts...)
}

// AESEncryptGCM encrypts plain data using AES-GCM.
func AESEncryptGCM(plain, key, nonce, additionalData []byte) ([]byte, error) {
	return cryptoimpl.AESEncryptGCM(plain, key, nonce, additionalData)
}

// AESEncryptGCMWithOptions encrypts plain data using AES-GCM with options.
func AESEncryptGCMWithOptions(plain, key, nonce, additionalData []byte, opts ...AESGCMOption) ([]byte, error) {
	return cryptoimpl.AESEncryptGCMWithOptions(plain, key, nonce, additionalData, opts...)
}

// AESDecryptGCM decrypts AES-GCM data.
func AESDecryptGCM(cipherText, key, nonce, additionalData []byte) ([]byte, error) {
	return cryptoimpl.AESDecryptGCM(cipherText, key, nonce, additionalData)
}

// AESDecryptGCMWithOptions decrypts AES-GCM data with options.
func AESDecryptGCMWithOptions(cipherText, key, nonce, additionalData []byte, opts ...AESGCMOption) ([]byte, error) {
	return cryptoimpl.AESDecryptGCMWithOptions(cipherText, key, nonce, additionalData, opts...)
}
