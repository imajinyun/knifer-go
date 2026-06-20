# vjwt Quickstart

`vjwt` provides JWT creation, parsing, signature verification, date-claim validation, and multiple signers including HMAC, RSA-PSS, and ECDSA.

## Create and verify tokens with HMAC

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vjwt"
)

func main() {
	key := []byte("secret")
	token, err := vjwt.CreateToken(map[string]any{vjwt.JWTPayloadSubject: "alice"}, key)
	if err != nil {
		panic(err)
	}

	parsed, err := vjwt.ParseToken(token)
	if err != nil {
		panic(err)
	}
	fmt.Println(parsed.Payload(vjwt.JWTPayloadSubject))
	fmt.Println(vjwt.Verify(token, key))
}
```

## Set headers, payloads, and algorithms with options

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vjwt"
)

func main() {
	token, err := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenHeaders(map[string]any{vjwt.JWTHeaderKeyID: "key-1"}),
		vjwt.WithTokenPayload(map[string]any{vjwt.JWTPayloadIssuer: "issuer", vjwt.JWTPayloadSubject: "alice"}),
		vjwt.WithTokenAlgorithm(vjwt.JWTAlgHS384),
		vjwt.WithTokenKey([]byte("secret")),
	)
	if err != nil {
		panic(err)
	}

	parsed, err := vjwt.JWTOf(token)
	if err != nil {
		panic(err)
	}
	fmt.Println(parsed.Header(vjwt.JWTHeaderKeyID), parsed.Algorithm())
}
```

## Build JWTs fluently and validate date claims

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vjwt"
)

func main() {
	now := time.Now()
	j := vjwt.New().
		SetKey([]byte("secret")).
		SetIssuer("go-knifer").
		SetSubject("alice").
		SetIssuedAt(now).
		SetNotBefore(now.Add(-time.Minute)).
		SetExpiresAt(now.Add(time.Hour))

	token, err := j.Sign()
	if err != nil {
		panic(err)
	}

	parsed, err := vjwt.ParseToken(token)
	if err != nil {
		panic(err)
	}
	fmt.Println(parsed.ValidateWithOptions(vjwt.WithValidateTime(now), vjwt.WithValidateLeeway(30)))
}
```

## Use an RSA-PSS signer

```go
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"

	"github.com/imajinyun/go-knifer/vjwt"
)

func main() {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	signer := vjwt.PS256(priv, &priv.PublicKey)

	token, err := vjwt.CreateTokenWithSigner(map[string]any{vjwt.JWTPayloadSubject: "alice"}, signer)
	if err != nil {
		panic(err)
	}

	fmt.Println(vjwt.VerifyWithSigner(token, signer))
}
```
