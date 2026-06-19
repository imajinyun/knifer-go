package vcrypto_test

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vcrypto"
)

func ExampleSHA256Hex() {
	fmt.Println(vcrypto.SHA256Hex("abc"))
	// Output: ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
}

func ExampleHMACSHA256Hex() {
	fmt.Println(vcrypto.HMACSHA256Hex([]byte("key"), []byte("data")))
	// Output: 5031fe3d989c6d1537a013fa6e739da23463fdaec3b70137d828e36ace221bd0
}

func ExampleValidateAESKey() {
	err := vcrypto.ValidateAESKey([]byte("too-short"))
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	fmt.Println(errors.Is(err, vcrypto.ErrInvalidKey))
	// Output:
	// true
	// true
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

func ExampleAESEncryptGCM() {
	key := []byte("0123456789abcdef")
	nonce := []byte("123456789012")
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
