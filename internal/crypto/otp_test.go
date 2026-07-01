package crypto

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"hash"
	"strings"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
)

func TestHOTPRFC4226Vectors(t *testing.T) {
	secret := []byte("12345678901234567890")
	want := []string{
		"755224",
		"287082",
		"359152",
		"969429",
		"338314",
		"254676",
		"287922",
		"162583",
		"399871",
		"520489",
	}
	for counter, code := range want {
		got, err := HOTP(secret, uint64(counter))
		if err != nil {
			t.Fatalf("HOTP(%d) error = %v", counter, err)
		}
		if got != code {
			t.Fatalf("HOTP(%d) = %q, want %q", counter, got, code)
		}
		ok, err := HOTPVerify(code, secret, uint64(counter))
		if err != nil || !ok {
			t.Fatalf("HOTPVerify(%d) = %v, %v", counter, ok, err)
		}
	}
}

func TestTOTPRFC6238Vectors(t *testing.T) {
	tests := []struct {
		name   string
		secret []byte
		hash   func() hash.Hash
		want   map[int64]string
	}{
		{
			name:   "sha1",
			secret: []byte("12345678901234567890"),
			hash:   sha1.New,
			want: map[int64]string{
				59:          "94287082",
				1111111109:  "07081804",
				1111111111:  "14050471",
				1234567890:  "89005924",
				2000000000:  "69279037",
				20000000000: "65353130",
			},
		},
		{
			name:   "sha256",
			secret: []byte("12345678901234567890123456789012"),
			hash:   sha256.New,
			want: map[int64]string{
				59:          "46119246",
				1111111109:  "68084774",
				1111111111:  "67062674",
				1234567890:  "91819424",
				2000000000:  "90698825",
				20000000000: "77737706",
			},
		},
		{
			name:   "sha512",
			secret: []byte("1234567890123456789012345678901234567890123456789012345678901234"),
			hash:   sha512.New,
			want: map[int64]string{
				59:          "90693936",
				1111111109:  "25091201",
				1111111111:  "99943326",
				1234567890:  "93441116",
				2000000000:  "38618901",
				20000000000: "47863826",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for unix, want := range tt.want {
				got, err := TOTP(tt.secret, time.Unix(unix, 0).UTC(), WithOTPDigits(8), WithOTPHash(tt.hash))
				if err != nil {
					t.Fatalf("TOTP(%d) error = %v", unix, err)
				}
				if got != want {
					t.Fatalf("TOTP(%d) = %q, want %q", unix, got, want)
				}
			}
		})
	}
}

func TestTOTPVerifyWindowAndClock(t *testing.T) {
	secret := []byte("12345678901234567890")
	now := time.Unix(60, 0).UTC()
	code, err := TOTP(secret, now.Add(-30*time.Second))
	if err != nil {
		t.Fatalf("TOTP error = %v", err)
	}
	ok, err := TOTPVerify(code, secret, now, WithTOTPWindow(1))
	if err != nil || !ok {
		t.Fatalf("TOTPVerify with window = %v, %v", ok, err)
	}
	ok, err = TOTPVerify(code, secret, now)
	if err != nil || ok {
		t.Fatalf("TOTPVerify without window = %v, %v", ok, err)
	}
	ok, err = TOTPVerifyNow(code, secret, WithOTPClock(func() time.Time { return now }), WithTOTPWindow(1))
	if err != nil || !ok {
		t.Fatalf("TOTPVerifyNow with clock = %v, %v", ok, err)
	}
}

func TestOTPSecretBase32AndAuthURL(t *testing.T) {
	secret := []byte("12345678901234567890")
	encoded := OTPSecretBase32(secret)
	decoded, err := ParseOTPSecretBase32(strings.ToLower(encoded[:4] + " " + encoded[4:] + "===="))
	if err != nil {
		t.Fatalf("ParseOTPSecretBase32 error = %v", err)
	}
	if !bytes.Equal(decoded, secret) {
		t.Fatalf("ParseOTPSecretBase32 = %q, want %q", decoded, secret)
	}
	u, err := OTPAuthURL("Example", "alice@example.com", secret, WithOTPDigits(8), WithOTPHash(sha256.New), WithTOTPStep(60*time.Second))
	if err != nil {
		t.Fatalf("OTPAuthURL error = %v", err)
	}
	for _, fragment := range []string{"otpauth://totp/Example:alice@example.com", "algorithm=SHA256", "digits=8", "issuer=Example", "period=60", "secret=" + encoded} {
		if !strings.Contains(u, fragment) {
			t.Fatalf("OTPAuthURL = %q, missing %q", u, fragment)
		}
	}
}

func TestOTPErrors(t *testing.T) {
	_, err := HOTP(nil, 0)
	if !errors.Is(err, ErrInvalidKey) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("HOTP nil secret error = %v", err)
	}
	_, err = HOTP([]byte("secret"), 0, WithOTPDigits(5))
	if !errors.Is(err, ErrInvalidOTP) || !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("HOTP invalid digits error = %v", err)
	}
	_, err = TOTP([]byte("secret"), time.Unix(0, 0), WithTOTPStep(time.Millisecond))
	if !errors.Is(err, ErrInvalidOTP) {
		t.Fatalf("TOTP invalid step error = %v", err)
	}
	if _, err = ParseOTPSecretBase32("not base32!"); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("ParseOTPSecretBase32 invalid error = %v", err)
	}
	if ok, err := HOTPVerify("12a456", []byte("secret"), 0); ok || !errors.Is(err, ErrInvalidOTP) {
		t.Fatalf("HOTPVerify invalid code = %v, %v", ok, err)
	}
	if _, err = OTPAuthURL("", "alice", []byte("secret")); !errors.Is(err, ErrInvalidOTP) {
		t.Fatalf("OTPAuthURL invalid issuer error = %v", err)
	}
	if _, err = GenerateOTPSecret(0); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("GenerateOTPSecret invalid size error = %v", err)
	}
}
