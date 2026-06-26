package vcrypto_test

import (
	"bytes"
	stdcrypto "crypto"
	"crypto/aes"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vcrypto"
)

func Example_recommendedDigestAndHMAC() {
	digest := vcrypto.SHA256Hex("message")
	mac := vcrypto.HMACBytes(sha256.New, []byte("public-demo-key"), []byte("message"))

	fmt.Println(digest[:16])
	fmt.Println(vcrypto.HMACEqual(mac, append([]byte(nil), mac...)))
	// Output:
	// ab530a13e4591498
	// true
}

func Example_recommendedAESGCM() {
	key := exampleAESKey()
	nonce := exampleGCMNonce()
	cipherText, err := vcrypto.AESEncryptGCM([]byte("message"), key, nonce, []byte("public-aad"))
	plain, openErr := vcrypto.AESOpenGCM(cipherText, key, nonce, []byte("public-aad"))

	fmt.Println(err == nil, openErr == nil, string(plain))
	// Output: true true message
}

func Example_recommendedRSAPSS() {
	digest := sha256.Sum256([]byte("message"))
	sig, err := vcrypto.RSASignPSS(exampleRSAPrivateKey, stdcrypto.SHA256, digest[:])
	verifyErr := vcrypto.RSAVerifyPSS(&exampleRSAPrivateKey.PublicKey, stdcrypto.SHA256, digest[:], sig)

	fmt.Println(err == nil, verifyErr == nil)
	// Output: true true
}

func ExampleAESDecryptGCM() {
	key := exampleAESKey()
	nonce := exampleGCMNonce()
	cipherText, _ := vcrypto.AESEncryptGCM([]byte("secret"), key, nonce, []byte("aad"))
	plain, err := vcrypto.AESDecryptGCM(cipherText, key, nonce, []byte("aad"))
	fmt.Println(err == nil, string(plain))
	// Output: true secret
}

func ExampleAESDecryptGCMWithOptions() {
	key := exampleAESKey()
	nonce := []byte("1234567890123456")
	cipherText, _ := vcrypto.AESEncryptGCMWithOptions(
		[]byte("secret"),
		key,
		nonce,
		nil,
		vcrypto.WithGCMNonceSize(len(nonce)),
	)
	plain, err := vcrypto.AESDecryptGCMWithOptions(
		cipherText,
		key,
		nonce,
		nil,
		vcrypto.WithGCMNonceSize(len(nonce)),
	)
	fmt.Println(err == nil, string(plain))
	// Output: true secret
}

func ExampleAESEncryptGCM() {
	key := exampleAESKey()
	nonce := exampleGCMNonce()
	cipherText, err := vcrypto.AESEncryptGCM([]byte("secret"), key, nonce, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, err := vcrypto.AESDecryptGCM(cipherText, key, nonce, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(plain))
	// Output: secret
}

func ExampleAESEncryptGCMWithOptions() {
	key := exampleAESKey()
	nonce := exampleGCMNonce()
	cipherText, err := vcrypto.AESEncryptGCMWithOptions(
		[]byte("secret"),
		key,
		nonce,
		nil,
		vcrypto.WithGCMBlockFactory(aes.NewCipher),
		vcrypto.WithGCMTagSize(16),
	)
	plain, decryptErr := vcrypto.AESDecryptGCM(cipherText, key, nonce, nil)
	fmt.Println(err == nil, decryptErr == nil, string(plain))
	// Output: true true secret
}

func ExampleAESOpenGCM() {
	key := exampleAESKey()
	nonce := exampleGCMNonce()
	cipherText, err := vcrypto.AESEncryptGCM([]byte("secret"), key, nonce, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = vcrypto.AESOpenGCM(cipherText, key, nonce, []byte("wrong aad"))
	fmt.Println(errors.Is(err, vcrypto.ErrInvalidCipherText))
	// Output: true
}

func ExampleAESOpenGCMWithOptions() {
	key := exampleAESKey()
	nonce := []byte("1234567890123456")
	cipherText, _ := vcrypto.AESEncryptGCMWithOptions(
		[]byte("secret"),
		key,
		nonce,
		[]byte("aad"),
		vcrypto.WithGCMNonceSize(len(nonce)),
	)
	plain, err := vcrypto.AESOpenGCMWithOptions(
		cipherText,
		key,
		nonce,
		[]byte("aad"),
		vcrypto.WithGCMNonceSize(len(nonce)),
	)
	fmt.Println(err == nil, string(plain))
	// Output: true secret
}

func ExampleAESSealGCM() {
	nonce, cipherText, err := vcrypto.AESSealGCM([]byte("secret"), exampleAESKey(), []byte("aad"))
	plain, openErr := vcrypto.AESOpenGCM(cipherText, exampleAESKey(), nonce, []byte("aad"))
	fmt.Println(err == nil, openErr == nil, len(nonce), string(plain))
	// Output: true true 12 secret
}

func ExampleAESSealGCMWithOptions() {
	nonce, cipherText, err := vcrypto.AESSealGCMWithOptions(
		[]byte("secret"),
		exampleAESKey(),
		[]byte("aad"),
		vcrypto.WithGCMRandomOptions(vcrypto.WithRandomReader(bytes.NewReader(bytes.Repeat([]byte{0x42}, 12)))),
	)
	plain, openErr := vcrypto.AESOpenGCM(cipherText, exampleAESKey(), nonce, []byte("aad"))
	fmt.Println(err == nil, openErr == nil, fmt.Sprintf("%x", nonce), string(plain))
	// Output: true true 424242424242424242424242 secret
}

func ExampleConstantTimeEqual() {
	fmt.Println(vcrypto.ConstantTimeEqual([]byte("token"), []byte("token")))
	fmt.Println(vcrypto.ConstantTimeEqual([]byte("token"), []byte("other")))
	// Output:
	// true
	// false
}

func ExampleDigest() {
	digest := vcrypto.Digest([]byte("abc"), sha256.New)
	fmt.Println(hex.EncodeToString(digest))
	// Output: ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
}

func ExampleDigestHex() {
	fmt.Println(vcrypto.DigestHex([]byte("abc"), sha512.New384))
	// Output: cb00753f45a35e8bb5a03d699ac65007272c32ab0eded1631a8b605a43ff5bed8086072ba1e7cc2358baeca134c825a7
}

func ExampleGenAESKey() {
	key, err := vcrypto.GenAESKey(16)
	fmt.Println(err == nil, len(key))
	// Output: true 16
}

func ExampleGenAESKeyWithOptions() {
	key, err := vcrypto.GenAESKeyWithOptions(
		16,
		vcrypto.WithRandomReader(bytes.NewReader(bytes.Repeat([]byte{0x42}, 16))),
	)
	fmt.Println(err == nil, fmt.Sprintf("%x", key))
	// Output: true 42424242424242424242424242424242
}

func ExampleGenRSAKey() {
	priv, err := vcrypto.GenRSAKey(1024)
	fmt.Println(err == nil, priv.N.BitLen() >= 1024)
	// Output: true true
}

func ExampleGenRSAKeyWithOptions() {
	priv, err := vcrypto.GenRSAKeyWithOptions(1024)
	fmt.Println(err == nil, priv.N.BitLen() >= 1024)
	// Output: true true
}

func ExampleHMACBytes() {
	mac := vcrypto.HMACBytes(sha256.New, []byte("key"), []byte("data"))
	fmt.Println(hex.EncodeToString(mac))
	// Output: 5031fe3d989c6d1537a013fa6e739da23463fdaec3b70137d828e36ace221bd0
}

func ExampleHMACEqual() {
	mac := vcrypto.HMACBytes(sha256.New, []byte("key"), []byte("data"))
	fmt.Println(vcrypto.HMACEqual(mac, append([]byte(nil), mac...)))
	// Output: true
}

func ExampleHMACHex() {
	fmt.Println(vcrypto.HMACHex(nil, []byte("key"), []byte("data")))
	// Output: 5031fe3d989c6d1537a013fa6e739da23463fdaec3b70137d828e36ace221bd0
}

func ExampleHMACSHA256Hex() {
	fmt.Println(vcrypto.HMACSHA256Hex([]byte("key"), []byte("data")))
	// Output: 5031fe3d989c6d1537a013fa6e739da23463fdaec3b70137d828e36ace221bd0
}

func ExampleHMACSHA384Hex() {
	fmt.Println(vcrypto.HMACSHA384Hex([]byte("key"), []byte("data")))
	// Output: c5f97ad9fd1020c174d7dc02cf83c4c1bf15ee20ec555b690ad58e62da8a00ee44ccdb65cb8c80acfd127ebee568958a
}

func ExampleHMACSHA512Hex() {
	fmt.Println(vcrypto.HMACSHA512Hex([]byte("key"), []byte("data")))
	// Output: 3c5953a18f7303ec653ba170ae334fafa08e3846f2efe317b87efce82376253cb52a8c31ddcde5a3a2eee183c2b34cb91f85e64ddbc325f7692b199473579c58
}

func ExamplePBKDF2() {
	key, err := vcrypto.PBKDF2([]byte("password"), []byte("salt"), 1, 32, sha256.New)
	fmt.Println(err == nil, hex.EncodeToString(key))
	// Output: true 120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b
}

func ExamplePBKDF2SHA256() {
	key, err := vcrypto.PBKDF2SHA256([]byte("password"), []byte("salt"), 1, 16)
	fmt.Println(err == nil, hex.EncodeToString(key))
	// Output: true 120fb6cffcf8b32c43e7225256c4f837
}

func ExampleParseRSAPrivateKeyPEM() {
	parsed, err := vcrypto.ParseRSAPrivateKeyPEM(vcrypto.PrivateKeyToPEM(exampleRSAPrivateKey))
	fmt.Println(err == nil, parsed.N.Cmp(exampleRSAPrivateKey.N) == 0)
	// Output: true true
}

func ExampleParseRSAPublicKeyPEM() {
	pubPEM, _ := vcrypto.PublicKeyToPEM(&exampleRSAPrivateKey.PublicKey)
	parsed, err := vcrypto.ParseRSAPublicKeyPEM(pubPEM)
	fmt.Println(err == nil, parsed.N.Cmp(exampleRSAPrivateKey.N) == 0)
	// Output: true true
}

func ExampleParseX509CertificatePEM() {
	cert, err := vcrypto.ParseX509CertificatePEM(exampleCertificatePEM())
	fmt.Println(err == nil, cert.Subject.CommonName)
	// Output: true knifer-go.test
}

func ExamplePrivateKeyToPEM() {
	pemBytes := vcrypto.PrivateKeyToPEM(exampleRSAPrivateKey)
	block, _ := pem.Decode(pemBytes)
	fmt.Println(block.Type)
	// Output: RSA PRIVATE KEY
}

func ExamplePrivateKeyToPKCS8PEM() {
	pemBytes, err := vcrypto.PrivateKeyToPKCS8PEM(exampleRSAPrivateKey)
	block, _ := pem.Decode(pemBytes)
	fmt.Println(err == nil, block.Type)
	// Output: true PRIVATE KEY
}

func ExamplePublicKeyFromCertificatePEM() {
	pub, err := vcrypto.PublicKeyFromCertificatePEM(exampleCertificatePEM())
	fmt.Println(err == nil, pub.N.Cmp(exampleRSAPrivateKey.N) == 0)
	// Output: true true
}

func ExamplePublicKeyToPEM() {
	pemBytes, err := vcrypto.PublicKeyToPEM(&exampleRSAPrivateKey.PublicKey)
	block, _ := pem.Decode(pemBytes)
	fmt.Println(err == nil, block.Type)
	// Output: true PUBLIC KEY
}

func ExamplePublicKeyToPKCS1PEM() {
	pemBytes := vcrypto.PublicKeyToPKCS1PEM(&exampleRSAPrivateKey.PublicKey)
	block, _ := pem.Decode(pemBytes)
	fmt.Println(block.Type)
	// Output: RSA PUBLIC KEY
}

func ExampleRSADecryptOAEP() {
	cipherText, _ := vcrypto.RSAEncryptOAEP([]byte("secret"), &exampleRSAPrivateKey.PublicKey, []byte("label"))
	plain, err := vcrypto.RSADecryptOAEP(cipherText, exampleRSAPrivateKey, []byte("label"))
	fmt.Println(err == nil, string(plain))
	// Output: true secret
}

func ExampleRSADecryptOAEPWithOptions() {
	cipherText, _ := vcrypto.RSAEncryptOAEPWithOptions(
		[]byte("secret"),
		&exampleRSAPrivateKey.PublicKey,
		[]byte("label"),
		vcrypto.WithRSAOAEPHash(sha256.New),
	)
	plain, err := vcrypto.RSADecryptOAEPWithOptions(
		cipherText,
		exampleRSAPrivateKey,
		[]byte("label"),
		vcrypto.WithRSAOAEPHash(sha256.New),
	)
	fmt.Println(err == nil, string(plain))
	// Output: true secret
}

func ExampleRSAEncryptOAEP() {
	cipherText, err := vcrypto.RSAEncryptOAEP([]byte("secret"), &exampleRSAPrivateKey.PublicKey, nil)
	fmt.Println(err == nil, len(cipherText) > 0)
	// Output: true true
}

func ExampleRSAEncryptOAEPWithOptions() {
	cipherText, err := vcrypto.RSAEncryptOAEPWithOptions(
		[]byte("secret"),
		&exampleRSAPrivateKey.PublicKey,
		nil,
		vcrypto.WithRSAOAEPHash(sha256.New),
	)
	fmt.Println(err == nil, len(cipherText) > 0)
	// Output: true true
}

func ExampleRSASignPSS() {
	digest := sha256.Sum256([]byte("message"))
	sig, err := vcrypto.RSASignPSS(exampleRSAPrivateKey, stdcrypto.SHA256, digest[:])
	verifyErr := vcrypto.RSAVerifyPSS(&exampleRSAPrivateKey.PublicKey, stdcrypto.SHA256, digest[:], sig)
	fmt.Println(err == nil, verifyErr == nil)
	// Output: true true
}

func ExampleRSASignPSSWithOptions() {
	digest := sha256.Sum256([]byte("message"))
	pssOptions := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: stdcrypto.SHA256}
	sig, err := vcrypto.RSASignPSSWithOptions(
		exampleRSAPrivateKey,
		stdcrypto.SHA256,
		digest[:],
		vcrypto.WithRSAPSSOptions(pssOptions),
	)
	verifyErr := vcrypto.RSAVerifyPSSWithOptions(
		&exampleRSAPrivateKey.PublicKey,
		stdcrypto.SHA256,
		digest[:],
		sig,
		vcrypto.WithRSAPSSOptions(pssOptions),
	)
	fmt.Println(err == nil, verifyErr == nil)
	// Output: true true
}

func ExampleRSAVerifyPSS() {
	digest := sha256.Sum256([]byte("message"))
	sig, _ := vcrypto.RSASignPSS(exampleRSAPrivateKey, stdcrypto.SHA256, digest[:])
	fmt.Println(vcrypto.RSAVerifyPSS(&exampleRSAPrivateKey.PublicKey, stdcrypto.SHA256, digest[:], sig) == nil)
	// Output: true
}

func ExampleRSAVerifyPSSWithOptions() {
	digest := sha256.Sum256([]byte("message"))
	pssOptions := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: stdcrypto.SHA256}
	sig, _ := vcrypto.RSASignPSSWithOptions(
		exampleRSAPrivateKey,
		stdcrypto.SHA256,
		digest[:],
		vcrypto.WithRSAPSSOptions(pssOptions),
	)
	err := vcrypto.RSAVerifyPSSWithOptions(
		&exampleRSAPrivateKey.PublicKey,
		stdcrypto.SHA256,
		digest[:],
		sig,
		vcrypto.WithRSAPSSOptions(pssOptions),
	)
	fmt.Println(err == nil)
	// Output: true
}

func ExampleRandomBytes() {
	b, err := vcrypto.RandomBytes(4)
	fmt.Println(err == nil, len(b))
	// Output: true 4
}

func ExampleRandomBytesWithOptions() {
	b, err := vcrypto.RandomBytesWithOptions(4, vcrypto.WithRandomReader(bytes.NewReader([]byte{1, 2, 3, 4})))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v\n", b)
	// Output: [1 2 3 4]
}

func ExampleSHA224() {
	fmt.Println(hex.EncodeToString(vcrypto.SHA224([]byte("abc"))))
	// Output: 23097d223405d8228642a477bda255b32aadbce4bda0b3f7e36c9da7
}

func ExampleSHA224Hex() {
	fmt.Println(vcrypto.SHA224Hex([]byte("abc")))
	// Output: 23097d223405d8228642a477bda255b32aadbce4bda0b3f7e36c9da7
}

func ExampleSHA256() {
	fmt.Println(hex.EncodeToString(vcrypto.SHA256([]byte("abc"))))
	// Output: ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
}

func ExampleSHA256Hex() {
	fmt.Println(vcrypto.SHA256Hex("abc"))
	// Output: ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
}

func ExampleSHA256HexBytes() {
	fmt.Println(vcrypto.SHA256HexBytes([]byte("abc")))
	// Output: ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
}

func ExampleSHA384() {
	fmt.Println(hex.EncodeToString(vcrypto.SHA384([]byte("abc"))))
	// Output: cb00753f45a35e8bb5a03d699ac65007272c32ab0eded1631a8b605a43ff5bed8086072ba1e7cc2358baeca134c825a7
}

func ExampleSHA384Hex() {
	fmt.Println(vcrypto.SHA384Hex([]byte("abc")))
	// Output: cb00753f45a35e8bb5a03d699ac65007272c32ab0eded1631a8b605a43ff5bed8086072ba1e7cc2358baeca134c825a7
}

func ExampleSHA512() {
	fmt.Println(hex.EncodeToString(vcrypto.SHA512([]byte("abc"))))
	// Output: ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f
}

func ExampleSHA512Hex() {
	fmt.Println(vcrypto.SHA512Hex("abc"))
	// Output: ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f
}

func ExampleSHA512HexBytes() {
	fmt.Println(vcrypto.SHA512HexBytes([]byte("abc")))
	// Output: ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f
}

func ExampleSignParams() {
	params := map[string]any{"b": 2, "a": 1, "skip": nil}
	fmt.Println(vcrypto.SignParams(params, vcrypto.SHA256HexBytes, "&", "=", true, "secret"))
	// Output: a425c0882a33d582a3b459297ce6e49b3bfcfae05756fa225511eeb3f670b298
}

func ExampleSignParamsSHA256() {
	fmt.Println(vcrypto.SignParamsSHA256(map[string]any{"b": 2, "a": 1}, "z"))
	// Output: c68ce2610f406ae3c70af2a5b5883eb13c9c8ae4b17a63f64d1c545b0934a0dc
}

func ExampleSignSHA256WithRSA() {
	sig, err := vcrypto.SignSHA256WithRSA([]byte("message"), exampleRSAPrivateKey)
	verifyErr := vcrypto.VerifySHA256WithRSA([]byte("message"), sig, &exampleRSAPrivateKey.PublicKey)
	fmt.Println(err == nil, verifyErr == nil)
	// Output: true true
}

func ExampleSignWithRSAOptions() {
	sig, err := vcrypto.SignWithRSAOptions(
		[]byte("message"),
		exampleRSAPrivateKey,
		vcrypto.WithRSADigestHash(stdcrypto.SHA256, sha256.New),
	)
	verifyErr := vcrypto.VerifyWithRSAOptions(
		[]byte("message"),
		sig,
		&exampleRSAPrivateKey.PublicKey,
		vcrypto.WithRSADigestHash(stdcrypto.SHA256, sha256.New),
	)
	fmt.Println(err == nil, verifyErr == nil)
	// Output: true true
}

func ExampleValidateAESGCMNonce() {
	fmt.Println(vcrypto.ValidateAESGCMNonce(exampleGCMNonce()) == nil)
	fmt.Println(errors.Is(vcrypto.ValidateAESGCMNonce([]byte("short")), vcrypto.ErrInvalidIV))
	// Output:
	// true
	// true
}

func ExampleValidateAESIV() {
	fmt.Println(vcrypto.ValidateAESIV(bytes.Repeat([]byte{0}, aes.BlockSize)) == nil)
	fmt.Println(errors.Is(vcrypto.ValidateAESIV([]byte("short")), vcrypto.ErrInvalidIV))
	// Output:
	// true
	// true
}

func ExampleValidateAESKey() {
	err := vcrypto.ValidateAESKey([]byte("too-short"))
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	fmt.Println(errors.Is(err, vcrypto.ErrInvalidKey))
	// Output:
	// true
	// true
}

func ExampleVerifySHA256WithRSA() {
	sig, _ := vcrypto.SignSHA256WithRSA([]byte("message"), exampleRSAPrivateKey)
	fmt.Println(vcrypto.VerifySHA256WithRSA([]byte("message"), sig, &exampleRSAPrivateKey.PublicKey) == nil)
	// Output: true
}

func ExampleVerifyWithRSAOptions() {
	sig, _ := vcrypto.SignWithRSAOptions([]byte("message"), exampleRSAPrivateKey)
	fmt.Println(vcrypto.VerifyWithRSAOptions([]byte("message"), sig, &exampleRSAPrivateKey.PublicKey) == nil)
	// Output: true
}

func ExampleWithGCMBlockFactory() {
	key := exampleAESKey()
	nonce := exampleGCMNonce()
	cipherText, err := vcrypto.AESEncryptGCMWithOptions(
		[]byte("secret"),
		key,
		nonce,
		nil,
		vcrypto.WithGCMBlockFactory(aes.NewCipher),
	)
	plain, openErr := vcrypto.AESOpenGCM(cipherText, key, nonce, nil)
	fmt.Println(err == nil, openErr == nil, string(plain))
	// Output: true true secret
}

func ExampleWithGCMNonceSize() {
	key := exampleAESKey()
	nonce := []byte("1234567890123456")
	cipherText, err := vcrypto.AESEncryptGCMWithOptions(
		[]byte("secret"),
		key,
		nonce,
		nil,
		vcrypto.WithGCMNonceSize(len(nonce)),
	)
	plain, openErr := vcrypto.AESOpenGCMWithOptions(
		cipherText,
		key,
		nonce,
		nil,
		vcrypto.WithGCMNonceSize(len(nonce)),
	)
	fmt.Println(err == nil, openErr == nil, string(plain))
	// Output: true true secret
}

func ExampleWithGCMRandomOptions() {
	nonce, _, err := vcrypto.AESSealGCMWithOptions(
		[]byte("secret"),
		exampleAESKey(),
		nil,
		vcrypto.WithGCMRandomOptions(vcrypto.WithRandomReader(bytes.NewReader(bytes.Repeat([]byte{0x22}, 12)))),
	)
	fmt.Println(err == nil, fmt.Sprintf("%x", nonce))
	// Output: true 222222222222222222222222
}

func ExampleWithGCMTagSize() {
	key := exampleAESKey()
	nonce := exampleGCMNonce()
	cipherText, err := vcrypto.AESEncryptGCMWithOptions(
		[]byte("secret"),
		key,
		nonce,
		nil,
		vcrypto.WithGCMTagSize(16),
	)
	fmt.Println(err == nil, len(cipherText))
	// Output: true 22
}

func ExampleWithRSADigestHash() {
	sig, err := vcrypto.SignWithRSAOptions(
		[]byte("message"),
		exampleRSAPrivateKey,
		vcrypto.WithRSADigestHash(stdcrypto.SHA512, sha512.New),
	)
	verifyErr := vcrypto.VerifyWithRSAOptions(
		[]byte("message"),
		sig,
		&exampleRSAPrivateKey.PublicKey,
		vcrypto.WithRSADigestHash(stdcrypto.SHA512, sha512.New),
	)
	fmt.Println(err == nil, verifyErr == nil)
	// Output: true true
}

func ExampleWithRSADigestPSS() {
	pssOptions := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: stdcrypto.SHA256}
	sig, err := vcrypto.SignWithRSAOptions(
		[]byte("message"),
		exampleRSAPrivateKey,
		vcrypto.WithRSADigestPSS(pssOptions),
	)
	verifyErr := vcrypto.VerifyWithRSAOptions(
		[]byte("message"),
		sig,
		&exampleRSAPrivateKey.PublicKey,
		vcrypto.WithRSADigestPSS(pssOptions),
	)
	fmt.Println(err == nil, verifyErr == nil)
	// Output: true true
}

func ExampleWithRSADigestRandomReader() {
	sig, err := vcrypto.SignWithRSAOptions(
		[]byte("message"),
		exampleRSAPrivateKey,
		vcrypto.WithRSADigestRandomReader(bytes.NewReader(bytes.Repeat([]byte{0x33}, 128))),
	)
	fmt.Println(err == nil, len(sig) > 0)
	// Output: true true
}

func ExampleWithRSAOAEPHash() {
	cipherText, _ := vcrypto.RSAEncryptOAEPWithOptions(
		[]byte("secret"),
		&exampleRSAPrivateKey.PublicKey,
		nil,
		vcrypto.WithRSAOAEPHash(sha256.New),
	)
	plain, err := vcrypto.RSADecryptOAEPWithOptions(
		cipherText,
		exampleRSAPrivateKey,
		nil,
		vcrypto.WithRSAOAEPHash(sha256.New),
	)
	fmt.Println(err == nil, string(plain))
	// Output: true secret
}

func ExampleWithRSAPSSOptions() {
	digest := sha256.Sum256([]byte("message"))
	pssOptions := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: stdcrypto.SHA256}
	sig, _ := vcrypto.RSASignPSSWithOptions(
		exampleRSAPrivateKey,
		stdcrypto.SHA256,
		digest[:],
		vcrypto.WithRSAPSSOptions(pssOptions),
	)
	fmt.Println(vcrypto.RSAVerifyPSSWithOptions(
		&exampleRSAPrivateKey.PublicKey,
		stdcrypto.SHA256,
		digest[:],
		sig,
		vcrypto.WithRSAPSSOptions(pssOptions),
	) == nil)
	// Output: true
}

func ExampleWithRSARandomReader() {
	cipherText, err := vcrypto.RSAEncryptOAEPWithOptions(
		[]byte("secret"),
		&exampleRSAPrivateKey.PublicKey,
		nil,
		vcrypto.WithRSARandomReader(bytes.NewReader(bytes.Repeat([]byte{0x44}, 64))),
	)
	fmt.Println(err == nil, len(cipherText) > 0)
	// Output: true true
}

func ExampleWithRandomReader() {
	b, err := vcrypto.RandomBytesWithOptions(3, vcrypto.WithRandomReader(bytes.NewReader([]byte{7, 8, 9})))
	fmt.Printf("%v %v\n", err == nil, b)
	// Output: true [7 8 9]
}

var exampleRSAPrivateKey = mustExampleRSAKey()

func exampleAESKey() []byte {
	return []byte("0123456789abcdef")
}

func exampleGCMNonce() []byte {
	return []byte("123456789012")
}

func mustExampleRSAKey() *rsa.PrivateKey {
	priv, err := vcrypto.GenRSAKey(1024)
	if err != nil {
		panic(err)
	}
	return priv
}

func exampleCertificatePEM() []byte {
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "knifer-go.test"},
		NotBefore:    time.Unix(0, 0),
		NotAfter:     time.Unix(3600, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certDER, err := x509.CreateCertificate(
		bytes.NewReader(bytes.Repeat([]byte{0x42}, 1024)),
		template,
		template,
		&exampleRSAPrivateKey.PublicKey,
		exampleRSAPrivateKey,
	)
	if err != nil {
		panic(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
}
