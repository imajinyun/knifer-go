package jwt

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Matches the utility toolkit-jwt JWTValidator.

// JWTValidator JWT data validator used for:
//   - algorithm consistency
//   - signature correctness
//   - field validity, such as unexpired time fields
//
// Chainable API similar to the utility toolkit:
//
//	err := gkjwt.OfValidator(token).
//	    ValidateAlgorithm(signer).
//	    ValidateDate(time.Now(), 0).
//	    Err()
type JWTValidator struct {
	jwt *JWT
	err error
}

// OfValidator creates a validator from a token string.
func OfValidator(token string) *JWTValidator {
	j, err := Of(token)
	v := &JWTValidator{jwt: j, err: err}
	return v
}

// OfValidatorJWT creates a validator from a JWT object.
func OfValidatorJWT(j *JWT) *JWTValidator { return &JWTValidator{jwt: j} }

// JWT returns the underlying JWT object.
func (v *JWTValidator) JWT() *JWT { return v.jwt }

// Err returns the first failure reason; nil means all checks passed.
func (v *JWTValidator) Err() error { return v.err }

// ValidateAlgorithm validates that header alg matches signer and verifies the signature. When signer is nil, the JWT signer is used; validation fails if it is still nil.
func (v *JWTValidator) ValidateAlgorithm(signer JWTSigner) *JWTValidator {
	if v.err != nil || v.jwt == nil {
		return v
	}
	if signer == nil {
		signer = v.jwt.Signer()
	}
	alg := v.jwt.Algorithm()
	if alg == "" {
		v.err = NewJWTError("No algorithm defined in header!")
		return v
	}
	if signer == nil {
		v.err = NewJWTError("No Signer for validate algorithm!")
		return v
	}
	if alg != signer.Algorithm() {
		v.err = JWTErrorf("Algorithm [%s] defined in header doesn't match to [%s]!", alg, signer.Algorithm())
		return v
	}
	if !v.jwt.VerifyWith(signer) {
		v.err = NewJWTError("Signature verification failed!")
	}
	return v
}

// ValidateDate validates nbf, exp, and iat time fields; leeway is the allowed leeway in seconds.
func (v *JWTValidator) ValidateDate(now time.Time, leeway int64) *JWTValidator {
	if v.err != nil || v.jwt == nil {
		return v
	}
	if e := ValidateDate(v.jwt, now, leeway); e != nil {
		v.err = e
	}
	return v
}

// ValidateAlgorithm validates that the JWT algorithm matches expectations and verifies the signature as a package-level helper.
func ValidateAlgorithm(token string, signer JWTSigner) error {
	return OfValidator(token).ValidateAlgorithm(signer).Err()
}

// ValidateDate validates time fields: nbf must not be after now, exp must not be before now, and iat must not be after now.
// leeway is in seconds and acts as tolerance.
func ValidateDate(j *JWT, now time.Time, leeway int64) error {
	if j == nil {
		return NewJWTError("jwt is nil")
	}
	nowSec := now.Unix()
	if v, ok, err := payloadAsUnix(j.Payload(PayloadNotBefore)); err != nil {
		return wrapJWTError(err, "invalid nbf claim")
	} else if ok {
		if nowSec+leeway < v {
			return NewJWTError("the token is not yet valid (nbf)")
		}
	}
	if v, ok, err := payloadAsUnix(j.Payload(PayloadExpiresAt)); err != nil {
		return wrapJWTError(err, "invalid exp claim")
	} else if ok {
		if nowSec-leeway > v {
			return NewJWTError("the token is expired (exp)")
		}
	}
	if v, ok, err := payloadAsUnix(j.Payload(PayloadIssuedAt)); err != nil {
		return wrapJWTError(err, "invalid iat claim")
	} else if ok {
		if nowSec+leeway < v {
			return NewJWTError("the token issued time is in future (iat)")
		}
	}
	return nil
}

// payloadAsUnix converts a time field in payload, possibly float64, int64, or string, into Unix seconds.
func payloadAsUnix(v any) (int64, bool, error) {
	if v == nil {
		return 0, false, nil
	}
	switch x := v.(type) {
	case int:
		return int64(x), true, nil
	case int8:
		return int64(x), true, nil
	case int16:
		return int64(x), true, nil
	case int32:
		return int64(x), true, nil
	case int64:
		return x, true, nil
	case uint:
		return uintToInt64(uint64(x))
	case uint8:
		return int64(x), true, nil
	case uint16:
		return int64(x), true, nil
	case uint32:
		return int64(x), true, nil
	case uint64:
		return uintToInt64(x)
	case float32:
		return floatToInt64(float64(x))
	case float64:
		return floatToInt64(x)
	case string:
		return stringToInt64(x)
	case time.Time:
		return x.Unix(), true, nil
	}
	return 0, false, fmt.Errorf("unsupported type %T", v)
}

func uintToInt64(v uint64) (int64, bool, error) {
	if v > math.MaxInt64 {
		return 0, false, fmt.Errorf("value overflows int64")
	}
	return int64(v), true, nil
}

func floatToInt64(v float64) (int64, bool, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false, fmt.Errorf("value is not finite")
	}
	if math.Trunc(v) != v {
		return 0, false, fmt.Errorf("value is not an integer")
	}
	if v < math.MinInt64 || v > math.MaxInt64 {
		return 0, false, fmt.Errorf("value overflows int64")
	}
	return int64(v), true, nil
}

func stringToInt64(v string) (int64, bool, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false, fmt.Errorf("value is blank")
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return n, true, nil
}
