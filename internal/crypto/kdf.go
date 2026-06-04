package crypto

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"hash"
)

// PBKDF2 derives a key from password and salt using PBKDF2.
func PBKDF2(password, salt []byte, iterations, keyLen int, fn func() hash.Hash) ([]byte, error) {
	if iterations <= 0 || keyLen <= 0 || fn == nil {
		return nil, ErrInvalidKey
	}
	h := fn()
	hLen := h.Size()
	nBlocks := (keyLen + hLen - 1) / hLen
	derived := make([]byte, 0, nBlocks*hLen)
	var blockIndex [4]byte
	for block := 1; block <= nBlocks; block++ {
		blockIndex[0] = byte(block >> 24)
		blockIndex[1] = byte(block >> 16)
		blockIndex[2] = byte(block >> 8)
		blockIndex[3] = byte(block)

		u := hmac.New(fn, password)
		_, _ = u.Write(salt)
		_, _ = u.Write(blockIndex[:])
		sum := u.Sum(nil)
		t := append([]byte(nil), sum...)

		for i := 1; i < iterations; i++ {
			u = hmac.New(fn, password)
			_, _ = u.Write(sum)
			sum = u.Sum(nil)
			for j := range t {
				t[j] ^= sum[j]
			}
		}
		derived = append(derived, t...)
	}
	return derived[:keyLen], nil
}

// PBKDF2SHA1 derives a key using PBKDF2-HMAC-SHA1.
func PBKDF2SHA1(password, salt []byte, iterations, keyLen int) ([]byte, error) {
	return PBKDF2(password, salt, iterations, keyLen, sha1.New)
}

// PBKDF2SHA256 derives a key using PBKDF2-HMAC-SHA256.
func PBKDF2SHA256(password, salt []byte, iterations, keyLen int) ([]byte, error) {
	return PBKDF2(password, salt, iterations, keyLen, sha256.New)
}
