package jwt

import (
	"encoding/json"
	"maps"
	"strings"
	"time"
)

// JWT represents a JWT object composed of Header, Payload, and Signer.
//
// matches the utility toolkit-jwt JWT.
type JWT struct {
	header  map[string]any
	payload map[string]any
	signer  JWTSigner
	tokens  []string // stores the three raw base64 segments when parsed
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
	return func(c *jsonConfig) {
		if marshal != nil {
			c.marshal = marshal
		}
	}
}

// WithJSONUnmarshalFunc sets the JSON unmarshal provider used by JWT parsing helpers.
func WithJSONUnmarshalFunc(unmarshal func([]byte, any) error) JSONOption {
	return func(c *jsonConfig) {
		if unmarshal != nil {
			c.unmarshal = unmarshal
		}
	}
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

// New creates an empty JWT.
func New() *JWT {
	return &JWT{
		header:  map[string]any{},
		payload: map[string]any{},
	}
}

// Of parses an existing token string into a JWT object.
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

// Parse parses a token string into the current JWT.
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

// SetKey sets the key with an HMAC algorithm, defaulting to HS256.
//
// If the header declares an HMAC algorithm, that algorithm is used; the none
// algorithm is always rejected to avoid alg=none authentication bypasses. This
// compatibility helper falls back to HS256 when signer creation fails. New code
// should prefer SetKeyE, SetKeyStrict, or SetKeyWithAlgorithm when configuration
// errors must be reported explicitly.
func (j *JWT) SetKey(key []byte) *JWT {
	alg := j.Algorithm()
	if alg == "" {
		alg = AlgHS256
	}
	if isNoneAlg(alg) {
		j.signer = nil
		return j
	}
	signer, err := CreateSigner(alg, key)
	if err != nil {
		// Fall back to HS256 when creation fails.
		signer, _ = NewHMACSigner(AlgHS256, key)
	}
	return j.SetSigner(signer)
}

// SetKeyE sets the signer using the current header alg and returns signer creation errors.
// It defaults a blank alg to HS256 and never silently falls back to another algorithm.
func (j *JWT) SetKeyE(key []byte) error {
	return j.SetKeyWithAlgorithm(key, j.Algorithm())
}

// SetKeyWithAlgorithm sets an HMAC signer with an explicit algorithm and returns any signer creation error.
// The none algorithm is always rejected.
func (j *JWT) SetKeyWithAlgorithm(key []byte, algorithm string) error {
	return j.setKeyWithAlgorithm(key, algorithm)
}

func (j *JWT) setKeyWithAlgorithm(key []byte, algorithm string) error {
	algorithm = normalizeAlgorithm(algorithm)
	if algorithm == "" {
		algorithm = AlgHS256
	}
	if isNoneAlg(algorithm) {
		return unsupportedJWTErrorf("jwt alg=none is not supported")
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
// The none algorithm is always rejected.
func (j *JWT) SetKeyStrict(key []byte) error {
	return j.SetKeyWithAlgorithm(key, j.Algorithm())
}

// SetKeyStrictWithMinLength sets an HMAC signer and enforces the recommended minimum key length.
func (j *JWT) SetKeyStrictWithMinLength(key []byte) error {
	algorithm := normalizeAlgorithm(j.Algorithm())
	if algorithm == "" {
		algorithm = AlgHS256
	}
	if isNoneAlg(algorithm) {
		return unsupportedJWTErrorf("jwt alg=none is not supported")
	}
	signer, err := CreateSignerStrict(algorithm, key)
	if err != nil {
		return err
	}
	j.SetHeader(HeaderAlgorithm, algorithm)
	j.SetSigner(signer)
	return nil
}

func normalizeAlgorithm(algorithm string) string {
	algorithm = strings.TrimSpace(algorithm)
	if isNoneAlg(algorithm) {
		return AlgNone
	}
	return strings.ToUpper(algorithm)
}

// SetSigner sets the signer and writes alg automatically when the header has no alg field.
// Passing nil clears the signer and leaves headers unchanged.
func (j *JWT) SetSigner(signer JWTSigner) *JWT {
	j.signer = signer
	if signer == nil {
		return j
	}
	if _, ok := j.header[HeaderAlgorithm]; !ok {
		j.header[HeaderAlgorithm] = signer.Algorithm()
	}
	return j
}

// SetSignerE sets the signer and reports nil signer as an explicit input error.
func (j *JWT) SetSignerE(signer JWTSigner) error {
	if signer == nil {
		return NewJWTError("jwt signer must not be nil")
	}
	j.SetSigner(signer)
	return nil
}

// Signer returns the current signer.
func (j *JWT) Signer() JWTSigner { return j.signer }

// Header operations.

// Headers returns a copy of all header fields.
func (j *JWT) Headers() map[string]any {
	return maps.Clone(j.header)
}

// Header gets a header field.
func (j *JWT) Header(name string) any { return j.header[name] }

// SetHeader sets a header field.
func (j *JWT) SetHeader(name string, value any) *JWT {
	j.header[name] = value
	return j
}

// AddHeaders adds header fields in bulk.
func (j *JWT) AddHeaders(headers map[string]any) *JWT {
	for k, v := range headers {
		j.header[k] = v
	}
	return j
}

// Algorithm gets the header alg field.
func (j *JWT) Algorithm() string {
	if v, ok := j.header[HeaderAlgorithm].(string); ok {
		return v
	}
	return ""
}

// Type gets the header typ field.
func (j *JWT) Type() string {
	if v, ok := j.header[HeaderType].(string); ok {
		return v
	}
	return ""
}

// Payload operations.

// Payloads returns a copy of all payload fields.
func (j *JWT) Payloads() map[string]any {
	return maps.Clone(j.payload)
}

// Payload gets a payload field.
func (j *JWT) Payload(name string) any { return j.payload[name] }

// SetPayload sets a payload field.
func (j *JWT) SetPayload(name string, value any) *JWT {
	j.payload[name] = value
	return j
}

// AddPayloads adds payload fields in bulk.
func (j *JWT) AddPayloads(payloads map[string]any) *JWT {
	for k, v := range payloads {
		j.payload[k] = v
	}
	return j
}

// Convenience methods for registered Payload fields.

// SetIssuer sets iss.
func (j *JWT) SetIssuer(issuer string) *JWT { return j.SetPayload(PayloadIssuer, issuer) }

// SetSubject sets sub.
func (j *JWT) SetSubject(subject string) *JWT { return j.SetPayload(PayloadSubject, subject) }

// SetAudience sets aud.
func (j *JWT) SetAudience(audience ...string) *JWT {
	if len(audience) == 1 {
		return j.SetPayload(PayloadAudience, audience[0])
	}
	return j.SetPayload(PayloadAudience, audience)
}

// SetExpiresAt sets exp in Unix seconds.
func (j *JWT) SetExpiresAt(t time.Time) *JWT {
	return j.SetPayload(PayloadExpiresAt, t.Unix())
}

// SetNotBefore sets nbf in Unix seconds.
func (j *JWT) SetNotBefore(t time.Time) *JWT {
	return j.SetPayload(PayloadNotBefore, t.Unix())
}

// SetIssuedAt sets iat in Unix seconds.
func (j *JWT) SetIssuedAt(t time.Time) *JWT {
	return j.SetPayload(PayloadIssuedAt, t.Unix())
}

// SetJWTID sets jti.
func (j *JWT) SetJWTID(jwtID string) *JWT { return j.SetPayload(PayloadJWTID, jwtID) }

// Sign signs and generates a JWT string, adding typ=JWT automatically.
func (j *JWT) Sign() (string, error) {
	return j.SignOpts(true)
}

// SignWith signs with the specified signer, adding typ=JWT automatically.
func (j *JWT) SignWith(signer JWTSigner) (string, error) {
	j.SetSigner(signer)
	return j.SignOpts(true)
}

// SignOpts signs; when addTypeIfNot is true, it fills typ=JWT if typ is absent.
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
	if sig == "" {
		return "", NewJWTError("signer returned empty signature")
	}
	return headerB64 + "." + payloadB64 + "." + sig, nil
}

// MustSign panics when signing fails.
func (j *JWT) MustSign() string {
	s, err := j.Sign()
	if err != nil {
		panic(err)
	}
	return s
}

// Verify verifies the token with the current signer and returns false when no signer is explicitly set.
func (j *JWT) Verify() bool { return j.VerifyWith(j.signer) }

// VerifyWith verifies with the specified signer.
//
// Verification rules:
//   - nil signer returns false,
//   - alg=none always returns false,
//   - the token alg must match signer.Algorithm().
func (j *JWT) VerifyWith(signer JWTSigner) bool {
	if signer == nil {
		return false
	}
	if len(j.tokens) != 3 {
		return false
	}
	if isNoneAlg(j.Algorithm()) {
		return false
	}
	if normalizeAlgorithm(j.Algorithm()) != normalizeAlgorithm(signer.Algorithm()) {
		return false
	}
	return signer.Verify(j.tokens[0], j.tokens[1], j.tokens[2])
}

// Validate validates nbf, exp, and iat time fields after Verify.
// leeway is the allowed leeway in seconds.
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
