package vcrypto

import (
	"hash"
	"time"

	cryptoimpl "github.com/imajinyun/knifer-go/internal/crypto"
)

// OTPOption customizes HOTP and TOTP helpers.
type OTPOption = cryptoimpl.OTPOption

// WithOTPDigits sets the number of decimal digits in generated OTP codes.
func WithOTPDigits(digits int) OTPOption { return cryptoimpl.WithOTPDigits(digits) }

// WithOTPHash sets the HMAC hash used by HOTP/TOTP helpers.
func WithOTPHash(fn func() hash.Hash) OTPOption { return cryptoimpl.WithOTPHash(fn) }

// WithTOTPStep sets the TOTP time step.
func WithTOTPStep(step time.Duration) OTPOption { return cryptoimpl.WithTOTPStep(step) }

// WithTOTPWindow sets the number of adjacent time steps accepted by TOTPVerify.
func WithTOTPWindow(window int) OTPOption { return cryptoimpl.WithTOTPWindow(window) }

// WithOTPClock sets the clock used by TOTPNow and TOTPVerifyNow.
func WithOTPClock(clock func() time.Time) OTPOption { return cryptoimpl.WithOTPClock(clock) }

// GenerateOTPSecret returns random bytes suitable for HOTP/TOTP secrets.
func GenerateOTPSecret(size int, opts ...RandomOption) ([]byte, error) {
	return cryptoimpl.GenerateOTPSecret(size, opts...)
}

// OTPSecretBase32 encodes a binary OTP secret using unpadded Base32.
func OTPSecretBase32(secret []byte) string { return cryptoimpl.OTPSecretBase32(secret) }

// ParseOTPSecretBase32 decodes an unpadded or padded Base32 OTP secret.
func ParseOTPSecretBase32(secret string) ([]byte, error) {
	return cryptoimpl.ParseOTPSecretBase32(secret)
}

// HOTP generates an HMAC-based one-time password for counter.
func HOTP(secret []byte, counter uint64, opts ...OTPOption) (string, error) {
	return cryptoimpl.HOTP(secret, counter, opts...)
}

// HOTPVerify verifies an HOTP code for counter using constant-time comparison.
func HOTPVerify(code string, secret []byte, counter uint64, opts ...OTPOption) (bool, error) {
	return cryptoimpl.HOTPVerify(code, secret, counter, opts...)
}

// TOTP generates a time-based one-time password for t.
func TOTP(secret []byte, t time.Time, opts ...OTPOption) (string, error) {
	return cryptoimpl.TOTP(secret, t, opts...)
}

// TOTPNow generates a time-based one-time password using the configured clock.
func TOTPNow(secret []byte, opts ...OTPOption) (string, error) {
	return cryptoimpl.TOTPNow(secret, opts...)
}

// TOTPVerify verifies a TOTP code for t and the configured time-step window.
func TOTPVerify(code string, secret []byte, t time.Time, opts ...OTPOption) (bool, error) {
	return cryptoimpl.TOTPVerify(code, secret, t, opts...)
}

// TOTPVerifyNow verifies a TOTP code using the configured clock.
func TOTPVerifyNow(code string, secret []byte, opts ...OTPOption) (bool, error) {
	return cryptoimpl.TOTPVerifyNow(code, secret, opts...)
}

// OTPAuthURL returns an otpauth:// URL for provisioning TOTP authenticators.
func OTPAuthURL(issuer, account string, secret []byte, opts ...OTPOption) (string, error) {
	return cryptoimpl.OTPAuthURL(issuer, account, secret, opts...)
}
