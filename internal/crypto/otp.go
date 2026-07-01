package crypto

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	knifer "github.com/imajinyun/knifer-go"
)

const (
	defaultOTPDigits = 6
	defaultTOTPStep  = 30 * time.Second
)

type otpConfig struct {
	digits int
	hash   func() hash.Hash
	step   time.Duration
	window int
	clock  func() time.Time
}

// OTPOption customizes HOTP and TOTP helpers.
type OTPOption func(*otpConfig)

// WithOTPDigits sets the number of decimal digits in generated OTP codes.
func WithOTPDigits(digits int) OTPOption {
	return func(c *otpConfig) { c.digits = digits }
}

// WithOTPHash sets the HMAC hash used by HOTP/TOTP helpers.
func WithOTPHash(fn func() hash.Hash) OTPOption {
	return func(c *otpConfig) {
		if fn != nil {
			c.hash = fn
		}
	}
}

// WithTOTPStep sets the TOTP time step.
func WithTOTPStep(step time.Duration) OTPOption {
	return func(c *otpConfig) { c.step = step }
}

// WithTOTPWindow sets the number of adjacent time steps accepted by TOTPVerify.
func WithTOTPWindow(window int) OTPOption {
	return func(c *otpConfig) { c.window = window }
}

// WithOTPClock sets the clock used by TOTPNow and TOTPVerifyNow.
func WithOTPClock(clock func() time.Time) OTPOption {
	return func(c *otpConfig) {
		if clock != nil {
			c.clock = clock
		}
	}
}

func applyOTPOptions(opts []OTPOption) otpConfig {
	cfg := otpConfig{
		digits: defaultOTPDigits,
		hash:   sha1.New,
		step:   defaultTOTPStep,
		clock:  time.Now,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.hash == nil {
		cfg.hash = sha1.New
	}
	if cfg.clock == nil {
		cfg.clock = time.Now
	}
	return cfg
}

func validateOTPConfig(secret []byte, cfg otpConfig) error {
	if len(secret) == 0 {
		return knifer.WrapError(knifer.ErrCodeInvalidInput, "otp secret must not be empty", ErrInvalidKey)
	}
	if cfg.digits < 6 || cfg.digits > 9 {
		return knifer.WrapError(knifer.ErrCodeInvalidInput, "otp digits must be between 6 and 9", ErrInvalidOTP)
	}
	if cfg.hash == nil {
		return knifer.WrapError(knifer.ErrCodeInvalidInput, "otp hash must not be nil", ErrInvalidOTP)
	}
	if cfg.step < time.Second {
		return knifer.WrapError(knifer.ErrCodeInvalidInput, "totp step must be at least one second", ErrInvalidOTP)
	}
	if cfg.window < 0 {
		return knifer.WrapError(knifer.ErrCodeInvalidInput, "totp window must not be negative", ErrInvalidOTP)
	}
	return nil
}

// GenerateOTPSecret returns random bytes suitable for HOTP/TOTP secrets.
func GenerateOTPSecret(size int, opts ...RandomOption) ([]byte, error) {
	if size <= 0 {
		return nil, knifer.WrapError(knifer.ErrCodeInvalidInput, "otp secret size must be positive", ErrInvalidKey)
	}
	return RandomBytesWithOptions(size, opts...)
}

// OTPSecretBase32 encodes a binary OTP secret using unpadded Base32.
func OTPSecretBase32(secret []byte) string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)
}

// ParseOTPSecretBase32 decodes an unpadded or padded Base32 OTP secret.
func ParseOTPSecretBase32(secret string) ([]byte, error) {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(secret), " ", ""))
	if normalized == "" {
		return nil, knifer.WrapError(knifer.ErrCodeInvalidInput, "otp secret must not be empty", ErrInvalidKey)
	}
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.TrimRight(normalized, "="))
	if err != nil {
		decoded, err = base32.StdEncoding.DecodeString(normalized)
	}
	if err != nil || len(decoded) == 0 {
		return nil, knifer.WrapError(knifer.ErrCodeInvalidInput, "otp secret must be valid base32", ErrInvalidKey)
	}
	return decoded, nil
}

// HOTP generates an HMAC-based one-time password for counter.
func HOTP(secret []byte, counter uint64, opts ...OTPOption) (string, error) {
	cfg := applyOTPOptions(opts)
	if err := validateOTPConfig(secret, cfg); err != nil {
		return "", err
	}
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	mac := hmac.New(cfg.hash, secret)
	_, _ = mac.Write(buf[:])
	sum := mac.Sum(nil)
	if len(sum) < 20 {
		return "", knifer.WrapError(knifer.ErrCodeInvalidInput, "otp hash output is too short", ErrInvalidOTP)
	}
	offset := sum[len(sum)-1] & 0x0f
	binCode := (uint32(sum[offset])&0x7f)<<24 |
		(uint32(sum[offset+1])&0xff)<<16 |
		(uint32(sum[offset+2])&0xff)<<8 |
		(uint32(sum[offset+3]) & 0xff)
	mod := uint32(math.Pow10(cfg.digits))
	code := binCode % mod
	return fmt.Sprintf("%0*d", cfg.digits, code), nil
}

// HOTPVerify verifies an HOTP code for counter using constant-time comparison.
func HOTPVerify(code string, secret []byte, counter uint64, opts ...OTPOption) (bool, error) {
	cfg := applyOTPOptions(opts)
	if err := validateOTPConfig(secret, cfg); err != nil {
		return false, err
	}
	if err := validateOTPCode(code, cfg.digits); err != nil {
		return false, err
	}
	expected, err := HOTP(secret, counter, opts...)
	if err != nil {
		return false, err
	}
	return otpCodeEqual(code, expected), nil
}

// TOTP generates a time-based one-time password for t.
func TOTP(secret []byte, t time.Time, opts ...OTPOption) (string, error) {
	cfg := applyOTPOptions(opts)
	if err := validateOTPConfig(secret, cfg); err != nil {
		return "", err
	}
	counter := uint64(t.Unix() / int64(cfg.step/time.Second))
	return HOTP(secret, counter, opts...)
}

// TOTPNow generates a time-based one-time password using the configured clock.
func TOTPNow(secret []byte, opts ...OTPOption) (string, error) {
	cfg := applyOTPOptions(opts)
	return TOTP(secret, cfg.clock(), opts...)
}

// TOTPVerify verifies a TOTP code for t and the configured time-step window.
func TOTPVerify(code string, secret []byte, t time.Time, opts ...OTPOption) (bool, error) {
	cfg := applyOTPOptions(opts)
	if err := validateOTPConfig(secret, cfg); err != nil {
		return false, err
	}
	if err := validateOTPCode(code, cfg.digits); err != nil {
		return false, err
	}
	baseCounter := t.Unix() / int64(cfg.step/time.Second)
	for offset := -cfg.window; offset <= cfg.window; offset++ {
		counter := baseCounter + int64(offset)
		if counter < 0 {
			continue
		}
		expected, err := HOTP(secret, uint64(counter), opts...)
		if err != nil {
			return false, err
		}
		if otpCodeEqual(code, expected) {
			return true, nil
		}
	}
	return false, nil
}

// TOTPVerifyNow verifies a TOTP code using the configured clock.
func TOTPVerifyNow(code string, secret []byte, opts ...OTPOption) (bool, error) {
	cfg := applyOTPOptions(opts)
	return TOTPVerify(code, secret, cfg.clock(), opts...)
}

// OTPAuthURL returns an otpauth:// URL for provisioning TOTP authenticators.
func OTPAuthURL(issuer, account string, secret []byte, opts ...OTPOption) (string, error) {
	cfg := applyOTPOptions(opts)
	if err := validateOTPConfig(secret, cfg); err != nil {
		return "", err
	}
	issuer = strings.TrimSpace(issuer)
	account = strings.TrimSpace(account)
	if issuer == "" || account == "" {
		return "", knifer.WrapError(knifer.ErrCodeInvalidInput, "otp issuer and account must be non-empty", ErrInvalidOTP)
	}
	u := url.URL{
		Scheme: "otpauth",
		Host:   "totp",
		Path:   "/" + issuer + ":" + account,
	}
	q := u.Query()
	q.Set("secret", OTPSecretBase32(secret))
	q.Set("issuer", issuer)
	q.Set("algorithm", otpAlgorithm(cfg.hash))
	q.Set("digits", strconv.Itoa(cfg.digits))
	q.Set("period", strconv.FormatInt(int64(cfg.step/time.Second), 10))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func validateOTPCode(code string, digits int) error {
	if len(code) != digits {
		return knifer.WrapError(knifer.ErrCodeInvalidInput, "otp code has invalid length", ErrInvalidOTP)
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return knifer.WrapError(knifer.ErrCodeInvalidInput, "otp code must contain only decimal digits", ErrInvalidOTP)
		}
	}
	return nil
}

func otpCodeEqual(a, b string) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func otpAlgorithm(fn func() hash.Hash) string {
	if fn == nil {
		return "SHA1"
	}
	switch fn().Size() {
	case sha1.Size:
		return "SHA1"
	case sha256.Size:
		return "SHA256"
	case sha512.Size:
		return "SHA512"
	default:
		return "SHA1"
	}
}
