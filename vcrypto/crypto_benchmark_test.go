package vcrypto_test

import (
	"crypto/sha256"
	"testing"

	"github.com/imajinyun/knifer-go/vcrypto"
)

var (
	benchmarkCryptoPayload = []byte("knifer-go crypto benchmark payload for hashing hmac and authenticated encryption")
	benchmarkCryptoKey     = []byte("0123456789abcdef0123456789abcdef")
	benchmarkCryptoNonce   = []byte("123456789012")
	benchmarkCryptoAAD     = []byte("benchmark-aad")
	benchmarkCryptoBytes   []byte
	benchmarkCryptoErr     error
)

func BenchmarkSHA256Digest(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchmarkCryptoBytes = vcrypto.SHA256(benchmarkCryptoPayload)
	}
}

func BenchmarkHMACSHA256Signing(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchmarkCryptoBytes = vcrypto.HMACBytes(sha256.New, benchmarkCryptoKey, benchmarkCryptoPayload)
	}
}

func BenchmarkAESGCMEncrypt(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchmarkCryptoBytes, benchmarkCryptoErr = vcrypto.AESEncryptGCM(
			benchmarkCryptoPayload,
			benchmarkCryptoKey,
			benchmarkCryptoNonce,
			benchmarkCryptoAAD,
		)
		if benchmarkCryptoErr != nil {
			b.Fatalf("AESEncryptGCM: %v", benchmarkCryptoErr)
		}
	}
}

func BenchmarkAESGCMDecrypt(b *testing.B) {
	cipherText, err := vcrypto.AESEncryptGCM(
		benchmarkCryptoPayload,
		benchmarkCryptoKey,
		benchmarkCryptoNonce,
		benchmarkCryptoAAD,
	)
	if err != nil {
		b.Fatalf("AESEncryptGCM setup: %v", err)
	}

	b.ReportAllocs()
	for b.Loop() {
		benchmarkCryptoBytes, benchmarkCryptoErr = vcrypto.AESDecryptGCM(
			cipherText,
			benchmarkCryptoKey,
			benchmarkCryptoNonce,
			benchmarkCryptoAAD,
		)
		if benchmarkCryptoErr != nil {
			b.Fatalf("AESDecryptGCM: %v", benchmarkCryptoErr)
		}
	}
}
