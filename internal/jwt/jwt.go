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

type validateConfig struct {
	now    func() time.Time
	leeway int64
}

type jsonConfig struct {
	marshal   func(any) ([]byte, error)
	unmarshal func([]byte, any) error
}

// ValidateOption customizes JWT ValidateWithOptions.
type ValidateOption func(*validateConfig)

// JSONOption customizes JWT JSON encoding and decoding per call.
type JSONOption func(*jsonConfig)

// WithJSONMarshalFunc sets the JSON marshal provider used by JWT signing helpers.
func WithJSONMarshalFunc(marshal func(any) ([]byte, error)) JSONOption {
	return func(c *jsonConfig) { c.marshal = marshal }
}

// WithJSONUnmarshalFunc sets the JSON unmarshal provider used by JWT parsing helpers.
func WithJSONUnmarshalFunc(unmarshal func([]byte, any) error) JSONOption {
	return func(c *jsonConfig) { c.unmarshal = unmarshal }
}

func applyJSONOptions(opts []JSONOption) jsonConfig {
	cfg := jsonConfig{marshal: json.Marshal, unmarshal: json.Unmarshal}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.marshal == nil {
		cfg.marshal = json.Marshal
	}
	if cfg.unmarshal == nil {
		cfg.unmarshal = json.Unmarshal
	}
	return cfg
}

// WithValidateTime sets the time used by ValidateWithOptions.
func WithValidateTime(now time.Time) ValidateOption {
	return func(c *validateConfig) { c.now = func() time.Time { return now } }
}

// WithValidateClock sets the clock used by ValidateWithOptions.
func WithValidateClock(clock func() time.Time) ValidateOption {
	return func(c *validateConfig) {
		if clock != nil {
			c.now = clock
		}
	}
}

// WithValidateLeeway sets the leeway in seconds used by ValidateWithOptions.
func WithValidateLeeway(leeway int64) ValidateOption {
	return func(c *validateConfig) { c.leeway = leeway }
}

func applyValidateOptions(opts []ValidateOption) validateConfig {
	cfg := validateConfig{now: time.Now}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.now == nil {
		cfg.now = time.Now
	}
	return cfg
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
	return OfWithOptions(token)
}

// OfWithOptions parses an existing token string with JSON options.
func OfWithOptions(token string, opts ...JSONOption) (*JWT, error) {
	j := New()
	if err := j.ParseWithOptions(token, opts...); err != nil {
		return nil, err
	}
	return j, nil
}

// Parse 解析 token 字符串到当前 JWT。
func (j *JWT) Parse(token string) error {
	return j.ParseWithOptions(token)
}

// ParseWithOptions parses token string to current JWT with JSON options.
func (j *JWT) ParseWithOptions(token string, opts ...JSONOption) error {
	cfg := applyJSONOptions(opts)
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
		if err := cfg.unmarshal(headerJSON, &header); err != nil {
			return wrapJWTError(err, "unmarshal header")
		}
	}
	payload := map[string]any{}
	if len(payloadJSON) > 0 {
		if err := cfg.unmarshal(payloadJSON, &payload); err != nil {
			return wrapJWTError(err, "unmarshal payload")
		}
	}
	j.header = header
	j.payload = payload
	j.tokens = parts
	return nil
}

// SetKey 使用 HMAC 算法（默认 HS256）设置密钥。
// 若 header 已声明 HMAC 算法则使用该算法；为避免 alg=none 认证绕过，none 算法不会被隐式接受。
// 如需处理已信任的 none token，请显式使用 SetSigner(NoneSigner()) 或 SetKeyAllowNoneForTrustedToken。
func (j *JWT) SetKey(key []byte) *JWT {
	alg := j.Algorithm()
	if alg == "" {
		alg = AlgHS256
	}
	if IsNoneAlg(alg) {
		j.signer = nil
		return j
	}
	signer, err := CreateSigner(alg, key)
	if err != nil {
		// 无法创建时降级为 HS256
		signer, _ = NewHMACSigner(AlgHS256, key)
	}
	return j.SetSigner(signer)
}

// SetKeyWithAlgorithm sets an HMAC signer with an explicit algorithm and returns any signer creation error.
// The none algorithm is rejected by default; use SetSigner(NoneSigner()) or SetKeyAllowNoneForTrustedToken for explicit opt-in.
func (j *JWT) SetKeyWithAlgorithm(key []byte, algorithm string) error {
	return j.setKeyWithAlgorithm(key, algorithm, false)
}

// SetKeyAllowNoneForTrustedToken sets the signer using the requested/header algorithm and explicitly opts in to alg=none.
// Only use this for already trusted tokens; untrusted tokens should use SetKeyWithAlgorithm or SetKeyStrict.
func (j *JWT) SetKeyAllowNoneForTrustedToken(key []byte) error {
	return j.setKeyWithAlgorithm(key, j.Algorithm(), true)
}

func (j *JWT) setKeyWithAlgorithm(key []byte, algorithm string, allowNone bool) error {
	algorithm = normalizeAlgorithm(algorithm)
	if algorithm == "" {
		algorithm = AlgHS256
	}
	if IsNoneAlg(algorithm) && !allowNone {
		return NewJWTError("jwt alg=none requires explicit none signer opt-in")
	}
	signer, err := CreateSigner(algorithm, key)
	if allowNone {
		signer, err = CreateSignerAllowNoneForTrustedToken(algorithm, key)
	}
	if err != nil {
		return err
	}
	j.SetHeader(HeaderAlgorithm, algorithm)
	j.SetSigner(signer)
	return nil
}

// SetKeyStrict sets the signer using the header alg without silently falling back.
// The none algorithm is rejected unless explicitly opted in with SetSigner(NoneSigner()) or SetKeyAllowNoneForTrustedToken.
func (j *JWT) SetKeyStrict(key []byte) error {
	return j.SetKeyWithAlgorithm(key, j.Algorithm())
}

func normalizeAlgorithm(algorithm string) string {
	algorithm = strings.TrimSpace(algorithm)
	if IsNoneAlg(algorithm) {
		return AlgNone
	}
	return strings.ToUpper(algorithm)
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
	return j.SignOptsWithOptions(addTypeIfNot)
}

// SignOptsWithOptions signs the token with JSON options.
func (j *JWT) SignOptsWithOptions(addTypeIfNot bool, opts ...JSONOption) (string, error) {
	cfg := applyJSONOptions(opts)
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
	hb, err := cfg.marshal(j.header)
	if err != nil {
		return "", wrapJWTError(err, "marshal header")
	}
	pb, err := cfg.marshal(j.payload)
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

// Verify 使用当前 signer 校验 token 是否合法；未显式设置 signer 时返回 false。
func (j *JWT) Verify() bool { return j.VerifyWith(j.signer) }

// VerifyWith 使用指定 signer 校验。
//
// Verification rules:
//   - nil signer returns false,
//   - alg=none with a non-None signer returns false,
//   - alg!=none with NoneSigner also returns false.
func (j *JWT) VerifyWith(signer JWTSigner) bool {
	if signer == nil {
		return false
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
	return j.ValidateWithOptions(WithValidateLeeway(leeway))
}

// ValidateWithOptions validates signature and time claims using custom validation options.
func (j *JWT) ValidateWithOptions(opts ...ValidateOption) bool {
	cfg := applyValidateOptions(opts)
	return j.validateAt(cfg.now(), cfg.leeway)
}

// ValidateAt validates signature and time claims at the provided time.
func (j *JWT) ValidateAt(now time.Time, leeway int64) bool {
	return j.validateAt(now, leeway)
}

func (j *JWT) validateAt(now time.Time, leeway int64) bool {
	if !j.Verify() {
		return false
	}
	return ValidateDate(j, now, leeway) == nil
}
