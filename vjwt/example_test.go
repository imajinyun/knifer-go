package vjwt_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vjwt"
)

func ExampleCreateJWTToken() {
	token, err := vjwt.CreateJWTToken(exampleJWTPayload(), exampleJWTKey())
	parsed, parseErr := vjwt.ParseJWT(token)
	fmt.Println(err == nil, parseErr == nil, parsed.Payload(vjwt.JWTPayloadSubject))
	// Output: true true alice
}

func ExampleCreateJWTTokenWithSigner() {
	signer := vjwt.HS256(exampleJWTKey())
	token, err := vjwt.CreateJWTTokenWithSigner(exampleJWTPayload(), signer)
	fmt.Println(err == nil, vjwt.VerifyJWTWithSigner(token, signer))
	// Output: true true
}

func ExampleCreateToken() {
	token, err := vjwt.CreateToken(exampleJWTPayload(), exampleJWTKey())
	fmt.Println(err == nil, vjwt.Verify(token, exampleJWTKey()))
	// Output: true true
}

func ExampleCreateTokenWithHeaders() {
	token, err := vjwt.CreateTokenWithHeaders(
		map[string]any{vjwt.JWTHeaderKeyID: "kid-1"},
		exampleJWTPayload(),
		exampleJWTKey(),
	)
	parsed, parseErr := vjwt.ParseToken(token)
	fmt.Println(err == nil, parseErr == nil, parsed.Header(vjwt.JWTHeaderKeyID))
	// Output: true true kid-1
}

func ExampleCreateTokenWithAlgorithm() {
	token, err := vjwt.CreateTokenWithAlgorithm(exampleJWTPayload(), exampleJWTKey(), vjwt.JWTAlgHS384)
	parsed, parseErr := vjwt.ParseToken(token)
	fmt.Println(err == nil, parseErr == nil, parsed.Algorithm(), vjwt.VerifyWithSigner(token, vjwt.HS384(exampleJWTKey())))
	// Output: true true HS384 true
}

func ExampleCreateTokenWithHeadersAndAlgorithm() {
	token, err := vjwt.CreateTokenWithHeadersAndAlgorithm(
		map[string]any{vjwt.JWTHeaderKeyID: "kid-2"},
		exampleJWTPayload(),
		exampleJWTKey(),
		vjwt.JWTAlgHS512,
	)
	parsed, parseErr := vjwt.ParseToken(token)
	fmt.Println(err == nil, parseErr == nil, parsed.Header(vjwt.JWTHeaderKeyID), parsed.Algorithm())
	// Output: true true kid-2 HS512
}

func ExampleCreateTokenWithSigner() {
	signer := vjwt.HS256(exampleJWTKey())
	token, err := vjwt.CreateTokenWithSigner(exampleJWTPayload(), signer)
	fmt.Println(err == nil, vjwt.VerifyWithSigner(token, signer))
	// Output: true true
}

func ExampleCreateTokenWithHeadersAndSigner() {
	signer := vjwt.HS256(exampleJWTKey())
	token, err := vjwt.CreateTokenWithHeadersAndSigner(
		map[string]any{vjwt.JWTHeaderKeyID: "kid-3"},
		exampleJWTPayload(),
		signer,
	)
	parsed, parseErr := vjwt.ParseToken(token)
	fmt.Println(err == nil, parseErr == nil, parsed.Header(vjwt.JWTHeaderKeyID), vjwt.VerifyWithSigner(token, signer))
	// Output: true true kid-3 true
}

func ExampleCreateTokenWithOptions() {
	token, err := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenHeaders(map[string]any{vjwt.JWTHeaderKeyID: "kid-4"}),
		vjwt.WithTokenPayload(exampleJWTPayload()),
		vjwt.WithTokenKey(exampleJWTKey()),
		vjwt.WithTokenAlgorithm(vjwt.JWTAlgHS256),
	)
	parsed, parseErr := vjwt.ParseToken(token)
	fmt.Println(err == nil, parseErr == nil, parsed.Payload(vjwt.JWTPayloadSubject), parsed.Header(vjwt.JWTHeaderKeyID))
	// Output: true true alice kid-4
}

func ExampleCreateTokenWithOptions_strictKey() {
	token, err := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenPayload(exampleJWTPayload()),
		vjwt.WithTokenKey([]byte("weak")),
		vjwt.WithTokenStrictKey(),
	)
	fmt.Println(token == "", err != nil)
	// Output: true true
}

func ExampleCreateTokenWithOptions_signer() {
	signer := vjwt.HS384(exampleJWTKey())
	token, err := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenPayload(exampleJWTPayload()),
		vjwt.WithTokenSigner(signer),
	)
	parsed, parseErr := vjwt.ParseToken(token)
	fmt.Println(err == nil, parseErr == nil, parsed.Algorithm(), vjwt.VerifyWithSigner(token, signer))
	// Output: true true HS384 true
}

func ExampleCreateTokenWithOptions_jsonOptions() {
	marshalCalled := false
	marshal := func(v any) ([]byte, error) {
		marshalCalled = true
		return json.Marshal(v)
	}
	token, err := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenPayload(exampleJWTPayload()),
		vjwt.WithTokenKey(exampleJWTKey()),
		vjwt.WithTokenJSONOptions(vjwt.WithJSONMarshalFunc(marshal)),
	)
	fmt.Println(err == nil, marshalCalled, vjwt.Verify(token, exampleJWTKey()))
	// Output: true true true
}

func ExampleParseJWT() {
	parsed, err := vjwt.ParseJWT(exampleJWTToken())
	fmt.Println(err == nil, parsed.Payload(vjwt.JWTPayloadSubject))
	// Output: true alice
}

func ExampleJWTOf() {
	parsed, err := vjwt.JWTOf(exampleJWTToken())
	fmt.Println(err == nil, parsed.Algorithm(), parsed.Type())
	// Output: true HS256 JWT
}

func ExampleJWTOfWithOptions() {
	unmarshalCalled := false
	unmarshal := func(data []byte, v any) error {
		unmarshalCalled = true
		return json.Unmarshal(data, v)
	}
	parsed, err := vjwt.JWTOfWithOptions(exampleJWTToken(), vjwt.WithJSONUnmarshalFunc(unmarshal))
	fmt.Println(err == nil, unmarshalCalled, parsed.Payload(vjwt.JWTPayloadSubject))
	// Output: true true alice
}

func ExampleParseToken() {
	parsed, err := vjwt.ParseToken(exampleJWTToken())
	fmt.Println(err == nil, parsed.Header(vjwt.JWTHeaderAlgorithm), parsed.Payload(vjwt.JWTPayloadSubject))
	// Output: true HS256 alice
}

func ExampleParseTokenWithOptions() {
	unmarshalCalled := false
	unmarshal := func(data []byte, v any) error {
		unmarshalCalled = true
		return json.Unmarshal(data, v)
	}
	parsed, err := vjwt.ParseTokenWithOptions(exampleJWTToken(), vjwt.WithJSONUnmarshalFunc(unmarshal))
	fmt.Println(err == nil, unmarshalCalled, parsed.Payload(vjwt.JWTPayloadSubject))
	// Output: true true alice
}

func ExampleNewJWT() {
	j := vjwt.NewJWT().
		SetHeader(vjwt.JWTHeaderKeyID, "kid-5").
		SetSubject("alice").
		SetIssuer("issuer-1").
		SetAudience("web", "api").
		SetJWTID("jwt-1").
		SetKey(exampleJWTKey())
	token, err := j.Sign()
	fmt.Println(err == nil, strings.Count(token, "."), j.Header(vjwt.JWTHeaderKeyID), j.Payload(vjwt.JWTPayloadIssuer))
	// Output: true 2 kid-5 issuer-1
}

func ExampleNew() {
	j := vjwt.New().AddHeaders(map[string]any{vjwt.JWTHeaderKeyID: "kid-6"}).AddPayloads(exampleJWTPayload())
	fmt.Println(j.Header(vjwt.JWTHeaderKeyID), j.Payload(vjwt.JWTPayloadSubject))
	// Output: kid-6 alice
}

func ExampleHS256() {
	signer := vjwt.HS256(exampleJWTKey())
	fmt.Println(signer.Algorithm())
	// Output: HS256
}

func ExampleHS384() {
	signer := vjwt.HS384(exampleJWTKey())
	fmt.Println(signer.Algorithm())
	// Output: HS384
}

func ExampleHS512() {
	signer := vjwt.HS512(exampleJWTKey())
	fmt.Println(signer.Algorithm())
	// Output: HS512
}

func ExampleJWTSignerHS256() {
	signer := vjwt.JWTSignerHS256(exampleJWTKey())
	fmt.Println(signer.Algorithm())
	// Output: HS256
}

func ExampleJWTSignerHMAC() {
	signer, err := vjwt.JWTSignerHMAC(vjwt.JWTAlgHS384, exampleJWTKey())
	fmt.Println(err == nil, signer.Algorithm())
	// Output: true HS384
}

func ExampleNewHMACSigner() {
	signer, err := vjwt.NewHMACSigner(vjwt.JWTAlgHS512, exampleJWTKey())
	fmt.Println(err == nil, signer.Algorithm())
	// Output: true HS512
}

func ExampleNewHMACSignerStrict() {
	signer, err := vjwt.NewHMACSignerStrict(vjwt.JWTAlgHS256, bytes.Repeat([]byte("k"), vjwt.MinHMACKeyBytesHS256))
	fmt.Println(err == nil, signer.Algorithm())
	// Output: true HS256
}

func ExampleMustHMACSigner() {
	signer := vjwt.MustHMACSigner(vjwt.JWTAlgHS512, exampleJWTKey())
	fmt.Println(signer.Algorithm())
	// Output: HS512
}

func ExampleCreateSigner() {
	signer, err := vjwt.CreateSigner(vjwt.JWTAlgHS256, exampleJWTKey())
	fmt.Println(err == nil, signer.Algorithm())
	// Output: true HS256
}

func ExampleCreateSignerStrict() {
	signer, err := vjwt.CreateSignerStrict(vjwt.JWTAlgHS256, bytes.Repeat([]byte("k"), vjwt.MinHMACKeyBytesHS256))
	fmt.Println(err == nil, signer.Algorithm())
	// Output: true HS256
}

func ExampleMinHMACKeyBytes() {
	n, err := vjwt.MinHMACKeyBytes(vjwt.JWTAlgHS256)
	fmt.Println(n, err == nil)
	// Output: 32 true
}

func ExampleAlgorithmName() {
	fmt.Println(vjwt.AlgorithmName(vjwt.JWTAlgPS256))
	// Output: SHA256withRSA_PSS
}

func ExampleJWTSignerECDSA() {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer, err := vjwt.JWTSignerECDSA(vjwt.JWTAlgES256, priv, &priv.PublicKey)
	fmt.Println(err == nil, signer.Algorithm())
	// Output: true ES256
}

func ExampleJWTSignerES256() {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer := vjwt.JWTSignerES256(priv, &priv.PublicKey)
	fmt.Println(signer.Algorithm())
	// Output: ES256
}

func ExampleNewECDSASigner() {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer, err := vjwt.NewECDSASigner(vjwt.JWTAlgES256, priv, &priv.PublicKey)
	fmt.Println(err == nil, signer.Algorithm())
	// Output: true ES256
}

func ExampleNewECDSASignerWithOptions() {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer, err := vjwt.NewECDSASignerWithOptions(
		vjwt.JWTAlgES256,
		priv,
		&priv.PublicKey,
		vjwt.WithSignerRandomReader(zeroReader{}),
	)
	fmt.Println(err == nil, signer.Algorithm())
	// Output: true ES256
}

func ExampleES256() {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer := vjwt.ES256(priv, &priv.PublicKey)
	fmt.Println(signer.Algorithm())
	// Output: ES256
}

func ExampleES256WithOptions() {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer := vjwt.ES256WithOptions(priv, &priv.PublicKey, vjwt.WithSignerRandomReader(zeroReader{}))
	fmt.Println(signer.Algorithm())
	// Output: ES256
}

func ExampleES384() {
	priv, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	signer := vjwt.ES384(priv, &priv.PublicKey)
	fmt.Println(signer.Algorithm())
	// Output: ES384
}

func ExampleES384WithOptions() {
	priv, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	signer := vjwt.ES384WithOptions(priv, &priv.PublicKey, vjwt.WithSignerRandomReader(zeroReader{}))
	fmt.Println(signer.Algorithm())
	// Output: ES384
}

func ExampleES512() {
	priv, _ := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	signer := vjwt.ES512(priv, &priv.PublicKey)
	fmt.Println(signer.Algorithm())
	// Output: ES512
}

func ExampleES512WithOptions() {
	priv, _ := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	signer := vjwt.ES512WithOptions(priv, &priv.PublicKey, vjwt.WithSignerRandomReader(zeroReader{}))
	fmt.Println(signer.Algorithm())
	// Output: ES512
}

func ExampleNewRSAPSSSigner() {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer, err := vjwt.NewRSAPSSSigner(vjwt.JWTAlgPS256, priv, &priv.PublicKey)
	fmt.Println(err == nil, signer.Algorithm())
	// Output: true PS256
}

func ExampleNewRSAPSSSignerWithOptions() {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer, err := vjwt.NewRSAPSSSignerWithOptions(
		vjwt.JWTAlgPS256,
		priv,
		&priv.PublicKey,
		vjwt.WithSignerRandomReader(zeroReader{}),
		vjwt.WithRSAPSSOptions(&rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto}),
	)
	fmt.Println(err == nil, signer.Algorithm())
	// Output: true PS256
}

func ExamplePS256() {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer := vjwt.PS256(priv, &priv.PublicKey)
	fmt.Println(signer.Algorithm())
	// Output: PS256
}

func ExamplePS256WithOptions() {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer := vjwt.PS256WithOptions(priv, &priv.PublicKey, vjwt.WithSignerRandomReader(zeroReader{}))
	fmt.Println(signer.Algorithm())
	// Output: PS256
}

func ExamplePS384() {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer := vjwt.PS384(priv, &priv.PublicKey)
	fmt.Println(signer.Algorithm())
	// Output: PS384
}

func ExamplePS384WithOptions() {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer := vjwt.PS384WithOptions(priv, &priv.PublicKey, vjwt.WithSignerRandomReader(zeroReader{}))
	fmt.Println(signer.Algorithm())
	// Output: PS384
}

func ExamplePS512() {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer := vjwt.PS512(priv, &priv.PublicKey)
	fmt.Println(signer.Algorithm())
	// Output: PS512
}

func ExamplePS512WithOptions() {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer := vjwt.PS512WithOptions(priv, &priv.PublicKey, vjwt.WithSignerRandomReader(zeroReader{}))
	fmt.Println(signer.Algorithm())
	// Output: PS512
}

func ExampleVerifyJWT() {
	token := exampleJWTToken()
	fmt.Println(vjwt.VerifyJWT(token, exampleJWTKey()))
	// Output: true
}

func ExampleVerifyJWTWithSigner() {
	token := exampleJWTToken()
	fmt.Println(vjwt.VerifyJWTWithSigner(token, vjwt.HS256(exampleJWTKey())))
	// Output: true
}

func ExampleVerify() {
	token := exampleJWTToken()
	fmt.Println(vjwt.Verify(token, exampleJWTKey()))
	// Output: true
}

func ExampleVerifyStrict() {
	token := exampleJWTToken()
	fmt.Println(vjwt.VerifyStrict(token, exampleJWTKey()))
	// Output: true
}

func ExampleVerifyWithSigner() {
	token := exampleJWTToken()
	fmt.Println(vjwt.VerifyWithSigner(token, vjwt.HS256(exampleJWTKey())))
	// Output: true
}

func ExampleValidateAlgorithm() {
	token := exampleJWTToken()
	fmt.Println(vjwt.ValidateAlgorithm(token, vjwt.HS256(exampleJWTKey())) == nil)
	// Output: true
}

func ExampleValidateDate() {
	now := time.Unix(1_700_000_000, 0)
	j := vjwt.New().
		SetSubject("alice").
		SetNotBefore(now.Add(-time.Minute)).
		SetIssuedAt(now.Add(-time.Second)).
		SetExpiresAt(now.Add(time.Minute)).
		SetKey(exampleJWTKey())
	token, _ := j.Sign()
	parsed, _ := vjwt.ParseToken(token)
	parsed.SetKey(exampleJWTKey())
	fmt.Println(vjwt.ValidateDate(parsed, now, 0) == nil, parsed.ValidateAt(now, 0))
	// Output: true true
}

func ExampleValidateJWTDate() {
	now := time.Unix(1_700_000_000, 0)
	j := vjwt.New().SetExpiresAt(now.Add(time.Minute)).SetKey(exampleJWTKey())
	token, _ := j.Sign()
	parsed, _ := vjwt.ParseToken(token)
	parsed.SetKey(exampleJWTKey())
	fmt.Println(vjwt.ValidateJWTDate(parsed, now, 0) == nil)
	// Output: true
}

func ExampleOfValidator() {
	token := exampleJWTToken()
	validator := vjwt.OfValidator(token).ValidateAlgorithm(vjwt.HS256(exampleJWTKey()))
	fmt.Println(validator.Err() == nil, validator.JWT().Payload(vjwt.JWTPayloadSubject))
	// Output: true alice
}

func ExampleOfValidatorJWT() {
	parsed, _ := vjwt.ParseToken(exampleJWTToken())
	validator := vjwt.OfValidatorJWT(parsed).ValidateAlgorithm(vjwt.HS256(exampleJWTKey()))
	fmt.Println(validator.Err() == nil, validator.JWT() == parsed)
	// Output: true true
}

func ExampleWithValidateTime() {
	now := time.Unix(1_700_000_000, 0)
	parsed := exampleTimedJWT(now)
	fmt.Println(parsed.ValidateWithOptions(vjwt.WithValidateTime(now)))
	// Output: true
}

func ExampleWithValidateClock() {
	now := time.Unix(1_700_000_000, 0)
	parsed := exampleTimedJWT(now)
	fmt.Println(parsed.ValidateWithOptions(vjwt.WithValidateClock(func() time.Time { return now })))
	// Output: true
}

func ExampleWithValidateLeeway() {
	now := time.Unix(1_700_000_000, 0)
	j := vjwt.New().SetExpiresAt(now.Add(-time.Second)).SetKey(exampleJWTKey())
	token, _ := j.Sign()
	parsed, _ := vjwt.ParseToken(token)
	parsed.SetKey(exampleJWTKey())
	fmt.Println(parsed.ValidateWithOptions(vjwt.WithValidateTime(now), vjwt.WithValidateLeeway(2)))
	// Output: true
}

func ExampleWithJSONMarshalFunc() {
	marshalCalled := false
	j := vjwt.New().SetSubject("alice").SetKey(exampleJWTKey())
	_, err := j.SignOptsWithOptions(true, vjwt.WithJSONMarshalFunc(func(v any) ([]byte, error) {
		marshalCalled = true
		return json.Marshal(v)
	}))
	fmt.Println(err == nil, marshalCalled)
	// Output: true true
}

func ExampleWithJSONUnmarshalFunc() {
	unmarshalCalled := false
	parsed, err := vjwt.ParseTokenWithOptions(exampleJWTToken(), vjwt.WithJSONUnmarshalFunc(func(data []byte, v any) error {
		unmarshalCalled = true
		return json.Unmarshal(data, v)
	}))
	fmt.Println(err == nil, unmarshalCalled, parsed.Payload(vjwt.JWTPayloadSubject))
	// Output: true true alice
}

func ExampleWithTokenHeaders() {
	token, _ := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenHeaders(map[string]any{vjwt.JWTHeaderKeyID: "kid-7"}),
		vjwt.WithTokenPayload(exampleJWTPayload()),
		vjwt.WithTokenKey(exampleJWTKey()),
	)
	parsed, _ := vjwt.ParseToken(token)
	fmt.Println(parsed.Header(vjwt.JWTHeaderKeyID))
	// Output: kid-7
}

func ExampleWithTokenPayload() {
	token, _ := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenPayload(map[string]any{vjwt.JWTPayloadSubject: "bob"}),
		vjwt.WithTokenKey(exampleJWTKey()),
	)
	parsed, _ := vjwt.ParseToken(token)
	fmt.Println(parsed.Payload(vjwt.JWTPayloadSubject))
	// Output: bob
}

func ExampleWithTokenKey() {
	token, _ := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenPayload(exampleJWTPayload()),
		vjwt.WithTokenKey(exampleJWTKey()),
	)
	fmt.Println(vjwt.Verify(token, exampleJWTKey()))
	// Output: true
}

func ExampleWithTokenAlgorithm() {
	token, _ := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenPayload(exampleJWTPayload()),
		vjwt.WithTokenKey(exampleJWTKey()),
		vjwt.WithTokenAlgorithm(vjwt.JWTAlgHS512),
	)
	parsed, _ := vjwt.ParseToken(token)
	fmt.Println(parsed.Algorithm())
	// Output: HS512
}

func ExampleWithTokenSigner() {
	token, _ := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenPayload(exampleJWTPayload()),
		vjwt.WithTokenSigner(vjwt.HS384(exampleJWTKey())),
	)
	parsed, _ := vjwt.ParseToken(token)
	fmt.Println(parsed.Algorithm())
	// Output: HS384
}

func ExampleWithTokenStrictKey() {
	_, err := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenPayload(exampleJWTPayload()),
		vjwt.WithTokenKey([]byte("weak")),
		vjwt.WithTokenStrictKey(),
	)
	fmt.Println(err != nil)
	// Output: true
}

func ExampleWithTokenJSONOptions() {
	marshalCalled := false
	_, err := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenPayload(exampleJWTPayload()),
		vjwt.WithTokenKey(exampleJWTKey()),
		vjwt.WithTokenJSONOptions(vjwt.WithJSONMarshalFunc(func(v any) ([]byte, error) {
			marshalCalled = true
			return json.Marshal(v)
		})),
	)
	fmt.Println(err == nil, marshalCalled)
	// Output: true true
}

func ExampleWithSignerRandomReader() {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer := vjwt.PS256WithOptions(priv, &priv.PublicKey, vjwt.WithSignerRandomReader(zeroReader{}))
	fmt.Println(signer.Algorithm())
	// Output: PS256
}

func ExampleWithRSAPSSOptions() {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer := vjwt.PS256WithOptions(
		priv,
		&priv.PublicKey,
		vjwt.WithRSAPSSOptions(&rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto}),
	)
	fmt.Println(signer.Algorithm())
	// Output: PS256
}

func ExampleNewJWTError() {
	err := vjwt.NewJWTError("token must not be blank")
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	// Output: true
}

func ExampleJWTErrorf() {
	err := vjwt.JWTErrorf("code %d: %s", 400, "bad request")
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput), err.Error())
	// Output: true code 400: bad request
}

func exampleJWTKey() []byte {
	return []byte("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
}

func exampleJWTPayload() map[string]any {
	return map[string]any{vjwt.JWTPayloadSubject: "alice"}
}

func exampleJWTToken() string {
	token, err := vjwt.CreateToken(exampleJWTPayload(), exampleJWTKey())
	if err != nil {
		panic(err)
	}
	return token
}

func exampleTimedJWT(now time.Time) *vjwt.JWT {
	j := vjwt.New().
		SetSubject("alice").
		SetNotBefore(now.Add(-time.Minute)).
		SetIssuedAt(now.Add(-time.Second)).
		SetExpiresAt(now.Add(time.Minute)).
		SetKey(exampleJWTKey())
	token, err := j.Sign()
	if err != nil {
		panic(err)
	}
	parsed, err := vjwt.ParseToken(token)
	if err != nil {
		panic(err)
	}
	return parsed.SetKey(exampleJWTKey())
}
