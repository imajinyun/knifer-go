package vcrypto_test

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vcrypto"
)

func TestFacadeHOTPTOTP(t *testing.T) {
	secret := []byte("12345678901234567890")
	hotp, err := vcrypto.HOTP(secret, 0)
	if err != nil {
		t.Fatalf("HOTP error = %v", err)
	}
	if hotp != "755224" {
		t.Fatalf("HOTP = %q", hotp)
	}
	ok, err := vcrypto.HOTPVerify("755224", secret, 0)
	if err != nil || !ok {
		t.Fatalf("HOTPVerify = %v, %v", ok, err)
	}
	totp, err := vcrypto.TOTP(secret, time.Unix(59, 0), vcrypto.WithOTPDigits(8), vcrypto.WithOTPHash(sha1.New))
	if err != nil {
		t.Fatalf("TOTP error = %v", err)
	}
	if totp != "94287082" {
		t.Fatalf("TOTP = %q", totp)
	}
	ok, err = vcrypto.TOTPVerifyNow(
		totp,
		secret,
		vcrypto.WithOTPClock(func() time.Time { return time.Unix(59, 0) }),
		vcrypto.WithOTPDigits(8),
		vcrypto.WithOTPHash(sha1.New),
	)
	if err != nil || !ok {
		t.Fatalf("TOTPVerifyNow = %v, %v", ok, err)
	}
}

func TestFacadeOTPSecretsAndErrors(t *testing.T) {
	secret, err := vcrypto.GenerateOTPSecret(4, vcrypto.WithRandomReader(bytes.NewReader([]byte{1, 2, 3, 4})))
	if err != nil {
		t.Fatalf("GenerateOTPSecret error = %v", err)
	}
	encoded := vcrypto.OTPSecretBase32(secret)
	decoded, err := vcrypto.ParseOTPSecretBase32(encoded)
	if err != nil {
		t.Fatalf("ParseOTPSecretBase32 error = %v", err)
	}
	if !bytes.Equal(decoded, secret) {
		t.Fatalf("ParseOTPSecretBase32 = %x, want %x", decoded, secret)
	}
	u, err := vcrypto.OTPAuthURL("Example", "alice@example.com", secret)
	if err != nil || u == "" {
		t.Fatalf("OTPAuthURL = %q, %v", u, err)
	}
	if _, err := vcrypto.HOTP(nil, 0); !errors.Is(err, vcrypto.ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("HOTP invalid secret error = %v", err)
	}
	if _, err := vcrypto.HOTP([]byte("secret"), 0, vcrypto.WithOTPDigits(5)); !errors.Is(err, vcrypto.ErrInvalidOTP) {
		t.Fatalf("HOTP invalid digits error = %v", err)
	}
}
