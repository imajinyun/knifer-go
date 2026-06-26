package vjwt

import jwtimpl "github.com/imajinyun/knifer-go/internal/jwt"

// NewJWT creates a new JWT object.
func NewJWT() *JWT { return jwtimpl.New() }

// ParseJWT parses a token string.
func ParseJWT(token string) (*JWT, error) { return jwtimpl.ParseToken(token) }

// JWTOf parses a token string.
func JWTOf(token string) (*JWT, error) { return jwtimpl.Of(token) }

// JWTOfWithOptions parses a token string with JSON options.
func JWTOfWithOptions(token string, opts ...JSONOption) (*JWT, error) {
	return jwtimpl.OfWithOptions(token, opts...)
}

// New creates a new JWT object.
func New() *JWT {
	return jwtimpl.New()
}

// CreateJWTToken creates a signed token using HMAC key.
func CreateJWTToken(payload map[string]any, key []byte) (string, error) {
	return jwtimpl.CreateToken(payload, key)
}

// CreateJWTTokenWithSigner creates a signed token using signer.
func CreateJWTTokenWithSigner(payload map[string]any, signer JWTSigner) (string, error) {
	return jwtimpl.CreateTokenWithSigner(payload, signer)
}

// CreateToken creates an HS256 token from payload and HMAC key.
func CreateToken(payload map[string]any, key []byte) (string, error) {
	return jwtimpl.CreateToken(payload, key)
}

// CreateTokenWithHeaders creates an HS256 token with custom headers, payload, and HMAC key.
func CreateTokenWithHeaders(headers, payload map[string]any, key []byte) (string, error) {
	return jwtimpl.CreateTokenWithHeaders(headers, payload, key)
}

// CreateTokenWithAlgorithm creates a token with an explicit HMAC algorithm.
func CreateTokenWithAlgorithm(payload map[string]any, key []byte, algorithm string) (string, error) {
	return jwtimpl.CreateTokenWithAlgorithm(payload, key, algorithm)
}

// CreateTokenWithHeadersAndAlgorithm creates a token with headers and an explicit HMAC algorithm.
func CreateTokenWithHeadersAndAlgorithm(headers, payload map[string]any, key []byte, algorithm string) (string, error) {
	return jwtimpl.CreateTokenWithHeadersAndAlgorithm(headers, payload, key, algorithm)
}

// CreateTokenWithSigner creates a token from payload using signer.
func CreateTokenWithSigner(payload map[string]any, signer JWTSigner) (string, error) {
	return jwtimpl.CreateTokenWithSigner(payload, signer)
}

// CreateTokenWithHeadersAndSigner creates a token with custom headers and signer.
func CreateTokenWithHeadersAndSigner(headers, payload map[string]any, signer JWTSigner) (string, error) {
	return jwtimpl.CreateTokenWithHeadersAndSigner(headers, payload, signer)
}

// WithTokenHeaders sets JWT header fields for CreateTokenWithOptions.
func WithTokenHeaders(headers map[string]any) TokenOption { return jwtimpl.WithTokenHeaders(headers) }

// WithTokenPayload sets JWT payload fields for CreateTokenWithOptions.
func WithTokenPayload(payload map[string]any) TokenOption { return jwtimpl.WithTokenPayload(payload) }

// WithTokenKey sets the HMAC key used by CreateTokenWithOptions.
func WithTokenKey(key []byte) TokenOption { return jwtimpl.WithTokenKey(key) }

// WithTokenStrictKey makes CreateTokenWithOptions enforce the recommended HMAC key length.
func WithTokenStrictKey() TokenOption { return jwtimpl.WithTokenStrictKey() }

// WithTokenAlgorithm sets the HMAC algorithm used by CreateTokenWithOptions.
func WithTokenAlgorithm(algorithm string) TokenOption { return jwtimpl.WithTokenAlgorithm(algorithm) }

// WithTokenSigner sets the signer used by CreateTokenWithOptions and takes precedence over key/algorithm options.
func WithTokenSigner(signer JWTSigner) TokenOption { return jwtimpl.WithTokenSigner(signer) }

// WithJSONMarshalFunc sets the JSON marshal provider used by JWT signing helpers.
func WithJSONMarshalFunc(marshal func(any) ([]byte, error)) JSONOption {
	return jwtimpl.WithJSONMarshalFunc(marshal)
}

// WithJSONUnmarshalFunc sets the JSON unmarshal provider used by JWT parsing helpers.
func WithJSONUnmarshalFunc(unmarshal func([]byte, any) error) JSONOption {
	return jwtimpl.WithJSONUnmarshalFunc(unmarshal)
}

// WithTokenJSONOptions sets JSON codec options used when signing in CreateTokenWithOptions.
func WithTokenJSONOptions(opts ...JSONOption) TokenOption {
	return jwtimpl.WithTokenJSONOptions(opts...)
}

// CreateTokenWithOptions creates a token from functional options and avoids adding more overload variants.
func CreateTokenWithOptions(opts ...TokenOption) (string, error) {
	return jwtimpl.CreateTokenWithOptions(opts...)
}

// ParseToken parses a token string.
func ParseToken(token string) (*JWT, error) {
	return jwtimpl.ParseToken(token)
}

// ParseTokenWithOptions parses a token string with JSON options.
func ParseTokenWithOptions(token string, opts ...JSONOption) (*JWT, error) {
	return jwtimpl.ParseTokenWithOptions(token, opts...)
}

// OfValidator creates a validator from token string.
func OfValidator(token string) *JWTValidator {
	return jwtimpl.OfValidator(token)
}

// OfValidatorJWT creates a validator from an existing JWT object.
func OfValidatorJWT(j *JWT) *JWTValidator {
	return jwtimpl.OfValidatorJWT(j)
}
