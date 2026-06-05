package jwt

import (
	"encoding/json"
	"strings"
	"time"
)

// JWT 表示一个 JWT 对象，由 Header + Payload + Signer 组成。
//
// 对应 the utility toolkit-jwt JWT。
type JWT struct {
	header  map[string]any
	payload map[string]any
	signer  JWTSigner
	tokens  []string // 解析时保存的三段原始 base64 串
}

// New 创建一个空 JWT。
func New() *JWT {
	return &JWT{
		header:  map[string]any{},
		payload: map[string]any{},
	}
}

// Of 解析已有 token 字符串，得到 JWT 对象。
func Of(token string) (*JWT, error) {
	j := New()
	if err := j.Parse(token); err != nil {
		return nil, err
	}
	return j, nil
}

// Parse 解析 token 字符串到当前 JWT。
func (j *JWT) Parse(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return NewJWTError("token must not be blank")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return JWTErrorf("the token was expected 3 parts, but got %d", len(parts))
	}
	headerJSON, err := b64URLDecode(parts[0])
	if err != nil {
		return wrapJWTError(err, "decode header")
	}
	payloadJSON, err := b64URLDecode(parts[1])
	if err != nil {
		return wrapJWTError(err, "decode payload")
	}
	header := map[string]any{}
	if len(headerJSON) > 0 {
		if err := json.Unmarshal(headerJSON, &header); err != nil {
			return wrapJWTError(err, "unmarshal header")
		}
	}
	payload := map[string]any{}
	if len(payloadJSON) > 0 {
		if err := json.Unmarshal(payloadJSON, &payload); err != nil {
			return wrapJWTError(err, "unmarshal payload")
		}
	}
	j.header = header
	j.payload = payload
	j.tokens = parts
	return nil
}

// SetKey 使用 HMAC 算法（默认 HS256）设置密钥。
// 若 header 已声明算法则使用该算法。
func (j *JWT) SetKey(key []byte) *JWT {
	alg := j.Algorithm()
	if alg == "" {
		alg = AlgHS256
	}
	signer, err := CreateSigner(alg, key)
	if err != nil {
		// 无法创建时降级为 HS256
		signer, _ = NewHMACSigner(AlgHS256, key)
	}
	return j.SetSigner(signer)
}

// SetKeyWithAlgorithm sets the signer with an explicit algorithm and returns any signer creation error.
func (j *JWT) SetKeyWithAlgorithm(key []byte, algorithm string) error {
	algorithm = strings.ToUpper(strings.TrimSpace(algorithm))
	if algorithm == "" {
		algorithm = AlgHS256
	}
	signer, err := CreateSigner(algorithm, key)
	if err != nil {
		return err
	}
	j.SetHeader(HeaderAlgorithm, algorithm)
	j.SetSigner(signer)
	return nil
}

// SetKeyStrict sets the signer using the header alg without silently falling back.
func (j *JWT) SetKeyStrict(key []byte) error {
	return j.SetKeyWithAlgorithm(key, j.Algorithm())
}

// SetSigner 设置签名器；若 header 中无 alg 字段则自动写入。
func (j *JWT) SetSigner(signer JWTSigner) *JWT {
	j.signer = signer
	if _, ok := j.header[HeaderAlgorithm]; !ok {
		j.header[HeaderAlgorithm] = signer.Algorithm()
	}
	return j
}

// Signer 返回当前签名器。
func (j *JWT) Signer() JWTSigner { return j.signer }

// Header 操作。

// Headers 返回头部全部字段（拷贝）。
func (j *JWT) Headers() map[string]any {
	out := make(map[string]any, len(j.header))
	for k, v := range j.header {
		out[k] = v
	}
	return out
}

// Header 取头字段。
func (j *JWT) Header(name string) any { return j.header[name] }

// SetHeader 设置头字段。
func (j *JWT) SetHeader(name string, value any) *JWT {
	j.header[name] = value
	return j
}

// AddHeaders 批量添加头字段。
func (j *JWT) AddHeaders(headers map[string]any) *JWT {
	for k, v := range headers {
		j.header[k] = v
	}
	return j
}

// Algorithm 取得头部 alg 字段。
func (j *JWT) Algorithm() string {
	if v, ok := j.header[HeaderAlgorithm].(string); ok {
		return v
	}
	return ""
}

// Type 取得头部 typ。
func (j *JWT) Type() string {
	if v, ok := j.header[HeaderType].(string); ok {
		return v
	}
	return ""
}

// Payload 操作。

// Payloads 返回载荷全部字段（拷贝）。
func (j *JWT) Payloads() map[string]any {
	out := make(map[string]any, len(j.payload))
	for k, v := range j.payload {
		out[k] = v
	}
	return out
}

// Payload 取载荷字段。
func (j *JWT) Payload(name string) any { return j.payload[name] }

// SetPayload 设置载荷字段。
func (j *JWT) SetPayload(name string, value any) *JWT {
	j.payload[name] = value
	return j
}

// AddPayloads 批量添加载荷字段。
func (j *JWT) AddPayloads(payloads map[string]any) *JWT {
	for k, v := range payloads {
		j.payload[k] = v
	}
	return j
}

// 注册的 Payload 字段便捷方法。

// SetIssuer 设置 iss。
func (j *JWT) SetIssuer(issuer string) *JWT { return j.SetPayload(PayloadIssuer, issuer) }

// SetSubject 设置 sub。
func (j *JWT) SetSubject(subject string) *JWT { return j.SetPayload(PayloadSubject, subject) }

// SetAudience 设置 aud。
func (j *JWT) SetAudience(audience ...string) *JWT {
	if len(audience) == 1 {
		return j.SetPayload(PayloadAudience, audience[0])
	}
	return j.SetPayload(PayloadAudience, audience)
}

// SetExpiresAt 设置 exp（按 Unix 秒）。
func (j *JWT) SetExpiresAt(t time.Time) *JWT {
	return j.SetPayload(PayloadExpiresAt, t.Unix())
}

// SetNotBefore 设置 nbf（按 Unix 秒）。
func (j *JWT) SetNotBefore(t time.Time) *JWT {
	return j.SetPayload(PayloadNotBefore, t.Unix())
}

// SetIssuedAt 设置 iat（按 Unix 秒）。
func (j *JWT) SetIssuedAt(t time.Time) *JWT {
	return j.SetPayload(PayloadIssuedAt, t.Unix())
}

// SetJWTID 设置 jti。
func (j *JWT) SetJWTID(jwtID string) *JWT { return j.SetPayload(PayloadJWTID, jwtID) }

// Sign 进行签名生成 JWT 字符串（自动补 typ=JWT）。
func (j *JWT) Sign() (string, error) {
	return j.SignOpts(true)
}

// SignWith 使用指定签名器签名（自动补 typ=JWT）。
func (j *JWT) SignWith(signer JWTSigner) (string, error) {
	j.SetSigner(signer)
	return j.SignOpts(true)
}

// SignOpts 进行签名；addTypeIfNot=true 时若无 typ 字段则补 JWT。
func (j *JWT) SignOpts(addTypeIfNot bool) (string, error) {
	if j.signer == nil {
		return "", NewJWTError("no signer provided")
	}
	if addTypeIfNot {
		if _, ok := j.header[HeaderType]; !ok {
			j.header[HeaderType] = "JWT"
		}
	}
	if _, ok := j.header[HeaderAlgorithm]; !ok {
		j.header[HeaderAlgorithm] = j.signer.Algorithm()
	}
	hb, err := json.Marshal(j.header)
	if err != nil {
		return "", wrapJWTError(err, "marshal header")
	}
	pb, err := json.Marshal(j.payload)
	if err != nil {
		return "", wrapJWTError(err, "marshal payload")
	}
	headerB64 := b64URLEncode(hb)
	payloadB64 := b64URLEncode(pb)
	sig := j.signer.Sign(headerB64, payloadB64)
	return headerB64 + "." + payloadB64 + "." + sig, nil
}

// MustSign 签名失败时 panic。
func (j *JWT) MustSign() string {
	s, err := j.Sign()
	if err != nil {
		panic(err)
	}
	return s
}

// Verify 使用当前 signer 校验 token 是否合法。
func (j *JWT) Verify() bool { return j.VerifyWith(j.signer) }

// VerifyWith 使用指定 signer 校验。
//
// Verification rules:
//   - nil signer is treated as NoneSigner,
//   - alg=none with a non-None signer returns false,
//   - alg!=none with NoneSigner also returns false.
func (j *JWT) VerifyWith(signer JWTSigner) bool {
	if signer == nil {
		signer = NoneSigner()
	}
	if len(j.tokens) != 3 {
		return false
	}
	alg := j.Algorithm()
	if IsNoneAlg(alg) {
		if _, isNone := signer.(noneSigner); !isNone {
			return false
		}
	} else {
		if _, isNone := signer.(noneSigner); isNone {
			return false
		}
	}
	return signer.Verify(j.tokens[0], j.tokens[1], j.tokens[2])
}

// Validate 在 Verify 基础上校验 nbf/exp/iat 时间字段。
// leeway 为容忍秒数。
func (j *JWT) Validate(leeway int64) bool {
	if !j.Verify() {
		return false
	}
	return j.ValidateAt(time.Now(), leeway)
}

// ValidateAt validates signature and time claims at the provided time.
func (j *JWT) ValidateAt(now time.Time, leeway int64) bool {
	if !j.Verify() {
		return false
	}
	return ValidateDate(j, now, leeway) == nil
}
