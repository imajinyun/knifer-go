package jwt

import "time"

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
		if _, ok := signer.(noneSigner); ok {
			return v
		}
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
	nowSec := now.Unix()
	if v, ok := payloadAsUnix(j.Payload(PayloadNotBefore)); ok {
		if nowSec+leeway < v {
			return NewJWTError("the token is not yet valid (nbf)")
		}
	}
	if v, ok := payloadAsUnix(j.Payload(PayloadExpiresAt)); ok {
		if nowSec-leeway > v {
			return NewJWTError("the token is expired (exp)")
		}
	}
	if v, ok := payloadAsUnix(j.Payload(PayloadIssuedAt)); ok {
		if nowSec+leeway < v {
			return NewJWTError("the token issued time is in future (iat)")
		}
	}
	return nil
}

// payloadAsUnix 将 payload 中的时间字段（可能是 float64/int64/string）转为 Unix 秒。
func payloadAsUnix(v any) (int64, bool) {
	if v == nil {
		return 0, false
	}
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int64:
		return x, true
	case float32:
		return int64(x), true
	case float64:
		return int64(x), true
	case time.Time:
		return x.Unix(), true
	}
	return 0, false
}
