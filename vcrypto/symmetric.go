package vcrypto

import cryptoimpl "github.com/imajinyun/go-knifer/internal/crypto"

// DESEncryptCBC encrypts plain data using DES-CBC with PKCS#7 padding.
func DESEncryptCBC(plain, key, iv []byte) ([]byte, error) {
	return cryptoimpl.DESEncryptCBC(plain, key, iv)
}

// DESDecryptCBC decrypts DES-CBC data using PKCS#7 padding.
func DESDecryptCBC(cipherText, key, iv []byte) ([]byte, error) {
	return cryptoimpl.DESDecryptCBC(cipherText, key, iv)
}

// TripleDESEncryptCBC encrypts plain data using 3DES-CBC with PKCS#7 padding.
func TripleDESEncryptCBC(plain, key, iv []byte) ([]byte, error) {
	return cryptoimpl.TripleDESEncryptCBC(plain, key, iv)
}

// TripleDESDecryptCBC decrypts 3DES-CBC data using PKCS#7 padding.
func TripleDESDecryptCBC(cipherText, key, iv []byte) ([]byte, error) {
	return cryptoimpl.TripleDESDecryptCBC(cipherText, key, iv)
}

// RC4Crypt encrypts or decrypts data using RC4.
func RC4Crypt(data, key []byte) ([]byte, error) { return cryptoimpl.RC4Crypt(data, key) }

// VigenereEncrypt encrypts printable ASCII text using the Vigenere helper algorithm.
func VigenereEncrypt(data, cipherKey string) (string, error) {
	return cryptoimpl.VigenereEncrypt(data, cipherKey)
}

// VigenereDecrypt decrypts text encrypted by VigenereEncrypt.
func VigenereDecrypt(data, cipherKey string) (string, error) {
	return cryptoimpl.VigenereDecrypt(data, cipherKey)
}

// XXTEAEncrypt encrypts data using XXTEA.
func XXTEAEncrypt(data, key []byte) []byte { return cryptoimpl.XXTEAEncrypt(data, key) }

// XXTEADecrypt decrypts data using XXTEA.
func XXTEADecrypt(data, key []byte) ([]byte, error) { return cryptoimpl.XXTEADecrypt(data, key) }
