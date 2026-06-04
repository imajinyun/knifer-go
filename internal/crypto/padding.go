package crypto

import "bytes"

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	return append(append([]byte(nil), data...), bytes.Repeat([]byte{byte(padding)}, padding)...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, ErrInvalidCipherText
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, ErrInvalidCipherText
	}
	for _, b := range data[len(data)-padding:] {
		if int(b) != padding {
			return nil, ErrInvalidCipherText
		}
	}
	return append([]byte(nil), data[:len(data)-padding]...), nil
}
