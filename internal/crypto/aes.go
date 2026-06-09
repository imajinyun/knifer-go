package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

type aesGCMConfig struct {
	nonceSize int
	tagSize   int
	blockFunc func([]byte) (cipher.Block, error)
	random    []RandomOption
}

// AESGCMOption customizes AES-GCM helper behavior.
type AESGCMOption func(*aesGCMConfig)

// WithGCMNonceSize sets a custom nonce size for AES-GCM helpers.
func WithGCMNonceSize(size int) AESGCMOption {
	return func(c *aesGCMConfig) { c.nonceSize = size }
}

// WithGCMTagSize sets a custom tag size for AES-GCM helpers.
func WithGCMTagSize(size int) AESGCMOption {
	return func(c *aesGCMConfig) { c.tagSize = size }
}

// WithGCMBlockFactory sets the cipher block factory used by AES-GCM helpers.
func WithGCMBlockFactory(factory func([]byte) (cipher.Block, error)) AESGCMOption {
	return func(c *aesGCMConfig) { c.blockFunc = factory }
}

// WithGCMRandomOptions sets the entropy source options used when AESSealGCM generates a nonce.
func WithGCMRandomOptions(opts ...RandomOption) AESGCMOption {
	return func(c *aesGCMConfig) { c.random = append([]RandomOption(nil), opts...) }
}

func applyAESGCMOptions(opts []AESGCMOption) aesGCMConfig {
	cfg := aesGCMConfig{blockFunc: aes.NewCipher}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.blockFunc == nil {
		cfg.blockFunc = aes.NewCipher
	}
	return cfg
}

func newGCM(block cipher.Block, cfg aesGCMConfig) (cipher.AEAD, error) {
	if cfg.nonceSize > 0 && cfg.tagSize > 0 {
		return nil, errors.New("crypto: cannot set both GCM nonce size and tag size")
	}
	if cfg.nonceSize > 0 {
		return cipher.NewGCMWithNonceSize(block, cfg.nonceSize)
	}
	if cfg.tagSize > 0 {
		return cipher.NewGCMWithTagSize(block, cfg.tagSize)
	}
	return cipher.NewGCM(block)
}

// AESSealGCM encrypts plain data using AES-GCM and a freshly generated nonce.
// Prefer this helper for new encryption because AES-GCM authenticates both the
// ciphertext and additionalData. The returned nonce is not secret, but it must be
// stored or transmitted with the ciphertext and must never be reused with the
// same key.
func AESSealGCM(plain, key, additionalData []byte) (nonce, cipherText []byte, err error) {
	return AESSealGCMWithOptions(plain, key, additionalData)
}

// AESSealGCMWithOptions encrypts plain data using AES-GCM and a freshly generated nonce.
func AESSealGCMWithOptions(plain, key, additionalData []byte, opts ...AESGCMOption) (nonce, cipherText []byte, err error) {
	cfg := applyAESGCMOptions(opts)
	block, err := cfg.blockFunc(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := newGCM(block, cfg)
	if err != nil {
		return nil, nil, err
	}
	nonce, err = RandomBytesWithOptions(gcm.NonceSize(), cfg.random...)
	if err != nil {
		return nil, nil, err
	}
	return nonce, gcm.Seal(nil, nonce, plain, additionalData), nil
}

// AESOpenGCM decrypts data produced by AESSealGCM or AESEncryptGCM.
func AESOpenGCM(cipherText, key, nonce, additionalData []byte) ([]byte, error) {
	return AESDecryptGCM(cipherText, key, nonce, additionalData)
}

// AESOpenGCMWithOptions decrypts AES-GCM data with options.
func AESOpenGCMWithOptions(cipherText, key, nonce, additionalData []byte, opts ...AESGCMOption) ([]byte, error) {
	return AESDecryptGCMWithOptions(cipherText, key, nonce, additionalData, opts...)
}

// AESEncryptGCM encrypts plain data using AES-GCM.
func AESEncryptGCM(plain, key, nonce, additionalData []byte) ([]byte, error) {
	return AESEncryptGCMWithOptions(plain, key, nonce, additionalData)
}

// AESEncryptGCMWithOptions encrypts plain data using AES-GCM with options.
func AESEncryptGCMWithOptions(plain, key, nonce, additionalData []byte, opts ...AESGCMOption) ([]byte, error) {
	cfg := applyAESGCMOptions(opts)
	block, err := cfg.blockFunc(key)
	if err != nil {
		return nil, err
	}
	gcm, err := newGCM(block, cfg)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, ErrInvalidIV
	}
	return gcm.Seal(nil, nonce, plain, additionalData), nil
}

// AESDecryptGCM decrypts AES-GCM data.
func AESDecryptGCM(cipherText, key, nonce, additionalData []byte) ([]byte, error) {
	return AESDecryptGCMWithOptions(cipherText, key, nonce, additionalData)
}

// AESDecryptGCMWithOptions decrypts AES-GCM data with options.
func AESDecryptGCMWithOptions(cipherText, key, nonce, additionalData []byte, opts ...AESGCMOption) ([]byte, error) {
	cfg := applyAESGCMOptions(opts)
	block, err := cfg.blockFunc(key)
	if err != nil {
		return nil, err
	}
	gcm, err := newGCM(block, cfg)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, ErrInvalidIV
	}
	return gcm.Open(nil, nonce, cipherText, additionalData)
}
