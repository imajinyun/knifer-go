package crypto

import (
	"crypto/aes"
	"crypto/cipher"
)

// AESEncryptCBC encrypts plain data using AES-CBC with PKCS#7 padding.
func AESEncryptCBC(plain, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return encryptCBC(block, plain, iv)
}

// AESDecryptCBC decrypts AES-CBC data using PKCS#7 padding.
func AESDecryptCBC(cipherText, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return decryptCBC(block, cipherText, iv)
}

// AESEncryptECB encrypts plain data using AES-ECB with PKCS#7 padding.
func AESEncryptECB(plain, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return encryptECB(block, plain), nil
}

// AESDecryptECB decrypts AES-ECB data using PKCS#7 padding.
func AESDecryptECB(cipherText, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return decryptECB(block, cipherText)
}

// AESEncryptCTR encrypts or decrypts data using AES-CTR.
func AESEncryptCTR(data, key, iv []byte) ([]byte, error) {
	return aesStream(data, key, iv, cipher.NewCTR)
}

// AESDecryptCTR decrypts or encrypts data using AES-CTR.
func AESDecryptCTR(data, key, iv []byte) ([]byte, error) { return AESEncryptCTR(data, key, iv) }

// AESEncryptCFB encrypts data using AES-CFB.
func AESEncryptCFB(data, key, iv []byte) ([]byte, error) {
	return aesCFB(data, key, iv, false)
}

// AESDecryptCFB decrypts data using AES-CFB.
func AESDecryptCFB(data, key, iv []byte) ([]byte, error) {
	return aesCFB(data, key, iv, true)
}

// AESEncryptOFB encrypts or decrypts data using AES-OFB.
func AESEncryptOFB(data, key, iv []byte) ([]byte, error) {
	return aesOFB(data, key, iv)
}

// AESDecryptOFB decrypts or encrypts data using AES-OFB.
func AESDecryptOFB(data, key, iv []byte) ([]byte, error) { return AESEncryptOFB(data, key, iv) }

// AESEncryptGCM encrypts plain data using AES-GCM.
func AESEncryptGCM(plain, key, nonce, additionalData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
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
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, ErrInvalidIV
	}
	return gcm.Open(nil, nonce, cipherText, additionalData)
}

func aesBlockWithIV(key, iv []byte) (cipher.Block, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) != block.BlockSize() {
		return nil, ErrInvalidIV
	}
	return block, nil
}

func aesStream(data, key, iv []byte, newStream func(cipher.Block, []byte) cipher.Stream) ([]byte, error) {
	block, err := aesBlockWithIV(key, iv)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(data))
	newStream(block, iv).XORKeyStream(out, data)
	return out, nil
}

func aesCFB(data, key, iv []byte, decrypt bool) ([]byte, error) {
	block, err := aesBlockWithIV(key, iv)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(data))
	feedback := append([]byte(nil), iv...)
	stream := make([]byte, block.BlockSize())
	for pos := 0; pos < len(data); pos += block.BlockSize() {
		block.Encrypt(stream, feedback)
		n := min(block.BlockSize(), len(data)-pos)
		for i := 0; i < n; i++ {
			out[pos+i] = data[pos+i] ^ stream[i]
		}
		if n == block.BlockSize() {
			if decrypt {
				copy(feedback, data[pos:pos+n])
			} else {
				copy(feedback, out[pos:pos+n])
			}
		}
	}
	return out, nil
}

func aesOFB(data, key, iv []byte) ([]byte, error) {
	block, err := aesBlockWithIV(key, iv)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(data))
	feedback := append([]byte(nil), iv...)
	for pos := 0; pos < len(data); pos += block.BlockSize() {
		block.Encrypt(feedback, feedback)
		n := min(block.BlockSize(), len(data)-pos)
		for i := 0; i < n; i++ {
			out[pos+i] = data[pos+i] ^ feedback[i]
		}
	}
	return out, nil
}

func encryptCBC(block cipher.Block, plain, iv []byte) ([]byte, error) {
	if len(iv) != block.BlockSize() {
		return nil, ErrInvalidIV
	}
	plain = pkcs7Pad(plain, block.BlockSize())
	out := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, plain)
	return out, nil
}

func decryptCBC(block cipher.Block, cipherText, iv []byte) ([]byte, error) {
	if len(iv) != block.BlockSize() {
		return nil, ErrInvalidIV
	}
	if len(cipherText) == 0 || len(cipherText)%block.BlockSize() != 0 {
		return nil, ErrInvalidCipherText
	}
	out := make([]byte, len(cipherText))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, cipherText)
	return pkcs7Unpad(out, block.BlockSize())
}

func encryptECB(block cipher.Block, plain []byte) []byte {
	plain = pkcs7Pad(plain, block.BlockSize())
	out := make([]byte, len(plain))
	for bs, be := 0, block.BlockSize(); bs < len(plain); bs, be = bs+block.BlockSize(), be+block.BlockSize() {
		block.Encrypt(out[bs:be], plain[bs:be])
	}
	return out
}

func decryptECB(block cipher.Block, cipherText []byte) ([]byte, error) {
	if len(cipherText) == 0 || len(cipherText)%block.BlockSize() != 0 {
		return nil, ErrInvalidCipherText
	}
	out := make([]byte, len(cipherText))
	for bs, be := 0, block.BlockSize(); bs < len(cipherText); bs, be = bs+block.BlockSize(), be+block.BlockSize() {
		block.Decrypt(out[bs:be], cipherText[bs:be])
	}
	return pkcs7Unpad(out, block.BlockSize())
}
