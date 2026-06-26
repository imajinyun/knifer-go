package vjwt

import (
	"time"

	jwtimpl "github.com/imajinyun/knifer-go/internal/jwt"
)

// VerifyJWT verifies a token using HMAC key.
func VerifyJWT(token string, key []byte) bool { return jwtimpl.Verify(token, key) }

// VerifyJWTWithSigner verifies a token using signer.
func VerifyJWTWithSigner(token string, signer JWTSigner) bool {
	return jwtimpl.VerifyWithSigner(token, signer)
}

// Verify verifies a token using an HMAC key.
// The none algorithm is always rejected.
func Verify(token string, key []byte) bool {
	return jwtimpl.Verify(token, key)
}

// VerifyStrict verifies a token using the header algorithm without fallback.
// The none algorithm is always rejected.
func VerifyStrict(token string, key []byte) bool {
	return jwtimpl.VerifyStrict(token, key)
}

// VerifyWithSigner verifies a token using signer.
func VerifyWithSigner(token string, signer JWTSigner) bool {
	return jwtimpl.VerifyWithSigner(token, signer)
}

// ValidateAlgorithm checks whether token's alg header matches signer.
func ValidateAlgorithm(token string, signer JWTSigner) error {
	return jwtimpl.ValidateAlgorithm(token, signer)
}

// ValidateJWTDate validates time based JWT claims.
func ValidateJWTDate(j *JWT, now time.Time, leeway int64) error {
	return jwtimpl.ValidateDate(j, now, leeway)
}

// ValidateDate validates time based JWT claims.
func ValidateDate(j *JWT, now time.Time, leeway int64) error {
	return jwtimpl.ValidateDate(j, now, leeway)
}

// WithValidateTime sets the time used by JWT.ValidateWithOptions.
func WithValidateTime(now time.Time) ValidateOption { return jwtimpl.WithValidateTime(now) }

// WithValidateClock sets the clock used by JWT.ValidateWithOptions.
func WithValidateClock(clock func() time.Time) ValidateOption {
	return jwtimpl.WithValidateClock(clock)
}

// WithValidateLeeway sets the leeway in seconds used by JWT.ValidateWithOptions.
func WithValidateLeeway(leeway int64) ValidateOption { return jwtimpl.WithValidateLeeway(leeway) }
