# vcrypto Quickstart

`vcrypto` provides common cryptographic helpers, including digests, HMAC, AES-GCM, random bytes, PBKDF2, RSA encryption/decryption/signing, and PEM conversion.

## SHA and HMAC

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcrypto"
)

func main() {
	digest := vcrypto.SHA256Hex("hello")
	mac := vcrypto.HMACSHA256Hex([]byte("secret"), []byte("hello"))

	fmt.Println(digest)
	fmt.Println(mac)
}
```

When the hash factory is `nil`, `HMACHex` and `HMACBytes` fall back to SHA-256 instead of panicking. Prefer passing an explicit hash when interoperability matters.

## AES-GCM encryption and decryption

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcrypto"
)

func main() {
	key, err := vcrypto.GenAESKey(32)
	if err != nil {
		panic(err)
	}

	nonce, cipherText, err := vcrypto.AESSealGCM([]byte("secret data"), key, []byte("aad"))
	if err != nil {
		panic(err)
	}

	plain, err := vcrypto.AESOpenGCM(cipherText, key, nonce, []byte("aad"))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(plain))
}
```

Authentication failures from `AESOpenGCM` / `AESDecryptGCM` match `vcrypto.ErrInvalidCipherText`, so callers can distinguish tampering or wrong AAD from nonce-length validation errors.

## Derive keys with PBKDF2

```go
package main

import (
	"crypto/sha256"
	"fmt"

	"github.com/imajinyun/go-knifer/vcrypto"
)

func main() {
	key, err := vcrypto.PBKDF2([]byte("password"), []byte("salt"), 10000, 32, sha256.New)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(key))
}
```

## RSA signing and verification

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcrypto"
)

func main() {
	priv, err := vcrypto.GenRSAKey(2048)
	if err != nil {
		panic(err)
	}

	data := []byte("message")
	sig, err := vcrypto.SignSHA256WithRSA(data, priv)
	if err != nil {
		panic(err)
	}

	err = vcrypto.VerifySHA256WithRSA(data, sig, &priv.PublicKey)
	fmt.Println(err == nil)
}
```
