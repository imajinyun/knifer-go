package jwt

// 包级便捷函数（对应 the utility toolkit-jwt JWTUtil）。

type tokenConfig struct {
	headers map[string]any
	payload map[string]any
	key     []byte
	alg     string
	signer  JWTSigner
	json    []JSONOption
}

// TokenOption customizes CreateTokenWithOptions.
type TokenOption func(*tokenConfig)

// WithTokenHeaders sets JWT header fields for CreateTokenWithOptions.
func WithTokenHeaders(headers map[string]any) TokenOption {
	return func(c *tokenConfig) { c.headers = headers }
}

// WithTokenPayload sets JWT payload fields for CreateTokenWithOptions.
func WithTokenPayload(payload map[string]any) TokenOption {
	return func(c *tokenConfig) { c.payload = payload }
}

// WithTokenKey sets the HMAC key used by CreateTokenWithOptions.
func WithTokenKey(key []byte) TokenOption {
	return func(c *tokenConfig) { c.key = append([]byte(nil), key...) }
}

// WithTokenAlgorithm sets the HMAC algorithm used by CreateTokenWithOptions.
func WithTokenAlgorithm(algorithm string) TokenOption {
	return func(c *tokenConfig) { c.alg = algorithm }
}

// WithTokenSigner sets the signer used by CreateTokenWithOptions and takes precedence over key/algorithm options.
func WithTokenSigner(signer JWTSigner) TokenOption { return func(c *tokenConfig) { c.signer = signer } }

// WithTokenJSONOptions sets JSON codec options used when signing in CreateTokenWithOptions.
func WithTokenJSONOptions(opts ...JSONOption) TokenOption {
	return func(c *tokenConfig) { c.json = append(c.json, opts...) }
}

func applyTokenOptions(opts []TokenOption) tokenConfig {
	cfg := tokenConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// CreateToken 用 HS256 创建 token。
func CreateToken(payload map[string]any, key []byte) (string, error) {
	return CreateTokenWithHeaders(nil, payload, key)
}

// CreateTokenWithHeaders 创建带 header 的 token（HS256）。
func CreateTokenWithHeaders(headers, payload map[string]any, key []byte) (string, error) {
	j := New().AddHeaders(headers).AddPayloads(payload).SetKey(key)
	return j.Sign()
}

// CreateTokenWithAlgorithm creates a token with an explicit HMAC algorithm.
func CreateTokenWithAlgorithm(payload map[string]any, key []byte, algorithm string) (string, error) {
	return CreateTokenWithHeadersAndAlgorithm(nil, payload, key, algorithm)
}

// CreateTokenWithHeadersAndAlgorithm creates a token with headers and an explicit HMAC algorithm.
func CreateTokenWithHeadersAndAlgorithm(headers, payload map[string]any, key []byte, algorithm string) (string, error) {
	j := New().AddHeaders(headers).AddPayloads(payload)
	if err := j.SetKeyWithAlgorithm(key, algorithm); err != nil {
		return "", err
	}
	return j.Sign()
}

// CreateTokenWithSigner 使用自定义签名器创建 token。
func CreateTokenWithSigner(payload map[string]any, signer JWTSigner) (string, error) {
	return CreateTokenWithHeadersAndSigner(nil, payload, signer)
}

// CreateTokenWithHeadersAndSigner 使用自定义签名器与 header 创建 token。
func CreateTokenWithHeadersAndSigner(headers, payload map[string]any, signer JWTSigner) (string, error) {
	j := New().AddHeaders(headers).AddPayloads(payload).SetSigner(signer)
	return j.Sign()
}

// CreateTokenWithOptions creates a token from functional options and avoids adding more overload variants.
func CreateTokenWithOptions(opts ...TokenOption) (string, error) {
	cfg := applyTokenOptions(opts)
	j := New().AddHeaders(cfg.headers).AddPayloads(cfg.payload)
	if cfg.signer != nil {
		return j.SetSigner(cfg.signer).SignOptsWithOptions(true, cfg.json...)
	}
	if cfg.alg != "" {
		if err := j.SetKeyWithAlgorithm(cfg.key, cfg.alg); err != nil {
			return "", err
		}
		return j.SignOptsWithOptions(true, cfg.json...)
	}
	return j.SetKey(cfg.key).SignOptsWithOptions(true, cfg.json...)
}

// ParseToken 解析 token。
func ParseToken(token string) (*JWT, error) { return Of(token) }

// ParseTokenWithOptions parses a token with JSON options.
func ParseTokenWithOptions(token string, opts ...JSONOption) (*JWT, error) {
	return OfWithOptions(token, opts...)
}

// Verify verifies a token using the algorithm declared by the token header.
// The none algorithm is always rejected.
func Verify(token string, key []byte) bool {
	return VerifyStrict(token, key)
}

// VerifyStrict verifies a token using the header algorithm without fallback.
// The none algorithm is always rejected.
func VerifyStrict(token string, key []byte) bool {
	j, err := Of(token)
	if err != nil {
		return false
	}
	if err := j.SetKeyStrict(key); err != nil {
		return false
	}
	return j.Verify()
}

// VerifyWithSigner 使用自定义 signer 校验 token。
func VerifyWithSigner(token string, signer JWTSigner) bool {
	j, err := Of(token)
	if err != nil {
		return false
	}
	return j.VerifyWith(signer)
}
