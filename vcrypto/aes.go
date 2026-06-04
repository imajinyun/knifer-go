package vcrypto

import cryptoimpl "github.com/imajinyun/go-knifer/internal/crypto"

// AESEncryptCBC encrypts plain data using AES-CBC with PKCS#7 padding.
func AESEncryptCBC(plain, key, iv []byte) ([]byte, error) {
	return cryptoimpl.AESEncryptCBC(plain, key, iv)
}

// AESDecryptCBC decrypts AES-CBC data using PKCS#7 padding.
func AESDecryptCBC(cipherText, key, iv []byte) ([]byte, error) {
	return cryptoimpl.AESDecryptCBC(cipherText, key, iv)
}

// AESEncryptECB encrypts plain data using AES-ECB with PKCS#7 padding.
func AESEncryptECB(plain, key []byte) ([]byte, error) { return cryptoimpl.AESEncryptECB(plain, key) }

// AESDecryptECB decrypts AES-ECB data using PKCS#7 padding.
func AESDecryptECB(cipherText, key []byte) ([]byte, error) {
	return cryptoimpl.AESDecryptECB(cipherText, key)
}

// AESEncryptCTR encrypts or decrypts data using AES-CTR.
func AESEncryptCTR(data, key, iv []byte) ([]byte, error) {
	return cryptoimpl.AESEncryptCTR(data, key, iv)
}

// AESDecryptCTR decrypts or encrypts data using AES-CTR.
func AESDecryptCTR(data, key, iv []byte) ([]byte, error) {
	return cryptoimpl.AESDecryptCTR(data, key, iv)
}

// AESEncryptCFB encrypts data using AES-CFB.
func AESEncryptCFB(data, key, iv []byte) ([]byte, error) {
	return cryptoimpl.AESEncryptCFB(data, key, iv)
}

// AESDecryptCFB decrypts data using AES-CFB.
func AESDecryptCFB(data, key, iv []byte) ([]byte, error) {
	return cryptoimpl.AESDecryptCFB(data, key, iv)
}

// AESEncryptOFB encrypts or decrypts data using AES-OFB.
func AESEncryptOFB(data, key, iv []byte) ([]byte, error) {
	return cryptoimpl.AESEncryptOFB(data, key, iv)
}

// AESDecryptOFB decrypts or encrypts data using AES-OFB.
func AESDecryptOFB(data, key, iv []byte) ([]byte, error) {
	return cryptoimpl.AESDecryptOFB(data, key, iv)
}

// AESEncryptGCM encrypts plain data using AES-GCM.
func AESEncryptGCM(plain, key, nonce, additionalData []byte) ([]byte, error) {
	return cryptoimpl.AESEncryptGCM(plain, key, nonce, additionalData)
}

// AESDecryptGCM decrypts AES-GCM data.
func AESDecryptGCM(cipherText, key, nonce, additionalData []byte) ([]byte, error) {
	return cryptoimpl.AESDecryptGCM(cipherText, key, nonce, additionalData)
}
