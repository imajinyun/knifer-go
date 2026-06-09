package jwt

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// 对应 the utility toolkit-jwt JWTValidator。

// JWTValidator JWT 数据校验器，用于：
//   - 算法是否一致
//   - 算法签名是否正确
//   - 字段值是否有效（例如时间未过期等）
//
// 使用方式与 the utility toolkit 类似的链式 API：
//
//	err := gkjwt.OfValidator(token).
//	    ValidateAlgorithm(signer).
//	    ValidateDate(time.Now(), 0).
//	    Err()
type JWTValidator struct {
	jwt *JWT
	err error
}

// OfValidator 由 token 字符串创建校验器。
func OfValidator(token string) *JWTValidator {
	j, err := Of(token)
	v := &JWTValidator{jwt: j, err: err}
	return v
}

// OfValidatorJWT 由 JWT 对象创建校验器。
func OfValidatorJWT(j *JWT) *JWTValidator { return &JWTValidator{jwt: j} }

// JWT 返回底层 JWT 对象。
func (v *JWTValidator) JWT() *JWT { return v.jwt }

// Err 返回首个失败原因；nil 表示全部通过。
func (v *JWTValidator) Err() error { return v.err }

// ValidateAlgorithm 校验头部 alg 与 signer 一致，并验证签名。signer=nil 时使用 JWT 自带 signer；仍为空则失败。
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

// ValidateDate 校验 nbf/exp/iat 时间字段；leeway 为容忍秒数。
func (v *JWTValidator) ValidateDate(now time.Time, leeway int64) *JWTValidator {
	if v.err != nil || v.jwt == nil {
		return v
	}
	if e := ValidateDate(v.jwt, now, leeway); e != nil {
		v.err = e
	}
	return v
}

// ValidateAlgorithm 校验 JWT 算法是否符合预期，并校验签名（包级版本）。
func ValidateAlgorithm(token string, signer JWTSigner) error {
	return OfValidator(token).ValidateAlgorithm(signer).Err()
}

// ValidateDate 校验时间字段：nbf 不能晚于当前时间，exp 不能早于当前时间，iat 不能晚于当前时间。
// leeway 单位为秒，作为容忍空间。
func ValidateDate(j *JWT, now time.Time, leeway int64) error {
	if j == nil {
		return NewJWTError("jwt is nil")
	}
	nowSec := now.Unix()
	if v, ok, err := payloadAsUnix(j.Payload(PayloadNotBefore)); err != nil {
		return NewJWTError("invalid nbf claim: " + err.Error())
	} else if ok {
		if nowSec+leeway < v {
			return NewJWTError("the token is not yet valid (nbf)")
		}
	}
	if v, ok, err := payloadAsUnix(j.Payload(PayloadExpiresAt)); err != nil {
		return NewJWTError("invalid exp claim: " + err.Error())
	} else if ok {
		if nowSec-leeway > v {
			return NewJWTError("the token is expired (exp)")
		}
	}
	if v, ok, err := payloadAsUnix(j.Payload(PayloadIssuedAt)); err != nil {
		return NewJWTError("invalid iat claim: " + err.Error())
	} else if ok {
		if nowSec+leeway < v {
			return NewJWTError("the token issued time is in future (iat)")
		}
	}
	return nil
}

// payloadAsUnix 将 payload 中的时间字段（可能是 float64/int64/string）转为 Unix 秒。
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
