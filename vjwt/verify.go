package vjwt

import (
	"time"

	jwtimpl "github.com/imajinyun/go-knifer/internal/jwt"
)

// VerifyJWT verifies a token using HMAC key.
func VerifyJWT(token string, key []byte) bool { return jwtimpl.Verify(token, key) }

// VerifyJWTWithSigner verifies a token using signer.
func VerifyJWTWithSigner(token string, signer JWTSigner) bool {
	return jwtimpl.VerifyWithSigner(token, signer)
}

// Verify delegates to the internal jwt implementation.
func Verify(token string, key []byte) bool {
	return jwtimpl.Verify(token, key)
}

// VerifyWithSigner delegates to the internal jwt implementation.
func VerifyWithSigner(token string, signer JWTSigner) bool {
	return jwtimpl.VerifyWithSigner(token, signer)
}

// ValidateAlgorithm delegates to the internal jwt implementation.
func ValidateAlgorithm(token string, signer JWTSigner) error {
	return jwtimpl.ValidateAlgorithm(token, signer)
}

// ValidateJWTDate validates time based JWT claims.
func ValidateJWTDate(j *JWT, now time.Time, leeway int64) error {
	return jwtimpl.ValidateDate(j, now, leeway)
}

// ValidateDate delegates to the internal jwt implementation.
func ValidateDate(j *JWT, now time.Time, leeway int64) error {
	return jwtimpl.ValidateDate(j, now, leeway)
}
