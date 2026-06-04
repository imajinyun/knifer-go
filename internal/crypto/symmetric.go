package crypto

import (
	"crypto/des"
	"crypto/rc4"
	"encoding/binary"
)

const xxteaDelta uint32 = 0x9E3779B9

// DESEncryptCBC encrypts plain data using DES-CBC with PKCS#7 padding.
func DESEncryptCBC(plain, key, iv []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return encryptCBC(block, plain, iv)
}

// DESDecryptCBC decrypts DES-CBC data using PKCS#7 padding.
func DESDecryptCBC(cipherText, key, iv []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return decryptCBC(block, cipherText, iv)
}

// TripleDESEncryptCBC encrypts plain data using 3DES-CBC with PKCS#7 padding.
func TripleDESEncryptCBC(plain, key, iv []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	return encryptCBC(block, plain, iv)
}

// TripleDESDecryptCBC decrypts 3DES-CBC data using PKCS#7 padding.
func TripleDESDecryptCBC(cipherText, key, iv []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	return decryptCBC(block, cipherText, iv)
}

// RC4Crypt encrypts or decrypts data using RC4.
func RC4Crypt(data, key []byte) ([]byte, error) {
	c, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(data))
	c.XORKeyStream(out, data)
	return out, nil
}

// VigenereEncrypt encrypts printable ASCII text using the Vigenere helper algorithm.
func VigenereEncrypt(data, cipherKey string) (string, error) {
	if cipherKey == "" {
		return "", ErrInvalidKey
	}
	dataRunes, keyRunes := []rune(data), []rune(cipherKey)
	out := make([]rune, len(dataRunes))
	for i, r := range dataRunes {
		k := keyRunes[i%len(keyRunes)]
		out[i] = (r+k-64)%95 + 32
	}
	return string(out), nil
}

// VigenereDecrypt decrypts text encrypted by VigenereEncrypt.
func VigenereDecrypt(data, cipherKey string) (string, error) {
	if cipherKey == "" {
		return "", ErrInvalidKey
	}
	dataRunes, keyRunes := []rune(data), []rune(cipherKey)
	out := make([]rune, len(dataRunes))
	for i, r := range dataRunes {
		k := keyRunes[i%len(keyRunes)]
		diff := r - k
		if diff >= 0 {
			out[i] = diff%95 + 32
		} else {
			out[i] = (diff+95)%95 + 32
		}
	}
	return string(out), nil
}

// XXTEAEncrypt encrypts data using XXTEA.
func XXTEAEncrypt(data, key []byte) []byte {
	if len(data) == 0 {
		return append([]byte(nil), data...)
	}
	return xxteaToByteArray(xxteaEncryptWords(xxteaToIntArray(data, true), xxteaToIntArray(xxteaFixKey(key), false)), false)
}

// XXTEADecrypt decrypts data using XXTEA.
func XXTEADecrypt(data, key []byte) ([]byte, error) {
	if len(data) == 0 {
		return append([]byte(nil), data...), nil
	}
	out := xxteaToByteArray(xxteaDecryptWords(xxteaToIntArray(data, false), xxteaToIntArray(xxteaFixKey(key), false)), true)
	if out == nil {
		return nil, ErrInvalidCipherText
	}
	return out, nil
}

func xxteaEncryptWords(v, k []uint32) []uint32 {
	n := len(v) - 1
	if n < 1 {
		return v
	}
	z := v[n]
	q := 6 + 52/uint32(n+1)
	var sum uint32
	for q > 0 {
		q--
		sum += xxteaDelta
		e := (sum >> 2) & 3
		for p := 0; p < n; p++ {
			y := v[p+1]
			v[p] += xxteaMX(sum, y, z, uint32(p), e, k)
			z = v[p]
		}
		y := v[0]
		v[n] += xxteaMX(sum, y, z, uint32(n), e, k)
		z = v[n]
	}
	return v
}

func xxteaDecryptWords(v, k []uint32) []uint32 {
	n := len(v) - 1
	if n < 1 {
		return v
	}
	y := v[0]
	q := 6 + 52/uint32(n+1)
	sum := q * xxteaDelta
	for sum != 0 {
		e := (sum >> 2) & 3
		for p := n; p > 0; p-- {
			z := v[p-1]
			v[p] -= xxteaMX(sum, y, z, uint32(p), e, k)
			y = v[p]
		}
		z := v[n]
		v[0] -= xxteaMX(sum, y, z, 0, e, k)
		y = v[0]
		sum -= xxteaDelta
	}
	return v
}

func xxteaMX(sum, y, z, p, e uint32, k []uint32) uint32 {
	return ((z>>5 ^ y<<2) + (y>>3 ^ z<<4)) ^ ((sum ^ y) + (k[p&3^e] ^ z))
}

func xxteaFixKey(key []byte) []byte {
	fixed := make([]byte, 16)
	copy(fixed, key)
	return fixed
}

func xxteaToIntArray(data []byte, includeLength bool) []uint32 {
	n := (len(data) + 3) >> 2
	if includeLength {
		n++
	}
	result := make([]uint32, n)
	for i, b := range data {
		result[i>>2] |= uint32(b) << ((i & 3) << 3)
	}
	if includeLength {
		result[n-1] = uint32(len(data))
	}
	return result
}

func xxteaToByteArray(data []uint32, includeLength bool) []byte {
	n := len(data) << 2
	if includeLength {
		m := int(data[len(data)-1])
		n -= 4
		if m < n-3 || m > n {
			return nil
		}
		n = m
	}
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		var word [4]byte
		binary.LittleEndian.PutUint32(word[:], data[i>>2])
		result[i] = word[i&3]
	}
	return result
}
