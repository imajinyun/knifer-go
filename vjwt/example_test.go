package vjwt_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vjwt"
)

func ExampleNewJWTError() {
	err := vjwt.NewJWTError("token must not be blank")
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	// Output: true
}

func ExampleCreateTokenWithOptions() {
	token, err := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenHeaders(map[string]any{vjwt.JWTHeaderAlgorithm: vjwt.JWTAlgHS256}),
		vjwt.WithTokenPayload(map[string]any{vjwt.JWTPayloadSubject: "alice"}),
		vjwt.WithTokenKey([]byte("secret")),
	)
	parsed, parseErr := vjwt.ParseToken(token)
	fmt.Println(err == nil, parseErr == nil, parsed.Payload(vjwt.JWTPayloadSubject))
	// Output: true true alice
}

func ExampleNewRSAPSSSignerWithOptions() {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer, err := vjwt.NewRSAPSSSignerWithOptions(
		vjwt.JWTAlgPS256,
		priv,
		nil,
		vjwt.WithRSAPSSOptions(&rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto}),
	)
	fmt.Println(err == nil, signer.Algorithm())
	// Output: true PS256
}

func ExampleMinHMACKeyBytes() {
	n, err := vjwt.MinHMACKeyBytes(vjwt.JWTAlgHS256)
	fmt.Println(n, err == nil)
	// Output: 32 true
}

func ExampleNewHMACSignerStrict() {
	signer, err := vjwt.NewHMACSignerStrict(vjwt.JWTAlgHS256, bytes.Repeat([]byte("k"), 32))
	fmt.Println(err == nil, signer.Algorithm())
	// Output: true HS256
}
